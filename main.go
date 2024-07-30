package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

func APIKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("Authorization")
		if len(apiKey) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authorization header is required"})
			return
		}

		if apiKey != Env.ApiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid API key"})
			return
		}

		c.Next()
	}
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	api := r.Group("/api", APIKeyAuthMiddleware())
	api.POST("/drts", validateLP)

	return r
}

func main() {
	r := setupRouter()
	exit := make(chan os.Signal, 2)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		if err := r.Run(":" + Env.Port); err != nil {
			Log.Println(err)
			exit <- syscall.Signal(0)
		}
	}()

	if !gin.IsDebugging() {
		Log.Printf("Listening and serving HTTP on %s\n", Env.Port)
	}
	<-exit
	Log.Println("Shutting down server...")
}

type ValidateLP struct {
	RefNo string `json:"ref_no" binding:"required"`
}

func validateLP(c *gin.Context) {
	var reqBody ValidateLP

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		Log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if len(reqBody.RefNo) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ref_no is required"})
		return
	}

	reqBuf := bytes.Buffer{}
	if err := json.NewEncoder(&reqBuf).Encode(reqBody); err != nil {
		Log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	resp, err := http.Post(Env.DRTS_API, "application/json", &reqBuf)
	if err != nil {
		Log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unknown"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		Log.Println("DRTS api failed:", resp.Status)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unknown"})
		return
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		Log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	data := make(map[string]any)
	json.NewDecoder(bytes.NewBuffer(buf)).Decode(&data)
	c.JSON(http.StatusOK, data)
}
