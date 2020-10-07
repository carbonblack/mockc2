package cli

import "github.com/carbonblack/mockc2/internal/log"

func debugCommand(cmd []string) {
	if len(cmd) == 2 {
		if cmd[1] == "on" {
			log.DebugEnabled = true
			log.Success("Debug output on")
			return
		} else if cmd[1] == "off" {
			log.DebugEnabled = false
			log.Success("Debug output off")
			return
		}
	}

	log.Warn("Invalid command")
	log.Info("debug [on|off]")
}
