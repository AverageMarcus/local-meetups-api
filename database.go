package main

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var JavascriptISOString = "2006-01-02T15:04:05.999Z07:00"
var instance *sql.DB
var once sync.Once

type MeetupRow struct {
	ID           string
	Title        string
	Created      string
	Updated      string
	Persisted    string
	Description  string
	URL          string
	RsvpCount    int
	RsvpLimit    int
	Time         string
	Status       string
	GroupName    string
	GroupUrlName string
	VenueName    string
	VenueAddress string
	VenueCity    string
	VenueCountry string
}

func getDB() (*sql.DB, error) {
	var err error
	once.Do(func() {
		instance, err = sql.Open("mysql", os.Getenv("MYSQL_USER")+":"+os.Getenv("MYSQL_PASSWORD")+"@tcp("+os.Getenv("MYSQL_HOST")+":3306)/"+os.Getenv("MYSQL_DATABASE")+"?charset=utf8mb4")
	})
	return instance, err
}

func hydrateRows(rows *sql.Rows) ([]Meetup, error) {
	var meetups []Meetup

	if !rows.Next() {
		return nil, fmt.Errorf("No meetups returned")
	}

	for rows.Next() {
		var meetup Meetup
		var created, updated, persisted, meetupTime string
		if err := rows.Scan(
			&meetup.ID, &meetup.Title, &created, &updated, &persisted,
			&meetup.Description, &meetup.URL, &meetup.RsvpCount, &meetup.RsvpLimit, &meetupTime,
			&meetup.Status, &meetup.Group.Name, &meetup.Group.UrlName, &meetup.Venue.Name,
			&meetup.Venue.Address, &meetup.Venue.City, &meetup.Venue.Country,
		); err != nil {
			return nil, err
		}
		meetup.Created, _ = time.Parse(JavascriptISOString, created)
		meetup.Updated, _ = time.Parse(JavascriptISOString, updated)
		meetup.Persisted, _ = time.Parse(JavascriptISOString, persisted)
		meetup.Time, _ = time.Parse(JavascriptISOString, meetupTime)
		meetups = append(meetups, meetup)
	}

	return meetups, nil
}
