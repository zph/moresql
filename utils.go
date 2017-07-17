package moresql

import (
	"encoding/json"
	"flag"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/tidwall/gjson"

	rollus "github.com/heroku/rollrus"
	"github.com/rwynn/gtm"
)

func FetchEnvsAndFlags() (e Env) {
	e = Env{}
	e.urls.mongo = os.Getenv("MONGO_URL")
	e.urls.postgres = os.Getenv("POSTGRES_URL")
	var x = *flag.String("mongo-url", "", "`MONGO_URL` aka connection string")
	var p = *flag.String("postgres-url", "", "`POSTGRES_URL` aka connection string")
	flag.StringVar(&e.configFile, "config-file", "moresql.json", "Configuration file to use")
	flag.BoolVar(&e.sync, "full-sync", false, "Run full sync for each db.collection in config")
	flag.BoolVar(&e.allowDeletes, "allow-deletes", true, "Allow deletes to propagate from Mongo -> PG")
	flag.BoolVar(&e.tail, "tail", false, "Tail mongodb for each db.collection in config")
	flag.StringVar(&e.SSLCert, "ssl-cert", "", "SSL PEM cert for Mongodb")
	flag.StringVar(&e.appName, "app-name", "moresql", "AppName used in Checkpoint table")
	flag.BoolVar(&e.monitor, "enable-monitor", false, "Run expvarmon endpoint")
	flag.BoolVar(&e.checkpoint, "checkpoint", false, "Store and restore from checkpoints in PG table: moresql_metadata")
	flag.BoolVar(&e.createTableSQL, "create-table-sql", false, "Print out the necessary SQL for creating metadata table required for checkpointing")
	flag.BoolVar(&e.validatePostgres, "validate", false, "Validate the postgres table structures and exit")
	flag.StringVar(&e.errorReporting, "error-reporting", "", "Error reporting tool to use (currently only supporting Rollbar)")
	flag.StringVar(&e.memprofile, "memprofile", "", "Profile memory usage. Supply filename for output of memory usage")
	defaultDuration := time.Duration(0 * time.Second)
	flag.DurationVar(&e.replayDuration, "replay-duration", defaultDuration, "Last x to replay ie '1s', '5m', etc as parsed by Time.ParseDuration. Will be subtracted from time.Now()")
	flag.Int64Var(&e.replaySecond, "replay-second", 0, "Replay a specific epoch second of the oplog and forward from there.")
	flag.BoolVar(&e.SSLInsecureSkipVerify, "ssl-insecure-skip-verify", false, "Skip verification of Mongo SSL certificate ala sslAllowInvalidCertificates")
	flag.Parse()
	e.reportingToken = os.Getenv("ERROR_REPORTING_TOKEN")
	e.appEnvironment = os.Getenv("APP_ENV")
	if e.appEnvironment == "" {
		e.appEnvironment = "production"
	}
	if e.replayDuration != defaultDuration && e.replaySecond != 0 {
		e.replayOplog = true
	} else {
		e.replayOplog = false
	}
	if x != "" {
		e.urls.mongo = x
	}
	if p != "" {
		e.urls.postgres = p
	}
	log.Debugf("Configuration: %+v", e)
	if e.memprofile != "" {
		f, err := os.Create(e.memprofile)
		if err != nil {
			log.Fatal(err)
		}
		wg.Add(1)
		go func() {
			defer f.Close()
			tick := time.Tick(time.Duration(20) * time.Second)
			for {
				select {
				case <-tick:
					pprof.WriteHeapProfile(f)
				}
			}
		}()
	}
	return
}

func SetupLogger(env Env) {
	// Alter logging pattern for heroku
	log.SetOutput(os.Stdout)
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	if os.Getenv("DYNO") != "" {
		formatter.FullTimestamp = false
		log.SetLevel(log.InfoLevel)
		log.SetFormatter(&log.JSONFormatter{})
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		l, err := log.ParseLevel(v)
		if err != nil {
			log.WithField("level", v).Warn("LOG_LEVEL invalid, choose from debug, info, warn, fatal.")
		} else {
			log.SetLevel(l)
		}
	}
	switch env.errorReporting {
	case "rollbar":
		rollus.SetupLogging(env.reportingToken, env.appEnvironment)
	}

	log.WithField("logLevel", log.GetLevel()).Debug("Log Settings")
}

func IsInsertUpdateDelete(op *gtm.Op) bool {
	return isActionableOperation(op.IsInsert, op.IsUpdate, op.IsDelete)
}

func isActionableOperation(filters ...func() bool) bool {
	for _, fn := range filters {
		if fn() {
			return true
		}
	}
	return false
}

// SanitizeData handles type inconsistency between mongo and pg
// and flattens the data from a potentially nested data struct
// into a flattened struct using gjson.
func SanitizeData(pgFields Fields, op *gtm.Op) map[string]interface{} {
	if !IsInsertUpdateDelete(op) {
		return make(map[string]interface{})
	}

	newData, err := json.Marshal(op.Data)
	parsed := gjson.ParseBytes(newData)
	output := make(map[string]interface{})
	if err != nil {
		log.Errorf("Failed to marshal op.Data into json %s", err.Error())
	}
	// Normalize data map to always include the Id with conversion
	if op.Id != nil {
		output["_id"] = op.Id
	}

	for k, v := range pgFields {
		// Dot notation extraction
		maybe := parsed.Get(k)
		if !maybe.Exists() {
			continue
		}

		output[v.Postgres.Name] = maybe.Value()
	}
	return output
}

func createFanKey(db string, collection string) string {
	return db + "." + collection
}

func splitFanKey(key string) (string, string) {
	s := strings.Split(key, ".")
	return s[0], s[1]
}

// EnsureOpHasAllFields: Ensure that required keys are present will null value
func EnsureOpHasAllFields(op *gtm.Op, keysToEnsure []string) *gtm.Op {
	// Guard against assignment into nil map
	if op.Data == nil {
		op.Data = make(map[string]interface{})
	}
	for _, k := range keysToEnsure {
		if _, ok := op.Data[k]; !ok {
			op.Data[k] = nil
		}
	}
	return op
}

func ExitUnlessValidEnv(e Env) {
	if e.validatePostgres {
		return
	}

	if e.createTableSQL {
		c := Commands{}
		c.CreateTableSQL()
	}
	if e.urls.mongo == "" || e.urls.postgres == "" {
		log.Warnf(`Missing required variable. Both MONGO_URL and POSTGRES_URL must be set.
		            See the following usage instructions for setting those variables.`)
		flag.Usage()
		os.Exit(1)
	}
	if !(e.sync || e.tail) {
		flag.Usage()
		os.Exit(1)
	}
}
