package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func router(server *gin.Engine) {
	server.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"title": "Oxford Technology Meetups API",
			"about": "An API providing the details of upcoming meetups local to Oxford, UK",
			"routes": gin.H{
				"/":                     "This page",
				"/ping":                 "Health check",
				"/upcoming":             "Lists the next meetup of all local groups",
				"/upcoming/{groupName}": "List all upcoming meetups for the given group name",
				"/upcoming/all":         "List all upcoming meetups for all groups",
				"/next":                 "Fetches the details for the next scheduled meetup",
				"/next/{groupName}":     "Fetches the details for the next scheduled meetup for the given group name",
			},
		})
	})

	server.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Writer.Header().Set("Pragma", "no-cache")
		c.Writer.Header().Set("Expires", "0")
		c.Writer.Header().Set("access-control-allow-origin", "*")
		c.Writer.Header().Set("access-control-allow-headers", "Origin, X-Requested-With, Content-Type, Accept, authorization")

		c.Next()
	})

	server.GET("/ping", func(c *gin.Context) {
		db, err := getDB()
		if err != nil {
			c.JSON(500, "Error connecting to database")
		}
		if err = db.Ping(); err != nil {
			c.JSON(500, "Error pinging to database")
		}
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	server.GET("/upcoming", func(c *gin.Context) {
		meetups, err := getAllNextMeetups()
		if err != nil {
			fmt.Println(err)
			c.JSON(404, "No upcoming meetups found")
			return
		}
		c.JSON(200, meetups)
	})

	server.GET("/upcoming/:group", func(c *gin.Context) {
		group := c.Param("group")
		meetups := []Meetup{}
		var err error
		if group == "all" {
			meetups, err = getAllMeetups()
		} else {
			meetups, err = getAllMeetupsForGroup(group)
		}
		if err != nil {
			c.JSON(404, "No upcoming meetups found")
			return
		}
		c.JSON(200, meetups)
	})

	server.GET("/next", func(c *gin.Context) {
		meetup, err := getNextMeetup()
		if err != nil {
			c.JSON(404, "No upcoming meetup found")
			return
		}
		c.JSON(200, meetup)
	})

	server.GET("/next/:group", func(c *gin.Context) {
		group := c.Param("group")
		meetup, err := getNextMeetupForGroup(group)
		if err != nil {
			c.JSON(404, "No upcoming meetup found")
			return
		}
		c.JSON(200, meetup)
	})
}
