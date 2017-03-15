package moresql

import (
	"flag"
	"net/http"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

// workerCount dedicated workers per collection
const workerCount = 5

// workerCountOverflow Threads in Golang 1.6+ are ~4kb to start
// 500 * 4k = ~2MB ram usage due to heap of each routine
const workerCountOverflow = 500

// reportFrequency the timing for how often to report activity
const reportFrequency = 60 // seconds

// checkpointFrequency frequency at which checkpointing is saved to DB
const checkpointFrequency = time.Duration(30) * time.Second

var wg sync.WaitGroup

func Run() {
	c := Commands{}
	env := FetchEnvsAndFlags()
	SetupLogger(env)
	ExitUnlessValidEnv(env)

	config := LoadConfig(env.configFile)
	pg := GetPostgresConnection(env)
	defer pg.Close()

	// TODO: should this run for each execution of application
	if env.validatePostgres {
		c.ValidateTablesAndColumns(config, pg)
	}

	session := GetMongoConnection(env)
	defer session.Close()
	log.Info("Connected to postgres")
	log.Info("Connected to mongo")

	if env.monitor {
		go http.ListenAndServe(":1234", nil)
	}

	switch {
	case env.sync:
		FullSync(config, pg, session)
	case env.tail:
		Tail(config, pg, session, env)
	default:
		flag.Usage()
	}
}
