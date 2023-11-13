package bootstrap

import (
	"context"
	"github.com/AhegaoHD/WBT/config"
	"github.com/AhegaoHD/WBT/pkg/postgres"
	"log"
)

func Run(cfg *config.Config) {
	pg, err := postgres.New(postgres.GetConnString(&cfg.Db), postgres.MaxPoolSize(cfg.Db.MaxPoolSize))
	if err != nil {
		log.Fatal("APP - START - POSTGRES INI PROBLEM: %v", err)
	}
	defer pg.Close()

	err = pg.Pool.Ping(context.Background())
	if err != nil {
		log.Fatal("APP - START - POSTGRES INI PROBLEM: %v", err)
		return
	}
}
