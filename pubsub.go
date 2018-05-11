package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/streadway/amqp"
)

func getPubSub() (*amqp.Connection, error) {
	return amqp.Dial("amqp://" + os.Getenv("PUBSUB_USER") + ":" + os.Getenv("PUBSUB_PASSWORD") + "@local-meetups-api-mssaging:5672/")
}

func publish(meetup Meetup) {
	conn, _ := getPubSub()
	defer conn.Close()
	ch, _ := conn.Channel()
	defer ch.Close()

	ch.ExchangeDeclare(
		"meetup-announcement",
		"topic",
		false,
		false,
		false,
		false,
		nil,
	)

	body, _ := json.Marshal(meetup)

	fmt.Println("Announcing " + meetup.Group.Name + " - " + meetup.Title)

	ch.Publish(
		"meetup-announcement",
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
}
