package lib

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Tnze/go-mc/level"
	"github.com/Tnze/go-mc/save"
	"github.com/Tnze/go-mc/save/region"
)

///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////
////
//// datamine world
////
//// this part is RLLY bad
////

func datamineWorld(ctx context.Context, job GodSeed, dmDone chan GodSeed, dmErr chan error) {

	///////////////////////////////////////////////////////////////////////////
	// overworld checks

	shipwrecksWithIron := []string{}
	exposedRavineBlocks := 0

	for quadrant := 0; quadrant < 4; quadrant++ {
		// todo swap x/z
		regionX := (quadrant % 2) - 1
		regionZ := (quadrant / 2) - 1

		x1 := job.RavineAreaX1()
		z1 := job.RavineAreaZ1()
		x2 := job.RavineAreaX2()
		z2 := job.RavineAreaZ2()

		if regionX < 0 {
			if x1 >= 0 {
				log.Printf("info no overlap with -X regions")
				goto NextQuadrant
			}
			if x2 >= 0 {
				x2 = -1
			}
		} else {
			if x2 < 0 {
				log.Printf("info no overlap with +X regions")
				goto NextQuadrant
			}
			if x1 < 0 {
				x1 = 0
			}
		}
		if regionZ < 0 {
			if z1 >= 0 {
				log.Printf("info no overlap with -Z regions")
				goto NextQuadrant
			}
			if z2 >= 0 {
				z2 = -1
			}
		} else {
			if z2 < 0 {
				log.Printf("info no overlap with +Z regions")
				goto NextQuadrant
			}
			if z1 < 0 {
				z1 = 0
			}
		}
		goto OpenRegion

	NextQuadrant:
		log.Printf(
			"info skipping region %d,%d due to no overlap with ravine area around shipwreck %d %d %d %d",
			regionX, regionZ,
			job.RavineAreaX1(), job.RavineAreaZ1(), job.RavineAreaX2(), job.RavineAreaZ2(),
		)
		continue

	OpenRegion:
		region, err := region.Open(fmt.Sprintf(
			"%s/tmp/sfw0/data/world/region/r.%d.%d.mca",
			MustString(os.Getwd()), regionX, regionZ,
		))
		if err != nil {
			dmErr <- err
			return
		}

		for xC := x1 / 16; xC < (x2+1)/16; xC++ {
			for zC := z1 / 16; zC < (z2+1)/16; zC++ {
				// todo check sector exists
				// if !region.ExistSector(0, 0) {
				// 	log.Fatal("sector no existo")
				// }

				data, err := region.ReadSector(ToSector(xC), ToSector(zC))
				if err != nil {
					dmErr <- err
					return
				}

				var chunkSave save.Chunk
				err = chunkSave.Load(data)
				if err != nil {
					dmErr <- err
					return
				}

				chunkLevel, err := level.ChunkFromSave(&chunkSave)
				if err != nil {
					dmErr <- err
					return
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

		x1 = job.ShipwreckAreaX1()
		z1 = job.ShipwreckAreaZ1()
		x2 = job.ShipwreckAreaX2()
		z2 = job.ShipwreckAreaZ2()

		if regionX < 0 {
			if x1 >= 0 {
				region.Close()
			}
			if x2 >= 0 {
				x2 = -1
			}
		} else {
			if x2 < 0 {
				region.Close()
			}
			if x1 < 0 {
				x1 = 0
			}
		}
		if regionZ < 0 {
			if z1 >= 0 {
				region.Close()
			}
			if z2 >= 0 {
				z2 = -1
			}
		} else {
			if z2 < 0 {
				region.Close()
			}
			if z1 < 0 {
				z1 = 0
			}
		}

		for xC := x1 / 16; xC < (x2+1)/16; xC++ {
			for zC := z1 / 16; zC < (z2+1)/16; zC++ {
				data, err := region.ReadSector(ToSector(xC), ToSector(zC))
				if err != nil {
					dmErr <- err
					return
				}
				var chunkSave save.Chunk
				err = chunkSave.Load(data)
				if err != nil {
					dmErr <- err
					return
				}
				if len(chunkSave.Level.Structures.Starts.Shipwreck.Children) < 1 {
					continue
				}
				for _, v := range chunkSave.Level.Structures.Starts.Shipwreck.Children {
					if v.Template == "minecraft:shipwreck/rightsideup_backhalf" ||
						v.Template == "minecraft:shipwreck/rightsideup_backhalf_degraded" ||
						v.Template == "minecraft:shipwreck/rightsideup_full" ||
						v.Template == "minecraft:shipwreck/rightsideup_full_degraded" ||
						v.Template == "minecraft:shipwreck/sideways_backhalf" ||
						v.Template == "minecraft:shipwreck/sideways_backhalf_degraded" ||
						v.Template == "minecraft:shipwreck/sideways_full" ||
						v.Template == "minecraft:shipwreck/sideways_full_degraded" ||
						v.Template == "minecraft:shipwreck/upsidedown_backhalf" ||
						v.Template == "minecraft:shipwreck/upsidedown_backhalf_degraded" ||
						v.Template == "minecraft:shipwreck/upsidedown_full" ||
						v.Template == "minecraft:shipwreck/upsidedown_full_degraded" ||
						v.Template == "minecraft:shipwreck/with_mast" ||
						v.Template == "minecraft:shipwreck/with_mast_degraded" {
						shipwrecksWithIron = append(shipwrecksWithIron, v.Template)
					}
				}
			}
		}
	}

	///////////////////////////////////////////////////////////////////////////
	// nether checks
	// todo ckeck for lava lake

	netherChunkCoords := job.NetherChunksToBastion()

	region, err := region.Open(fmt.Sprintf(
		"%s/tmp/sfw0/data/world/DIM-1/region/r.%d.%d.mca",
		MustString(os.Getwd()), netherChunkCoords[0].X, netherChunkCoords[0].Z,
	))
	if err != nil {
		dmErr <- err
		return
	}

	percentageOfAir := []int{}
	percentageOfAirAvg := 0
	for _, v := range netherChunkCoords {
		// log.Printf("info chunk %d %d", (v.X), (v.Z))
		// log.Printf("info sector %d %d", ToSector(v.X), ToSector(v.Z))
		data, err := region.ReadSector(ToSector(v.X), ToSector(v.Z))
		if err != nil {
			dmErr <- err
			return
		}

		var chunkSave save.Chunk
		err = chunkSave.Load(data)
		if err != nil {
			dmErr <- err
			return
		}

		chunkLevel, err := level.ChunkFromSave(&chunkSave)
		if err != nil {
			dmErr <- err
			return
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

	log.Printf("info > *+*+*+*+* GENERATED GOD SEED *+*+*+*+*")
	log.Printf("info > seed: %s", *job.Seed)
	log.Printf("info > shipwrecks with iron: %d (%v)", len(shipwrecksWithIron), shipwrecksWithIron)
	log.Printf(
		"info > exposed ravine blocks within %d chunk radius: %d",
		Cfg.Worldgen.RavineProximity,
		exposedRavineBlocks,
	)
	log.Printf("info > pc.s of air toward bastion: %v", percentageOfAir)
	if len(percentageOfAir) > 0 {
		percentageOfAirAvg = percentageOfAirAvg / len(percentageOfAir)
		log.Printf("info > avg. pc. of air toward bastion: %d", percentageOfAirAvg)
	}
	log.Printf("info > *+*+*+*+*+*+*+*+*+*+*+*+*+*+*+*+*+*+*")

	job.RavineProximity = ToIntRef(Cfg.Worldgen.RavineProximity)
	job.ExposedRavineBlocks = ToIntRef(exposedRavineBlocks)
	job.IronShipwrecks = ToIntRef(len(shipwrecksWithIron))
	job.AvgBastionAir = ToIntRef(percentageOfAirAvg)

	dmDone <- job
}
