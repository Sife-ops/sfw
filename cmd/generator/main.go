package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sfw/db"
	"sfw/lib"
	"strconv"
	"time"
)

var RavineProximity = 4 // chunk radius
var RavineOffsetNegative = RavineProximity * 16
var RavineOffsetPositive = RavineOffsetNegative + 15

var FlagThreads = flag.Int("t", 2, "threads")
var FlagJobs = flag.Int("j", 2, "jobs")
var FlagPrune = flag.Bool("p", false, "prune")

var CubiomesDone chan struct{}
var CubiomesOut chan db.GodSeed

var WorldgenDone chan struct{}
var WorldgenRecovering = make(chan db.GodSeed, 1)

func init() {
	flag.Parse()
	CubiomesDone = make(chan struct{}, *FlagJobs)
	CubiomesOut = make(chan db.GodSeed, *FlagJobs)
	WorldgenDone = make(chan struct{}, *FlagJobs)
}

// todo delete old cubiomes/worldgen/container
// todo move everything to /lib
// todo html
// todo ability to load unfinished seeds
// todo c interop w/ cubiomes https://karthikkaranth.me/blog/calling-c-code-from-go/
func main() {
	defer close(CubiomesDone)
	defer close(CubiomesOut)
	defer close(WorldgenDone)
	defer close(WorldgenRecovering)

	go func() {
		sigchan := make(chan os.Signal)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		// log.Printf("info cleaning up")
		os.Exit(0)
	}()

	for i := 0; i < *FlagThreads; i++ {
		go func() {
			for len(CubiomesDone) < *FlagJobs {
				gs, err := lib.Cubiomes()
				if err != nil {
					// 33 means seed was successfully filtered out
					// if err.Error() != "exit status 33" {
					// 	log.Fatalf("error %s", err.Error())
					// }
					continue
				}

				// log.Printf("info saving cubiomes results %v", godSeed)
				if _, err := db.Db.NamedExec(
					`INSERT INTO seed 
						(seed, spawn_x, spawn_z, bastion_x, bastion_z, shipwreck_x, shipwreck_z, fortress_x, fortress_z, finished_cubiomes)
					VALUES 
						(:seed, :spawn_x, :spawn_z, :bastion_x, :bastion_z, :shipwreck_x, :shipwreck_z, :fortress_x, :fortress_z, :finished_cubiomes)`,
					&gs,
				); err != nil {
					log.Fatalf("error saving cubiomes results %s", err.Error())
				}

				if len(CubiomesDone) < *FlagJobs {
					CubiomesDone <- struct{}{}
					CubiomesOut <- gs
					log.Printf("info finished %d cubiomes jobs", len(CubiomesDone))
				}
			}
		}()
	}

	for len(WorldgenDone) < *FlagJobs {
		select {
		case j := <-WorldgenRecovering:
			// todo force a recovery
			log.Printf("########### WORLDGEN RECOVERING ###########")
			log.Printf("job: %v", j)

			PromptIndex := make(chan int)
			go func() {
				fmt.Printf(">>> select action\n")
				fmt.Printf(">>> 1) go next (default)\n")
				fmt.Printf(">>> 2) add to end of queue\n")
				fmt.Printf(">>> 3) quit with %d worldgen jobs remaining\n", *FlagJobs-len(WorldgenDone))
				var action string
				fmt.Scanln(&action)
				actionInt, err := strconv.Atoi(action)
				if err != nil || actionInt < 1 || actionInt > 3 {
					PromptIndex <- 0
					return
				}
				PromptIndex <- actionInt - 1
			}()

			select {
			case <-time.After(30 * time.Second):
				WorldgenDone <- struct{}{}
				log.Printf("info continuing")
			case promptIndex := <-PromptIndex:
				switch promptIndex {
				case 0:
					WorldgenDone <- struct{}{}
					log.Printf("info continuing")
				case 1:
					CubiomesOut <- j
					log.Printf("info added to queue")
				case 2:
					fallthrough
				default:
					log.Printf("info exiting")
					return
				}
			}

		default:
			Worldgen()
		}
	}

	if *FlagPrune {
		Prune()
	}
}

func Prune() {
	if _, err := db.Db.Exec(
		`DELETE FROM seed
		WHERE iron_shipwrecks<1`,
	); err != nil {
		log.Fatalf("error %v", err)
	}
	if _, err := db.Db.Exec(
		`DELETE FROM seed
		WHERE ravine_chunks<5`,
	); err != nil {
		log.Fatalf("error %v", err)
	}
	if _, err := db.Db.Exec(
		`DELETE FROM seed
		WHERE ravine_proximity IS NULL`,
	); err != nil {
		log.Fatalf("error %v", err)
	}
}
