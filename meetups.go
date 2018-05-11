package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	resty "gopkg.in/resty.v1"
)

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
			Announced: mysql.NullTime{},
		}

		// 2. get existing meetup
		existingMeetup, err := getMeetup(meetup.ID)
		isNew := err != nil || &existingMeetup.Persisted == nil
		isUpdate := !isNew && !existingMeetup.Updated.Equal(meetup.Updated)
		_, noUpcoming := getNextMeetupForGroup(meetup.Group.Name)

		// 3. Post to message queue

		if !meetup.Announced.Valid && noUpcoming != nil {
			meetup.Announced = mysql.NullTime{
				Time:  now,
				Valid: true,
			}
			go publish(meetup)
		}

		// 4. save
		if isNew || isUpdate {
			go saveMeetup(meetup)
		}
	}
}

func saveMeetup(meetup Meetup) {
	db, err := getDB()
	if err != nil {
		panic(err.Error())
	}
	stmtIns, err := db.Prepare("INSERT INTO meetup VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE ID = values(ID), Title = values(Title), Created = values(Created), Updated = values(Updated), Persisted = values(Persisted), Description = values(Description), URL = values(URL), RsvpCount = values(RsvpCount), RsvpLimit = values(RsvpLimit), Time = values(Time), Status = values(Status), GroupName = values(GroupName), GroupUrlName = values(GroupUrlName), VenueName = values(VenueName), VenueAddress = values(VenueAddress), VenueCity = values(VenueCity), VenueCountry = values(VenueCountry), Announced = values(Announced)")
	if err != nil {
		panic(err.Error())
	}
	defer stmtIns.Close()

	fmt.Println("Saving " + meetup.Group.Name + " - " + meetup.Title + " - " + meetup.Time.Format(JavascriptISOString) + " (Last updated: " + meetup.Updated.Format(JavascriptISOString) + ")")

	announcedTime := ""
	if meetup.Announced.Valid {
		announcedTime = meetup.Announced.Time.Format(JavascriptISOString)
	}

	_, err = stmtIns.Exec(
		meetup.ID, meetup.Title, meetup.Created.Format(JavascriptISOString), meetup.Updated.Format(JavascriptISOString), meetup.Persisted.Format(JavascriptISOString),
		meetup.Description, meetup.URL, meetup.RsvpCount, meetup.RsvpLimit, meetup.Time.Format(JavascriptISOString),
		meetup.Status, meetup.Group.Name, meetup.Group.UrlName, meetup.Venue.Name,
		meetup.Venue.Address, meetup.Venue.City, meetup.Venue.Country, announcedTime,
	)
	if err != nil {
		panic(err.Error())
	}
}

func getMeetup(id string) (Meetup, error) {
	db, err := getDB()
	if err != nil {
		panic(err.Error())
	}
	rows, err := db.Query("SELECT * FROM meetup WHERE ID = ?", id)
	if err != nil {
		return Meetup{}, err
	}

	meetups, err := hydrateRows(rows)

	if len(meetups) > 0 {
		return meetups[0], err
	}
	return Meetup{}, fmt.Errorf("Meetup not found")
}

func getAllMeetups() ([]Meetup, error) {
	db, err := getDB()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(JavascriptISOString)
	rows, err := db.Query("SELECT * FROM meetup WHERE Time > ? ORDER BY Time asc", now)
	if err != nil {
		return nil, err
	}

	return hydrateRows(rows)
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

func getAllMeetupsForGroup(group string) ([]Meetup, error) {
	db, err := getDB()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(JavascriptISOString)
	rows, err := db.Query("SELECT * FROM meetup WHERE Time > ? AND GroupName = ? ORDER BY Time asc", now, group)
	if err != nil {
		return nil, err
	}

	return hydrateRows(rows)
}

func getNextMeetup() (Meetup, error) {
	meetups, err := getAllMeetups()
	if err != nil {
		return Meetup{}, err
	}

	return meetups[0], nil
}

func getNextMeetupForGroup(group string) (Meetup, error) {
	meetups, err := getAllMeetupsForGroup(group)
	if err != nil {
		return Meetup{}, err
	}

	return meetups[0], nil
}
