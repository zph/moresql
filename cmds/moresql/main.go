package main

import (
	_ "expvar"
	"fmt"
	"os"

	"flag"

	_ "github.com/lib/pq"
	"github.com/zph/moresql"
)

var GitRef, version, BuildDate, GitSHA string

func usage() func() {
	return func() {
		fmt.Fprintf(os.Stderr, "%s\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Version %s, Git %s, Git SHA %s, BuildDate %s\n", version, GitRef, GitSHA, BuildDate)
		fmt.Fprintln(os.Stderr, "Repo https://github.com/zph/moresql")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Usage = usage()
	moresql.Run()
}
