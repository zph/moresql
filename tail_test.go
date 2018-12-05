package moresql_test

import (
	"time"

	bson "github.com/globalsign/mgo/bson"
	m "github.com/zph/moresql"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestTimestamp(c *C) {
	startTime := time.Duration(1485144398995 * time.Millisecond)
	f := func() time.Time { return time.Unix(0, startTime.Nanoseconds()) }
	t := m.OpTimestampWrapper(f, -1*time.Hour)(nil, nil)
	expected := bson.MongoTimestamp(6378662081129873409)
	c.Check(t, Equals, expected)
}

func (s *MySuite) TestNewOptionsWithEpoch(c *C) {
	tail := m.Tailer{}
	opts, err := tail.NewOptions(m.EpochTimestamp(1485144398), time.Duration(0))
	actual := opts.After(nil, nil)
	expected := bson.MongoTimestamp(6378646619247607809)
	c.Check(actual, Equals, expected)
	c.Check(err, Equals, nil)
}

func (s *MySuite) TestMsLag(c *C) {
	var table = []struct {
		ts      int32
		now     int64
		nowNano int64
		out     int64
	}{
		{1486898048, 1486898048, 0, 0},
		{1486898047, 1486898048, 0, 1000},
		{1486898047, 0, 1486898048001000000, 1001},
	}
	for _, tt := range table {
		ts := tt.ts
		f := func() time.Time { return time.Unix(tt.now, tt.nowNano) }
		t := m.Tailer{}
		actual := t.MsLag(ts, f)
		c.Check(actual, Equals, int64(tt.out))
	}
}
