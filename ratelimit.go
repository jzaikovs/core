package core

import (
	"time"
)

const (
	Header_X_Rate_Limit_Limit     = `X-Rate-Limit-Limit`
	Header_X_Rate_Limit_Remaining = `X-Rate-Limit-Remaining`
)

type ratelimit struct {
	rate      float32 // requests
	per       float32 // seconds
	allowance float32
	lastCheck time.Time
}

func newRateLimit(rate, per float32) *ratelimit {
	return &ratelimit{rate: rate, per: per, allowance: rate}
}

// thx http://stackoverflow.com/a/668327
func (rate *ratelimit) test(current time.Time) bool {
	timePassed := time.Since(current)
	rate.lastCheck = current
	rate.allowance += float32(timePassed.Seconds()) * (rate.rate / rate.per)

	if rate.allowance > rate.rate {
		rate.allowance = rate.rate
	}

	if rate.allowance < 1.0 {
		return false
	}

	rate.allowance -= 1.0
	return true
}
