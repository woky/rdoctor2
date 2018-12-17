package main

import (
	"os"
)

func main() {
	if len(os.Args) < 2 {
		Die("Usage: rdoctor <command>")
	}
	config := LoadConfig()
	RunSetup(config)
	forwarder := ConnectForwarder(config)
	lines := make(chan CapturedLine)
	StartMainProgram(os.Args[1:], lines)
	forwarder.ForwardLines(lines)
}
