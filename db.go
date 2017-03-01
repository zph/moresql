package moresql

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"math"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func GetMongoConnection(env Env) (session *mgo.Session) {
	if env.UseSSL() {
		clientCert, err := ioutil.ReadFile(env.SSLCert)
		if err != nil {
			log.Fatalln("Unable to read ssl certificate")
		}
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM([]byte(clientCert))
		if !ok {
			log.Fatalln("failed to parse root certificate")
		}

		tlsConfig := &tls.Config{RootCAs: roots}

		c, err := mgo.ParseURL(env.urls.mongo)
		if err != nil {
			log.Fatalf("Unable to parse mongo url")
		}
		dialInfo := &mgo.DialInfo{
			Addrs:    c.Addrs,
			Database: c.Database,
			Source:   c.Source,
			Username: c.Username,
			Password: c.Password,
			DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
				return tls.Dial("tcp", addr.String(), tlsConfig)
			},
			Timeout: time.Second * 10,
		}
		session, err = mgo.DialWithInfo(dialInfo)
		if err != nil {
			log.Fatal(err)
		}
		session.SetMode(mgo.Monotonic, true)
	} else {
		var err error
		session, err = mgo.Dial(env.urls.mongo)
		if err != nil {
			log.Fatal(err)
		}
		session.SetMode(mgo.Monotonic, true)
	}
	return
}

func GetPostgresConnection(env Env) (pg *sqlx.DB) {
	var err error
	pg, err = sqlx.Connect("postgres", env.urls.postgres)
	setupPgDefaults(pg)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// setupPgDefaults: Set safe cap so workers do not overwhelm server
func setupPgDefaults(pg *sqlx.DB) {
	pg.SetMaxIdleConns(50)
	pg.SetMaxOpenConns(50)
}

// Credit: https://github.com/go-mgo/mgo/pull/202/files#diff-b47d6566744e81abad9312022bdc8896R374
// From @mwmahlberg
func NewMongoTimestamp(t time.Time, c uint32) (bson.MongoTimestamp, error) {
	var tv uint32
	u := t.Unix()
	if u < 0 || u > math.MaxUint32 {
		return -1, errors.New("invalid value for time")
	}
	tv = uint32(u)
	buf := bytes.Buffer{}
	binary.Write(&buf, binary.BigEndian, tv)
	binary.Write(&buf, binary.BigEndian, c)
	i := int64(binary.BigEndian.Uint64(buf.Bytes()))
	return bson.MongoTimestamp(i), nil
}
