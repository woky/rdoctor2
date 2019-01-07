package main

import (
	"os"
)

const DevVersion = "git"

var Version string = DevVersion

func main() {
	SayOut("Version: %s", Version)
	if len(os.Args) < 2 {
		Die("Usage: rdoctor <command>")
	}
	config := LoadConfig()
	CheckForUpdate(config)
	RunSetup(config)
	forwarder := ConnectForwarder(config)
	lines := make(chan CapturedLine)
	StartMainProgram(os.Args[1:], lines)
	forwarder.ForwardLines(lines)
}
