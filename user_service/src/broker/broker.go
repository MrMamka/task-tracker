package broker

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

type Broker struct {
	producer sarama.SyncProducer
}

func New() (*Broker, func()) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // ?
	config.Producer.Return.Successes = true

	// TODO: use env
	brokers := []string{"kafka:29092"}
	var producer sarama.SyncProducer
	var err error
	for {
		producer, err = sarama.NewSyncProducer(brokers, config)
		if err != nil {
			fmt.Println("Wait for Kafka")
			time.Sleep(5 * time.Second)
			continue
		}
		fmt.Println("Kafka is ready")
		break
	}

	close := func() {
		producer.Close()
	}

	return &Broker{producer: producer}, close
}

type Statistic struct {
	Login  string `json:"login"`
	TaskID uint   `json:"task_id"`
}

func (b *Broker) sendStat(stat Statistic, key string) error {
	messageBytes, err := json.Marshal(stat)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: "Stat",
		Key: sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(messageBytes),
	}
	_, _, err = b.producer.SendMessage(msg)
	return err
}

func (b *Broker) SendLike(stat Statistic) error {
	return b.sendStat(stat, "Like")
}

func (b *Broker) SendView(stat Statistic) error {
	return b.sendStat(stat, "View")
}
