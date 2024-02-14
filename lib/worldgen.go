package lib

import (
	"context"
)

///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////
////
//// worldgen task
////

func WorldgenTask(ctx context.Context, job GodSeed) (GodSeed, error) {
	if err := generateWorld(ctx, job); err != nil {
		return job, nil
	}

	dmDone := make(chan GodSeed)
	dmErr := make(chan error)

	// todo doesnt need to be goroutine?
	go datamineWorld(ctx, job, dmDone, dmErr)

	var gs GodSeed
	select {
	case <-ctx.Done():
		return job, nil
	case err := <-dmErr:
		return job, err
	case g := <-dmDone:
		gs = g
	}

	return gs, nil
}
