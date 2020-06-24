package cli

import (
	"strconv"

	"megaman.genesis.local/sknight/mockc2/internal/log"
	"megaman.genesis.local/sknight/mockc2/pkg/c2"
	"megaman.genesis.local/sknight/mockc2/pkg/protocol"
)

var serverList map[uint16]*c2.Server

func listenerCommand(cmd []string) {
	if serverList == nil {
		serverList = make(map[uint16]*c2.Server)
	}

	if len(cmd) >= 3 {
		if cmd[1] == "start" {
			c, err := strconv.ParseUint(cmd[3], 0, 16)
			if err != nil {
				return
			}
			port := uint16(c)

			if serverList[port] != nil {
				log.Warn("Already listening on port %d", port)
				return
			}

			handler := protocol.HandlerByName(cmd[2])
			if handler != nil {
				s, err := c2.NewServer(port, handler)
				if err != nil {
					log.Warn(err.Error())
					return
				}
				serverList[port] = s
				return
			}
		} else if cmd[1] == "stop" {
			c, err := strconv.ParseUint(cmd[2], 0, 16)
			if err != nil {
				return
			}
			port := uint16(c)

			if s, ok := serverList[port]; ok {
				s.Shutdown()
				delete(serverList, port)
			} else {
				log.Warn("Nothing listening on port %d", port)
			}
		}
		return
	}

	log.Warn("Invalid command")
	log.Info("listener [start|stop] <port>")
}
