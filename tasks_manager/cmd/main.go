package main

import (
	"flag"
	"fmt"

	"tasksmanager/src/server"
	"tasksmanager/src/database"
)

func main() {
	port := flag.Int("port", 8080, "Port of tasks manager server.")
	flag.Parse()

	db := database.New()
	server := server.New(db)
	server.Register()

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	server.Listen(addr)
}
