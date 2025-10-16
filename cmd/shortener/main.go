package main

import (
	"log"

	"github.com/SerzhLimon/ReductionURL/internal/config"
	"github.com/SerzhLimon/ReductionURL/internal/server"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalln(err)
	}
	s := server.NewServer(cfg)
	s.Run()
}
