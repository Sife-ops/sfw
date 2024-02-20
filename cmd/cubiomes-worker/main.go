package main

import (
	"context"
	"errors"
	"io"
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
	log.SetOutput(io.MultiWriter(os.Stdout, lib.SockLogger{}))
	signal.Notify(sigC, os.Interrupt)
	threadsC = make(chan struct{}, *lib.FlagThreads)
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("%v", err)
	}
}

func run() error {
	log.Printf("info starting cubiomes worker on %d threads", *lib.FlagThreads)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go saveWorld(ctx)

	for len(threadsC) < *lib.FlagThreads {
		threadsC <- struct{}{}
		go loopCubiomes(ctx)
	}

	select {
	case <-sigC:
		return nil
	case err := <-asyncErrC:
		return err
	}
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

func saveWorld(ctx context.Context) {
	for {
		world := <-worldC

	Retry:
		ctxTo, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		savedC := make(chan struct{})
		errC := make(chan error, 1)
		go func() {
			if _, err := lib.Db.NamedExecContext(
				ctxTo,
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
			// todo ctxTo no require check err???
			savedC <- struct{}{}
		}()

		select {
		case <-ctxTo.Done():
			log.Printf("info database not responding")
			// cancel()
			goto Retry
		case <-savedC:
		case err := <-errC:
			asyncErrC <- err
		case <-ctx.Done():
			return
		}
	}
}
