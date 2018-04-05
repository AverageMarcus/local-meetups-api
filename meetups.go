package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rhinoman/couchdb-go"
	resty "gopkg.in/resty.v1"
)

var timeout = time.Duration(500 * time.Millisecond)
var port, _ = strconv.Atoi(os.Getenv("COUCHDB_PORT"))
var client, _ = couchdb.NewConnection(os.Getenv("COUCHDB_HOST"), port, timeout)
var auth = couchdb.BasicAuth{Username: os.Getenv("COUCHDB_USER"), Password: os.Getenv("COUCHDB_PASSWORD")}
var db = client.SelectDB("local-meetups", &auth)

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
		// TODO: Convert meetup to a more friendly type

		// 1. Check for existing meetup saved
		savedMeetup, err := getSavedMeetup(meetup.ID)
		if err != nil {
			fmt.Println(err)
		}
		// TODO: Check if meetup has updated since being saved
		if savedMeetup.Persisted != nil {
			continue
		}

		// 2. If not found, save
		now := time.Now()
		meetup.Persisted = &now
		saveMeetup(meetup)

		// TODO: 3. Post to message queue

	}
}

func getSavedMeetup(id string) (MeetupEvent, error) {
	var meetup = MeetupEvent{}
	_, err := db.Read(id, &meetup, nil)
	if err != nil {
		return MeetupEvent{}, err
	}
	return meetup, nil
}

func saveMeetup(meetup MeetupEvent) {
	_, err := db.Save(meetup, meetup.ID, "")
	if err != nil {
		panic(err)
	}
}

type FindResponse struct {
	Docs []MeetupEvent `json:"docs"`
}

func getNextMeetupForGroup(group string) (MeetupEvent, error) {
	meetups, err := getMeetupsForGroup(group)
	meetup := MeetupEvent{}
	if len(meetups) > 0 {
		meetup = meetups[0]
	}
	return meetup, err
}

func getMeetupsForGroup(group string) ([]MeetupEvent, error) {
	meetups := FindResponse{}

	params := couchdb.FindQueryParams{
		Selector: map[string]interface{}{
			"$and": [2]interface{}{
				map[string]interface{}{"group.name": map[string]interface{}{
					"$regex": "(?i)" + group,
				}},
				map[string]interface{}{"time": map[string]interface{}{
					"$gt": time.Now().Unix() * 1000,
				}},
			},
		},
		Sort: [1]interface{}{map[string]interface{}{"time": "desc"}},
	}

	err := db.Find(&meetups, &params)
	if err != nil {
		panic(err)
	}

	var notFoundErr error
	if len(meetups.Docs) <= 0 {
		notFoundErr = errors.New("No upcoming meetup found")
	}

	return meetups.Docs, notFoundErr
}

func getNextMeetup() (MeetupEvent, error) {
	meetup := MeetupEvent{}
	meetups, err := getAllMeetups()
	if err == nil && len(meetups) > 0 {
		meetup = meetups[0]
	}

	return meetup, err
}

func getAllMeetups() ([]MeetupEvent, error) {
	response := FindResponse{}

	params := couchdb.FindQueryParams{
		Selector: map[string]interface{}{
			"time": map[string]interface{}{
				"$gt": time.Now().Unix() * 1000,
			},
		},
		Sort: [1]interface{}{map[string]interface{}{"time": "desc"}},
	}

	err := db.Find(&response, &params)
	if err != nil {
		panic(err)
	}

	var notFoundError error
	if len(response.Docs) <= 0 {
		notFoundError = errors.New("No upcoming meetup found")
	}

	return response.Docs, notFoundError
}

func cleanupPastMeetups() {
	meetups := FindResponse{}

	params := couchdb.FindQueryParams{
		Selector: map[string]interface{}{
			"time": map[string]interface{}{
				"$lt": time.Now().Unix() * 1000,
			},
		},
	}

	err := db.Find(&meetups, &params)
	if err != nil {
		fmt.Println("Failed to cleanup past meetups")
		return
	}

	if len(meetups.Docs) > 0 {
		for _, meetup := range meetups.Docs {
			_, err := db.Delete(meetup.ID, "")
			if err != nil {
				fmt.Println("Failed to remove past meetup: " + meetup.ID)
			}
		}
	}
}
