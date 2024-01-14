package server

import (
	"net/http"
	"time"

	videoUploader "absolute.video/internal/controller"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
		  return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	  }))
	r.GET("/health", s.healthHandler)
	r.POST("/upload-video", videoUploader.VideoHandler )
	r.GET("/test", s.testRoute)
	return r
}


func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) testRoute(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "hello world"})
}