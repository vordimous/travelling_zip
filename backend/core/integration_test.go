package core

import (
	"sort"
	"testing"
)

// TestIntegrationFullDayDefaultConfig runs the simulator end-to-end against
// the real hospitals.csv and orders.csv with default config and asserts the
// scheduler invariants:
//
//   - every order is fulfilled by midnight
//   - no flight's cumulative route distance exceeds the range limit
//   - no more than NumZips zips are airborne at any instant
//   - Emergency orders, on average, are launched sooner after receipt than
//     Resupply orders (priority is observable in delivery delay).
func TestIntegrationFullDayDefaultConfig(t *testing.T) {
	hospitals := LoadHospitals("../../data/hospitals.csv")
	orders := LoadOrders("../../data/orders.csv")
	config := DefaultConfig()
	scheduler := NewZipScheduler(hospitals, config)
	runner := NewRunner(append([]Order{}, orders...), hospitals, scheduler)

	flights := runner.Simulate(false)

	if got := len(scheduler.UnfulfilledOrders()); got != 0 {
		t.Errorf("UnfulfilledOrders after full-day simulation = %d, want 0", got)
	}

	for _, flight := range flights {
		distance := scheduler.routeDistance(flight.HospitalNames)
		if distance > float64(config.ZipMaxCumulativeRangeM) {
			t.Errorf("flight %v exceeds range: %.0f m > %d m",
				flight.HospitalNames, distance, config.ZipMaxCumulativeRangeM)
		}
	}

	maxConcurrent := computeMaxConcurrentFlights(scheduler, flights)
	if maxConcurrent > config.NumZips {
		t.Errorf("max concurrent flights = %d, exceeds fleet size %d",
			maxConcurrent, config.NumZips)
	}

	emergencyDelay, resupplyDelay := computeMeanDelays(orders, flights)
	if emergencyDelay >= resupplyDelay {
		t.Errorf("Emergency mean delay (%.0f s) should be less than Resupply (%.0f s)",
			emergencyDelay, resupplyDelay)
	}

	t.Logf("flights=%d unfulfilled=0 maxConcurrent=%d Emean=%.0fs Rmean=%.0fs",
		len(flights), maxConcurrent, emergencyDelay, resupplyDelay)
}

// computeMaxConcurrentFlights replays each flight's launch and computed return
// time as a sweep-line; returns the peak count of in-flight zips.
func computeMaxConcurrentFlights(scheduler *ZipScheduler, flights []Flight) int {
	type event struct {
		t       int
		isStart bool
	}
	events := make([]event, 0, 2*len(flights))
	for _, flight := range flights {
		distance := scheduler.routeDistance(flight.HospitalNames)
		duration := int(distance / float64(scheduler.zipSpeedMps))
		events = append(events,
			event{t: flight.LaunchTime, isStart: true},
			event{t: flight.LaunchTime + duration, isStart: false},
		)
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].t != events[j].t {
			return events[i].t < events[j].t
		}
		// process completions before starts at the same instant
		return !events[i].isStart && events[j].isStart
	})
	concurrent, peak := 0, 0
	for _, e := range events {
		if e.isStart {
			concurrent++
			if concurrent > peak {
				peak = concurrent
			}
		} else {
			concurrent--
		}
	}
	return peak
}

func computeMeanDelays(orders []Order, flights []Flight) (emergency float64, resupply float64) {
	receivedTime := make(map[string]int, len(orders))
	priorityByID := make(map[string]string, len(orders))
	for _, order := range orders {
		receivedTime[order.ID] = order.Time
		priorityByID[order.ID] = order.Priority
	}
	var emergencySum, emergencyCount, resupplySum, resupplyCount int
	for _, flight := range flights {
		for _, id := range flight.OrderIDs {
			delay := flight.LaunchTime - receivedTime[id]
			if priorityByID[id] == Emergency {
				emergencySum += delay
				emergencyCount++
			} else {
				resupplySum += delay
				resupplyCount++
			}
		}
	}
	if emergencyCount > 0 {
		emergency = float64(emergencySum) / float64(emergencyCount)
	}
	if resupplyCount > 0 {
		resupply = float64(resupplySum) / float64(resupplyCount)
	}
	return emergency, resupply
}
