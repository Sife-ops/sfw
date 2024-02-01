package lib

import (
	"context"
	"fmt"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
)

// todo return continue value???
func Cubiomes() (GodSeed, error) {
	seed := rand.Uint64()
	// log.Printf("info checking seed %d", int64(seed))

	execCubiomes := exec.CommandContext(
		context.TODO(),
		"./bin/cubiomes", fmt.Sprintf("%d", int64(seed)),
	)
	outCubiomes, err := execCubiomes.Output()
	if err != nil {
		return GodSeed{}, err
	}

	// log.Printf("info cubiomes output: %s", string(outCubiomes))
	// log.Printf("info %v passed cubiomes", seed)

	outCubiomesArr := strings.Split(string(outCubiomes), ":")
	godSeed := GodSeed{
		Seed:             ToStringRef(fmt.Sprintf("%d", int64(seed))),
		SpawnX:           MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[0], ",")[0])),
		SpawnZ:           MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[0], ",")[1])),
		ShipwreckX:       MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[1], ",")[0])),
		ShipwreckZ:       MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[1], ",")[1])),
		BastionX:         MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[2], ",")[0])),
		BastionZ:         MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[2], ",")[1])),
		FortressX:        MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[3], ",")[0])),
		FortressZ:        MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[3], ",")[1])),
		FinishedCubiomes: ToIntRef(1),
	}

	return godSeed, nil
}
