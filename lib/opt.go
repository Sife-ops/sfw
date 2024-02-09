package lib

import (
	"flag"
)

var FlagCwLim = flag.Bool("cw_lim", true, "limit cubiomes workers")
var FlagHost = flag.String("db_host", "10.0.0.10:5432", "host")
var FlagInst = flag.String("inst", "sfw0", "instance diff") // todo not good
var FlagLogSrv = flag.String("log_srv", "10.0.0.10:1337", "listener addr")
var FlagName = flag.String("db_name", "seed", "db user")
var FlagPass = flag.String("db_pass", "seed", "db password")
var FlagThreads = flag.Int("t", 1, "threads")
var FlagUser = flag.String("db_user", "seed", "db user")
var FlagWebSrv = flag.String("web_srv", "10.0.0.10:3000", "web addr")

func FlagParse() {
	flag.Parse()
}
