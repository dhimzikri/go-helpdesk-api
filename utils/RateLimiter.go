// ratelimiter.go
package utils

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type RateLimiter struct {
	Cache      *cache.Cache
	RateLimit  int
	TimeWindow time.Duration
}

// NewRateLimiter creates a new rate limiter with custom rate limit and time window.
func NewRateLimiter(rateLimit int, timeWindow time.Duration) *RateLimiter {
	return &RateLimiter{
		Cache:      cache.New(timeWindow, 10*time.Minute), // 10 minute cleanup interval
		RateLimit:  rateLimit,
		TimeWindow: timeWindow,
	}
}

// CheckRateLimit checks if the rate limit has been exceeded for a given key (IP or user).
func (rl *RateLimiter) CheckRateLimit(key string) bool {
	// Get the current timestamp
	now := time.Now()

	// Retrieve the request timestamps for the provided key (e.g., IP address)
	requests, found := rl.Cache.Get(key)
	if !found {
		requests = []time.Time{}
	}

	// Filter out timestamps that are older than the time window
	validRequests := []time.Time{}
	for _, reqTime := range requests.([]time.Time) {
		if now.Sub(reqTime) <= rl.TimeWindow {
			validRequests = append(validRequests, reqTime)
		}
	}

	// If the number of valid requests exceeds the rate limit, return true (indicating exceeded limit)
	if len(validRequests) >= rl.RateLimit {
		return true
	}

	// Otherwise, add the current timestamp and update the cache
	validRequests = append(validRequests, now)
	rl.Cache.Set(key, validRequests, cache.DefaultExpiration)

	return false
}
