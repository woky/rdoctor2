package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

const prefix = "rdoctor: "

func PrintOut(message string, args ...interface{}) {
	var buffer bytes.Buffer
	buffer.WriteString(prefix)
	buffer.WriteString(fmt.Sprintf(message, args...))
	fmt.Print(buffer.String())
}

func PrintErr(message string, args ...interface{}) {
	var buffer bytes.Buffer
	buffer.WriteString(prefix)
	buffer.WriteString("ERROR: ")
	buffer.WriteString(fmt.Sprintf(message, args...))
	fmt.Fprint(os.Stderr, buffer.String())
}

func SayOut(message string, args ...interface{}) {
	var buffer bytes.Buffer
	buffer.WriteString(prefix)
	buffer.WriteString(fmt.Sprintf(message, args...))
	buffer.WriteString("\n")
	fmt.Print(buffer.String())
}

func sayStderr(level string, message string, args ...interface{}) {
	var buffer bytes.Buffer
	buffer.WriteString(prefix)
	buffer.WriteString(level)
	buffer.WriteString(fmt.Sprintf(message, args...))
	buffer.WriteString("\n")
	fmt.Fprint(os.Stderr, buffer.String())
}

func SayErr(message string, args ...interface{}) {
	sayStderr("ERROR: ", message, args...)
}

func Warn(message string, args ...interface{}) {
	sayStderr("WARNING: ", message, args...)
}

func Die(message string, args ...interface{}) {
	SayErr(message, args...)
	os.Exit(1)
}

func Prompt(prompt string) string {
	PrintOut("%s: ", prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		SayOut("")
		err := scanner.Err()
		if err != nil {
			Die("Could not read input: %s", err)
		} else {
			Die("EOF while waiting for input")
		}
	}
	return scanner.Text()
}
