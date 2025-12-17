package main

import (
	"github.com/sirupsen/logrus"

	"github.com/SerzhLimon/ReductionURL/internal/config"
	"github.com/SerzhLimon/ReductionURL/internal/config/db"
	"github.com/SerzhLimon/ReductionURL/internal/server"
	"github.com/SerzhLimon/ReductionURL/migrations"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		logrus.Fatalln(err)
	}
	psql, err := db.InitPostgresClient(cfg)
	if err != nil {
		logrus.Warn(err)
	}

	logrus.Info("Running migrations...")
	err = migrations.Up(psql)
	if err != nil {
		logrus.Warn(err)
	} else {
		logrus.Info("Migrations applied successfully")
	}
	defer func() {
		//save data
		// migrations.Down(psql)
		// logrus.Info("Migrations down")
	}()

	s, err := server.NewServer(cfg, psql)
	if err != nil {
		logrus.Fatalln(err)
	}
	s.Run()
}
