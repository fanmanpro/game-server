package main

import (
	"fmt"

	"github.com/fanmanpro/game-server/server"
)

func main() {
	gameServer := server.NewServer()

	err := gameServer.Start()
	fmt.Println(err)
	defer gameServer.Stop()
}
