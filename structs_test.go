package moresql_test

import (
	m "github.com/zph/moresql"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestUseSSL(c *C) {
	var table = []struct {
		env    m.Env
		result bool
	}{
		{m.Env{SSLCert: ""}, false},
		{m.Env{SSLCert: "cert.pem"}, true},
	}
	for _, t := range table {
		actual := t.env.UseSSL()
		c.Check(actual, DeepEquals, t.result)
	}
}
