package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"megaman.genesis.local/sknight/mockc2/pkg/c2"
)

func listCommand(cmd []string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	fmt.Fprintf(w, "Agent ID\tIP\tLast Seen\n")

	for _, a := range c2.Agents() {
		fmt.Fprintf(w, "%s\t%v\t%s\n", a.ID, a.Addr, a.LastSeen.UTC().Format(time.RFC3339))
	}

	w.Flush()
	fmt.Println("")
}
