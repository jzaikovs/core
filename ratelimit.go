package core

import (
	"time"
)

type ratelimit struct {
	rate       float32 // requests
	per        float32 // seconds
	allowance  float32
	lass_check time.Time
}

func new_ratelimit(rate, per float32) *ratelimit {
	return &ratelimit{rate: rate, per: per, allowance: rate}
}

// thx http://stackoverflow.com/a/668327
func (this *ratelimit) test(current time.Time) bool {
	time_passed := time.Since(current)
	this.lass_check = current
	this.allowance += float32(time_passed.Seconds()) * (this.rate / this.per)

	if this.allowance > this.rate {
		this.allowance = this.rate
	}

	if this.allowance < 1.0 {
		return false
	}

	this.allowance -= 1.0
	return true
}
