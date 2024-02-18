package lib

import (
	"context"
)

func WorldgenTask(ctx context.Context, world World) (World, error) {
	if err := generateWorld(ctx, world); err != nil {
		return world, err
	}

	worldDatamined, err := datamineWorld(ctx, world)
	if err != nil {
		return world, err
	}

	return worldDatamined, nil
}
