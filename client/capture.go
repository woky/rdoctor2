package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type CapturedLine struct {
	Line       string
	LineNumber uint64
	Stderr     bool
	Eof        bool
}

func (c CapturedLine) String() string {
	origin := "STDOUT"
	if c.Stderr {
		origin = "STDERR"
	}
	format := "%s L%03d: %s"
	if c.Eof {
		format = "%s L%03d EOF: %s"
	}
	return fmt.Sprintf(format, origin, c.LineNumber, c.Line)
}

func readLines(pipe io.ReadCloser, lines chan CapturedLine, stderr bool) {
	defer pipe.Close()
	defer close(lines)
	copyOut := os.Stdout
	if stderr {
		copyOut = os.Stderr
	}
	scanner := bufio.NewScanner(pipe)
	var lineNumber uint64 = 0
	for {
		eof := !scanner.Scan()
		capturedLine := CapturedLine{
			Line:       scanner.Text(),
			LineNumber: lineNumber,
			Stderr:     stderr,
			Eof:        eof,
		}
		fmt.Fprintln(copyOut, capturedLine.Line)
		lines <- capturedLine
		if eof {
			break
		}
		lineNumber++
	}
	if err := scanner.Err(); err != nil {
		SayErr("Could not read from pipe: %s", err)
	}
}

func combineOutputs(stdout, stderr io.ReadCloser, lines chan CapturedLine) {
	defer close(lines)
	stdoutLines := make(chan CapturedLine)
	stderrLines := make(chan CapturedLine)
	go readLines(stdout, stdoutLines, false)
	go readLines(stderr, stderrLines, true)
	for {
		select {
		case line, ok := <-stdoutLines:
			if ok {
				lines <- line
			} else {
				stdoutLines = nil
			}
		case line, ok := <-stderrLines:
			if ok {
				lines <- line
			} else {
				stderrLines = nil
			}
		}
		if stdoutLines == nil && stderrLines == nil {
			break
		}
	}
}

func StartMainProgram(cmdLine []string, lines chan CapturedLine) {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, cmdLine[0], cmdLine[1:]...)
	var stdout, stderr io.ReadCloser
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		Die("Could not create pipe: %s", err)
	}
	stderr, err = cmd.StderrPipe()
	if err != nil {
		Die("Could not create pipe: %s", err)
	}
	go combineOutputs(stdout, stderr, lines)
	interrupts := make(chan os.Signal)
	signal.Notify(interrupts, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for {
			<-interrupts
			cancel()
		}
	}()
	err = cmd.Start()
	if err != nil {
		Die("Could not create process: %s", err)
	}
}
