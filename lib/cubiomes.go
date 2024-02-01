package lib

import (
	"context"
	"fmt"
	"math/rand"
	"os/exec"
	"sfw/db"
	"strconv"
	"strings"
)

// todo return continue value???
func Cubiomes() (db.GodSeed, error) {
	seed := rand.Uint64()
	// log.Printf("info checking seed %d", int64(seed))

	execCubiomes := exec.CommandContext(
		context.TODO(),
		"./bin/cubiomes", fmt.Sprintf("%d", int64(seed)),
	)
	outCubiomes, err := execCubiomes.Output()
	if err != nil {
		return db.GodSeed{}, err
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

	return godSeed, nil
}
