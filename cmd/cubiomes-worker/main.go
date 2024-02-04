package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sfw/lib"
	"time"
)

var loopCubiomesErrC = make(chan error)
var loopCubiomesIdleC = make(chan struct{}, 1)
var sigC = make(chan os.Signal, 1)
var threadsC chan struct{}

func init() {
	lib.FlagParse()
	threadsC = make(chan struct{}, *lib.FlagThreads)
	signal.Notify(sigC, os.Interrupt)
	loopCubiomesIdleC <- struct{}{}
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error %v", err)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	for {
		go loopPollDb(ctx)
		for len(threadsC) < *lib.FlagThreads {
			threadsC <- struct{}{}
			go loopCubiomes(ctx)
		}

		select {
		case err := <-loopCubiomesErrC:
			log.Printf("fatal error %v", err)
			cancel()
			log.Printf("info trying again in 3 seconds")
			select {
			case <-time.After(3 * time.Second):
				ctx, cancel = context.WithCancel(context.Background())
			case <-sigC:
				return nil
			}

		case <-sigC:
			cancel()
			return nil
		}
	}
}

func loopCubiomes(ctx context.Context) {
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
				loopCubiomesErrC <- err
				break Loop0
			}
		default:
			if len(loopCubiomesIdleC) > 0 {
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
				loopCubiomesErrC <- err
				return
			}
			switch {
			case len(godSeeds) < 6:
				if len(loopCubiomesIdleC) > 0 {
					<-loopCubiomesIdleC
					log.Printf("info changed idle to false")
				}
			case len(godSeeds) > 9:
				if len(loopCubiomesIdleC) < 1 {
					loopCubiomesIdleC <- struct{}{}
					log.Printf("info changed idle to true")
				}
			}
		}
	}
}
