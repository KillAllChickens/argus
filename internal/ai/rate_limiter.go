package ai

import (
	"argus/internal/helpers"
	"argus/internal/shared"
	"argus/internal/vars"

	"sync"
	"time"
)

// TokenRateLimiter manages token usage to stay within a defined limit per minute.
// It is safe for concurrent use.
type TokenRateLimiter struct {
	mu                sync.Mutex
	limit             int // Tokens per minute
	highWaterMark     int // Threshold to start waiting
	currentTokenCount int
	windowStartTime   time.Time
}

// NewTokenRateLimiter creates and initializes a new rate limiter.
func NewTokenRateLimiter(limit int, highWaterMark int) *TokenRateLimiter {
	return &TokenRateLimiter{
		limit:             limit,
		highWaterMark:     highWaterMark,
		windowStartTime:   time.Now(),
		currentTokenCount: 0,
	}
}

// waitIfNearLimit checks the current token count and sleeps if it's near the limit.
// This is the core logic that fulfills the user's request.
func (l *TokenRateLimiter) waitIfNearLimit() {
	l.mu.Lock()
	defer l.mu.Unlock()

	// This loop is crucial. After sleeping, we might need to re-evaluate.
	for {
		now := time.Now()
		elapsed := now.Sub(l.windowStartTime)

		// 1. Check if the 1-minute window has passed. If so, reset everything.
		if elapsed >= (30*time.Second) {
			if vars.Verbose {
				shared.Bar.Clear()
				helpers.V("Rate limit window reset.")
			}
			l.windowStartTime = now
			l.currentTokenCount = 0
			return // Window is fresh, no need to wait.
		}

		// 2. Check if we've passed the high-water mark.
		if l.currentTokenCount >= l.highWaterMark {
			// Calculate how much time is left in the current 1-minute window.
			timeLeftInWindow := (30*time.Second) - elapsed
			// Add a safety buffer as requested.
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

			// IMPORTANT: We must unlock the mutex before sleeping to allow other goroutines
			// to run. We will re-lock and re-evaluate after waking up.
			l.mu.Unlock()
			time.Sleep(sleepDuration)
			l.mu.Lock() // Re-acquire the lock to re-check the condition in the loop.

			// 'continue' will cause the loop to restart, re-checking the window.
			continue
		}

		// 3. If we are below the high-water mark, we can proceed.
		return
	}
}

// recordUsage adds the tokens used by a successful API call to the current window's count.
func (l *TokenRateLimiter) recordUsage(tokens int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// It's possible the window reset while we were waiting for the API call to finish.
	// Only add to the count if we are still in the same window.
	if time.Since(l.windowStartTime) < (30*time.Second) {
		l.currentTokenCount += tokens
		if vars.Verbose {
			shared.Bar.Clear()
			helpers.V("Recorded %d tokens. Total for this minute: %d", tokens, l.currentTokenCount)
		}
	} else {
		// The window has already reset, so this call's cost applies to the new window.
		l.windowStartTime = time.Now()
		l.currentTokenCount = tokens
		if vars.Verbose {
			shared.Bar.Clear()
			helpers.V("Window reset during API call. Recorded %d tokens for new window.", tokens)
		}
	}
}
