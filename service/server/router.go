package main

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine) {
	r.GET("/status", getStatus)
}

func getStatus(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func getStoredScripts(c *gin.Context) {

}

func getRunningTaskList(c *gin.Context) {

}
