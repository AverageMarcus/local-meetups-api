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
var JavascriptISOString = "2006-01-02T15:04:05.999Z07:00"

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

	now := time.Now()

	for _, meetupEvent := range results.Results {
		// 1: Convert meetup to a more friendly type
		meetup := Meetup{
			ID:          meetupEvent.ID,
			Title:       meetupEvent.Name,
			Created:     time.Unix(meetupEvent.Created/1000, 0),
			Updated:     time.Unix(meetupEvent.Updated/1000, 0),
			Persisted:   now,
			Description: meetupEvent.Description,
			URL:         meetupEvent.EventURL,
			RsvpCount:   meetupEvent.YesRsvpCount,
			RsvpLimit:   meetupEvent.RsvpLimit,
			Time:        time.Unix(meetupEvent.Time/1000, 0),
			Status:      meetupEvent.Status,
			Group: MeetupGroup{
				Name:    meetupEvent.Group.Name,
				UrlName: meetupEvent.Group.UrlName,
			},
			Venue: MeetupVenue{
				Name:    meetupEvent.Venue.Name,
				Address: meetupEvent.Venue.Address,
				City:    meetupEvent.Venue.City,
				Country: meetupEvent.Venue.Country,
			},
		}

		// 2. save
		saveMeetup(meetup)

		// TODO: 3. Post to message queue

	}
}

func saveMeetup(meetup Meetup) {
	revision := getSavedMeetupRevision(meetup.ID)
	fmt.Println("Updating meetup " + meetup.ID + " - Revision: " + revision)
	_, err := db.Save(meetup, meetup.ID, revision)
	if err != nil {
		panic(err)
	}
}

func getSavedMeetupRevision(id string) string {
	revision, _ := db.Read(id, &Meetup{}, nil)
	return revision
}

type FindResponse struct {
	Docs []Meetup `json:"docs"`
}

func getNextMeetupForGroup(group string) (Meetup, error) {
	meetups, err := getMeetupsForGroup(group)
	meetup := Meetup{}
	if len(meetups) > 0 {
		meetup = meetups[0]
	}
	return meetup, err
}

func getMeetupsForGroup(group string) ([]Meetup, error) {
	meetups := FindResponse{}

	params := couchdb.FindQueryParams{
		Limit: 100,
		Selector: map[string]interface{}{
			"$and": [2]interface{}{
				map[string]interface{}{"Group.Name": map[string]interface{}{
					"$regex": "(?i)" + group,
				}},
				map[string]interface{}{"Time": map[string]interface{}{
					"$gt": time.Now().UTC().Format(JavascriptISOString),
				}},
			},
		},
		Sort: [1]interface{}{map[string]interface{}{"Time": "desc"}},
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

func getNextMeetup() (Meetup, error) {
	meetup := Meetup{}
	meetups, err := getAllMeetups()
	if err == nil && len(meetups) > 0 {
		meetup = meetups[0]
	}

	return meetup, err
}

func getAllMeetups() ([]Meetup, error) {
	response := FindResponse{}

	params := couchdb.FindQueryParams{
		Limit: 1000,
		Selector: map[string]interface{}{
			"Time": map[string]interface{}{
				"$gt": time.Now().UTC().Format(JavascriptISOString),
			},
		},
		Sort: [1]interface{}{map[string]interface{}{"Time": "desc"}},
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

func getAllNextMeetups() ([]Meetup, error) {
	allMeetups, err := getAllMeetups()
	if err != nil {
		return nil, err
	}

	var meetups = []Meetup{}

	for _, meetup := range allMeetups {
		var found = false
		for _, existingMeetup := range meetups {
			if existingMeetup.Group.Name == meetup.Group.Name {
				found = true
			}
		}
		if found == false {
			meetups = append(meetups, meetup)
		}
	}

	return meetups, nil
}

func cleanupPastMeetups() {
	meetups := FindResponse{}

	params := couchdb.FindQueryParams{
		Selector: map[string]interface{}{
			"Time": map[string]interface{}{
				"$lt": time.Now().UTC().Format(JavascriptISOString),
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
