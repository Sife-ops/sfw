package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sfw/lib"
	"time"
)

var asyncErrC = make(chan error, 1)
var sigC = make(chan os.Signal, 1)
var threadsC chan struct{}
var worldC = make(chan lib.World)

func init() {
	lib.FlagParse()
	signal.Notify(sigC, os.Interrupt)
	threadsC = make(chan struct{}, *lib.FlagThreads)
}

func main() {
	if err := run(); err != nil {
		log.Printf("%v", err)
	}
}

func run() error {
	log.Printf("info starting cubiomes worker on %d threads", *lib.FlagThreads)

	go saveWorld()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for len(threadsC) < *lib.FlagThreads {
		threadsC <- struct{}{}
		go loopCubiomes(ctx)
	}

	select {
	case <-sigC:
	case err := <-asyncErrC:
		return err
	}

	return nil
}

func loopCubiomes(ctx context.Context) {
	defer func() {
		<-threadsC
	}()

	for {
		world, err := lib.Cubiomes(ctx)
		switch {
		case errors.Is(ctx.Err(), context.Canceled):
			return
		case err != nil:
			asyncErrC <- err
			return
		}

		log.Printf("info generated cubiomes world %s", *world.Seed)
		worldC <- world
	}
}

func saveWorld() {
	for {
		world := <-worldC

		doneC := make(chan struct{})
		errC := make(chan error)
		go func() {
			if _, err := lib.Db.NamedExec(
				`INSERT INTO world (
					seed,
					spawn_x, spawn_z,
					bastion_x, bastion_z,
					shipwreck_x, shipwreck_z,
					fortress_x, fortress_z,
					finished_cubiomes
				)
				VALUES (
					:seed,
					:spawn_x, :spawn_z,
					:bastion_x, :bastion_z,
					:shipwreck_x, :shipwreck_z,
					:fortress_x, :fortress_z,
					:finished_cubiomes
				)`,
				&world,
			); err != nil {
				errC <- err
				return
			}
			doneC <- struct{}{}
		}()

		for {
			select {
			case <-doneC:
			case <-sigC:
				return
			case err := <-errC:
				asyncErrC <- err
				return
			case <-time.After(3 * time.Second):
				log.Printf("info database not responding")
				continue
			}
			break
		}
	}
}
