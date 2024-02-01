package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"sfw/lib"
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
	godSeed := lib.GodSeed{
		Seed:             lib.ToStringRef(fmt.Sprintf("%d", int64(seed))),
		SpawnX:           lib.MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[0], ",")[0])),
		SpawnZ:           lib.MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[0], ",")[1])),
		ShipwreckX:       lib.MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[1], ",")[0])),
		ShipwreckZ:       lib.MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[1], ",")[1])),
		BastionX:         lib.MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[2], ",")[0])),
		BastionZ:         lib.MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[2], ",")[1])),
		FortressX:        lib.MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[3], ",")[0])),
		FortressZ:        lib.MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[3], ",")[1])),
		FinishedCubiomes: lib.ToIntRef(1),
	}
	if _, err := lib.Db.NamedExec(
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
