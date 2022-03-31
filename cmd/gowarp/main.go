package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gowarp/pkg/app"
	"log"
	"net/http"
	"strings"
)

func warp(c *gin.Context) {
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		http.Error(c.Writer, "Server does not support Flusher!",
			http.StatusInternalServerError)
		return
	}

	ua := c.Request.UserAgent()

	if strings.Contains(ua, "Firefox/") {
		c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	} else {
		c.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	}
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	if err := app.Generate(c.Writer, flusher); err != nil {
		_, _ = fmt.Fprintln(c.Writer, err)
		return
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		warp(c)
	})
	log.Fatal(r.Run())
}
