package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func router(server *gin.Engine) {
	server.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	server.GET("/", func(c *gin.Context) {
		meetups, err := getAllMeetups()
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
		fmt.Println("Got meetup: " + meetup.Name)
		c.JSON(200, meetup)
	})

	server.GET("/next/:group", func(c *gin.Context) {
		group := c.Param("group")
		meetup, err := getNextMeetupForGroup(group)
		if err != nil {
			c.JSON(404, "No upcoming meetup found")
			return
		}
		fmt.Println("Got meetup: " + meetup.Name)
		c.JSON(200, meetup)
	})
}
