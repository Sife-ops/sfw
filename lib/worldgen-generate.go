package lib

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
)

///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////
////
//// generate world
////

func generateWorld(ctx context.Context, world World) error {
	log.Printf("info generating world files for %s", *world.Seed)

	log.Printf("info killing old container")
	if err := KillMcContainer(ctx); err != nil {
		if !strings.Contains(err.Error(), "is not running") {
			return err
		}
	}

	log.Printf("info removing old container")
	if err := RemoveMcContainer(ctx); err != nil {
		return err
	}

	log.Printf("info creating instance folder")
	cmdMkDirInst := exec.CommandContext(ctx, "mkdir", "-p",
		fmt.Sprintf("%s/tmp/sfw0/data", MustString(os.Getwd())),
	)
	if outMkDirInst, err := cmdMkDirInst.Output(); err != nil {
		log.Printf("info error creating instance folder: %s %v", string(outMkDirInst), err)
		return err
	}

	log.Printf("info deleting previous world folder")
	cmdRmRfWorld := exec.CommandContext(ctx, "sudo", "rm", "-rf", // todo susdo
		fmt.Sprintf("%s/tmp/sfw0/data/world", MustString(os.Getwd())),
	)
	if outRmRfWorld, err := cmdRmRfWorld.Output(); err != nil {
		log.Printf("info error deleting world folder: %s %v", string(outRmRfWorld), err)
		return err
	}

	log.Printf("info starting minecraft server container")
	mc, err := ContainerCreateMc(ctx, world.Seed)
	if err != nil {
		return err
	}
	if err := DockerClient.ContainerStart(context.TODO(), mc.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	McStarted := make(chan error)
	go AwaitMcStarted(ctx, McStarted, mc.ID)

	log.Printf("info waiting for minecraft server to start")
	if err := <-McStarted; err != nil {
		return err
	}

	McStopped := make(chan error)
	go AwaitMcStopped(ctx, McStopped, mc.ID)

	///////////////////////////////////////////////////////////////////////////
	// forceload chunks

	// overworld
	if ec, err := McExec(ctx, mc.ID, []string{"rcon-cli",
		fmt.Sprintf(
			"forceload add %d %d %d %d",
			world.RavineAreaX1(), world.RavineAreaZ1(), world.RavineAreaX2(), world.RavineAreaZ2(),
		),
	}); ec != 0 && err != nil {
		return err
	} else {
		log.Printf(
			"info rcon forceloaded overworld area %d %d %d %d",
			world.RavineAreaX1(), world.RavineAreaZ1(), world.RavineAreaX2(), world.RavineAreaZ2(),
		)
	}

	// nether
	forceloadedNetherChunks := []Coords{}
	for _, v := range world.NetherChunksToBastion() {
		forceloadedNetherChunks = append(forceloadedNetherChunks, Coords{X: v.X, Z: v.Z})
		if ec, err := McExec(ctx, mc.ID, []string{"rcon-cli",
			fmt.Sprintf(
				"execute in minecraft:the_nether run forceload add %d %d %d %d",
				v.X*16, v.Z*16, (v.X*16)+15, (v.Z*16)+15,
			),
		}); ec != 0 && err != nil {
			return err
		}
	}
	log.Printf("info rcon forceloaded nether chunks: %v", forceloadedNetherChunks)

	// stop server
	if ec, err := McExec(ctx, mc.ID, []string{"rcon-cli", "stop"}); ec != 0 && err != nil {
		return err
	} else {
		log.Printf("info rcon stopped server")
	}

	log.Printf("info waiting for minecraft server to stop")
	if err := <-McStopped; err != nil {
		return err
	}

	log.Printf("info chmod data folder")
	cmdChmodData := exec.CommandContext(ctx, "sudo", "chmod", "-R", "a+rw",
		fmt.Sprintf("%s/tmp/sfw0/data", MustString(os.Getwd())),
	)
	if outChmodData, err := cmdChmodData.Output(); err != nil {
		log.Printf("error chmod data folder: %s %v", string(outChmodData), err)
		return err
	}

	return nil
}
