package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// AppHandlers 包含应用的处理器和依赖
type AppHandler struct {
	Store   Storage
	BaseURL string //"http://localhost:8080/"
}

type ShortenRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

// HandleShortenURL 处理创建短链接的请求
func (h *AppHandler) HandleShortenURL(c *gin.Context) {
	var req ShortenRequest
	if err := c.ShouldBindJ

}
