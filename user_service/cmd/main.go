package main

import (
	"flag"
	"fmt"
	"userservice/src/database"
	"userservice/src/server"
)

func main() {
	port := flag.Int("port", 8080, "Port of user service's server.")
	flag.Parse()

	db := database.New()
	server, close := server.New(db)
	defer close()
	server.Register()

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	server.Listen(addr)
}
