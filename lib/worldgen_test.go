package lib

import (
	"testing"
)

func Test_Worldgen(t *testing.T) {
	t.Logf("lmaooooooooo")
	gs, err := Worldgen(GodSeed{
		// 1018
		Seed: ToStringRef("5881310871221101610"),
		// 0
		SpawnX:     ToIntRef(16),
		SpawnZ:     ToIntRef(120),
		BastionX:   ToIntRef(-96),
		BastionZ:   ToIntRef(32),
		ShipwreckX: ToIntRef(32),
		ShipwreckZ: ToIntRef(128),
		FortressX:  ToIntRef(96),
		FortressZ:  ToIntRef(-128),
		// 1
		// 2024-02-01 14:47:43
	}, 4)
	if err != nil {
		t.FailNow()
	}

	t.Logf("info worldgen output %v", gs)
}
