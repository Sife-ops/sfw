package lib

import (
	"flag"
)

var FlagHost = flag.String("db_host", "10.0.0.10:5432", "host")
var FlagInst = flag.String("inst", "sfw0", "instance diff") // todo not good
var FlagName = flag.String("db_name", "seed", "db user")
var FlagPass = flag.String("db_pass", "seed", "db password")
var FlagThreads = flag.Int("t", 1, "threads")
var FlagUser = flag.String("db_user", "seed", "db user")

func FlagParse() {
	flag.Parse()
}
