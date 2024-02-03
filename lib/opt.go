package lib

import (
	"flag"
)

var FlagWorker = flag.String("w", "127.0.0.1:3100", "worker addr")
var FlagServer = flag.String("s", "0.0.0.0:3100", "server addr")
var FlagThreads = flag.Int("t", 1, "threads")
var FlagHost = flag.String("db_host", "127.0.0.1", "host")
var FlagUser = flag.String("db_user", "", "db user")
var FlagPass = flag.String("db_pass", "", "db password")

func FlagParse() {
	flag.Parse()
}
