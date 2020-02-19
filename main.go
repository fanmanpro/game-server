package main

import "github.com/fanmanpro/game-server/server"

func main() {
	gameServer := server.NewServer()

	gameServer.Start()
}
