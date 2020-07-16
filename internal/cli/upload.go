package cli

import (
	"os"

	"megaman.genesis.local/sknight/mockc2/internal/log"
	"megaman.genesis.local/sknight/mockc2/pkg/c2"
)

func uploadCommand(agentID string, cmd []string) {
	if len(cmd) != 3 {
		log.Warn("Invalid command")
		log.Info("upload <source> <destination>")
		return
	}

	if _, err := os.Stat(cmd[1]); os.IsNotExist(err) {
		log.Warn("Source file not found")
		log.Info("upload <source> <destination>")
		return
	}

	a := c2.AgentByID(agentID)

	command := c2.UploadCommand{
		Source:      cmd[1],
		Destination: cmd[2],
	}

	a.SendCommand(command)
}
