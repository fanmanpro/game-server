package main

import (
	"fmt"

	"github.com/fanmanpro/game-server/server"
)

func main() {
	err := server.Start()
	if err != nil {
		fmt.Println(err)
	}
}
