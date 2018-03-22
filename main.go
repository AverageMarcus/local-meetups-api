package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	couchdb "github.com/rhinoman/couchdb-go"
	"github.com/robfig/cron"
)

func main() {
	checkEnv()
	waitForDB()
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

	if os.Getenv("COUCHDB_HOST") == "" {
		fmt.Println("No CouchDB HOST specified")
		errors = true
	}

	if errors {
		os.Exit(1)
	}
}

func waitForDB() {
	err := retry(10, time.Duration(1*time.Second), func() error {
		timeout := time.Duration(500 * time.Millisecond)
		client, connectErr := couchdb.NewConnection(os.Getenv("COUCHDB_HOST"), 5984, timeout)
		if connectErr != nil {
			return connectErr
		}
		client.SelectDB("local-meetups", nil)
		dbErr := db.DbExists()
		return dbErr
	})

	if err != nil {
		panic(err)
	}
}

func retry(attempts int, sleep time.Duration, fn func() error) (err error) {
	for i := 0; ; i++ {
		err = fn()
		if err == nil {
			return
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(sleep)

		fmt.Println("Retrying...")
	}
	return fmt.Errorf("After %d attempts, last error: %s", attempts, err)
}

func setupCron() {
	go fetchMeetups()
	c := cron.New()
	c.AddFunc("@every 30m", fetchMeetups)
	c.AddFunc("@every 5m", cleanupPastMeetups)
	c.Start()
}

func setupServer() {
	server := gin.Default()
	router(server)
	server.Run(":80")
}
