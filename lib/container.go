package lib

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

var DockerClient *client.Client

// var ContainerName = "McServerTodo"

func init() {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalf("error %v", err)
	}
	DockerClient = dockerClient
}

func KillMcContainer(ctx context.Context) error {
	cl, err := DockerClient.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return err
	}
	for _, v := range cl {
		if strings.Contains(v.Names[0], ctx.Value("inst").(string)) {
			if err := DockerClient.ContainerKill(ctx, v.ID, ""); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func RemoveMcContainer(ctx context.Context) error {
	cl, err := DockerClient.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return err
	}
	for _, v := range cl {
		if strings.Contains(v.Names[0], ctx.Value("inst").(string)) {
			if err := DockerClient.ContainerRemove(ctx, v.ID, types.ContainerRemoveOptions{}); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func ContainerCreateMc(ctx context.Context, seed *string) (container.CreateResponse, error) {
	return DockerClient.ContainerCreate(
		ctx,
		&container.Config{
			Image:     "itzg/minecraft-server",
			Tty:       true,
			OpenStdin: true,
			Env: []string{
				"EULA=true",
				"VERSION=1.16.1",
				fmt.Sprintf("SEED=%s", *seed),
				"MEMORY=2G",
			},
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: fmt.Sprintf("%s/tmp/%s/data", MustString(os.Getwd()), ctx.Value("inst").(string)),
					Target: "/data",
				},
			},
		},
		&network.NetworkingConfig{},
		&ocispec.Platform{},
		ctx.Value("inst").(string),
	)
}

// todo remove container at start
func AwaitMcStopped(ctx context.Context, ms chan error, cid string) {
	for true {
		if ci, err := DockerClient.ContainerInspect(ctx, cid); err != nil {
			ms <- err
			return
		} else if !ci.State.Running {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	ms <- nil
}

// todo be able 2c rcon stdout???
func AwaitMcStarted(ctx context.Context, ms chan error, cid string) {
	for true {
		if ec, err := McExec(ctx, cid, []string{"rcon-cli", "msg @p echo"}); err != nil {
			ms <- err
			return
		} else if ec == 0 {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	ms <- nil
}

// todo return IDResponse?
func McExec(ctx context.Context, cid string, cmd []string) (int, error) {
	ec, err := DockerClient.ContainerExecCreate(ctx, cid, types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		Tty:          true,
		Cmd:          cmd,
	})
	if err != nil {
		return -1, err
	}

	if err := DockerClient.ContainerExecStart(ctx, ec.ID, types.ExecStartCheck{}); err != nil {
		if err != nil {
			return -1, err
		}
	}
	for true { // todo monka
		ei, err := DockerClient.ContainerExecInspect(ctx, ec.ID)
		if err != nil {
			return -1, err
		}
		if !ei.Running {
			break
		}
	}

	ei, err := DockerClient.ContainerExecInspect(ctx, ec.ID)
	if err != nil {
		return -1, err
	}
	return ei.ExitCode, nil
}
