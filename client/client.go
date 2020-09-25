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

func MakeConn(addr string, port int, rtimeout, wtimeout time.Duration) (*Conn, error) {
	tcpcon, err := net.Dial("tcp", addr+":"+strconv.Itoa(port))
	if err != nil {
		return &Conn{}, err
	}
	conn := Conn{
		sync.Mutex{},
		tcpcon,
		time.Second * rtimeout,
		time.Second * wtimeout,
		bufio.NewReader(tcpcon),
		bufio.NewWriter(tcpcon),
	}
	return &conn, nil
}

func (c *Conn) writeBytes(b []byte) error {
	_, err := c.bw.Write(b)
	if err != nil {
		return err
	}
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
}

func (c *Conn) write(cmd string, args ...interface{}) error {
	err := c.writeArg([]byte(cmd))
	//	log.Println(cmd)
	if err != nil {
		return err
	}
	for _, arg := range args {
		c.writeBytes([]byte{' '})
		//		log.Println(arg)
		err := c.writeArg(arg)
		if err != nil {
			return err
		}
	}
	_, err = c.bw.WriteString("\n")
	if err != nil {
		return err
	}
	return c.bw.Flush()
}

func (c *Conn) readLine() (string, error) {
	return c.br.ReadString('\n')
}
func (c *Conn) readStatus() (string, error) {
	stat, err := c.br.ReadString(':')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(stat, ":"), err
}

/*
ok:ashdha;\n
asddafa;\n
dff\n
*/

func (c *Conn) readBody() (string, error) {
	ret := []string{}
	for {
		line, err := c.readLine()
		if err != nil {
			return "", err
		}
		ret = append(ret, strings.Replace(line, ";\n", "\n", 1))
		if len(line) < 2 || line[len(line)-2] != ';' {
			break
		}
	}
	return strings.Join(ret, ""), nil
}

func (c *Conn) read() (string, error) {
	stat, err := c.readStatus()
	if err != nil {
		return "", err
	}
	if stat == "ok" {
		return c.readBody()
	} else {
		body, err := c.readBody()
		if err != nil {
			return "", err
		}
		return "", errors.New(body)
	}
}

func (c *Conn) Do(cmd string, args ...interface{}) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	err := c.write(cmd, args...)
	if err != nil {
		return "", err
	}
	resp, err := c.read()
	if err != nil {
		return "", err
	}

	return resp, nil
}
