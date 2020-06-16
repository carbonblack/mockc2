package cli

import "fmt"

func mainMenuCommand(cmd []string) {
	fmt.Println("Main Menu Help")
	fmt.Println("")
	fmt.Println("  debug       Enable or disable debug output [on/off]")
	fmt.Println("  exit        Exit and shut down mockc2")
	fmt.Println("  help        Print the main menu help")
	fmt.Println("  interact    Interact with connected agents")
	fmt.Println("  listener    Start or shutdown a protocol listener")
	fmt.Println("  list        List connected agents")
	fmt.Println("  version     Print the mockc2 server version")
	fmt.Println("")
}

func agentMenuCommand(cmd []string) {
	fmt.Println("Agent Menu Help")
	fmt.Println("")
	fmt.Println("  exec        Execute a command on the agent")
	fmt.Println("  exit        Exit and shut down mockc2")
	fmt.Println("  help        Print the agent menu help")
	fmt.Println("  download    Download a file from the agent")
	fmt.Println("  main        Return to the main menu")
	fmt.Println("  upload      Upload a file to the agent")
	fmt.Println("")
}
