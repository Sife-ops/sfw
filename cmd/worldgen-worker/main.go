package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sfw/lib"
	"strconv"
	"time"
)

var rlStartC = make(chan error, 1)
var sendIdleC = make(chan error, 1)
var sigC = make(chan os.Signal, 1)
var sockClient = lib.SockClient{}
var sockErrC = make(chan error, 1)
var wgBusyC = make(chan error, 1)
var wgStartC = make(chan lib.GodSeed, 1)

var connErrC = make(chan error)
var generatingC = make(chan struct{}, 1)
var noErrC = make(chan struct{}, 1)

func init() {
	lib.FlagParse()
	signal.Notify(sigC, os.Interrupt)
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error %v", err)
	}
}

func NewCtx() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.WithValue(context.Background(), "inst", *lib.FlagInst))
}

func run() error {
	ctx, cancel := NewCtx()
	for {
		select {
		case <-time.After(3 * time.Second):
			if len(generatingC) < 1 {
				generatingC <- struct{}{}
				go generate(ctx)
			}

		case err := <-connErrC:
			cancel()
			for len(generatingC) > 0 {
				<-generatingC
			}
			ctx, cancel = NewCtx()
			log.Printf("warning error %v", err)
			log.Printf("info trying again in 3 seconds")
			select {
			case <-time.After(3 * time.Second):
			case <-sigC:
				return nil
			}

		case <-noErrC:
			cancel()
			for len(generatingC) > 0 {
				<-generatingC
			}
			ctx, cancel = NewCtx()

		case <-sigC:
			cancel()
			return nil
		}
	}
}

func generate(ctx context.Context) {
	tx, err := lib.Db.BeginTxx(ctx, nil)
	if err != nil {
		connErrC <- err
		return
	}

	cubiomesSeedC := make(chan lib.GodSeed, 1)
	go func() {
		cs := []lib.GodSeed{}
		if err := tx.Select(&cs,
			`SELECT * 
			FROM seed 
			WHERE finished_worldgen IS NULL`,
		); err != nil {
			connErrC <- err
			return
		}
		if len(cs) < 1 {
			noErrC <- struct{}{}
			return
		}
		if _, err := tx.Exec(
			`UPDATE seed
			SET finished_worldgen=0
			WHERE seed=$1`,
			cs[0].Seed,
		); err != nil {
			connErrC <- err
			return
		}
		cubiomesSeedC <- cs[0]
	}()

	var cubiomesSeed lib.GodSeed
	select {
	case <-ctx.Done():
		return
	case cs := <-cubiomesSeedC:
		cubiomesSeed = cs
	}

	godSeedC := make(chan lib.GodSeed, 1)
	go func() {
	Dilate:
		// todo params
		// todo context
		gs, err := lib.Worldgen(ctx, cubiomesSeed, 4)
		if err != nil {
			fmt.Printf(">>> ***** WORLDGEN IS DILATING *****\n")
			fmt.Printf(">>> reason: %v\n", err)
			fmt.Printf(">>> 1) next\n")
			fmt.Printf(">>> Enter) dilate\n")

			action := make(chan string)
			go func() {
				var a string
				fmt.Scanln(&a)
				action <- a
			}()

			select {
			case a := <-action:
				aInt, err := strconv.Atoi(a)
				if err != nil || aInt != 1 {
					goto Dilate
				}
			case <-time.After(30 * time.Second):
			}

			if err := tx.Commit(); err != nil {
				connErrC <- err
				return
			}

			noErrC <- struct{}{}
			return
		}
		godSeedC <- gs
	}()

	var godSeed lib.GodSeed
	select {
	case <-ctx.Done():
		return
	case gs := <-godSeedC:
		godSeed = gs
	}

	if _, err := tx.NamedExec(
		`UPDATE 
			seed 
		SET 
			ravine_chunks=:ravine_chunks,
			iron_shipwrecks=:iron_shipwrecks,
			ravine_proximity=:ravine_proximity,
			avg_bastion_air=:avg_bastion_air,
			finished_worldgen=1 
		WHERE 
			seed=:seed`,
		&godSeed,
	); err != nil {
		connErrC <- err
		return
	}

	if err := tx.Commit(); err != nil {
		connErrC <- err
		return
	}

	noErrC <- struct{}{}
}
