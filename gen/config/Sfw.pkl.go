// Code generated from Pkl module `sfw`. DO NOT EDIT.
package config

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Sfw struct {
	// ////////////////////////////////////////////////////////////////////////////
	Postgres Postgres `pkl:"postgres"`

	Log Log `pkl:"log"`

	Web Web `pkl:"web"`

	Websocket Websocket `pkl:"websocket"`

	Lmao Websocket `pkl:"lmao"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Sfw
func LoadFromPath(ctx context.Context, path string) (ret *Sfw, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return nil, err
	}
	defer func() {
		cerr := evaluator.Close()
		if err == nil {
			err = cerr
		}
	}()
	ret, err = Load(ctx, evaluator, pkl.FileSource(path))
	return ret, err
}

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Sfw
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (*Sfw, error) {
	var ret Sfw
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
