package ai

import (
	"argus/internal/helpers"
	"argus/internal/shared"
	"argus/internal/vars"

	"sync"
	"time"
)

type TokenRateLimiter struct {
	mu                sync.Mutex
	limit             int
	highWaterMark     int
	currentTokenCount int
	windowStartTime   time.Time
}

func NewTokenRateLimiter(limit int, highWaterMark int) *TokenRateLimiter {
	return &TokenRateLimiter{
		limit:             limit,
		highWaterMark:     highWaterMark,
		windowStartTime:   time.Now(),
		currentTokenCount: 0,
	}
}

func (l *TokenRateLimiter) waitIfNearLimit() {
	l.mu.Lock()
	defer l.mu.Unlock()

	for {
		now := time.Now()
		elapsed := now.Sub(l.windowStartTime)

		if elapsed >= (30 * time.Second) {
			if vars.Verbose {
				shared.Bar.Clear()
				helpers.V("Rate limit window reset.")
			}
			l.windowStartTime = now
			l.currentTokenCount = 0
			return
		}

		if l.currentTokenCount >= l.highWaterMark {

			timeLeftInWindow := (30 * time.Second) - elapsed

			sleepDuration := timeLeftInWindow + (10 * time.Second)
			if vars.Verbose {
				shared.Bar.Clear()
				helpers.V(
					"Rate limit high-water mark (%d) reached. Current count: %d. Sleeping for %v...",
					l.highWaterMark,
					l.currentTokenCount,
					sleepDuration,
				)
			}

			l.mu.Unlock()
			time.Sleep(sleepDuration)
			l.mu.Lock()

			continue
		}

		return
	}
}

func (l *TokenRateLimiter) recordUsage(tokens int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if time.Since(l.windowStartTime) < (30 * time.Second) {
		l.currentTokenCount += tokens
		if vars.Verbose {
			shared.Bar.Clear()
			helpers.V("Recorded %d tokens. Total for this minute: %d", tokens, l.currentTokenCount)
		}
	} else {

		l.windowStartTime = time.Now()
		l.currentTokenCount = tokens
		if vars.Verbose {
			shared.Bar.Clear()
			helpers.V("Window reset during API call. Recorded %d tokens for new window.", tokens)
		}
	}
}
