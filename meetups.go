package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rhinoman/couchdb-go"
	resty "gopkg.in/resty.v1"
)

var timeout = time.Duration(500 * time.Millisecond)
var client, _ = couchdb.NewConnection(os.Getenv("COUCHDB_HOST"), 5984, timeout)
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

func hasUpcoming(group string) bool {
	_, err := getNextMeetupForGroup(group)
	if err != nil {
		return false
	}

	return true
}

type FindResponse struct {
	Docs []MeetupEvent `json:"docs"`
}

func getNextMeetupForGroup(group string) (*MeetupEvent, error) {
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

	if len(meetups.Docs) > 0 {
		return &meetups.Docs[0], nil
	}

	return nil, errors.New("No upcoming meetup found")
}

func getNextMeetup() (*MeetupEvent, error) {
	meetups := FindResponse{}

	params := couchdb.FindQueryParams{
		Selector: map[string]interface{}{
			"time": map[string]interface{}{
				"$gt": time.Now().Unix() * 1000,
			},
		},
		Sort: [1]interface{}{map[string]interface{}{"time": "desc"}},
	}

	err := db.Find(&meetups, &params)
	if err != nil {
		panic(err)
	}

	if len(meetups.Docs) > 0 {
		return &meetups.Docs[0], nil
	}

	return nil, errors.New("No upcoming meetup found")
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
