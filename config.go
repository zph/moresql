package moresql

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"

	"strings"

	log "github.com/Sirupsen/logrus"
)

func LoadConfigString(s string) (Config, error) {
	config := Config{}
	var configDelayed ConfigDelayed
	err := json.Unmarshal([]byte(s), &configDelayed)
	if err != nil {
		log.Fatalln(err)
	}
	for k, v := range configDelayed {
		db := DB{}
		collections := Collections{}
		db.Collections = collections
		for k, v := range v.Collections {
			coll := Collection{Name: v.Name, PgTable: v.PgTable}
			var fields Fields
			fields, err = JsonToFields(string(v.Fields))
			if err != nil {
				log.Warnf("JSON Config decoding error: ", err)
				return nil, fmt.Errorf("Unable to decode %s", err)
			}
			coll.Fields = fields
			db.Collections[k] = coll
		}
		config[k] = db
	}
	return config, nil
}

func LoadConfig(path string) Config {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	config, err := LoadConfigString(string(b))
	if err != nil {
		panic(err)
	}
	return config
}

func mongoToPostgresTypeConversion(mongoType string) string {
	// Coerce "id" bsonId types into text since Postgres doesn't have type for BSONID
	switch strings.ToLower(mongoType) {
	case "id":
		return "text"
	}
	return mongoType
}

func normalizeDotNotationToPostgresNaming(key string) string {
	re := regexp.MustCompile("\\.")
	return re.ReplaceAllString(key, "_")
}

func JsonToFields(s string) (Fields, error) {
	var init FieldsWrapper
	var err error
	result := Fields{}
	err = json.Unmarshal([]byte(s), &init)
	for k, v := range init {
		field := Field{}
		str := ""
		if err := json.Unmarshal(v, &field); err == nil {
			result[k] = field
		} else if err := json.Unmarshal(v, &str); err == nil {
			// Convert shorthand to longhand Field
			f := Field{
				Mongo{k, str},
				Postgres{normalizeDotNotationToPostgresNaming(k), mongoToPostgresTypeConversion(str)},
			}
			result[k] = f
		} else {
			errLong := json.Unmarshal(v, &field)
			errShort := json.Unmarshal(v, &str)
			err = errors.New(fmt.Sprintf("Could not decode Field. Long decoding %+v. Short decoding %+v", errLong, errShort))
			return nil, err
		}

	}

	return result, err
}
