package cli

import (
	"fmt"
	"strconv"

	"megaman.genesis.local/sknight/mockc2/internal/log"
	"megaman.genesis.local/sknight/mockc2/pkg/c2"
)

var serverList map[uint16]*c2.Server

func listenerCommand(cmd []string) {
	if serverList == nil {
		serverList = make(map[uint16]*c2.Server)
	}

	if len(cmd) >= 3 {
		if cmd[1] == "start" && len(cmd) >= 4 {
			c, err := strconv.ParseUint(cmd[3], 0, 16)
			if err != nil {
				return
			}
			port := uint16(c)

			if serverList[port] != nil {
				log.Warn("Already listening on port %d", port)
				return
			}

			protocol := cmd[2]
			address := fmt.Sprintf(":%d", port)

			s, err := c2.NewServer(protocol, address)
			if err != nil {
				log.Warn(err.Error())
				return
			}

			serverList[port] = s
			return
		} else if cmd[1] == "stop" && len(cmd) >= 3 {
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

			return
		}
	}

	log.Warn("Invalid command")
	log.Info("listener [start|stop] <protocol> <port>")
}
