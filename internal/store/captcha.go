// Package store untuk menyimpan captcha
package store

import (
	"strings"
	"sync"
	"time"
)

type CaptchaItem struct {
	Answer string
	Expiry time.Time
}

type CaptchaStore struct {
	m map[string]CaptchaItem
	sync.Mutex
}

var Store = &CaptchaStore{
	m: make(map[string]CaptchaItem),
}

func (s *CaptchaStore) Add(id, answer string, ttl time.Duration) {
	s.Lock()
	defer s.Unlock() // lebih aman pakai defer
	s.m[id] = CaptchaItem{
		Answer: strings.ToUpper(strings.TrimSpace(answer)),
		Expiry: time.Now().Add(ttl),
	}
}

func (s *CaptchaStore) Verify(id, answer string) bool {
	s.Lock()
	defer s.Unlock()

	item, ok := s.m[id]
	if !ok {
		return false
	}

	// hapus segera → one-time use
	delete(s.m, id)

	if time.Now().After(item.Expiry) {
		return false
	}

	userAnswer := strings.ToUpper(strings.TrimSpace(answer))
	return userAnswer == item.Answer
}
