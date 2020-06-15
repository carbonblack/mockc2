package cli

import (
	"os"

	"megaman.genesis.local/sknight/mockc2/internal/log"
)

func exitCommand(cmd []string) {
	log.Warn("Shutting down")
	os.Exit(0)
}
