package data

import (
	"math"
	"time"
)

type PricingRule struct {
	Version         string
	StartPrice      float64
	IncludedKm      float64
	PerKmPrice      float64
	PerMinutePrice  float64
	MinimumPrice    float64
	TimeMultiplier  float64
}

func DefaultPricingRule() PricingRule {
	return PricingRule{Version: "local-default-v1", StartPrice: 8, IncludedKm: 3, PerKmPrice: 1.5, PerMinutePrice: 0.15, MinimumPrice: 8, TimeMultiplier: 1}
}

func CalculatePrice(rule PricingRule, distanceMeters, durationSeconds int, depart time.Time) float64 {
	if rule.TimeMultiplier <= 0 { rule.TimeMultiplier = 1 }
	kilometers := float64(distanceMeters) / 1000
	minutes := float64(durationSeconds) / 60
	extraKm := math.Max(0, kilometers-rule.IncludedKm)
	price := rule.StartPrice + extraKm*rule.PerKmPrice + minutes*rule.PerMinutePrice
	price = math.Max(price, rule.MinimumPrice) * rule.TimeMultiplier
	return math.Round(price*100) / 100
}
