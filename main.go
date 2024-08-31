package main

import (
	"net/http"

	"github.com/lixiaoping0821/workflow/node"

	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.POST("/SendEmailAttach", func(c *gin.Context) {
		// c.JSON(http.StatusOK, gin.H{
		// 	"message": "pong",
		// })
		var smtp node.SmtpNode
		smtp.SendEmailAttach(c)

	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

}
