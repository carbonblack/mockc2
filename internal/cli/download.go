package cli

import (
	"os"

	"github.com/carbonblack/mockc2/internal/log"
	"github.com/carbonblack/mockc2/pkg/c2"
)

func downloadCommand(agentID string, cmd []string) {
	if len(cmd) != 3 {
		log.Warn("Invalid command")
		log.Info("download <source> <destination>")
		return
	}

	if _, err := os.Stat(cmd[2]); !os.IsNotExist(err) {
		log.Warn("Destination file already exists")
		log.Info("download <source> <destination>")
		return
	}

	a := c2.AgentByID(agentID)

	command := c2.DownloadCommand{
		Source:      cmd[1],
		Destination: cmd[2],
	}

	a.SendCommand(command)
}
