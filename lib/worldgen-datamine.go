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

func datamineWorld(ctx context.Context, godSeed GodSeed) (GodSeed, error) {
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
			return godSeed, err
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
			godSeed.RavineAreaX1(), godSeed.RavineAreaZ1(),
			godSeed.RavineAreaX2(), godSeed.RavineAreaZ2(),
		)
		if err != nil {
			log.Printf("info ravine %s", err.Error())
			continue
		}

		for xc := x1 / 16; xc < (x2+1)/16; xc++ {
			for zc := z1 / 16; zc < (z2+1)/16; zc++ {
				data, err := regions[rx+1][rz+1].ReadSector(ToSector(xc), ToSector(zc))
				if err != nil {
					return godSeed, err
				}

				var chunkSave save.Chunk
				err = chunkSave.Load(data)
				if err != nil {
					return godSeed, err
				}

				chunkLevel, err := level.ChunkFromSave(&chunkSave)
				if err != nil {
					return godSeed, err
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
	}

	shipwrecksWithIron := []string{}
	for q := 0; q < 4; q++ {
		rx := (q / 2) - 1
		rz := (q % 2) - 1

		x1, z1, x2, z2, err := toRegionArea(
			rx, rz,
			godSeed.ShipwreckAreaX1(), godSeed.ShipwreckAreaZ1(),
			godSeed.ShipwreckAreaX2(), godSeed.ShipwreckAreaZ2(),
		)
		if err != nil {
			log.Printf("info shipwreck %s", err.Error())
			continue
		}

		for xc := x1 / 16; xc < (x2+1)/16; xc++ {
			for zc := z1 / 16; zc < (z2+1)/16; zc++ {
				data, err := regions[rx+1][rz+1].ReadSector(ToSector(xc), ToSector(zc))
				if err != nil {
					return godSeed, err
				}

				var chunkSave save.Chunk
				err = chunkSave.Load(data)
				if err != nil {
					return godSeed, err
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
	}

	/*


	   nether


	*/

	netherChunkCoords := godSeed.NetherChunksToBastion()

	region, err := region.Open(fmt.Sprintf(
		"%s/tmp/sfw0/data/world/DIM-1/region/r.%d.%d.mca",
		MustString(os.Getwd()), netherChunkCoords[0].X, netherChunkCoords[0].Z,
	))
	if err != nil {
		return godSeed, err
	}

	percentageOfAir := []int{}
	percentageOfAirAvg := 0
	for _, v := range netherChunkCoords {
		data, err := region.ReadSector(ToSector(v.X), ToSector(v.Z))
		if err != nil {
			return godSeed, err
		}

		var chunkSave save.Chunk
		err = chunkSave.Load(data)
		if err != nil {
			return godSeed, err
		}

		chunkLevel, err := level.ChunkFromSave(&chunkSave)
		if err != nil {
			return godSeed, err
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

	log.Printf("info *+* seed: %s", *godSeed.Seed)
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

	godSeed.RavineProximity = ToIntRef(Cfg.Worldgen.RavineProximity)
	godSeed.ExposedRavineBlocks = ToIntRef(exposedRavineBlocks)
	godSeed.IronShipwrecks = ToIntRef(len(shipwrecksWithIron))
	godSeed.AvgBastionAir = ToIntRef(percentageOfAirAvg)

	return godSeed, nil
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
