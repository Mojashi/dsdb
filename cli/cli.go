package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Mojashi/dsdb/client"
)

var portNum = 5003
var conn client.Conn
var addr = "localhost"

func main() {
	conn, err := client.MakeConn(addr, portNum, 500, 500)
	if err != nil {
		fmt.Println("failed to connect")
		fmt.Println(err)
		return
	}
	fmt.Println("connected!")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("user> ")
		for !scanner.Scan() {
		}
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		args := []interface{}{}
		ls := strings.Split(line, " ")
		for _, arg := range ls {
			log.Print(arg)
			args = append(args, arg)
		}
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("sent " + line)

		out, err := conn.Do(ls[0], args[1:]...)
		fmt.Print(addr + ":" + strconv.Itoa(portNum) + "> ")
		if err != nil {
			fmt.Print("ERR:" + err.Error())
		} else {
			fmt.Print(out)
		}
	}
}
