package broker

import (
	"encoding/json"
	"fmt"
	"log"
	"statistics/src/database"
	"time"

	"github.com/IBM/sarama"
)

type Broker struct {
	Consumer sarama.PartitionConsumer
}

func New() (*Broker, func()) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// TODO: use env
	brokers := []string{"kafka:29092"}
	var master sarama.Consumer
	var err error
	for {
		master, err = sarama.NewConsumer(brokers, config)
		if err != nil {
			fmt.Println("Wait for Kafka")
			time.Sleep(5 * time.Second) // TODO: decrease?
			continue
		}
		fmt.Println("Kafka is ready")
		break
	}
	close := func() {
		master.Close()
	}

	consumer, err := master.ConsumePartition("Stat", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalln("error creating broker", err)
	}
	// TODO: close

	return &Broker{
		Consumer: consumer,
	}, close
}

func (b *Broker) Consume(db *database.DataBase) {
	for msg := range b.Consumer.Messages() {
		var stat Statistic
		json.Unmarshal(msg.Value, &stat)
		key := string(msg.Key)
		fmt.Printf("Got: %s %#v\n", string(msg.Key), stat)

		if key == "Like" {
			db.EnsureLike(database.Statistic{
				Login:  stat.Login,
				TaskID: stat.TaskID,
			})
		} else if key == "View" {
			db.EnsureView(database.Statistic{
				Login:  stat.Login,
				TaskID: stat.TaskID,
			})
		}
	}
}

type Statistic struct {
	Login  string `json:"login"`
	TaskID uint   `json:"task_id"`
}
