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

	server.GET("/next/:group", func(c *gin.Context) {
		group := c.Param("group")
		meetup, err := getNextMeetupForGroup(group)
		if err != nil {
			panic(err)
		}
		fmt.Println("Got meetup: " + meetup.Name)
		c.JSON(200, meetup)
	})

	server.GET("/next", func(c *gin.Context) {
		meetup, err := getNextMeetup()
		if err != nil {
			panic(err)
		}
		fmt.Println("Got meetup: " + meetup.Name)
		c.JSON(200, meetup)
	})
}
