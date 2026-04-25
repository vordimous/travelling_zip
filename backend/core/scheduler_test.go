package core

import "testing"

func newTestScheduler(numZips int) *ZipScheduler {
	hospitals := map[string]Hospital{
		"Near": {Name: "Near", NorthM: 1000, EastM: 0},
		"Mid":  {Name: "Mid", NorthM: 5000, EastM: 0},
		"Far":  {Name: "Far", NorthM: 20000, EastM: 0},
	}
	config := SimulationConfig{
		NumZips:                numZips,
		MaxPackagesPerZip:      3,
		ZipSpeedMps:            30,
		ZipMaxCumulativeRangeM: 160000,
	}
	return NewZipScheduler(hospitals, config)
}

func TestAvailableZipsStartsAtFleetSize(t *testing.T) {
	scheduler := newTestScheduler(5)
	if got := scheduler.availableZips(0); got != 5 {
		t.Errorf("availableZips(0) on idle fleet = %d, want 5", got)
	}
}

func TestMarkZipLaunchedReducesAvailable(t *testing.T) {
	scheduler := newTestScheduler(3)
	scheduler.markZipLaunched(100)
	scheduler.markZipLaunched(200)
	if got := scheduler.availableZips(50); got != 1 {
		t.Errorf("availableZips with 2 in flight = %d, want 1", got)
	}
}

func TestAvailableZipsReclaimsByReturnTime(t *testing.T) {
	scheduler := newTestScheduler(3)
	scheduler.markZipLaunched(100)
	scheduler.markZipLaunched(200)
	scheduler.markZipLaunched(300)
	// At t=150 the first zip has returned; the other two are still flying.
	if got := scheduler.availableZips(150); got != 1 {
		t.Errorf("availableZips(150) = %d, want 1", got)
	}
	// At t=300 all have returned (return time t means landed by t).
	if got := scheduler.availableZips(300); got != 3 {
		t.Errorf("availableZips(300) = %d, want 3", got)
	}
}

// LaunchFlights must never return more flights than the number of zips
// available at currentTime.
func TestLaunchFlightsRespectsFleetSize(t *testing.T) {
	scheduler := newTestScheduler(2)
	// Queue many emergencies to different hospitals so each needs its own
	// flight (one stop per flight after dedupe; but more importantly enough
	// distinct orders to exercise the fleet cap).
	for i := 0; i < 6; i++ {
		scheduler.QueueOrder(Order{
			ID:           "o" + string(rune('1'+i)),
			Time:         0,
			HospitalName: []string{"Near", "Mid", "Far"}[i%3],
			Priority:     Emergency,
		})
	}
	flights := scheduler.LaunchFlights(0)
	if len(flights) > 2 {
		t.Errorf("LaunchFlights returned %d flights with 2 zips: %v", len(flights), flights)
	}
}

func TestLaunchFlightsReturnsEmptyWhenFleetExhausted(t *testing.T) {
	scheduler := newTestScheduler(1)
	scheduler.markZipLaunched(1000) // single zip is in flight
	scheduler.QueueOrder(Order{
		ID: "o1", Time: 0, HospitalName: "Near", Priority: Emergency,
	})
	if got := scheduler.LaunchFlights(0); len(got) != 0 {
		t.Errorf("LaunchFlights with no available zips returned %d flights, want 0", len(got))
	}
}
