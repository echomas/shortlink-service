package main

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"log"
)

var (
	ErrShortCodeNotFound = errors.New("short code not found")
	ErrShortCodeExists   = errors.New("short code already exists")
)

// Storage
type Storage interface {
	Save(shortCode, originURL string) error
	Get(shortCode string) (string, error)
}

//RedisStorage

type RedisStorage struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisStorage 创建了一个新的 RedisStorage 实例
func NewRedisStorage(addr, password string, db int) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return &RedisStorage{client: client, ctx: ctx}, nil
}

// Save 将短链接和原始 URL 保存到 Redis
func (s *RedisStorage) Save(shortCode, originalURL string) error {
	//假设 shortCode 应该是唯一的，如果已存在则报错
	//Redis 中，短链接作为 key， 长链接作为 value
	//检查 shortCode 是否已存在
	exists, err := s.client.Exists(s.ctx, "short:"+shortCode).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return ErrShortCodeExists
	}
	//保存 shortCode -> originalURL
	err = s.client.Set(s.ctx, "short:"+shortCode, originalURL, 0).Err()
	if err != nil {
		return err
	}
	//保存 originalURL -> shortCode
	err = s.client.Set(s.ctx, "url:"+originalURL, shortCode, 0).Err()
	return err
}

// Get 从 Redis 获取原始 URL
func (s *RedisStorage) Get(shortCode string) (string, error) {
	originalURL, err := s.client.Get(s.ctx, "short:"+shortCode).Result()
	if err == redis.Nil {
		return "", ErrShortCodeNotFound
	} else if err != nil {
		return "", err
	}
	return originalURL, nil
}

func (s *RedisStorage) GetShortCodeForURL(originalURL string) (string, bool) {
	shortCode, err := s.client.Get(s.ctx, "url:"+originalURL).Result()
	if err == redis.Nil {
		return "", false
	} else if err != nil {
		log.Printf("Redis error checking for original URL %s: %v", originalURL, err)
		return "", false
	}
	return shortCode, true
}

// 使用计时器生成 shortCode ID （配合 Base62）
func (s *RedisStorage) GetNextID() (int64, error) {
	return s.client.Incr(s.ctx, "shortlink:counter").Result()
}

/*
import (
	"errors"
	"sync"
)

var (
	ErrShortCodeNotFound = errors.New("short code not found")
	ErrShortCodeExists   = errors.New("short code already exists")
)

// Storage
type Storage interface {
	Save(shortCode, originURL string) error
	Get(shortCode string) (string, error)
}

type InMemoryStorage struct {
	mu    sync.RWMutex
	urls  map[string]string //key: shortCode value: originURL
	codes map[string]string //key: originURL value: shortCode
}

// NewInMemoryStorage 创建一个新的 InMemoryStorage 实例
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		urls:  make(map[string]string),
		codes: make(map[string]string),
	}
}

// Save 保存短代码和原始 URL 的映射
func (s *InMemoryStorage) Save(shortCode, originURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.urls[shortCode]; ok {
		return ErrShortCodeExists
	}
	s.urls[shortCode] = originURL
	s.codes[originURL] = shortCode //存储反向映射
	return nil
}

// Get 根据短代码获取原始 URL
func (s *InMemoryStorage) Get(shortCode string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	originURL, ok := s.urls[shortCode]
	if !ok {
		return "", ErrShortCodeNotFound
	}
	return originURL, nil
}

// 如果你想在生成短链接前检查长链接是否已存在
func (s *InMemoryStorage) GetShortCodeForURL(originalURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	shortCode, ok := s.codes[originalURL]
	return shortCode, ok
}
*/
