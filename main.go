package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
)

func main() {
	// 从环境变量或配置文件获取，这里为了简单硬编码
	baseURL := os.Getenv("BASE_URL")

	if baseURL == "" {
		baseURL = "http://127.0.0.1:8080/" //确保末尾有斜杠
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	//storage := NewInMemoryStorage()

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}

	storage, err := NewRedisStorage(redisAddr, "", 0)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	handlers := &AppHandlers{
		Store:   storage,
		BaseURL: baseURL,
	}

	router := gin.Default()

	//API 端点用于创建短链接
	router.POST("/shorten", handlers.HandleShortenURL)

	//重定向端点
	// :shortCode 是路径参数
	router.GET("/:shortCode", handlers.HandleRedirect)

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
		})
	})

	log.Printf("Starting server on port %s with base URL %s", port, baseURL)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}

}
