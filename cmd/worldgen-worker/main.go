package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sfw/lib"
	"strconv"
	"time"
)

var generatingC = make(chan struct{}, 1)
var generateErrC = make(chan error)
var generateResetC = make(chan struct{}, 1)
var sigC = make(chan os.Signal, 1)

func init() {
	lib.FlagParse()

	log.SetOutput(io.MultiWriter(os.Stdout, lib.SockLogger{}))

	signal.Notify(sigC, os.Interrupt)
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error %v", err)
	}
}

func NewCtx() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

func run() error {
	log.Printf("info starting worldgen worker")
	ctx, cancel := NewCtx()
	for {
		select {
		case <-time.After(3 * time.Second):
			if len(generatingC) < 1 {
				generatingC <- struct{}{}
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

var c1 = make(chan struct{})

// func f1(ctx context.Context) {
// 	for {
// 		world := []lib.World{}
// 		if err := lib.Db.SelectContext(ctx, &world,
// 			`SELECT * 
// 			FROM world
// 			WHERE finished_worldgen IS NULL`,
// 		); err != nil {
// 			generateErrC <- err
// 			return
// 		}
// 		c1 <- struct{}{}
// 	}
// }

func generate(ctx context.Context) {
	worldNotGeneratedC := make(chan lib.World, 1)
	go func() {
		tx, err := lib.Db.BeginTxx(ctx, nil)
		if err != nil {
			generateErrC <- err
			<-generatingC
			return
		}

		world := []lib.World{}
		if err := tx.Select(&world,
			`SELECT * 
			FROM world
			WHERE finished_worldgen IS NULL`,
			// // 9154804515642838022
			// `SELECT *
			// FROM world
			// WHERE seed='9154804515642838022'`,
		); err != nil {
			generateErrC <- err
			return
		}
		if len(world) < 1 {
			generateResetC <- struct{}{}
			return
		}
		if _, err := tx.Exec(
			`UPDATE world
			SET finished_worldgen=0
			WHERE seed=$1`,
			world[0].Seed,
		); err != nil {
			generateErrC <- err
			return
		}

		if err := tx.Commit(); err != nil {
			generateErrC <- err
			return
		}
		worldNotGeneratedC <- world[0]
	}()

	var worldNotGenerated lib.World
	select {
	case <-ctx.Done():
		<-generatingC
		return
	case world := <-worldNotGeneratedC:
		worldNotGenerated = world
	}

	tx, err := lib.Db.BeginTxx(ctx, nil)
	if err != nil {
		generateErrC <- err
		<-generatingC
		return
	}

	worldGeneratedC := make(chan lib.World, 1)
	go func() {
	Dilate:
		// todo more params
		worldGenerated, err := lib.WorldgenTask(ctx, worldNotGenerated)
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
		worldGeneratedC <- worldGenerated
	}()

	var worldGenerated lib.World
	select {
	case <-ctx.Done():
		<-generatingC
		return
	case s := <-worldGeneratedC:
		worldGenerated = s
	}

	if _, err := tx.NamedExec(
		`UPDATE 
			world
		SET 
			exposed_ravine_blocks=:exposed_ravine_blocks,
			iron_shipwrecks=:iron_shipwrecks,
			ravine_proximity=:ravine_proximity,
			avg_bastion_air=:avg_bastion_air,
			finished_worldgen=1 
		WHERE 
			seed=:seed`,
		&worldGenerated,
	); err != nil {
		generateErrC <- err
		<-generatingC
		return
	}

	if err := tx.Commit(); err != nil {
		generateErrC <- err
		<-generatingC
		return
	}

	generateResetC <- struct{}{}
	<-generatingC
}
