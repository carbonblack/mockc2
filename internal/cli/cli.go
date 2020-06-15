package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"megaman.genesis.local/sknight/mockc2/internal/log"
	"megaman.genesis.local/sknight/mockc2/pkg/version"
)

type shellMenu int

const (
	Main shellMenu = iota
	Agent
)

type Shell struct {
	rl             *readline.Instance
	mainCompleter  *readline.PrefixCompleter
	agentCompleter *readline.PrefixCompleter
	menu           shellMenu
}

func (s *Shell) initCompleters() {
	s.mainCompleter = readline.NewPrefixCompleter(
		readline.PcItem("debug",
			readline.PcItem("on"),
			readline.PcItem("off"),
		),
		readline.PcItem("exit"),
		readline.PcItem("help"),
		readline.PcItem("interact"),
		readline.PcItem("listener"),
		readline.PcItem("list"),
		readline.PcItem("version"),
	)

	s.agentCompleter = readline.NewPrefixCompleter(
		readline.PcItem("exec"),
		readline.PcItem("exit"),
		readline.PcItem("help"),
		readline.PcItem("download"),
		readline.PcItem("main"),
		readline.PcItem("upload"),
	)
}

func (s *Shell) completer() *readline.PrefixCompleter {
	switch s.menu {
	default:
		fallthrough
	case Main:
		return s.mainCompleter
	case Agent:
		return s.agentCompleter
	}
}

func (s *Shell) initReadline() {
	s.initCompleters()

	l, err := readline.NewEx(&readline.Config{
		Prompt:              "mockc2> ",
		HistoryFile:         "/tmp/mockc2.tmp",
		HistorySearchFold:   true,
		AutoComplete:        s.completer(),
		FuncFilterInputRune: filterInput,
	})

	if err != nil {
		panic(err)
	}

	s.rl = l
	s.setMenu(Main)
}

func (s *Shell) prompt() string {
	switch s.menu {
	default:
		fallthrough
	case Main:
		return "mockc2> "
	case Agent:
		return "agent[1]> "
	}
}

func (s *Shell) setMenu(menu shellMenu) {
	s.menu = menu
	s.rl.Config.AutoComplete = s.completer()
	s.rl.SetPrompt(s.prompt())
}

func (s *Shell) Run() {
	s.initReadline()
	defer s.rl.Close()

	for {
		line, err := s.rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		cmd := strings.Fields(line)

		if len(cmd) > 0 {
			switch s.menu {
			case Main:
				s.mainMenuHandler(cmd)
			case Agent:
				s.agentMenuHandler(cmd)
			}
		}
	}
}

func (s *Shell) mainMenuHandler(cmd []string) {
	switch cmd[0] {
	case "debug":
		debugCommand(cmd)
	case "exit", "quit":
		s.exit()
	case "help", "?":
		printMainMenuHelp()
	case "interact":
		s.setMenu(Agent)
	case "version":
		printVersion()
	default:
		log.Warn("Invalid command")
	}
}

func (s *Shell) agentMenuHandler(cmd []string) {
	switch cmd[0] {
	case "exit", "quit":
		s.exit()
	case "help", "?":
		printAgentMenuHelp()
	case "main":
		s.setMenu(Main)
	default:
		log.Warn("Invalid command")
	}
}

func (s *Shell) exit() {
	log.Warn("Shutting down")
	os.Exit(0)
}

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

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

func printVersion() {
	fmt.Printf("  Version   %s\n", version.Version)
	fmt.Printf("  BuildDate %s\n", version.BuildDate)
}

func printMainMenuHelp() {
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

func printAgentMenuHelp() {
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
