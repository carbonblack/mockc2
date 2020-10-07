package cli

import (
	"os"

	"github.com/carbonblack/mockc2/internal/log"
)

func exitCommand(cmd []string) {
	log.Warn("Shutting down")
	os.Exit(0)
}
