package moresql_test

import (
	"testing"

	_ "github.com/lib/pq"
	m "github.com/zph/moresql"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestBuildUpsertStatement(c *C) {
	mongo := m.Mongo{"_id", "id"}
	p := m.Postgres{"id", "text"}
	f := m.Field{mongo, p}
	f2 := m.Field{m.Mongo{"count", "text"}, m.Postgres{"count", "text"}}
	fields := m.Fields{"_id": f, "count": f2}
	collection := m.Collection{
		Name:    "categories",
		PgTable: "categories",
		Fields:  fields}
	o := m.Statement{collection}

	sql := o.BuildUpsert()
	expected := `INSERT INTO "categories" ("id", "count")
VALUES (:_id, :count)
ON CONFLICT ("id")
DO UPDATE SET "count" = :count;`
	c.Check(sql, Equals, expected)
}

func (s *MySuite) TestBuildInsertStatement(c *C) {
	mongo := m.Mongo{"_id", "id"}
	p := m.Postgres{"id", "text"}
	f := m.Field{mongo, p}
	f2 := m.Field{m.Mongo{"count", "text"}, m.Postgres{"count", "text"}}
	fields := m.Fields{"_id": f, "count": f2}
	collection := m.Collection{
		Name:    "categories",
		PgTable: "categories",
		Fields:  fields}
	o := m.Statement{collection}

	sql := o.BuildInsert()
	expected := `INSERT INTO "categories" ("id", "count")
VALUES (:_id, :count)`
	c.Check(sql, Equals, expected)
}

func (s *MySuite) TestBuildUpdateStatement(c *C) {
	mongo := m.Mongo{"_id", "id"}
	p := m.Postgres{"id", "id"}
	f := m.Field{mongo, p}
	f2 := m.Field{m.Mongo{"count", "text"}, m.Postgres{"count", "text"}}
	f3 := m.Field{m.Mongo{"avg", "text"}, m.Postgres{"avg", "text"}}
	fields := m.Fields{"_id": f, "count": f2, "avg": f3}
	collection := m.Collection{
		Name:    "categories",
		PgTable: "categories",
		Fields:  fields}
	o := m.Statement{collection}
	sql := o.BuildUpdate()
	expected := `UPDATE "categories"
SET "avg" = :avg, "count" = :count
WHERE "id" = :_id;`
	c.Check(sql, Equals, expected)
}

func (s *MySuite) TestBuildDeleteStatement(c *C) {
	mongo := m.Mongo{"_id", "id"}
	p := m.Postgres{"id", "id"}
	f := m.Field{mongo, p}
	f2 := m.Field{m.Mongo{"count", "text"}, m.Postgres{"count", "text"}}
	f3 := m.Field{m.Mongo{"avg", "text"}, m.Postgres{"avg", "text"}}
	fields := m.Fields{"_id": f, "count": f2, "avg": f3}
	collection := m.Collection{
		Name:    "categories",
		PgTable: "categories",
		Fields:  fields}
	o := m.Statement{collection}
	sql := o.BuildDelete()
	expected := `DELETE FROM "categories" WHERE "id" = :_id;`
	c.Check(sql, Equals, expected)
}
