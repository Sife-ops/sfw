package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sfw/lib"
	"time"
)

var connErrC = make(chan error)
var idleC = make(chan struct{}, 1)
var sigC = make(chan os.Signal, 1)
var threadsC chan struct{}

func init() {
	lib.FlagParse()
	threadsC = make(chan struct{}, *lib.FlagThreads)
	signal.Notify(sigC, os.Interrupt)
	idleC <- struct{}{}
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error %v", err)
	}
}

func run() error {
	for {
		ctx, cancel := context.WithCancel(context.Background())
		go LoopPollDb(ctx)
		for len(threadsC) < *lib.FlagThreads {
			threadsC <- struct{}{}
			go LoopCubiomes(ctx)
		}

		select {
		case err := <-connErrC:
			cancel()
			log.Printf("warning error %v", err)
			log.Printf("info trying again in 3 seconds")
			<-time.After(3 * time.Second)
			continue
		case <-sigC:
			cancel()
			return nil
		}
	}
}

func LoopCubiomes(ctx context.Context) {
	cubiomesSeedC := make(chan lib.GodSeed, 1)
Loop0:
	for {
		select {
		case <-ctx.Done():
			break Loop0
		case cs := <-cubiomesSeedC:
			if _, err := lib.Db.NamedExec(
				`INSERT INTO seed 
					(seed, spawn_x, spawn_z, bastion_x, bastion_z, shipwreck_x, shipwreck_z, fortress_x, fortress_z, finished_cubiomes)
				VALUES 
					(:seed, :spawn_x, :spawn_z, :bastion_x, :bastion_z, :shipwreck_x, :shipwreck_z, :fortress_x, :fortress_z, :finished_cubiomes)`,
				&cs,
			); err != nil {
				connErrC <- err
				break Loop0
			}
		default:
			if len(idleC) > 0 {
				<-time.After(1 * time.Second)
				continue
			}
			cubiomesSeed, err := lib.Cubiomes()
			// todo check exit code
			if err != nil {
				continue
			}
			cubiomesSeedC <- cubiomesSeed
		}
	}
	<-threadsC
}

func LoopPollDb(ctx context.Context) {
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
				connErrC <- err
				return
			}
			switch {
			case len(godSeeds) < 6:
				if len(idleC) > 0 {
					<-idleC
					log.Printf("info changed idle to false")
				}
			case len(godSeeds) > 9:
				if len(idleC) < 1 {
					idleC <- struct{}{}
					log.Printf("info changed idle to true")
				}
			}
		}
	}
}
