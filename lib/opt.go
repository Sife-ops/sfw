package lib

import (
	"flag"
)

var FlagCwLim = flag.Bool("cw_lim", true, "limit cubiomes workers")
var FlagInst = flag.String("inst", "sfw0", "instance diff") // todo not good
var FlagThreads = flag.Int("t", 1, "threads")

func FlagParse() {
	flag.Parse()
}
