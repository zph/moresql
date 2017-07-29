# MoreSQL

[](NOTE: README.md is a generated FILE changes belong in docs/README.template.md. Update with make docs)
[![Build Status](https://travis-ci.org/zph/moresql.svg?branch=master)](https://travis-ci.org/zph/moresql)
[![GoDoc](https://godoc.org/github.com/zph/moresql?status.svg)](https://godoc.org/github.com/zph/moresql)

## Introduction

MoreSQL streams changes occuring in Mongo database into a Postgres db. MoreSQL tails the oplog and generates appropriate actions against Postgres. MoreSQL has the ability to do full synchronizations using `UPSERTS`, with the benefit over `INSERTS` that this can be executed against tables with existing data.

MoreSQL gives you a chance to use more sql and less mongo query language.

# Usage

## Basic Use

### Tail

`./moresql -tail -config-file=moresql.json`

Tail is the primary run mode for MoreSQL. When tailing, the oplog is observed for novely and each INSERT/UPDATE/DELETE is translated to its SQL equivalent, then executed against Postgres.

Tail makes a best faith effort to do this and uses checkpoint markers to track last successfully applied Mongo Oplog event.

Given that `tail` mode executes `UPSERTS` instead of `INSERT || UPDATE`, we expect MoreSQL to be roughly eventually consistent. We're chosing to prioritize speed of execution (multiple workers) in lieu of some consistency. This helps to keep low latency with larger workloads. We currently partition workload among multiple workers but ensure that each `collection.id` combination will be routed to same worker in correct oplog order. This avoids the circumstance where two operations against same `collection.id` are executed by different workers, out of order.

### Full Sync

`./moresql -full-sync -config-file=moresql.json`

Full sync is useful when first setting up a MoreSQL installation to port the existing Mongo data to Postgres. We recommend setting up a tailing instance first. Once that's running, do a full sync in different process. This should put the Mongo and Postgres into identical states.

Given the nature of streaming replica data from Mongo -> Postgres, it's recommended to run full sync at intervals in order to offset losses that may have occured during network issues, system downtime, etc.

### Documentation

https://zph.github.io/moresql/

[![GoDoc](https://godoc.org/github.com/zph/moresql?status.svg)](https://godoc.org/github.com/zph/moresql)

## QuickStart

### Introduction

* Create metadata table
* Setup moresql.json
* Setup any recipient tables in postgres
  * Validate with `./moresql -validate`
* Deploy binary to server
* Configure Environmental variables
* Run `./moresql -tail` to start transmitting novelty
* Run `./moresql -full-sync` to populate the database

### Table Setup

```sql
$SQL_OUTPUT
```

### Building Binary

Compile binary using `make build`

### Commandline Arguments / Usage

Execute `./moresql --help`

```
$MORESQL_USAGE
```

### Validation of Configuration + Postgres Schema

`./moresql -validate`

This will report any issues related to the postgres schema being a mis-match for the fields and tables setup in configuration.

# Requirements, Stability and Versioning

MoreSQL is expected and built with Golang 1.6, 1.7 and master in mind. Broken tests on these versions indicates a bug.

MoreSQL requires Postgres 9.5+ due to usage of UPSERTs. Using UPSERTs simplifies internal logic but also depends on UNIQUE indexes existing on each `_id` column in Postgres. See `moresql -validate` for advice.

# Miscellanea

### Error Reporting

Available through Rollbar. PRs welcome for other services. We currently use Rollus
which reports errors synchronously. If this is a performance bottleneck please PR or issue.

Enable this by two steps:

```
export ERROR_REPORTING_TOKEN=asdfasdfasdf
export APP_ENV=[production, development, or staging]
```

And when running application use the following flag to enable reporting:

`./moresql -tail -error-reporting "rollbar"`

If these steps are not followed, errors will be reported out solely via logging.

### Environmental Variables used in Moresql

```
$ENV_VARIABLES_FROM_GO
```

### Mongo types

We guard against a few of these for conversion into Postgres friendly types.

Objects and Arrays do not behave properly when inserting into Postgres. These will be automatically converted into their JSON representation before inserting into Postgres.

As of writing, any BsonID/ObjectId should be noted as `id` type in `Fields.Mongo.Type` to facilitate this. In the future we may assume that all fields ending in `_id` are Id based fields and require conversion.

## Converting from MoSQL

Run the ./bin/convert_config_from_mosql_to_moresql script in a folder with `collections.yml`

```
ruby ./bin/convert_config_from_mosql_to_moresql collection.yml
```

The generated file `moresql.json` will be in place ready for use.

## Unsupported Features

These features are part of mosql but not implemented in MoreSQL. PRs welcome.

* extra_props field for spare data
* Automatic creation of tables/columns (we require explicit actions from users after providing guidance)

## Dot notation

We support dot notation for extracting nested objects from Mongo.

`user.address` will perform a `(get_in map [:user :address])` type nested fetch.

See gjson project for full syntax details: https://github.com/tidwall/gjson#path-syntax
## Performance

During benchmarking when moresql is asked to replay existing events from oplog we've seen the following performance with the following configurations:

5 workers per collection
500 generic workers
On a Heroku 1X dyno

```
~ $ ./moresql -tail -replay-duration "5000m" | grep "Rate of"
{"level":"info","msg":"Rate of insert per min: 532","time":"2017-02-23T01:49:31Z"}
{"level":"info","msg":"Rate of update per min: 44089","time":"2017-02-23T01:49:31Z"}
{"level":"info","msg":"Rate of delete per min: 1","time":"2017-02-23T01:49:31Z"}
{"level":"info","msg":"Rate of read per min: 91209","time":"2017-02-23T01:49:31Z"}
{"level":"info","msg":"Rate of skipped per min: 46587","time":"2017-02-23T01:49:31Z"}
```

Approximately 700 updates/sec and 1500 reads/sec is our top observed throughput so far. Please submit PRs with further numbers using a similar command.

We expect the following bottlenecks: connection count in Postgres, pg connection limitations in Moresql (for safety), network bandwidth, worker availability.

At this level of throughput, Moresql uses ~90MB RAM. At low idle throughput of 10-20 req/sec it consumes ~30MB RAM.

In another benchmark when updating 28k documents simultaneously, we observed mean lag of ~ 500ms and 95% of requests arrived in <= 1194ms between when the document was updated in Mongo and when it arrived in Postgres.

See full [performance information](https://zph.github.io/moresql/performance/)

For a general discussion of UPSERT performance in Postgres: https://mark.zealey.org/2016/01/08/how-we-tweaked-postgres-upsert-performance-to-be-2-3-faster-than-mongodb

## Binaries

We release binaries for semvar tags on Github Releases page using `goreleaser` for the platforms listed in goreleaser.yml.

# Credit and Prior Art

 * [MoSQL](https://github.com/stripe/mosql) - the project we used for 3 yrs at work and then retired with MoreSQL. Thanks Stripe!
 * [GTM](https://github.com/rwynn/gtm) - the go library that builds on mgo to wrap the tailing and oplog interface in a pleasant API. rwynn was a large help with improving GTM's performance with varying levels of consistency guarantees.
