package main

import (
	"flag"
	"fmt"

	"tasksmanager/src/server"
	"tasksmanager/src/database"
)

func main() {
	port := flag.Int("port", 8081, "Port of tasks manager server.")
	flag.Parse()

	db := database.New()
	server := server.New(db)

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	server.RegisterAndListen(addr)
}
