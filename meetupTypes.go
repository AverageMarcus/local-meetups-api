package main

import "time"

type Meetup struct {
	ID          string
	Title       string
	Created     time.Time
	Updated     time.Time
	Persisted   time.Time
	Description string
	URL         string
	RsvpCount   int
	RsvpLimit   int
	Time        time.Time
	Status      string
	Group       MeetupGroup
	Venue       MeetupVenue
}

type MeetupGroup struct {
	Name    string
	UrlName string
}

type MeetupVenue struct {
	Name    string
	Address string
	City    string
	Country string
}

type RawMeetup struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Created      int64          `json:"created"`
	Updated      int64          `json:"updated"`
	Description  string         `json:"description"`
	EventURL     string         `json:"event_url"`
	RsvpLimit    int            `json:"rsvp_limit"`
	YesRsvpCount int            `json:"yes_rsvp_count"`
	Time         int64          `json:"time"`
	Status       string         `json:"status"`
	Venue        RawMeetupVenue `json:"venue"`
	Group        RawMeetupGroup `json:"group"`
	Persisted    *time.Time     `json:",omitempty"`
}

type RawMeetupVenue struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address_1"`
	City    string `json:"city"`
	Country string `json:"country"`
}

type RawMeetupGroup struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	UrlName string `json:"urlname"`
}

type MeetupMeta struct {
	Next        string `json:"next"`
	Method      string `json:"method"`
	TotalCount  int    `json:"total_count"`
	Link        string `json:"link"`
	Count       int    `json:"count"`
	Description string `json:"description"`
	Lon         string `json:"lon"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	ID          string `json:"id"`
	Updated     int    `json:"updated"`
	Lat         string `json:"lat"`
}

type MeetupResponse struct {
	Results []RawMeetup `json:"results"`
	Meta    MeetupMeta  `json:"meta"`
}
