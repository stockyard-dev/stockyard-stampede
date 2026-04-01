package server

import "github.com/stockyard-dev/stockyard-stampede/internal/license"

type Limits struct {
	MaxTests       int // 0 = unlimited
	MaxConcurrency int // max workers per test
	MaxDuration    int // max seconds per test
	RetentionDays  int
}

var freeLimits = Limits{
	MaxTests:       5,
	MaxConcurrency: 10,
	MaxDuration:    30,
	RetentionDays:  7,
}

var proLimits = Limits{
	MaxTests:       0,
	MaxConcurrency: 500,
	MaxDuration:    600,
	RetentionDays:  90,
}

func LimitsFor(info *license.Info) Limits {
	if info != nil && info.IsPro() {
		return proLimits
	}
	return freeLimits
}

func LimitReached(limit, current int) bool {
	if limit == 0 {
		return false
	}
	return current >= limit
}
