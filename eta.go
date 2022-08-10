package eta

import (
	"sync"
	"time"
)

// Calculator represents ETA calculator
type Calculator struct {
	startTime time.Time
	processed int

	// Expected processing count
	TotalCount int

	// Number of periods to store
	PeriodCount int

	periodDuration   time.Duration
	currentPeriod    time.Time
	currentProcessed int
	stats            []int

	mu sync.RWMutex
}

// New return new ETA calculator
func New(totalCount int) *Calculator {
	return NewCustom(totalCount, defaultPeriodDuration)
}

// NewCustom return new ETA calculator with custom params
func NewCustom(totalCount int, periodDuration time.Duration) *Calculator {
	now := time.Now()

	etaCalc := &Calculator{
		startTime:      now,
		TotalCount:     totalCount,
		PeriodCount:    defaultPeriodCount,
		currentPeriod:  now.Truncate(periodDuration),
		periodDuration: periodDuration}

	return etaCalc
}

// Increment increments processing count
func (ec *Calculator) Increment(n int) {
	if n <= 0 {
		return
	}

	now := time.Now()

	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.processed += n

	// -------------------------------------------------------------------------
	period := now.Truncate(ec.periodDuration)

	if ec.currentPeriod == period {
		ec.currentProcessed += n
		return
	} else {
		ec.stats = append(ec.stats, ec.currentProcessed)
		ec.currentProcessed = 0
		ec.currentPeriod = period
	}

	if len(ec.stats) > ec.PeriodCount {
		ec.stats = ec.stats[:ec.PeriodCount]
	}
}

// Last returns ETA based on last period processing speed
func (ec *Calculator) Last() time.Time {
	if ec.processed == 0 {
		return time.Time{}
	}

	ec.mu.RLock()
	defer ec.mu.RUnlock()

	lastPeriodSpeed := ec.periodDuration / time.Duration(ec.stats[len(ec.stats)-1])

	return time.Now().Add(lastPeriodSpeed * time.Duration(ec.TotalCount-ec.processed))
}

// Eta returns ETA based on total time and total processed items count
func (ec *Calculator) Eta() time.Time {
	if ec.processed == 0 {
		return time.Time{}
	}

	ec.mu.RLock()
	defer ec.mu.RUnlock()

	now := time.Now()
	elapsedTime := now.Sub(ec.startTime)
	avgSpeed := elapsedTime / time.Duration(ec.processed)

	return now.Add(avgSpeed * time.Duration(ec.TotalCount-ec.processed))
}

// Average returns ETA based on average processing speed of last periods
func (ec *Calculator) Average() time.Time {
	if len(ec.stats) == 0 {
		return ec.Eta()
	}

	ec.mu.RLock()
	defer ec.mu.RUnlock()

	now := time.Now()

	processed := ec.stats[len(ec.stats)-1]
	startPeriod := ec.currentPeriod.Add(-ec.periodDuration)

	for i := len(ec.stats) - 2; i >= 0; i-- {
		processed += ec.stats[i]
		startPeriod = startPeriod.Add(-ec.periodDuration)
	}

	if processed == 0 {
		return time.Time{}
	}

	avgSpeed := ec.currentPeriod.Sub(startPeriod) / time.Duration(processed)

	return now.Add(time.Duration(ec.TotalCount-ec.processed) * avgSpeed)
}

// Optimistic returns ETA based on detected maximum of processing speed
func (ec *Calculator) Optimistic() time.Time {
	if len(ec.stats) == 0 {
		return ec.Eta()
	}

	ec.mu.RLock()
	defer ec.mu.RUnlock()

	now := time.Now()

	var maxSpeed time.Duration
	if ec.stats[len(ec.stats)-1] > 0 {
		maxSpeed = ec.periodDuration / time.Duration(ec.stats[len(ec.stats)-1])
	} else {
		maxSpeed = 0
	}

	for i := len(ec.stats) - 2; i >= 1; i-- {
		if ec.stats[i-1] == 0 {
			continue
		}

		newMaxSpeed := ec.periodDuration / time.Duration(ec.stats[i-1])
		if newMaxSpeed < maxSpeed && newMaxSpeed > 0 {
			maxSpeed = newMaxSpeed
		}
	}

	return now.Add(time.Duration(ec.TotalCount-ec.processed) * maxSpeed)
}

// Pessimistic returns ETA based on detected minimum of processing speed
func (ec *Calculator) Pessimistic() time.Time {
	if len(ec.stats) == 0 {
		return ec.Eta()
	}

	ec.mu.RLock()
	defer ec.mu.RUnlock()

	now := time.Now()

	var minSpeed time.Duration
	if ec.stats[len(ec.stats)-1] > 0 {
		minSpeed = ec.periodDuration / time.Duration(ec.stats[len(ec.stats)-1])
	} else {
		minSpeed = 0
	}

	nulPeriods := 0

	for i := len(ec.stats) - 2; i >= 1; i-- {
		if ec.stats[i-1] == 0 {
			nulPeriods += 1
			continue
		}

		newMinSpeed := ec.periodDuration / time.Duration(ec.stats[i-1])
		if newMinSpeed > minSpeed {
			minSpeed = newMinSpeed
		}
	}

	return now.Add(time.Duration(ec.TotalCount-ec.processed) * minSpeed * time.Duration(1+nulPeriods))
}
