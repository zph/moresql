# Welcome to Moresql Documentation

For basic introduction: [README](/README/)

[Github Repository](https://github.com/zph/moresql)


MoreSQL streams changes occuring in Mongo database into a Postgres db. MoreSQL tails the oplog and generates appropriate actions against Postgres. MoreSQL has the ability to do full synchronizations using UPSERTS, with the benefit over INSERTS that this can be executed against tables with existing data.

MoreSQL gives you a chance to use more sql and less mongo query language.

## Commands

* `moresql -tail` - Start tailing the oplog from mongo and persist to Postgres.
* `moresql -full-sync` - Conduct a full sync based on configuration file from mongo->pg.
* `moresql -help` - Usage Instructions.
