package lib

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Tnze/go-mc/level"
	"github.com/Tnze/go-mc/save"
	"github.com/Tnze/go-mc/save/region"
)

func datamineWorld(ctx context.Context, world World) (World, error) {
	regions := [][]*region.Region{{nil, nil}, {nil, nil}}
	for q := 0; q < 4; q++ {
		rx := (q / 2) - 1
		rz := (q % 2) - 1

		mca := fmt.Sprintf(
			"%s/tmp/sfw0/data/world/region/r.%d.%d.mca",
			MustString(os.Getwd()), rx, rz,
		)

		fileInfo, err := os.Stat(mca)
		switch {
		case err != nil:
			log.Printf("info t.%d.%d.mca error %v", rx, rz, err)
			continue
		case fileInfo.Size() < 1:
			log.Printf("info t.%d.%d.mca is 0 bytes", rx, rz)
			continue
		}

		r, err := region.Open(mca)
		if err != nil {
			return world, err
		}

		defer func() {
			if err := r.Close(); err != nil {
				log.Printf("warning closing region %v", err)
			}
		}()

		regions[rx+1][rz+1] = r
	}

	/*


	   overworld


	*/

	exposedRavineBlocks := 0
	for q := 0; q < 4; q++ {
		rx := (q / 2) - 1
		rz := (q % 2) - 1

		x1, z1, x2, z2, err := toRegionArea(
			rx, rz,
			world.RavineAreaX1(), world.RavineAreaZ1(),
			world.RavineAreaX2(), world.RavineAreaZ2(),
		)
		if err != nil {
			log.Printf("info ravine %s", err.Error())
			continue
		}

		dx := (x2 - x1 + 1) / 16
		a := dx * (z2 - z1 + 1) / 16

		for c := 0; c < a; c++ {
			xc := c%dx + x1/16
			zc := c/dx + z1/16

			sector, err := regions[rx+1][rz+1].ReadSector(ToSector(xc), ToSector(zc))
			if err != nil {
				return world, err
			}

			var chunkSave save.Chunk
			err = chunkSave.Load(sector)
			if err != nil {
				return world, err
			}

			chunkLevel, err := level.ChunkFromSave(&chunkSave)
			if err != nil {
				return world, err
			}

			y10 := 256 * 10
			for i := y10; i < y10+256; i++ {
				x := chunkLevel.GetBlockID(1, i)
				if x == "minecraft:magma_block" || x == "minecraft:obsidian" {
					goodBlocksAbove := 0
					for j := 0; j < 52; j++ {
						yj := i + (j+1)*256
						s := (yj / 4096) + 1
						ys := (yj % 4096)
						x := chunkLevel.GetBlockID(s, ys)

						if x == "minecraft:water" || x == "minecraft:kelp_plant" {
							goodBlocksAbove++
						}
					}

					lavaBlocksBelow := 0
					if x := chunkLevel.GetBlockID(1, i-256); x == "minecraft:lava" {
						lavaBlocksBelow++
					}

					if goodBlocksAbove >= 50 && lavaBlocksBelow > 0 {
						exposedRavineBlocks++
					}
				}
			}
		}
	}

	shipwrecksWithIron := []string{}
	for q := 0; q < 4; q++ {
		rx := (q / 2) - 1
		rz := (q % 2) - 1

		x1, z1, x2, z2, err := toRegionArea(
			rx, rz,
			world.ShipwreckAreaX1(), world.ShipwreckAreaZ1(),
			world.ShipwreckAreaX2(), world.ShipwreckAreaZ2(),
		)
		if err != nil {
			log.Printf("info shipwreck %s", err.Error())
			continue
		}

		dx := (x2 - x1 + 1) / 16
		a := dx * (z2 - z1 + 1) / 16

		for c := 0; c < a; c++ {
			xc := c%dx + x1/16
			zc := c/dx + z1/16

			sector, err := regions[rx+1][rz+1].ReadSector(ToSector(xc), ToSector(zc))
			if err != nil {
				return world, err
			}

			var chunkSave save.Chunk
			err = chunkSave.Load(sector)
			if err != nil {
				return world, err
			}

			if len(chunkSave.Level.Structures.Starts.Shipwreck.Children) < 1 {
				continue
			}

			for _, v := range chunkSave.Level.Structures.Starts.Shipwreck.Children {
				for _, w := range goodShipwrecks {
					if v.Template == w {
						shipwrecksWithIron = append(shipwrecksWithIron, v.Template)
					}
				}
			}
		}
	}

	/*


	   nether


	*/

	netherChunkCoords := world.NetherChunksToBastion()

	region, err := region.Open(fmt.Sprintf(
		"%s/tmp/sfw0/data/world/DIM-1/region/r.%d.%d.mca",
		MustString(os.Getwd()), netherChunkCoords[0].X, netherChunkCoords[0].Z,
	))
	if err != nil {
		return world, err
	}

	percentageOfAir := []int{}
	percentageOfAirAvg := 0
	for _, v := range netherChunkCoords {
		data, err := region.ReadSector(ToSector(v.X), ToSector(v.Z))
		if err != nil {
			return world, err
		}

		var chunkSave save.Chunk
		err = chunkSave.Load(data)
		if err != nil {
			return world, err
		}

		chunkLevel, err := level.ChunkFromSave(&chunkSave)
		if err != nil {
			return world, err
		}

		airBlocks := 0
		for i := 1; i < 9; i++ {
			for j := 0; j < 16*16*16; j++ {
				x := chunkLevel.GetBlockID(i, j)
				if x == "minecraft:air" {
					airBlocks++
				}
			}
		}

		percentageOfAirChunk := int((float64(airBlocks) * 100) / 32768)
		percentageOfAir = append(percentageOfAir, percentageOfAirChunk)
		percentageOfAirAvg += percentageOfAirChunk
	}

	/*          */

	log.Printf("info *+* seed: %s", *world.Seed)
	log.Printf("info *+* shipwrecks with iron: %d (%v)", len(shipwrecksWithIron), shipwrecksWithIron)
	log.Printf(
		"info *+* exposed ravine blocks within %d chunk radius: %d",
		Cfg.Worldgen.RavineProximity,
		exposedRavineBlocks,
	)
	log.Printf("info *+* pc.s of air toward bastion: %v", percentageOfAir)
	if len(percentageOfAir) > 0 {
		percentageOfAirAvg = percentageOfAirAvg / len(percentageOfAir)
		log.Printf("info *+* avg. pc. of air toward bastion: %d", percentageOfAirAvg)
	}

	world.RavineProximity = ToIntRef(Cfg.Worldgen.RavineProximity)
	world.ExposedRavineBlocks = ToIntRef(exposedRavineBlocks)
	world.IronShipwrecks = ToIntRef(len(shipwrecksWithIron))
	world.AvgBastionAir = ToIntRef(percentageOfAirAvg)

	return world, nil
}

func toRegionArea(rx int, rz int, x1 int, z1 int, x2 int, z2 int) (int, int, int, int, error) {
	var err error

	area := fmt.Sprintf("%d %d %d %d", x1, z1, x2, z2)

	if rx < 0 {
		if x1 > rx {
			err = errors.New("")
		}
		if x2 > rx {
			x2 = -1
		}
	} else {
		if x2 < 0 {
			err = errors.New("")
		}
		if x1 < 0 {
			x1 = 0
		}
	}

	if rz < 0 {
		if z1 > rz {
			err = errors.New("")
		}
		if z2 > rz {
			z2 = -1
		}
	} else {
		if z2 < 0 {
			err = errors.New("")
		}
		if z1 < 0 {
			z1 = 0
		}
	}

	if err != nil {
		err = fmt.Errorf("area %s does not overlap region %d %d", area, rx, rz)
	}

	return x1, z1, x2, z2, err
}

var goodShipwrecks = []string{
	"minecraft:shipwreck/rightsideup_backhalf",
	"minecraft:shipwreck/rightsideup_backhalf_degraded",
	"minecraft:shipwreck/rightsideup_full",
	"minecraft:shipwreck/rightsideup_full_degraded",
	"minecraft:shipwreck/sideways_backhalf",
	"minecraft:shipwreck/sideways_backhalf_degraded",
	"minecraft:shipwreck/sideways_full",
	"minecraft:shipwreck/sideways_full_degraded",
	"minecraft:shipwreck/upsidedown_backhalf",
	"minecraft:shipwreck/upsidedown_backhalf_degraded",
	"minecraft:shipwreck/upsidedown_full",
	"minecraft:shipwreck/upsidedown_full_degraded",
	"minecraft:shipwreck/with_mast",
	"minecraft:shipwreck/with_mast_degraded",
}
