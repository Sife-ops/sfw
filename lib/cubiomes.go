package lib

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
)

// todo return 'continue' bool, only return error if not error 33
func Cubiomes(ctx context.Context) (GodSeed, error) {
	// seed := rand.Uint64()

	execCubiomes := exec.CommandContext(ctx, "./bin/cubiomes")
	outCubiomes, err := execCubiomes.Output()
	if err != nil {
		return GodSeed{}, err
	}

	outCubiomesArr := strings.Split(string(outCubiomes), ":")
	seed := GodSeed{
		Seed:             ToStringRef(outCubiomesArr[0]),
		SpawnX:           MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[1], ",")[0])),
		SpawnZ:           MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[1], ",")[1])),
		ShipwreckX:       MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[2], ",")[0])),
		ShipwreckZ:       MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[2], ",")[1])),
		BastionX:         MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[3], ",")[0])),
		BastionZ:         MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[3], ",")[1])),
		FortressX:        MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[4], ",")[0])),
		FortressZ:        MustIntRef(strconv.Atoi(strings.Split(outCubiomesArr[4], ",")[1])),
		FinishedCubiomes: ToIntRef(1),
	}

	return seed, nil
}
