package cli

import (
	"fmt"

	"github.com/carbonblack/mockc2/pkg/version"
)

func versionCommand(cmd []string) {
	fmt.Printf("  Version   %s\n", version.Version)
	fmt.Printf("  BuildDate %s\n", version.BuildDate)
}
