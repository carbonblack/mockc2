package cli

import (
	"fmt"

	"megaman.genesis.local/sknight/mockc2/pkg/agents"
)

func listCommand(cmd []string) {
	// Ignore all of the cmd strings

	// TODO: Need to print fixed width formatting and better timestamp for last seen
	fmt.Printf("Id               IP              Last Seen\n")
	for _, a := range agents.Agents() {
		fmt.Printf("%s     %v     %s\n", a.Id[:12], a.Addr, a.LastSeen)
	}
}
