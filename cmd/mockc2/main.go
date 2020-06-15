package main

import (
	"megaman.genesis.local/sknight/mockc2/internal/cli"
)

var Version = "0.0.1"

func main() {
	shell := new(cli.Shell)
	shell.Run()
}
