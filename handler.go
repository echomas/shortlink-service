package main

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// AppHandlers 包含应用的处理器和依赖
type AppHandlers struct {
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
func (h *AppHandlers) HandleShortenURL(c *gin.Context) {
	var req ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	//简单的 URL 格式校验
	if _, err := url.ParseRequestURI(req.URL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid URL format.",
		})
		return
	}

	//检查长链接是否已经有对应的短链接
	if memStore, ok := h.Store.(*RedisStorage); ok {
		if existingShortCode, found := memStore.GetShortCodeForURL(req.URL); found {
			c.JSON(http.StatusOK, ShortenResponse{
				ShortURL: h.BaseURL + existingShortCode,
			})
			return
		}
	}

	var shortCode string
	var err error
	//尝试生成唯一的短代码
	for i := 0; i < 5; i++ { //最多尝试5次
		shortCode, err = generateRandomShortCode()
		if err != nil {
			log.Printf("Error generating short code: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate short code"})
			return
		}

		//检查是否存在，如果不存在则保存
		if _, getErr := h.Store.Get(shortCode); errors.Is(getErr, ErrShortCodeNotFound) {
			err = h.Store.Save(shortCode, req.URL)
			if err == nil {
				break
			} else if errors.Is(err, ErrShortCodeExists) {
				//碰撞了，继续循环生成新的
				log.Printf("Short code %s collision, retrying...", shortCode)
				continue
			} else {
				//其他保存错误
				log.Printf("Error saving URL mapping: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save URL mapping"})
				return
			}
		} else if getErr == nil {
			//碰撞了，继续循环生成新的
			log.Printf("Short code %s collision (found via Get), retrying...", shortCode)
			continue
		} else {
			//Get 操作发生其他错误
			log.Printf("Error checking short code existence: %v", getErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check URL mapping"})
		}
		shortCode = "" //重置 shortCode 表示本次尝试失败
	}

	if shortCode == "" {
		log.Println("Failed to generate a unique short code after multiple retries")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate unique short code"})
		return
	}

	log.Printf("Generated short code: %s for URL:%s", shortCode, req.URL)
	c.JSON(http.StatusOK, ShortenResponse{
		ShortURL: h.BaseURL + shortCode,
	})
}

// HandleRedirect 处理短链接重定向
func (h *AppHandlers) HandleRedirect(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.String(http.StatusBadRequest, "Short code cannot be empty")
		return
	}

	originURL, err := h.Store.Get(shortCode)
	if err != nil {
		if errors.Is(err, ErrShortCodeNotFound) {
			log.Printf("Short code not found: %s", shortCode)
			c.String(http.StatusNotFound, "URL not found")
			return
		}
		log.Printf("Error retrieving URL for short code %s: %v", shortCode, err)
		c.String(http.StatusInternalServerError, "Internal server error")
		return
	}

	log.Printf("Redirecting %s to %s", shortCode, originURL)
	c.Redirect(http.StatusFound, originURL)
}
