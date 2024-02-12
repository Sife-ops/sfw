package lib

import (
	"context"
	"log"
	"sfw/gen/config"
)

var Cfg *config.Sfw

func init() {
	if err := loadConfig(); err != nil {
		log.Printf("%v", err)
	}
}

func loadConfig() error {
	cfg, err := config.LoadFromPath(context.Background(), "./pkl/amends.pkl")
	if err != nil {
		return err
	}
	Cfg = cfg
	return nil
}
