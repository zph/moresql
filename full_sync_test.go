package moresql_test

import (
	m "github.com/zph/moresql"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2/bson"
)

func BuildFields(sx ...string) m.Fields {
	f := m.Fields{}
	for _, s := range sx {
		var mon string
		if s == "_id" {
			mon = "id"
		} else {
			mon = "string"
		}
		f[s] = m.Field{
			m.Mongo{s, mon},
			m.Postgres{s, "string"},
		}
	}
	return f
}

func (s *MySuite) TestBuildOpFromMongo(c *C) {
	result := make(map[string]interface{})
	id := bson.ObjectId("123")
	result["_id"] = id
	result["name"] = "Alice"
	db := m.DBResult{"user", "user", result}
	fields := BuildFields("_id", "name", "age")
	coll := m.Collection{"user", "user", "moresql_schema", fields}
	op := m.BuildOpFromMgo([]string{"_id", "name", "age"}, db, coll)

	c.Check(op.Id, Equals, id)
	c.Check(op.Operation, Equals, "i")
	c.Check(op.Data["name"], Equals, "Alice")
	if val, ok := op.Data["age"]; ok {
		c.Check(ok, Equals, true)
		c.Check(val, Equals, nil)
	}
}
