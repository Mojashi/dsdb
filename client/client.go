package client

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Conn struct {
	mutex sync.Mutex
	conn  net.Conn

	readTimeout  time.Duration
	writeTimeout time.Duration
	br           *bufio.Reader
	bw           *bufio.Writer
}

func (c *Conn) writeBytes(b []byte) error {
	c.bw.Write(b)
	_, err := c.bw.WriteString("\r\n")
	return err
}

func (c *Conn) writeArg(arg interface{}) error {
	switch arg := arg.(type) {
	case string:
		return c.writeBytes([]byte(arg))
	case int:
		return c.writeBytes([]byte(strconv.Itoa(arg)))
	case int64:
		return c.writeBytes([]byte(strconv.Itoa(int(arg))))
	case []byte:
		return c.writeBytes(arg)
	case nil:
		return c.writeBytes([]byte{})
	default:
		var buf bytes.Buffer
		fmt.Fprint(&buf, arg)
		return c.writeBytes(buf.Bytes())
	}
	return nil
}

func (c *Conn) write(cmd string, args ...interface{}) error {
	c.writeArg([]byte(cmd))
	for _, arg := range args {
		err := c.writeArg(arg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Conn) readLine() (string, error) {
	return c.br.ReadString('\n')
}

func (c *Conn) read() (string, error) {
	line, err := c.readLine()
	if err != nil {
		return "", err
	}
	secs := strings.SplitN(line, ":", 2)
	if secs[0] == "ok" {
		return secs[1], nil
	} else {
		return "", errors.New(secs[1])
	}
}

func (c *Conn) Do(cmd string, args ...interface{}) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.write(cmd, args)
	resp, err := c.read()
	if err != nil {
		return "", err
	}

	return resp, nil
}
