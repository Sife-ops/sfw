package lib

import (
	"flag"
)

var FlagCwLim = flag.Bool("cw_lim", true, "limit cubiomes workers")
var FlagThreads = flag.Int("t", 1, "threads")

func FlagParse() {
	flag.Parse()
}
