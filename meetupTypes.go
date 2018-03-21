package main

import "time"

type MeetupEvent struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Created      int         `json:"created"`
	Updated      int         `json:"updated"`
	Description  string      `json:"description"`
	EventURL     string      `json:"event_url"`
	RsvpLimit    int         `json:"rsvp_limit"`
	YesRsvpCount int         `json:"yes_rsvp_count"`
	Time         int         `json:"time"`
	Status       string      `json:"status"`
	Venue        MeetupVenue `json:"venue"`
	Group        MeetupGroup `json:"group"`
	Persisted    *time.Time  `json:",omitempty"`
}

type MeetupVenue struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address_1"`
	City    string `json:"city"`
	Country string `json:"country"`
}

type MeetupGroup struct {
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
	Results []MeetupEvent `json:"results"`
	Meta    MeetupMeta    `json:"meta"`
}
