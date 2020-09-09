package main

import (
	"net"

	"github.com/Mojashi/dsdb/database"
	"github.com/Mojashi/dsdb/datastructures"
)

func main() {

	listener, err := net.Listen("tcp", "localhost:5003")
	if err != nil {
		panic(err)
	}

	db := database.MakeDB()
	db.Register(datastructures.Trie{})
	db.Register(datastructures.SegmentTree{})
	db.Run(listener)
}
