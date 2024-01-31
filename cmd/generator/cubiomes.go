package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"sfw/db"
	"strconv"
	"strings"
)

func Cubiomes() {
	seed := rand.Uint64()
	// log.Printf("info checking seed %d", int64(seed))

	execCubiomes := exec.CommandContext(
		context.TODO(),
		"./bin/cubiomes", fmt.Sprintf("%d", int64(seed)),
	)
	outCubiomes, err := execCubiomes.Output()
	if err != nil {
		if err.Error() != "exit status 33" {
			log.Printf("warning %s", err.Error())
		}
		return
	}

	// log.Printf("info cubiomes output: %s", string(outCubiomes))
	// log.Printf("info %v passed cubiomes", seed)

	outCubiomesArr := strings.Split(string(outCubiomes), ":")
	godSeed := db.GodSeed{
		Seed:             db.ToStringRef(fmt.Sprintf("%d", int64(seed))),
		SpawnX:           db.MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[0], ",")[0])),
		SpawnZ:           db.MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[0], ",")[1])),
		ShipwreckX:       db.MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[1], ",")[0])),
		ShipwreckZ:       db.MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[1], ",")[1])),
		BastionX:         db.MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[2], ",")[0])),
		BastionZ:         db.MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[2], ",")[1])),
		FortressX:        db.MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[3], ",")[0])),
		FortressZ:        db.MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[3], ",")[1])),
		FinishedCubiomes: db.ToIntRef(1),
	}
	if _, err := db.Db.NamedExec(
		`INSERT INTO seed 
			(seed, spawn_x, spawn_z, bastion_x, bastion_z, shipwreck_x, shipwreck_z, fortress_x, fortress_z, finished_cubiomes)
		VALUES 
			(:seed, :spawn_x, :spawn_z, :bastion_x, :bastion_z, :shipwreck_x, :shipwreck_z, :fortress_x, :fortress_z, :finished_cubiomes)`,
		&godSeed,
	); err != nil {
		log.Fatalf("error %v", err)
	}
	log.Printf("info saving cubiomes results %v", godSeed)

	if len(CubiomesDone) < *FlagJobs {
		CubiomesDone <- struct{}{}
		CubiomesOut <- godSeed
		log.Printf("info finished %d cubiomes jobs", len(CubiomesDone))
	}
}
