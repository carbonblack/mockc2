package cli

import (
	"fmt"

	"megaman.genesis.local/sknight/mockc2/pkg/c2"
)

func listCommand(cmd []string) {
	// Ignore all of the cmd strings

	// TODO: Need to print fixed width formatting and better timestamp for last seen
	fmt.Printf("Id                                                                 IP                  Last Seen\n")
	for _, a := range c2.Agents() {
		fmt.Printf("%s     %v     %s\n", a.ID, a.Addr, a.LastSeen)
	}
}
