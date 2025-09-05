package main

import "time"

type RateLimiter struct {
	clients   map[string]time.Time
	deltaTime time.Duration
}

func NewRateLimiter(deltaTime time.Duration) RateLimiter {
	return RateLimiter{
		clients:   make(map[string]time.Time),
		deltaTime: deltaTime,
	}
}

func (rl *RateLimiter) IsWithinLimit(ip string) bool {
	now := time.Now()
	if last, exists := rl.clients[ip]; exists {
		if now.Sub(last) < rl.deltaTime {
			return false
		}
	}
	rl.clients[ip] = now
	return true
}

func EvictOldEntries(rl *RateLimiter) {
	cutoff := time.Now().Add(-rl.deltaTime)
	for ip, last := range rl.clients {
		if last.Before(cutoff) {
			delete(rl.clients, ip)
		}
	}
}
