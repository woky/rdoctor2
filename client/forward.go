package main

import (
	"bytes"
	"strconv"

	"github.com/gorilla/websocket"
)

type Forwarder struct {
	conn *websocket.Conn
}

func ConnectForwarder(config *Config) Forwarder {
	url := config.GetSubmitLogUrl()
	for i := 0; i < 10; i++ {
		conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			if err == websocket.ErrBadHandshake && resp != nil && resp.StatusCode/100 == 3 {
				location, err := resp.Location()
				if err == nil {
					url = location.String()
					continue
				}
			}
			SayErr("Could not connect to remote websocket: %s", err)
			if resp != nil {
				SayErr("  HTTP response status: %s", resp.Status)
			}
			Die("Cannot continue")
		}
		return Forwarder{conn}
	}
	panic("")
}

func encodeLine(capturedLine *CapturedLine) []byte {
	var buffer bytes.Buffer
	if !capturedLine.Stderr {
		buffer.WriteString("O")
	} else {
		buffer.WriteString("E")
	}
	if !capturedLine.Eof {
		buffer.WriteString("L")
	} else {
		buffer.WriteString("C")
	}
	buffer.WriteString("|")
	tsMillis := capturedLine.Timestamp.UnixNano() / 1000000
	buffer.Write(strconv.AppendInt([]byte{}, tsMillis, 10))
	buffer.WriteString("|")
	buffer.Write(strconv.AppendUint([]byte{}, capturedLine.LineNumber, 10))
	buffer.WriteString("|")
	buffer.WriteString(capturedLine.Line)
	return buffer.Bytes()
}

func (fwd Forwarder) ForwardLines(lines chan CapturedLine) {
	var err error
	for capturedLine := range lines {
		err = fwd.conn.WriteMessage(websocket.TextMessage, encodeLine(&capturedLine))
		if err != nil {
			break
		}
	}
	fwd.conn.Close()
	if err != nil {
		Die("Could not send to remote websocket: %s", err)
	}
}
