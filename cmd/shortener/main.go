package main

import (
	"github.com/sirupsen/logrus"

	"github.com/SerzhLimon/ReductionURL/internal/config"
	"github.com/SerzhLimon/ReductionURL/internal/server"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		logrus.Fatalln(err)
	}
	s, err := server.NewServer(cfg)
	if err != nil {
		logrus.Fatalln(err)
	}
	s.Run()
}
