package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"statistics/src/broker"
	"statistics/src/database"
)

func main() {
	_ = flag.Int("port", 8082, "Port of statistics service server.")
	flag.Parse()

	b, close := broker.New()
	defer close()

	db := database.New()

	for msg := range b.Consumer.Messages() {
		var stat broker.Statistic
		json.Unmarshal(msg.Value, &stat)
		key := string(msg.Key)
		fmt.Printf("Got: %s %#v\n", string(msg.Key), stat)

		if key == "Like" {
			db.AddLike(database.Statistic{
				Login:  stat.Login,
				TaskID: stat.TaskID,
			})
		} else if key == "View" {
			db.AddView(database.Statistic{
				Login:  stat.Login,
				TaskID: stat.TaskID,
			})
		}
	}
}
