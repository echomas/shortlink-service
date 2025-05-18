package main

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
