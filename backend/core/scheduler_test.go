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

// fleetUnderReserveStress sets up a scheduler with 5 zips of which 4 are
// already in flight at currentTime=0, leaving exactly one free zip and a
// reserve size of one (5*80/100=4 cap → 1 reserved). A single Resupply order
// to "Near" is queued. The fixture isolates the reserve gating decision.
func fleetUnderReserveStress(t *testing.T, policy ReservePolicy) *ZipScheduler {
	t.Helper()
	hospitals := map[string]Hospital{
		"Near": {Name: "Near", NorthM: 1000, EastM: 0},
	}
	scheduler := NewZipScheduler(hospitals, SimulationConfig{
		NumZips:                5,
		MaxPackagesPerZip:      3,
		ZipSpeedMps:            30,
		ZipMaxCumulativeRangeM: 160000,
	})
	scheduler.reservePolicy = policy
	for i := 0; i < 4; i++ {
		scheduler.markZipLaunched(1000)
	}
	scheduler.QueueOrder(Order{ID: "r1", Time: 0, HospitalName: "Near", Priority: Resupply})
	return scheduler
}

func TestReserveHardBlocksResupplyWhenOnlyReserveFree(t *testing.T) {
	scheduler := fleetUnderReserveStress(t, ReserveHard)
	flights := scheduler.LaunchFlights(0)
	if len(flights) != 0 {
		t.Errorf("ReserveHard launched %d flights with only the reserve free; want 0",
			len(flights))
	}
	if got := len(scheduler.UnfulfilledOrders()); got != 1 {
		t.Errorf("Resupply queue size after blocked launch = %d, want 1", got)
	}
}

func TestReserveNoneLaunchesResupplyEvenIntoReserve(t *testing.T) {
	scheduler := fleetUnderReserveStress(t, ReserveNone)
	flights := scheduler.LaunchFlights(0)
	if len(flights) != 1 {
		t.Errorf("ReserveNone launched %d flights, want 1", len(flights))
	}
}

// At t = SecondsPerDay - 30 the seconds-remaining (30) is less than the round
// trip flight time to Near (~66 s), so resupplyAtRisk fires and the Soft
// policy permits borrowing the reserve.
func TestReserveSoftBorrowsReserveOnEoDRisk(t *testing.T) {
	scheduler := fleetUnderReserveStress(t, ReserveSoft)
	flights := scheduler.LaunchFlights(SecondsPerDay - 30)
	if len(flights) != 1 {
		t.Errorf("ReserveSoft did not borrow reserve at EoD: launched %d, want 1",
			len(flights))
	}
}

// Earlier in the day the at-risk threshold does not fire; Soft behaves like
// Hard and blocks Resupply.
func TestReserveSoftBlocksResupplyWhenNotAtRisk(t *testing.T) {
	scheduler := fleetUnderReserveStress(t, ReserveSoft)
	flights := scheduler.LaunchFlights(0)
	if len(flights) != 0 {
		t.Errorf("ReserveSoft launched %d flights at t=0; want 0 (not at risk)",
			len(flights))
	}
}

func TestResupplyAtRiskThreshold(t *testing.T) {
	scheduler := fleetUnderReserveStress(t, ReserveSoft)
	order := Order{HospitalName: "Near"}
	if scheduler.resupplyAtRisk(0, order) {
		t.Error("resupplyAtRisk(t=0) = true, want false")
	}
	if !scheduler.resupplyAtRisk(SecondsPerDay-10, order) {
		t.Error("resupplyAtRisk(near midnight) = false, want true")
	}
}

// schedulerWithRange returns a scheduler with the test hospitals and a custom
// cumulative range (in meters). Used to make range gating observable in
// small, deterministic tests.
func schedulerWithRange(rangeM int, maxPackages int) *ZipScheduler {
	hospitals := map[string]Hospital{
		"Near": {Name: "Near", NorthM: 1000, EastM: 0},
		"Mid":  {Name: "Mid", NorthM: 5000, EastM: 0},
		"Far":  {Name: "Far", NorthM: 20000, EastM: 0},
	}
	return NewZipScheduler(hospitals, SimulationConfig{
		NumZips:                1,
		MaxPackagesPerZip:      maxPackages,
		ZipSpeedMps:            30,
		ZipMaxCumulativeRangeM: rangeM,
	})
}

func TestBuildFlightRejectsOrdersBeyondRange(t *testing.T) {
	// With range = 30 km, a Far (20 km out) round trip is 40 km — over range
	// alone — and combining Far with anything else makes the route worse.
	// Near + Mid together is 10 km and must fit.
	scheduler := schedulerWithRange(30000, 3)
	candidates := []Order{
		{ID: "n", Time: 0, HospitalName: "Near", Priority: Emergency},
		{ID: "m", Time: 0, HospitalName: "Mid", Priority: Emergency},
		{ID: "f", Time: 0, HospitalName: "Far", Priority: Emergency},
	}

	flight, leftover := scheduler.buildFlight(0, candidates)

	if len(flight.OrderIDs) != 2 {
		t.Errorf("flight orderIDs = %v, want 2 (Near + Mid)", flight.OrderIDs)
	}
	stops := map[string]bool{}
	for _, s := range flight.HospitalNames {
		stops[s] = true
	}
	if stops["Far"] {
		t.Errorf("flight should not include Far (over range): stops=%v", flight.HospitalNames)
	}
	if len(leftover) != 1 || leftover[0].ID != "f" {
		t.Errorf("leftover = %v, want [f]", leftover)
	}
}

func TestBuildFlightCollapsesDuplicateStops(t *testing.T) {
	scheduler := schedulerWithRange(160000, 3)
	candidates := []Order{
		{ID: "a", Time: 0, HospitalName: "Near", Priority: Resupply},
		{ID: "b", Time: 0, HospitalName: "Near", Priority: Resupply},
		{ID: "c", Time: 0, HospitalName: "Near", Priority: Resupply},
	}

	flight, leftover := scheduler.buildFlight(0, candidates)

	if len(flight.HospitalNames) != 1 || flight.HospitalNames[0] != "Near" {
		t.Errorf("hospitalNames = %v, want [Near] (duplicates collapsed)", flight.HospitalNames)
	}
	if len(flight.OrderIDs) != 3 {
		t.Errorf("orderIDs = %v, want 3 packages on the single stop", flight.OrderIDs)
	}
	if len(leftover) != 0 {
		t.Errorf("leftover = %v, want []", leftover)
	}
}

func TestBuildFlightCapsAtMaxPackagesPerZip(t *testing.T) {
	// MaxPackages = 2; queue 3 orders to distinct in-range hospitals.
	scheduler := schedulerWithRange(160000, 2)
	candidates := []Order{
		{ID: "a", Time: 0, HospitalName: "Near", Priority: Emergency},
		{ID: "b", Time: 0, HospitalName: "Mid", Priority: Emergency},
		{ID: "c", Time: 0, HospitalName: "Far", Priority: Emergency},
	}

	flight, leftover := scheduler.buildFlight(0, candidates)

	if len(flight.OrderIDs) != 2 {
		t.Errorf("orderIDs = %d, want 2 (capped at MaxPackagesPerZip)", len(flight.OrderIDs))
	}
	if len(leftover) != 1 || leftover[0].ID != "c" {
		t.Errorf("leftover = %v, want [c] (third order deferred)", leftover)
	}
}

func TestPendingByPriorityEmergencyFirstFIFOWithin(t *testing.T) {
	scheduler := newTestScheduler(1)
	scheduler.QueueOrder(Order{ID: "r1", Time: 10, HospitalName: "Near", Priority: Resupply})
	scheduler.QueueOrder(Order{ID: "e1", Time: 20, HospitalName: "Mid", Priority: Emergency})
	scheduler.QueueOrder(Order{ID: "r2", Time: 30, HospitalName: "Near", Priority: Resupply})
	scheduler.QueueOrder(Order{ID: "e2", Time: 40, HospitalName: "Far", Priority: Emergency})

	got := scheduler.pendingByPriority()
	wantOrder := []string{"e1", "e2", "r1", "r2"}
	if len(got) != len(wantOrder) {
		t.Fatalf("pendingByPriority length = %d, want %d", len(got), len(wantOrder))
	}
	for i, id := range wantOrder {
		if got[i].ID != id {
			t.Errorf("pendingByPriority[%d].ID = %q, want %q", i, got[i].ID, id)
		}
	}
}

// With a single zip, an Emergency queued AFTER a Resupply must launch first.
func TestLaunchFlightsLaunchesEmergencyBeforeEarlierResupply(t *testing.T) {
	scheduler := newTestScheduler(1)
	scheduler.QueueOrder(Order{ID: "r1", Time: 10, HospitalName: "Near", Priority: Resupply})
	scheduler.QueueOrder(Order{ID: "e1", Time: 20, HospitalName: "Far", Priority: Emergency})

	flights := scheduler.LaunchFlights(30)
	if len(flights) != 1 {
		t.Fatalf("LaunchFlights returned %d flights, want 1", len(flights))
	}
	first := flights[0].OrderIDs[0]
	if first != "e1" {
		t.Errorf("first launched order = %q, want %q (emergency must precede earlier resupply)", first, "e1")
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
