package c2

// ExecuteCommand represents a command to run on an agent.
type ExecuteCommand struct {
	Name string
	Args []string
}

// UploadCommand represents the data needed to upload a file to a given agent.
type UploadCommand struct {
	Source      string
	Destination string
}

// DownloadCommand represents the data needed to download a file from a given agent.
type DownloadCommand struct {
	Source      string
	Destination string
}
