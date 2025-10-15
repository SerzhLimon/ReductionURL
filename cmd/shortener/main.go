package main

import (
	"github.com/SerzhLimon/ReductionURL/internal/server"
)

func main() {
	s := server.NewServer()
	s.Route()
	s.Run()
}
