package main

import (
	"fmt"
	"os"
)

func say(where *os.File, message string, args []interface{}) {
	fmt.Print("rdoctor: ")
	fmt.Printf(message, args...)
	fmt.Println()
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
