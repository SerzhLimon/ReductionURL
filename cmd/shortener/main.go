package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/SerzhLimon/ReductionURL/internal/config"
	"github.com/SerzhLimon/ReductionURL/internal/config/db"
	"github.com/SerzhLimon/ReductionURL/internal/server"
	"github.com/SerzhLimon/ReductionURL/migrations"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
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

	ctx, cancel := context.WithCancel(context.Background())
	go s.RunAudit(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		if err := s.Run(); err != nil {
			serverErr <- err
		}
	}()

	logrus.Printf("Build version: %s\n", buildVersion)
	logrus.Printf("Build date: %s\n", buildDate)
	logrus.Printf("Build commit: %s\n", buildCommit)

	select {
	case <-quit:
		logrus.Info("Shutdown signal received")
	case err := <-serverErr:
		logrus.WithError(err).Error("Server error occurred")
	}

	cancel()
	time.Sleep(500 * time.Millisecond)

	logrus.Info("Shutting down...")

}
