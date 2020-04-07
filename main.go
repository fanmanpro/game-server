package main

import (
	"fmt"

	"github.com/fanmanpro/game-server/server"
)

func main() {
	gameServer := server.NewServer()

	for {
		err := gameServer.Start()
		fmt.Println(err)
		defer gameServer.Stop(err)
	}
}
