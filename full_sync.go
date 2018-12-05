package moresql

import (
	"expvar"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	mgo "github.com/globalsign/mgo"
	"github.com/jmoiron/sqlx"
	"github.com/orcaman/concurrent-map"
	"github.com/paulbellamy/ratecounter"
	"github.com/rwynn/gtm"
)

type Syncer interface {
	Read() func()
	Write() func()
	BuildOpFromMgo() func(o Statement, e DBResult, coll Collection) gtm.Op
}

type FullSyncer struct {
	Config Config
	Output *sqlx.DB
	Mongo  *mgo.Session
	C      chan DBResult
	done   chan bool

	insertCounter *ratecounter.RateCounter
	readCounter   *ratecounter.RateCounter
}

func (z *FullSyncer) Read() {
	for dbName, v := range z.Config {
		db := z.Mongo.DB(dbName)
		for name := range v.Collections {
			coll := db.C(name)
			iter := coll.Find(nil).Iter()
			var result map[string]interface{}
			for iter.Next(&result) {
				z.readCounter.Incr(1)
				z.C <- DBResult{dbName, name, result}
				// Clear out result data for next round
				result = make(map[string]interface{})
			}
			if err := iter.Close(); err != nil {
				log.Error("Unable to close iterator: %s", err)
			}
		}
	}
	close(z.C)
	wg.Done()
}

func (z *FullSyncer) Write() {
	var workers [workerCountOverflow]int
	tables := z.buildTables()
	for _ = range workers {
		wg.Add(1)
		go z.writer(&tables)
	}
	wg.Done()
}

func BuildOpFromMgo(mongoFields []string, e DBResult, coll Collection) *gtm.Op {
	var op gtm.Op
	op.Data = e.Data
	opRef := EnsureOpHasAllFields(&op, mongoFields)
	opRef.Id = e.Data["_id"]
	// Set to I so we are consistent about these beings inserts
	// This avoids our guardclause in sanitize
	opRef.Operation = "i"
	data := SanitizeData(coll.Fields, opRef)
	opRef.Data = data
	return opRef
}

func (z *FullSyncer) writer(tables *cmap.ConcurrentMap) {
ForStatement:
	for {
		select {
		case e, more := <-z.C:
			if !more {
				break ForStatement
			}
			key := createFanKey(e.MongoDB, e.Collection)
			v, ok := tables.Get(key)
			if ok && !v.(bool) {
				// Table doesn't exist, skip
				break
			}
			o, coll := z.statementFromDbCollection(e.MongoDB, e.Collection)
			op := BuildOpFromMgo(o.mongoFields(), e, coll)
			s := o.BuildUpsert()
			log.WithFields(log.Fields{
				"collection": e.Collection,
				"id":         op.Id,
			}).Info("Syncing record")
			log.Debug("SQL Command ", s)
			log.Debug("Data ", op.Data)
			log.Debug("Executing statement: ", s)
			_, err := z.Output.NamedExec(s, op.Data)
			log.Debug("Statement executed successfully")
			z.insertCounter.Incr(1)
			if err != nil {
				log.WithFields(log.Fields{
					"description": err,
				}).Error("Error")
				if err.Error() == fmt.Sprintf(`pq: relation "%s" does not exist`, e.Collection) {
					tables.Set(key, false)
				}
			}
		}
	}
	wg.Done()
}
func (z *FullSyncer) statementFromDbCollection(db string, collectionName string) (Statement, Collection) {
	c := z.Config[db].Collections[collectionName]
	return Statement{c}, c
}

func (z *FullSyncer) buildTables() (tables cmap.ConcurrentMap) {
	tables = cmap.New()
	for dbName, db := range z.Config {
		for collectionName := range db.Collections {
			// Assume all tables are present
			tables.Set(createFanKey(dbName, collectionName), true)
		}
	}
	return
}

func NewSynchronizer(config Config, pg *sqlx.DB, mongo *mgo.Session) FullSyncer {
	c := make(chan DBResult)
	insertCounter := ratecounter.NewRateCounter(1 * time.Second)
	readCounter := ratecounter.NewRateCounter(1 * time.Second)
	expvar.Publish("insert/sec", insertCounter)
	expvar.Publish("read/sec", readCounter)
	done := make(chan bool, 2)
	sync := FullSyncer{config, pg, mongo, c, done, insertCounter, readCounter}
	return sync
}

func FullSync(config Config, pg *sqlx.DB, mongo *mgo.Session) {
	sync := NewSynchronizer(config, pg, mongo)
	wg.Add(2)
	log.Debug("Starting writer")
	go sync.Write()
	log.Debug("Starting reader")
	go sync.Read()

	wg.Wait()
}
