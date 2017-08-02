TODOs
==

TO DEPLOY and RELEASE
===
* [x] Add checkpointing in case of downtime
 * [x] Make it set on timer, ie every minute or configurable duration
 * [x] determine if we want to play catch up in oplog or not  (get oldest db.oplog.rs.find().sort({ts:-1}).limit(1))
 * [x] Make sure that multiple moresqls can insert their marker on same db meta table
* [x] Add conversion from MoSQL to MoreSQL
* [x] Add quickstart guide and gh pages or link to godoc
* [x] Add starter SQL script for checkpointing script
* [x] Add validation cmdline flag to compare configuration with PG tables/columns
* [x] Add documentation for deploying on Heroku via null buildpack
* [x] Verify that log statements are set at appropriate levels, ie warn/error/etc
* [x] Test in staging, then production
* [x] Bake binary in production
* [ ] Release and announce project to the world


DESIRED
==
* [x] Use dot notation for config description of nested maps, ie allow for `get_in("outerkey", "innerkey")` as `outerkey.innerkey` by using https://github.com/tidwall/gjson
* [x] Refactor tail to have a producer/consumer as Read/Write
* [x] Setup https://github.com/thejerf/suture wrappers on components
* [x] Add testing and refactor to make each bit fairly decoupled
 * [x] Include docs and scripts to transition from mosql to moresql
* [x] Setup formal worker pool along with overflow pool of workers
* [x] add tracking mechanism for missing/broken tables beyond "log it into abyss".
* [x] add error handling with rollbar/bugsnag/etc
* [ ] Improve library testing (unit and integration/system). Potentially using docker for full trip integration tests.
* [ ] Add validation for the moresql_metadata table
* [ ] Add configuration option to use configurable schema for metadata table and I/U/D
* [ ] Add `full-sync` option to only re-sync specific table
* [ ] Fix logging to include TIMESTAMP when deployed outside Heroku


SOMEDAYs
==
* [ ] Setup system tests (https://www.elastic.co/blog/code-coverage-for-your-golang-system-tests)
* [ ] Add basic auth and SSL for endpoint of expvarmon
* [ ] add signal handling for SIGTERM to flush existing content in buffers then exit
* [ ] Add way to reload configuration without dropping events?
* [ ] add expvar.Publish for backlog of all events waiting to process in `fan`
* [ ] time operates on int64, suggest that gtm.ParseTimestamp do likewise for interop
* [ ] Make library generic with regard to event destination. Could be expanded out as a bridge Mongo->{Kinesis,Kafka,Postgres,MySQL}
 * [ ] https://github.com/zph/moresql/blob/master/full_sync.go#L135
 * [ ] Make the writer function configurable with postgres as the default
 * [ ] Writers should fit the interface of accepting a pointer to tables struct and the channel of incoming operations
 * [ ] All of https://github.com/zph/moresql/blob/master/full_sync.go#L129-L136 should be inside the writer function as it will differ by output sink.
* [ ] Add persistance for oplog if desired by user via commandline flag
