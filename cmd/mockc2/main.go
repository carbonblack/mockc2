package main

import (
	"github.com/carbonblack/mockc2/internal/cli"
)

func main() {
	shell := new(cli.Shell)
	shell.Run()
}
