package main

import (
	"log"
	"os"
	"stock-monitor/internal/together/finnhub"
	"stock-monitor/internal/together/handlers"
	"stock-monitor/internal/together/scheduler"
	"stock-monitor/internal/together/store"

	"github.com/gin-gonic/gin"
)

func main() {
	// r := gin.Default()

	// r.GET("/ping", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{"message": "hello world"})
	// })

	// r.Run(":9000")
	apikey := os.Getenv("FINNHUB_API_KEY")

	if apikey == "" {
		log.Fatal("Finnhub api key not set")
	}

	//create new objects
	dataStore := store.New()
	finnhubClient := finnhub.New(apikey)
	shr := scheduler.New(dataStore, finnhubClient)
	h := handlers.New(dataStore, shr, finnhubClient)

	//create router
	r := gin.Default()

	//register routes
	r.POST("/start-monitoring", h.StartMonitoring)
	r.GET("/history", h.GetHistory)
	r.POST("/refresh", h.Refresh)
	r.DELETE("/history", h.Remove)

	//set port
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	//start http server
	log.Printf("Server listening on :%s", port)
	err := r.Run(":" + port)

	if err != nil {
		log.Fatal(err)
	}

}
