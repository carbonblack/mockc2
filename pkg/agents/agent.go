package agents

import (
	"net"
	"time"
)

var agents map[string]*Agent

func init() {
	agents = make(map[string]*Agent)
}

type Agent struct {
	Id       string
	LastSeen time.Time
	Addr     net.Addr
}

func Agents() []*Agent {
	results := make([]*Agent, len(agents))
	i := 0
	for _, v := range agents {
		results[i] = v
		i++
	}

	// TODO: Sort before returning

	return results
}

func AddAgent(agent *Agent) {
	if a, ok := agents[agent.Id]; ok {
		a.LastSeen = time.Now()
	} else {
		agent.LastSeen = time.Now()
		agents[agent.Id] = agent
	}
}
