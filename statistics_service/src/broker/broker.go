package broker

import (
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
)

type Broker struct {
	Consumer sarama.PartitionConsumer
}

// func consume(consumer sarama.Consumer, topic string) (chan *sarama.ConsumerMessage, error) {
// 	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// TODO: close

// 	for {
// 		select {
// 		case msg := <-partitionConsumer.Messages():
// 			fmt.Printf("Message received: key = %s, value = %s, topic = %s, partition = %d, offset = %d\n",
// 				string(msg.Key), string(msg.Value), msg.Topic, msg.Partition, msg.Offset)
// 		case err := <-partitionConsumer.Errors():
// 			log.Printf("Ошибка во время потребления: %v\n", err)
// 		}
// 	}

// 	return nil, nil
// }

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
			time.Sleep(5 * time.Second)
			continue
		}
		fmt.Println("Kafka is ready")
		break
	}
	close := func() {
		master.Close()
	}

	// partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	// if err != nil {
	// 	return nil, err
	// }
	consumer, err := master.ConsumePartition("Stat", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalln("error creating broker", err)
	}
	// TODO: close

	// viewsConsumer, err := master.ConsumePartition("View", 0, sarama.OffsetNewest)
	// if err != nil {
	// 	log.Fatalln("error creating broker", err)
	// }

	return &Broker{
		Consumer: consumer,
		//ViewsConsumer: viewsConsumer,
	}, close
}

type Statistic struct {
	Login  string `json:"login"`
	TaskID uint   `json:"task_id"`
}

// func (b *Broker) getStat(chan *sarama.ConsumerMessage) Statistic {
// 	msg := <-b.LikesConsumer.Messages()

// 	var stat Statistic
// 	json.Unmarshal(msg.Value, &stat)
// 	return stat
// }

// func (b *Broker) GetLikeStat() Statistic {
// 	return b.getStat(b.LikesConsumer)
// }

// func (b *Broker) GetViewStat() Statistic {
// 	return b.getStat(b.ViewsConsumer)
// }
