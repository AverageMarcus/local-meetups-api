package main

import (
	"github.com/gin-gonic/gin"
)

func router(server *gin.Engine) {
	server.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	server.GET("/upcoming", func(c *gin.Context) {
		meetups, err := getAllNextMeetups()
		if err != nil {
			c.JSON(404, "No upcoming meetups found")
			return
		}
		c.JSON(200, meetups)
	})

	server.GET("/upcoming/:group", func(c *gin.Context) {
		group := c.Param("group")
		meetups := []MeetupEvent{}
		var err error
		if group == "all" {
			meetups, err = getAllMeetups()
		} else {
			meetups, err = getMeetupsForGroup(group)
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
