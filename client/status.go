package main

import (
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
