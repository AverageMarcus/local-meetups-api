package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-stomp/stomp"
)

func getPubSub() (*stomp.Conn, error) {
	conn, err := stomp.Dial(
		"tcp",
		"local-meetups-api-messaging:61613",
		stomp.ConnOpt.Login(os.Getenv("PUBSUB_USER"), os.Getenv("PUBSUB_PASSWORD")),
	)
	return conn, err
}

func publish(meetup Meetup) {
	fmt.Println("Announcing " + meetup.Group.Name + " - " + meetup.Title)
	conn, _ := getPubSub()
	defer conn.Disconnect()

	body, _ := json.Marshal(meetup)
	conn.Send(
		"/topic/meetup-announcement",
		"application/json",
		body,
	)
}
