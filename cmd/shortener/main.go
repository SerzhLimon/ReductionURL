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
	psql, _ := db.InitPostgresClient(cfg)

	logrus.Info("Running migrations...")
	err = migrations.Up(psql)
	if err != nil {
		logrus.Warn(err)
	}
	defer func() {
		migrations.Down(psql)
		logrus.Info("Migrations down")
	}()
	logrus.Info("Migrations applied successfully")

	s, err := server.NewServer(cfg, psql)
	if err != nil {
		logrus.Fatalln(err)
	}
	s.Run()
}
