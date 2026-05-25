package api

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ChatLimits describes static chat quotas exposed via GET /api/chat/status.
type ChatLimits struct {
	PerMinute       int `json:"perMinute"`
	Daily           int `json:"daily"`
	MaxMessageChars int `json:"maxMessageChars"`
	MaxCodeChars    int `json:"maxCodeChars"`
	MaxBodyBytes    int `json:"maxBodyBytes"`
}

// RateLimitError is returned when a client exceeds chat quotas.
type RateLimitError struct {
	Message    string
	RetryAfter int
}

// ChatLimiter tracks per-client minute and daily counters for POST /api/chat.
type ChatLimiter struct {
	Limits ChatLimits

	mu     sync.Mutex
	minute map[string]bucketCount
	daily  map[string]bucketCount
}

type bucketCount struct {
	bucket int64
	count  int
}

// NewChatLimiterFromEnv builds a limiter using CHAT_* environment variables.
func NewChatLimiterFromEnv() *ChatLimiter {
	return &ChatLimiter{
		Limits: ChatLimits{
			PerMinute:       envInt("CHAT_RATE_PER_MIN", 5),
			Daily:           envInt("CHAT_RATE_DAILY", 50),
			MaxMessageChars: envInt("CHAT_MAX_MESSAGE_CHARS", 4000),
			MaxCodeChars:    envInt("CHAT_MAX_CODE_CHARS", 16000),
			MaxBodyBytes:    envInt("CHAT_MAX_BODY_BYTES", 65536),
		},
		minute: make(map[string]bucketCount),
		daily:  make(map[string]bucketCount),
	}
}

// Check records one chat request or returns a rate-limit error.
func (l *ChatLimiter) Check(client string) *RateLimitError {
	if l.Limits.PerMinute == 0 && l.Limits.Daily == 0 {
		return nil
	}

	now := time.Now().Unix()
	minuteBucket := now / 60
	dayBucket := now / 86400
	key := client
	if key == "" {
		key = "unknown"
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.pruneLocked(minuteBucket, dayBucket, now)

	if l.Limits.PerMinute > 0 {
		entry := l.minute[key]
		if entry.bucket != minuteBucket {
			entry = bucketCount{bucket: minuteBucket}
		}
		if entry.count >= l.Limits.PerMinute {
			retry := int(60 - (now % 60))
			if retry < 1 {
				retry = 1
			}
			return &RateLimitError{
				Message:    "rate limit exceeded: " + strconv.Itoa(l.Limits.PerMinute) + " requests per minute",
				RetryAfter: retry,
			}
		}
		entry.count++
		l.minute[key] = entry
	}

	if l.Limits.Daily > 0 {
		entry := l.daily[key]
		if entry.bucket != dayBucket {
			entry = bucketCount{bucket: dayBucket}
		}
		if entry.count >= l.Limits.Daily {
			retry := int(86400 - (now % 86400))
			if retry < 1 {
				retry = 1
			}
			return &RateLimitError{
				Message:    "rate limit exceeded: " + strconv.Itoa(l.Limits.Daily) + " requests per day",
				RetryAfter: retry,
			}
		}
		entry.count++
		l.daily[key] = entry
	}

	return nil
}

func (l *ChatLimiter) pruneLocked(minuteBucket, dayBucket, now int64) {
	if len(l.minute) > 10000 {
		next := make(map[string]bucketCount, len(l.minute))
		for k, v := range l.minute {
			if v.bucket == minuteBucket {
				next[k] = v
			}
		}
		l.minute = next
	}
	if len(l.daily) > 10000 {
		next := make(map[string]bucketCount, len(l.daily))
		for k, v := range l.daily {
			if v.bucket == dayBucket {
				next[k] = v
			}
		}
		l.daily = next
	}
	if now%600 == 0 {
		prevMinute := minuteBucket - 2
		prevDay := dayBucket - 2
		for k, v := range l.minute {
			if v.bucket < prevMinute {
				delete(l.minute, k)
			}
		}
		for k, v := range l.daily {
			if v.bucket < prevDay {
				delete(l.daily, k)
			}
		}
	}
}

func trustForwardedHeaders() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("TRUST_PROXY"))) {
	case "1", "true", "yes":
		return true
	}
	return strings.TrimSpace(os.Getenv("K_SERVICE")) != ""
}

func clientIP(r *http.Request) string {
	if trustForwardedHeaders() {
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			if first := strings.TrimSpace(strings.Split(forwarded, ",")[0]); first != "" {
				return first
			}
		}
		if real := strings.TrimSpace(r.Header.Get("X-Real-Ip")); real != "" {
			return real
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || host == "" {
		if r.RemoteAddr == "" {
			return "unknown"
		}
		return r.RemoteAddr
	}
	return host
}

func envInt(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return fallback
	}
	return n
}
