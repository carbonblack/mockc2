package cli

import (
	"fmt"

	"megaman.genesis.local/sknight/mockc2/pkg/version"
)

func versionCommand(cmd []string) {
	fmt.Printf("  Version   %s\n", version.Version)
	fmt.Printf("  BuildDate %s\n", version.BuildDate)
}
