package lib

import (
	"flag"
)

var FlagWorker = flag.String("h", "127.0.0.1:3100", "socket host addr") // todo remove this
var FlagServer = flag.String("s", "0.0.0.0:3100", "socket server addr") // todo remove this
var FlagThreads = flag.Int("t", 1, "threads")
var FlagHost = flag.String("db_host", "127.0.0.1:5432", "host")
var FlagUser = flag.String("db_user", "", "db user")
var FlagName = flag.String("db_name", "", "db user")
var FlagPass = flag.String("db_pass", "", "db password")

func FlagParse() {
	flag.Parse()
}
