package rest

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GetClientIP(c *gin.Context) (clientIP string) {
	clientIP = c.Request.Header.Get("X-Forwarded-For")

	if len(clientIP) == 0 {
		clientIP = c.Request.Header.Get("X-Real-IP")
	}

	if len(clientIP) == 0 {
		clientIP = c.Request.RemoteAddr
	}

	if strings.Contains(clientIP, ",") {
		clientIP = strings.Split(clientIP, ",")[0]
	}

	return
}

func GetDurationInMillseconds(start time.Time) float64 {
	end := time.Now()
	duration := end.Sub(start)
	milliseconds := float64(duration) / float64(time.Millisecond)
	rounded := float64(int(milliseconds*100+.5)) / 100
	return rounded
}
