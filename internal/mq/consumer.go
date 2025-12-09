package mq

import (
	"log"

	"provider_management/internal/service"

	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	url string
	svc *service.ComplaintService
	log *log.Logger
}

func NewConsumer(url string, svc *service.ComplaintService, l *log.Logger) *Consumer {
	return &Consumer{url: url, svc: svc, log: l}
}

func (c *Consumer) StartWorkers(workers int) {
	conn, err := amqp091.Dial(c.url)
	if err != nil {
		log.Fatal("RabbitMQ connect failed:", err)
	}
	ch, _ := conn.Channel()

	msgs, err := ch.Consume("complaint_queue", "", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	for w := 1; w <= workers; w++ {
		go func(id int) {
			for msg := range msgs {
				c.log.Println("Worker", id, "processing:", string(msg.Body))
				_ = c.svc // process with business logic
			}
		}(w)
	}
}
