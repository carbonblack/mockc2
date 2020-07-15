package c2

import (
	"net"
	"time"
)

var agents map[string]*Agent

func init() {
	agents = make(map[string]*Agent)
}

// An Agent represents a malware client that has connected to the server.
type Agent struct {
	ID       string
	LastSeen time.Time
	Addr     net.Addr
	conn     *c2Conn
}

// SendCommand instructs the agent to run the given command.
func (a *Agent) SendCommand(command interface{}) {
	a.conn.SendCommand(command)
}

// Agents returns the list of agents that have been seen.
func Agents() []*Agent {
	results := make([]*Agent, len(agents))
	i := 0
	for _, v := range agents {
		results[i] = v
		i++
	}

	return results
}

// AddAgent adds a new agent to the list of seen agents.
func AddAgent(agent *Agent) {
	agent.LastSeen = time.Now()
	agents[agent.ID] = agent
}

// AgentExists checks if a given agent ID is in the list of agents.
func AgentExists(ID string) bool {
	if _, ok := agents[ID]; ok {
		return true
	}

	return false
}

// AgentByID returns an agent instance given the corresponding Agent ID.
func AgentByID(ID string) *Agent {
	return agents[ID]
}
