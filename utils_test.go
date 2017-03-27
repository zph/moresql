package moresql_test

import (
	"github.com/rwynn/gtm"
	m "github.com/zph/moresql"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2/bson"
)

func (s *MySuite) TestSanitizeData(c *C) {
	bsonId := bson.ObjectId("123")
	withBson := map[string]interface{}{"_id": bsonId}
	withBsonResult := map[string]interface{}{"_id": "313233"}
	withSymbol := map[string]interface{}{"name": bson.Symbol("Alice")}
	withSymbolResult := map[string]interface{}{"name": "Alice"}
	withNonPrimaryKey := map[string]interface{}{"name": "Alice", "location_id": bsonId}
	withNonPrimaryKeyResult := map[string]interface{}{"name": "Alice", "location_id": "313233"}
	var table = []struct {
		op     *gtm.Op
		result map[string]interface{}
	}{
		{&gtm.Op{Operation: "i", Data: withBson}, withBsonResult},
		{&gtm.Op{Operation: "i", Data: withSymbol}, withSymbolResult},
		{&gtm.Op{Operation: "i", Data: withNonPrimaryKey}, withNonPrimaryKeyResult},
	}
	for _, t := range table {
		actual := m.SanitizeData(BuildFields("_id", "name", "age", "location_id"), t.op)
		c.Check(actual, DeepEquals, t.result)
	}
}

func (s *MySuite) TestEnsureOpHasAllFieldsWhenEmpty(c *C) {
	op := &gtm.Op{Operation: "i"}
	fields := []string{"_id", "name", "age"}
	actual := m.EnsureOpHasAllFields(op, fields)
	for _, f := range fields {
		val, ok := actual.Data[f]
		c.Check(ok, Equals, true)
		c.Check(val, Equals, nil)
	}
	c.Check(actual.Data, DeepEquals, map[string]interface{}{
		"_id":  interface{}(nil),
		"name": interface{}(nil),
		"age":  interface{}(nil),
	})
}

func (s *MySuite) TestEnsureOpHasAllFieldsWhenMissingField(c *C) {
	data := map[string]interface{}{
		"_id":  interface{}("123"),
		"name": interface{}("Alice"),
	}
	op := &gtm.Op{Operation: "i", Data: data}
	fields := []string{"_id", "name", "age"}
	actual := m.EnsureOpHasAllFields(op, fields)
	for _, f := range fields {
		_, ok := actual.Data[f]
		c.Check(ok, Equals, true)
	}
	c.Check(actual.Data, DeepEquals, map[string]interface{}{
		"_id":  interface{}("123"),
		"name": interface{}("Alice"),
		"age":  interface{}(nil),
	})
}

func (s *MySuite) TestIsInsertUpdateDelete(c *C) {
	var table = []struct {
		op     *gtm.Op
		result bool
	}{
		{&gtm.Op{Operation: "c"}, false},
		{&gtm.Op{Operation: "i"}, true},
		{&gtm.Op{Operation: "u"}, true},
		{&gtm.Op{Operation: "d"}, true},
	}
	for _, t := range table {
		actual := m.IsInsertUpdateDelete(t.op)
		c.Check(actual, Equals, t.result)
	}
}

// func (s *MySuite) TestCreateFanKey(c *C){

// }
