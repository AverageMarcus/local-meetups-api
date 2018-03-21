package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fjl/go-couchdb"
	resty "gopkg.in/resty.v1"
)

var client, _ = couchdb.NewClient("http://db:5984", nil)
var db = client.DB("local-meetups")

func buildURL() string {
	groups := strings.Replace(os.Getenv("GROUP_IDS"), ",", "%2C", -1)
	apiKey := strings.Replace(os.Getenv("MEETUP_KEY"), ",", "%2C", -1)
	return fmt.Sprintf("https://api.meetup.com/2/events?offset=0&format=json&limited_events=False&group_id=%s&photo-host=secure&fields=&order=time&status=upcoming&desc=false&key=%s", groups, apiKey)
}

func fetchMeetups() {
	resp, err := resty.R().Get(buildURL())
	if err != nil {
		fmt.Println("Failed to fetch latest meetups: " + err.Error())
	}
	results := MeetupResponse{}
	json.Unmarshal([]byte(resp.String()), &results)

	for _, meetup := range results.Results {
		// 1. Check for existing meetup saved
		savedMeetup, err := getSavedMeetup(meetup.ID)
		if err != nil {
			fmt.Println(err)
		}
		if savedMeetup.Persisted != nil {
			continue
		}

		// 2. Check for the next upcoming from the same group
		if hasUpcoming(meetup.Group.Name) {
			fmt.Println("We already have an upcoming meetup for '" + meetup.Group.Name + "' so skipping this for now")
			continue
		}

		// 3. If none found, save
		now := time.Now()
		meetup.Persisted = &now
		saveMeetup(meetup)

		// TODO: 4. Post to message queue

	}
}

func getSavedMeetup(id string) (MeetupEvent, error) {
	var meetup = MeetupEvent{}
	err := db.Get(id, &meetup, nil)
	if err != nil {
		return MeetupEvent{}, err
	}
	return meetup, nil
}

func saveMeetup(meetup MeetupEvent) {
	_, err := db.Put(meetup.ID, meetup, "")
	if err != nil {
		panic(err)
	}
}

func hasUpcoming(group string) bool {
	var meetups = []MeetupEvent{}
	err := db.AllDocs(&meetups, nil)
	if err != nil {
		panic(err)
	}

	for _, meetup := range meetups {
		if meetup.Group.Name == group || meetup.Group.UrlName == group {
			if time.Unix(int64(meetup.Time/1000), 0).After(time.Now()) {
				return true
			}
		}
	}

	return false
}
