package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sfw/lib"
	"time"
)

var asyncErrC = make(chan error)
var asyncIdleC = make(chan struct{}, 1)
var asyncStopC = make(chan struct{})
var hysteresisMin = 6
var hysteresisMax = 9
var sigC = make(chan os.Signal, 1)
var threadsC chan struct{}

func init() {
	lib.FlagParse()
	threadsC = make(chan struct{}, *lib.FlagThreads)
	signal.Notify(sigC, os.Interrupt)
	if *lib.FlagCwLim {
		asyncIdleC <- struct{}{}
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error %v", err)
	}
}

func run() error {
	for {
		ctx, cancel := context.WithCancel(context.Background())

		go loopPollDb(ctx)
		for len(threadsC) < *lib.FlagThreads {
			threadsC <- struct{}{}
			go loopCubiomes(ctx)
		}

		select {
		case err := <-asyncErrC:
			log.Printf("fatal error %v", err)
			cancel()
			log.Printf("info trying again in 3 seconds")
			select {
			case <-time.After(3 * time.Second):
			case <-sigC:
				return nil
			}

		case <-asyncStopC:
			cancel()
			for len(threadsC) > 0 {
				log.Printf("waiting for %d threads to finish", len(threadsC))
				<-time.After(1 * time.Second)
			}
			log.Printf("no more threads")

		case <-sigC:
			cancel()
			return nil
		}
	}
}

func loopCubiomes(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			<-threadsC
			return
		default:
			if len(asyncIdleC) > 0 {
				<-time.After(1 * time.Second)
				continue
			}

			cubiomesSeed, err := lib.Cubiomes(ctx)
			if err != nil {
				continue
			}

			log.Printf("saving potential god seed %s", *cubiomesSeed.Seed)
			if _, err := lib.Db.NamedExec(
				`INSERT INTO seed 
					(seed, spawn_x, spawn_z, bastion_x, bastion_z, shipwreck_x, shipwreck_z, fortress_x, fortress_z, finished_cubiomes)
				VALUES 
					(:seed, :spawn_x, :spawn_z, :bastion_x, :bastion_z, :shipwreck_x, :shipwreck_z, :fortress_x, :fortress_z, :finished_cubiomes)`,
				&cubiomesSeed,
			); err != nil {
				asyncErrC <- err
			}
		}
	}
}

func loopPollDb(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Second):
			godSeeds := []lib.GodSeed{}
			if err := lib.Db.Select(&godSeeds,
				`SELECT * 
				FROM seed 
				WHERE finished_worldgen IS NULL`,
			); err != nil {
				asyncErrC <- err
				return
			}
			switch {
			case len(godSeeds) < hysteresisMin:
				if len(asyncIdleC) > 0 {
					for len(asyncIdleC) > 0 {
						<-asyncIdleC
					}
					log.Printf("info changed idle to false")
				}
			case len(godSeeds) > hysteresisMax && *lib.FlagCwLim:
				if len(asyncIdleC) < 1 {
					asyncIdleC <- struct{}{}
					asyncStopC <- struct{}{}
					log.Printf("info changed idle to true")
				}
			}
		}
	}
}
