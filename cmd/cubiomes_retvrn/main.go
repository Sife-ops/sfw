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
Start:
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for len(threadsC) < *lib.FlagThreads {
		threadsC <- struct{}{}
		go loopCubiomes(ctx)
	}

	select {
	case <-sigC:
	case err := <-asyncErrC:
		cancel()
		log.Printf("info retrying after 3 seconds due to %v", err)
		<-time.After(3 * time.Second)
		for len(threadsC) > 0 {
			<-time.After(1 * time.Second)
		}
		for len(asyncErrC) > 0 {
			<-asyncErrC
		}
		goto Start
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

		log.Printf("info saving cubiomes world %s", *world.Seed)
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
			)`, &world,
		); err != nil {
			asyncErrC <- err
			return
		}
	}
}
