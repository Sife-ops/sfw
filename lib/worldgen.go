package lib

import (
	"context"
)

func WorldgenTask(ctx context.Context, seed GodSeed) (GodSeed, error) {
	if err := generateWorld(ctx, seed); err != nil {
		return seed, err
	}

	seedWorldgen, err := datamineWorld(ctx, seed)
	if err != nil {
		return seed, err
	}

	return seedWorldgen, nil
}
