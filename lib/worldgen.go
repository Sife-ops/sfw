package lib

import (
	"context"
)

func WorldgenTask(ctx context.Context, job GodSeed) (GodSeed, error) {
	if err := generateWorld(ctx, job); err != nil {
		return job, err
	}

	godSeed, err := datamineWorld(ctx, job)
	if err != nil {
		return job, err
	}

	return godSeed, nil
}
