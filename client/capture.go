package main

/*
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <fcntl.h>
*/
import "C"

import (
	"bufio"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"strings"
	"syscall"
	"time"
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
	format := "%s L%03d \"%s\""
	if c.Eof {
		format = "%s L%03d EOF \"%s\""
	}
	return fmt.Sprintf(format, origin, c.LineNumber, c.Line)
}

func captureOutput(fd uintptr, linesChan chan CapturedLine, stderr bool) {
	defer close(linesChan)
	_, err := unix.FcntlInt(fd, unix.F_SETFL, unix.O_NONBLOCK)
	if err != nil {
		Die("Could not make file descriptor non-blocking: %s", err)
	}
	file := os.NewFile(fd, "")
	defer file.Close()
	reader := bufio.NewReader(file)
	var lineNumber uint64 = 1
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		eof := false
		if err != nil {
			if err != io.EOF {
				SayErr("Could not read from pipe: %s", err)
				break
			}
			eof = true
		}
		capturedLine := CapturedLine{
			Line:       line,
			LineNumber: lineNumber,
			Stderr:     stderr,
			Eof:        eof,
		}
		linesChan <- capturedLine
		if eof {
			break
		}
		lineNumber++
	}
	SayErr("Exit")
}

//func captureOutput(fd uintptr, linesChan chan CapturedLine, stderr bool) {
//	defer close(linesChan)
//	_, err := unix.FcntlInt(fd, unix.F_SETFL, unix.O_NONBLOCK)
//	if err != nil {
//		Die("Could not make file descriptor non-blocking: %s", err)
//	}
//	file := os.NewFile(fd, "")
//	defer file.Close()
//	scanner := bufio.NewScanner(file)
//	var lineNumber uint64 = 1
//	for {
//		SayOut("Reading")
//		eof := !scanner.Scan()
//		capturedLine := CapturedLine{
//			Line:       scanner.Text(),
//			LineNumber: lineNumber,
//			Stderr:     stderr,
//			Eof:        eof,
//		}
//		SayOut("Writing to chan")
//		linesChan <- capturedLine
//		if eof {
//			break
//		}
//		lineNumber++
//	}
//	if err := scanner.Err(); err != nil {
//		SayErr("Could not read from pipe: %s", err)
//	}
//	SayErr("Exit")
//}

func captureBoth(stdoutFd uintptr, stderrFd uintptr, linesChan chan CapturedLine) {
	stdoutLinesChan := make(chan CapturedLine)
	stderrLinesChan := make(chan CapturedLine)
	go captureOutput(stdoutFd, stdoutLinesChan, false)
	go captureOutput(stderrFd, stderrLinesChan, true)
	for {
		select {
		case line, ok := <-stdoutLinesChan:
			if ok {
				linesChan <- line
			} else {
				stdoutLinesChan = nil
			}
		case line, ok := <-stderrLinesChan:
			if ok {
				linesChan <- line
			} else {
				stderrLinesChan = nil
			}
		}
		if stdoutLinesChan == nil && stderrLinesChan == nil {
			close(linesChan)
			break
		}
	}
}

func ForkMainProgram(cmdLine []string, linesChan chan CapturedLine) {
	var stdoutPipe, stderrPipe [2]int
	var err error
	var retInt _Ctype_int

	if len(cmdLine) == 0 {
		panic("Empty cmdLine")
	}

	err = syscall.Pipe(stdoutPipe[:])
	if err != nil {
		Die("Could not create Unix pipe: %s", err)
	}
	err = syscall.Pipe(stderrPipe[:])
	if err != nil {
		Die("Could not create Unix pipe: %s", err)
	}

	retInt, err = C.fork()
	if retInt == -1 {
		Die("Could not fork current process: %s", err)
		os.Exit(1)
	}
	childPid := int(retInt)

	if childPid == 0 {
		time.Sleep(1 * time.Second)
		SayOut("Reading...")
		syscall.Close(stdoutPipe[1])
		syscall.Close(stderrPipe[1])
		os.Stdin.Close()
		go captureBoth(uintptr(stdoutPipe[0]), uintptr(stderrPipe[0]), linesChan)
		return
	}

	// childPid > 0
	syscall.Close(stdoutPipe[0])
	syscall.Close(stderrPipe[0])
	os.Stdout.Close()
	os.Stderr.Close()
	err = syscall.Dup2(stdoutPipe[1], 1)
	if err != nil {
		Die("Could not duplicate file descriptor: %s", err)
	}
	err = syscall.Dup2(stderrPipe[1], 2)
	if err != nil {
		Die("Could not duplicate file descriptor: %s", err)
	}
	os.Stdout = os.NewFile(uintptr(1), "/dev/stdout")
	os.Stderr = os.NewFile(uintptr(2), "/dev/stderr")
	err = syscall.Exec(cmdLine[0], cmdLine, os.Environ())
	if err != nil {
		Die("Could not execute wrapped program: %s", err)
	}
	panic("Successful return from exec()")
}
