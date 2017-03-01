package moresql_test

import (
	m "github.com/zph/moresql"
	. "gopkg.in/check.v1"
)

func BuildFieldFromId(str string) m.Field {
	field := m.Field{}
	field.Mongo.Name = str
	field.Mongo.Type = "id"
	field.Postgres.Name = str
	field.Postgres.Type = "text"
	return field
}

func BuildTextField(str string) m.Field {
	field := m.Field{}
	field.Mongo.Name = str
	field.Mongo.Type = "text"
	field.Postgres.Name = str
	field.Postgres.Type = "text"
	return field
}

func (s *MySuite) TestConfigParsingForFields(c *C) {
	// Fields struct
	ex1 := `
          {"_id": {
            "mongo": {
              "name": "_id",
              "type": "id"
            },
            "postgres": {
              "name": "_id",
              "type": "text"
            }
          }}
          `
	field := BuildFieldFromId("_id")
	expected1 := m.Fields{"_id": field}

	ex2 := `
          {"name": {
            "mongo": {
              "name": "name",
              "type": "text"
            },
            "postgres": {
              "name": "name",
              "type": "text"
            }
          }}
    `
	expected2 := m.Fields{"name": BuildTextField("name")}

	ex3 := `{"name": "text"}`
	expected3 := m.Fields{"name": BuildTextField("name")}

	ex4 := `{"age": "integer", "name": "text"}`
	exField4 := BuildTextField("age")
	exField4.Mongo.Type = "integer"
	exField4.Postgres.Type = "integer"
	expected4 := m.Fields{"name": BuildTextField("name"), "age": exField4}

	var table = []struct {
		js       string
		expected m.Fields
	}{
		{ex1, expected1},
		{ex2, expected2},
		{ex3, expected3},
		{ex4, expected4},
	}
	for _, t := range table {
		f, err := m.JsonToFields(t.js)
		c.Check(err, Equals, nil)
		c.Check(f, DeepEquals, t.expected)
	}
}

func (s *MySuite) TestConfigParsingFull(c *C) {
	// Fields struct
	ex1 := `
{
  "company-production": {
    "collections": {
      "accounts": {
        "name": "users",
        "pg_table": "users",
        "fields": {
          "_id": {
            "mongo": {
              "name": "_id",
              "type": "id"
            },
            "postgres": {
              "name": "_id",
              "type": "text"
            }
          },
          "bio": {
            "mongo": {
              "name": "bio",
              "type": "text"
            },
            "postgres": {
              "name": "bio",
              "type": "text"
            }
          }
        }
      },
      "campaigns": {
        "name": "campaigns",
        "pg_table": "campaigns",
        "fields": {
          "_id": "id",
          "created_at": "text"
          }
      }
    }
  }
}
          `

	expected1 := m.Config{"company-production": m.DB{Collections: m.Collections{"accounts": m.Collection{Name: "users", PgTable: "users", Fields: m.Fields{"_id": m.Field{Mongo: m.Mongo{Name: "_id", Type: "id"}, Postgres: m.Postgres{Name: "_id", Type: "text"}}, "bio": m.Field{Mongo: m.Mongo{Name: "bio", Type: "text"}, Postgres: m.Postgres{Name: "bio", Type: "text"}}}}, "campaigns": m.Collection{Name: "campaigns", PgTable: "campaigns", Fields: m.Fields{"_id": m.Field{Mongo: m.Mongo{Name: "_id", Type: "id"}, Postgres: m.Postgres{Name: "_id", Type: "text"}}, "created_at": m.Field{Mongo: m.Mongo{Name: "created_at", Type: "text"}, Postgres: m.Postgres{Name: "created_at", Type: "text"}}}}}}}

	shorthand := `
{
  "company-production": {
    "collections": {
      "accounts": {
        "name": "users",
        "pg_table": "users",
        "fields": {
          "_id": "id",
          "bio": "text"
        }
      },
      "campaigns": {
        "name": "campaigns",
        "pg_table": "campaigns",
        "fields": {
          "_id": "id",
          "created_at": "text"
          }
      }
    }
  }
}
`
	var table = []struct {
		js       string
		expected m.Config
	}{
		{ex1, expected1},
		{shorthand, expected1},
	}
	for _, t := range table {
		f, err := m.LoadConfigString(t.js)
		c.Check(f, DeepEquals, t.expected)
		c.Check(err, Equals, nil)
	}
}
