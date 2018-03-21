package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
)

func main() {
	checkEnv()
	setupCron()
	setupServer()
}

func checkEnv() {
	errors := false
	if os.Getenv("GROUP_IDS") == "" {
		fmt.Println("No group IDs specified")
		errors = true
	}

	if os.Getenv("MEETUP_KEY") == "" {
		fmt.Println("No Meetup API Key specified")
		errors = true
	}

	if errors {
		os.Exit(1)
	}
}

func setupCron() {
	go fetchMeetups()
	c := cron.New()
	c.AddFunc("@every 30m", fetchMeetups)
	c.Start()
}

func setupServer() {
	server := gin.Default()
	router(server)
	server.Run(":8000")
}
