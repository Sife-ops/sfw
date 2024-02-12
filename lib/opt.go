package lib

import (
	"flag"
)

// todo REPLACE ALL THESE W PKL
var FlagHost = flag.String("db_host", "10.0.0.10:5432", "host")
var FlagName = flag.String("db_name", "seed", "db user")
var FlagUser = flag.String("db_user", "seed", "db user")
var FlagPass = flag.String("db_pass", "seed", "db password")
var FlagWebSrv = flag.String("web_srv", "10.0.0.10:3000", "web addr")
var FlagLogSrv = flag.String("log_srv", "10.0.0.10:1337", "listener addr")
var FlagWsSrv = flag.String("ws_srv", "10.0.0.10:51871", "ws addr")

var FlagCwLim = flag.Bool("cw_lim", true, "limit cubiomes workers")
var FlagInst = flag.String("inst", "sfw0", "instance diff") // todo not good
var FlagThreads = flag.Int("t", 1, "threads")

func FlagParse() {
	flag.Parse()
}
