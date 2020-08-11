package protocol

// A Delegate represents an external delegate that the protocolhandler can
// notify about data being processed.
type Delegate interface {
	SendData(data []byte)
	AgentConnected(id string)
	CloseConnection()
}

// A Handler represents a type capable of handling and decoding C2 traffic.
type Handler interface {
	NeedsTLS() bool
	SetDelegate(delegate Delegate)
	Accept()
	ReceiveData(data []byte)
	Execute(name string, args []string)
	Upload(source string, destination string)
	Download(source string, destination string)
	Close()
}
