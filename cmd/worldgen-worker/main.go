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

var generateC = make(chan struct{}, 1)
var generateErrC = make(chan error)
var generateResetC = make(chan struct{}, 1)
var sigC = make(chan os.Signal, 1)

func init() {
	log.SetOutput(lib.Logger{})
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
	log.Printf("info starting worldgen worker")
	ctx, cancel := NewCtx()
	for {
		select {
		case <-time.After(3 * time.Second):
			if len(generateC) < 1 {
				generateC <- struct{}{}
				go generate(ctx)
			}

		case err := <-generateErrC:
			log.Printf("fatal error %v", err)
			cancel()
			log.Printf("info trying again in 3 seconds")
			select {
			case <-time.After(3 * time.Second):
				ctx, cancel = NewCtx()
			case <-sigC:
				return nil
			}

		case <-generateResetC:
			cancel()
			ctx, cancel = NewCtx()

		case <-sigC:
			cancel()
			return nil
		}
	}
}

func generate(ctx context.Context) {
	cubiomesSeedC := make(chan lib.GodSeed, 1)
	go func() {
		tx, err := lib.Db.BeginTxx(ctx, nil)
		if err != nil {
			generateErrC <- err
			<-generateC
			return
		}

		cs := []lib.GodSeed{}
		if err := tx.Select(&cs,
			`SELECT * 
			FROM seed 
			WHERE finished_worldgen IS NULL`,
		); err != nil {
			generateErrC <- err
			return
		}
		if len(cs) < 1 {
			generateResetC <- struct{}{}
			return
		}
		if _, err := tx.Exec(
			`UPDATE seed
			SET finished_worldgen=0
			WHERE seed=$1`,
			cs[0].Seed,
		); err != nil {
			generateErrC <- err
			return
		}

		if err := tx.Commit(); err != nil {
			generateErrC <- err
			return
		}
		cubiomesSeedC <- cs[0]
	}()

	var cubiomesSeed lib.GodSeed
	select {
	case <-ctx.Done():
		<-generateC
		return
	case cs := <-cubiomesSeedC:
		cubiomesSeed = cs
	}

	tx, err := lib.Db.BeginTxx(ctx, nil)
	if err != nil {
		generateErrC <- err
		<-generateC
		return
	}

	godSeedC := make(chan lib.GodSeed, 1)
	go func() {
	Dilate:
		// todo more params
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
				generateErrC <- err
				return
			}

			generateResetC <- struct{}{}
			return
		}
		godSeedC <- gs
	}()

	var godSeed lib.GodSeed
	select {
	case <-ctx.Done():
		<-generateC
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
		generateErrC <- err
		<-generateC
		return
	}

	if err := tx.Commit(); err != nil {
		generateErrC <- err
		<-generateC
		return
	}

	generateResetC <- struct{}{}
	<-generateC
}
