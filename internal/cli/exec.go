package cli

import (
	"github.com/carbonblack/mockc2/pkg/c2"
)

func execCommand(agentID string, cmd []string) {
	a := c2.AgentByID(agentID)

	command := c2.ExecuteCommand{
		Name: cmd[1],
		Args: cmd[2:],
	}

	a.SendCommand(command)
}
