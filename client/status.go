package main

import (
	"bufio"
	"fmt"
	"os"
)

func say(where *os.File, message string, args []interface{}) {
	fmt.Fprint(where, "rdoctor: ")
	fmt.Fprintf(where, message, args...)
	fmt.Fprintln(where)
}

func SayOut(message string, args ...interface{}) {
	say(os.Stdout, message, args)
}

func SayErr(message string, args ...interface{}) {
	say(os.Stderr, message, args)
}

func Die(message string, args ...interface{}) {
	say(os.Stderr, message, args)
	os.Exit(1)
}

func Prompt(prompt string) string {
	fmt.Printf("rdoctor: %s: ", prompt)
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
