package main

import (
	"flag"
	"fmt"

	"statistics/src/broker"
	"statistics/src/database"
	"statistics/src/server"
)

func main() {
	port := flag.Int("port", 8082, "Port of statistics service server.")
	flag.Parse()

	db := database.New()

	b, close := broker.New()
	defer close()
	go b.Consume(db)

	server := server.New(db)
	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	server.RegisterAndListen(addr)
}
