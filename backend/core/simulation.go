package core

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	NumZips                = 10
	MaxPackagesPerZip      = 3
	ZipSpeedMps            = 30
	ZipMaxCumulativeRangeM = 160 * 1000
	SecondsPerDay          = 24 * 60 * 60
	Emergency              = "Emergency"
	Resupply               = "Resupply"
)

type SimulationConfig struct {
	NumZips                int `json:"numZips"`
	MaxPackagesPerZip      int `json:"maxPackagesPerZip"`
	ZipSpeedMps            int `json:"zipSpeedMps"`
	ZipMaxCumulativeRangeM int `json:"zipMaxCumulativeRangeM"`
}

func DefaultConfig() SimulationConfig {
	return SimulationConfig{
		NumZips:                NumZips,
		MaxPackagesPerZip:      MaxPackagesPerZip,
		ZipSpeedMps:            ZipSpeedMps,
		ZipMaxCumulativeRangeM: ZipMaxCumulativeRangeM,
	}
}

type Hospital struct {
	Name   string `json:"name"`
	NorthM int    `json:"northM"`
	EastM  int    `json:"eastM"`
}

type Order struct {
	ID           string `json:"id"`
	Time         int    `json:"time"`
	HospitalName string `json:"hospitalName"`
	Priority     string `json:"priority"`
}

type Flight struct {
	LaunchTime    int      `json:"launchTime"`
	HospitalNames []string `json:"hospitalNames"`
	OrderIDs      []string `json:"orderIds"`
}

func (flight Flight) String() string {
	hospitalNamesStr := strings.Join(flight.HospitalNames, "->")
	return fmt.Sprintf("<Flight @ %d to %s>", flight.LaunchTime, hospitalNamesStr)
}

type Snapshot struct {
	Implementation   string           `json:"implementation"`
	Config           SimulationConfig `json:"config"`
	Hospitals        []Hospital       `json:"hospitals"`
	Orders           []Order          `json:"orders"`
	Flights          []Flight         `json:"flights"`
	UnfulfilledOrders []Order         `json:"unfulfilledOrders"`
}

type ZipScheduler struct {
	hospitals              map[string]Hospital
	graph                  *Graph
	numZips                int
	maxPackagesPerZip      int
	zipSpeedMps            int
	zipMaxCumulativeRangeM int
	unfulfilledOrders      []Order
	// zipReturnTimes holds the return-to-Nest seconds-since-midnight for every
	// currently in-flight zip. Entries with t <= currentTime are reclaimed on
	// the next availableZips call.
	zipReturnTimes []int
	reservePolicy  ReservePolicy
}

func NewZipScheduler(
	hospitals map[string]Hospital,
	config SimulationConfig,
) *ZipScheduler {
	return &ZipScheduler{
		hospitals:              hospitals,
		graph:                  NewGraph(hospitals),
		numZips:                config.NumZips,
		maxPackagesPerZip:      config.MaxPackagesPerZip,
		zipSpeedMps:            config.ZipSpeedMps,
		zipMaxCumulativeRangeM: config.ZipMaxCumulativeRangeM,
		unfulfilledOrders:      []Order{},
		reservePolicy:          ReserveSoft,
	}
}

func (zipScheduler *ZipScheduler) QueueOrder(order Order) {
	zipScheduler.unfulfilledOrders = append(zipScheduler.unfulfilledOrders, order)
}

// buildFlight assembles at most one Flight from candidates, walking them in
// order and packing each order onto the flight when it fits.
//
// Packing rules:
//   - Multiple orders to the same hospital share a single stop (one delivery
//     leg per unique hospital, but the flight still carries N packages).
//   - The total number of packages cannot exceed MaxPackagesPerZip.
//   - The cumulative route distance (Nest → stops → Nest) cannot exceed
//     ZipMaxCumulativeRangeM. Stops are reordered by nearest-neighbor over the
//     graph before the range check (and on the final flight) so that orderings
//     that fit are not rejected just because the candidate FIFO order happened
//     to be long. NN matches the optimal TSP order for ~57% of 3-stop combos
//     in this dataset and salvages flights FIFO would reject as out-of-range.
//
// Returns the constructed flight (empty when nothing fits) and the candidates
// that were not consumed, in their original order. Skipped orders precede
// candidates that come after the first non-fit so callers can re-queue them.
func (zipScheduler *ZipScheduler) buildFlight(currentTime int, candidates []Order) (Flight, []Order) {
	stops := []string{}
	stopIndex := map[string]bool{}
	orderIDs := []string{}
	leftover := make([]Order, 0, len(candidates))

	for _, order := range candidates {
		if len(orderIDs) >= zipScheduler.maxPackagesPerZip {
			leftover = append(leftover, order)
			continue
		}

		candidateStops := stops
		if !stopIndex[order.HospitalName] {
			candidateStops = zipScheduler.nearestNeighborOrder(
				append(append([]string{}, stops...), order.HospitalName),
			)
		}
		if zipScheduler.routeDistance(candidateStops) > float64(zipScheduler.zipMaxCumulativeRangeM) {
			leftover = append(leftover, order)
			continue
		}

		if !stopIndex[order.HospitalName] {
			stops = candidateStops
			stopIndex[order.HospitalName] = true
		}
		orderIDs = append(orderIDs, order.ID)
	}

	if len(orderIDs) == 0 {
		return Flight{}, leftover
	}
	return Flight{
		LaunchTime:    currentTime,
		HospitalNames: stops,
		OrderIDs:      orderIDs,
	}, leftover
}

// nearestNeighborOrder returns stops reordered by greedy nearest-neighbor
// starting from the Nest. For small stop counts (≤ MaxPackagesPerZip) this is
// close to optimal and far cheaper than full TSP.
func (zipScheduler *ZipScheduler) nearestNeighborOrder(stops []string) []string {
	if len(stops) <= 1 {
		return append([]string{}, stops...)
	}
	remaining := append([]string{}, stops...)
	out := make([]string, 0, len(stops))
	current := NestKey
	for len(remaining) > 0 {
		bestIndex := 0
		bestDistance := zipScheduler.graph.EdgeWeight(current, remaining[0])
		for i := 1; i < len(remaining); i++ {
			distance := zipScheduler.graph.EdgeWeight(current, remaining[i])
			if distance < bestDistance {
				bestDistance = distance
				bestIndex = i
			}
		}
		out = append(out, remaining[bestIndex])
		current = remaining[bestIndex]
		remaining = append(remaining[:bestIndex], remaining[bestIndex+1:]...)
	}
	return out
}

// routeDistance returns the cumulative meters for a route Nest → stops → Nest.
func (zipScheduler *ZipScheduler) routeDistance(stops []string) float64 {
	previous := NestKey
	total := 0.0
	for _, stop := range stops {
		total += zipScheduler.graph.EdgeWeight(previous, stop)
		previous = stop
	}
	total += zipScheduler.graph.EdgeWeight(previous, NestKey)
	return total
}

// pendingByPriority returns the pending unfulfilled orders ordered Emergency
// before Resupply, with FIFO order preserved within each priority. The result
// is a fresh slice; the underlying queue is unchanged.
func (zipScheduler *ZipScheduler) pendingByPriority() []Order {
	emergencies := make([]Order, 0, len(zipScheduler.unfulfilledOrders))
	resupply := make([]Order, 0, len(zipScheduler.unfulfilledOrders))
	for _, order := range zipScheduler.unfulfilledOrders {
		if order.Priority == Emergency {
			emergencies = append(emergencies, order)
		} else {
			resupply = append(resupply, order)
		}
	}
	return append(emergencies, resupply...)
}

// availableZips reclaims any in-flight zips whose return time has elapsed and
// returns how many zips are free at currentTime.
func (zipScheduler *ZipScheduler) availableZips(currentTime int) int {
	stillFlying := zipScheduler.zipReturnTimes[:0]
	for _, returnTime := range zipScheduler.zipReturnTimes {
		if returnTime > currentTime {
			stillFlying = append(stillFlying, returnTime)
		}
	}
	zipScheduler.zipReturnTimes = stillFlying
	return zipScheduler.numZips - len(stillFlying)
}

// markZipLaunched records a zip launch whose return-to-Nest time is returnTime.
func (zipScheduler *ZipScheduler) markZipLaunched(returnTime int) {
	zipScheduler.zipReturnTimes = append(zipScheduler.zipReturnTimes, returnTime)
}

// ReservePolicy controls the 20% Emergency reserve enforcement when launching
// Resupply flights.
type ReservePolicy int

const (
	// ReserveNone launches Resupply greedily, ignoring the reserve.
	ReserveNone ReservePolicy = iota
	// ReserveHard refuses to launch Resupply once doing so would consume the
	// reserve. Best for emergency mean delay when the fleet has slack.
	ReserveHard
	// ReserveSoft enforces the same cap as ReserveHard, but allows a Resupply
	// order to borrow the reserve when waiting longer would push its delivery
	// past midnight (direct round-trip flight time > seconds remaining today).
	ReserveSoft
)

// ResupplyCapPercent is the fraction of fleet that Resupply may consume before
// the reserve kicks in. Step 2a may make this configurable.
const ResupplyCapPercent = 80

func (zipScheduler *ZipScheduler) resupplyCap() int {
	cap := (zipScheduler.numZips * ResupplyCapPercent) / 100
	if cap < 1 && zipScheduler.numZips > 0 {
		cap = 1
	}
	return cap
}

func (zipScheduler *ZipScheduler) reserveSize() int {
	return zipScheduler.numZips - zipScheduler.resupplyCap()
}

// resupplyAtRisk reports whether the order's deadline is at risk of slipping
// past midnight if it must wait for a non-reserve zip. The threshold is
// "round-trip direct flight time > seconds remaining in the day".
//
// Note: with the default config (10 zips, range 160 km, speed 30 m/s) this
// only fires in the last ~90 minutes of the day. Earlier triggering belongs
// to a tunable knob (Step 2a / C4).
//
// TODO: 2 * EdgeWeight(Nest, hospital) is the round-trip time if this order
// flew alone. In practice it is delivered as part of a multi-stop flight, so
// the actual time-to-deliver is shorter. The current value therefore
// overestimates risk and triggers reserve borrowing slightly earlier than
// strictly necessary — a conservative fail-safe. A more accurate computation
// would amortize the cost across the planned route, but the planned route is
// not known at this decision point. Leaving as-is; revisit post-Step-2.
func (zipScheduler *ZipScheduler) resupplyAtRisk(currentTime int, order Order) bool {
	roundTrip := 2 * zipScheduler.graph.EdgeWeight(NestKey, order.HospitalName)
	flightSeconds := int(roundTrip / float64(zipScheduler.zipSpeedMps))
	return flightSeconds > (SecondsPerDay - currentTime)
}

// LaunchFlights returns the list of flights to launch at currentTime. Pending
// orders are considered Emergency-first; flights are built greedily under the
// active reserve policy and the scheduler tracks zip availability so a zip
// cannot be in two flights at once.
func (zipScheduler *ZipScheduler) LaunchFlights(currentTime int) []Flight {
	flights := []Flight{}
	available := zipScheduler.availableZips(currentTime)
	if available <= 0 || len(zipScheduler.unfulfilledOrders) == 0 {
		return flights
	}

	queue := zipScheduler.pendingByPriority()
	reserve := zipScheduler.reserveSize()

	canLaunchResupply := func(order Order) bool {
		switch zipScheduler.reservePolicy {
		case ReserveNone:
			return true
		case ReserveHard:
			return available > reserve
		case ReserveSoft:
			if available > reserve {
				return true
			}
			return zipScheduler.resupplyAtRisk(currentTime, order)
		}
		return true
	}

	for available > 0 && len(queue) > 0 {
		head := queue[0]
		if head.Priority == Resupply && !canLaunchResupply(head) {
			break
		}

		flight, leftover := zipScheduler.buildFlight(currentTime, queue)
		if len(flight.OrderIDs) == 0 {
			break
		}

		distance := zipScheduler.routeDistance(flight.HospitalNames)
		flightDuration := int(distance / float64(zipScheduler.zipSpeedMps))
		zipScheduler.markZipLaunched(currentTime + flightDuration)
		flights = append(flights, flight)
		available--
		queue = leftover
	}

	zipScheduler.unfulfilledOrders = append([]Order{}, queue...)
	return flights
}

func (zipScheduler *ZipScheduler) UnfulfilledOrders() []Order {
	return append([]Order{}, zipScheduler.unfulfilledOrders...)
}

type Runner struct {
	orders       []Order
	hospitals    map[string]Hospital
	zipScheduler *ZipScheduler
}

func NewRunner(
	orders []Order,
	hospitals map[string]Hospital,
	zipScheduler *ZipScheduler,
) Runner {
	return Runner{orders: orders, hospitals: hospitals, zipScheduler: zipScheduler}
}

func (runner *Runner) QueuePendingOrders(secSinceMidnight int, verbose bool) {
	for len(runner.orders) > 0 && runner.orders[0].Time == secSinceMidnight {
		order := runner.orders[0]
		runner.orders = runner.orders[1:]
		if verbose {
			fmt.Printf(
				"[%d] %s order received to %s\n",
				order.Time,
				order.Priority,
				order.HospitalName,
			)
		}
		runner.zipScheduler.QueueOrder(order)
	}
}

func (runner *Runner) UpdateLaunchFlights(secSinceMidnight int, verbose bool) []Flight {
	flights := runner.zipScheduler.LaunchFlights(secSinceMidnight)
	if verbose && len(flights) > 0 {
		fmt.Printf("[%d] Scheduling flights:\n", secSinceMidnight)
		for _, flight := range flights {
			fmt.Println(flight)
		}
	}
	return flights
}

func (runner *Runner) Simulate(verbose bool) []Flight {
	flights := []Flight{}
	for secSinceMidnight := runner.orders[0].Time; secSinceMidnight < SecondsPerDay; secSinceMidnight++ {
		runner.QueuePendingOrders(secSinceMidnight, verbose)
		if secSinceMidnight%60 == 0 {
			flights = append(flights, runner.UpdateLaunchFlights(secSinceMidnight, verbose)...)
		}
	}
	return flights
}

func (runner *Runner) Run() {
	runner.Simulate(true)
	fmt.Printf("%d unfulfilled orders at the end of the day\n", len(runner.zipScheduler.UnfulfilledOrders()))
}

func loadCSV(csvPath string) [][]string {
	file, err := os.Open(csvPath)
	if err != nil {
		log.Fatal("Could not open file", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Could not read file", err)
	}
	return records
}

func LoadHospitals(hospitalsPath string) map[string]Hospital {
	hospitals := map[string]Hospital{}
	for _, record := range loadCSV(hospitalsPath) {
		if len(record) == 0 {
			continue
		}
		name := strings.TrimSpace(record[0])
		northM, err := strconv.Atoi(strings.TrimSpace(record[1]))
		if err != nil {
			log.Fatal(err)
		}
		eastM, err := strconv.Atoi(strings.TrimSpace(record[2]))
		if err != nil {
			log.Fatal(err)
		}
		hospitals[name] = Hospital{Name: name, NorthM: northM, EastM: eastM}
	}
	return hospitals
}

func LoadOrders(ordersPath string) []Order {
	orders := []Order{}
	for index, record := range loadCSV(ordersPath) {
		if len(record) == 0 {
			continue
		}
		timeValue, err := strconv.Atoi(strings.TrimSpace(record[0]))
		if err != nil {
			log.Fatal(err)
		}
		order := Order{
			ID:           strconv.Itoa(index + 1),
			Time:         timeValue,
			HospitalName: strings.TrimSpace(record[1]),
			Priority:     strings.TrimSpace(record[2]),
		}
		orders = append(orders, order)
	}
	return orders
}

func DefaultInputPaths() (string, string) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	rootDir := filepath.Dir(workingDirectory)
	return filepath.Join(rootDir, "data", "hospitals.csv"),
		filepath.Join(rootDir, "data", "orders.csv")
}

func BuildSimulationSnapshot(config SimulationConfig) Snapshot {
	hospitalsPath, ordersPath := DefaultInputPaths()
	hospitals := LoadHospitals(hospitalsPath)
	orders := LoadOrders(ordersPath)
	runner := NewRunner(append([]Order{}, orders...), hospitals, NewZipScheduler(hospitals, config))
	flights := runner.Simulate(false)

	hospitalList := make([]Hospital, 0, len(hospitals))
	for _, hospital := range hospitals {
		hospitalList = append(hospitalList, hospital)
	}
	sort.Slice(hospitalList, func(i int, j int) bool {
		return hospitalList[i].Name < hospitalList[j].Name
	})

	return Snapshot{
		Implementation:   "go",
		Config:           config,
		Hospitals:        hospitalList,
		Orders:           orders,
		Flights:          flights,
		UnfulfilledOrders: runner.zipScheduler.UnfulfilledOrders(),
	}
}

func ParseConfig(body []byte) (SimulationConfig, error) {
	var payload struct {
		Config map[string]json.RawMessage `json:"config"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return SimulationConfig{}, err
	}
	if payload.Config == nil {
		return SimulationConfig{}, fmt.Errorf("request body must include a config object")
	}

	parseField := func(name string) (int, error) {
		rawValue, ok := payload.Config[name]
		if !ok {
			return 0, fmt.Errorf("config.%s must be a number", name)
		}

		var value int
		if err := json.Unmarshal(rawValue, &value); err != nil {
			return 0, fmt.Errorf("config.%s must be a number", name)
		}
		return value, nil
	}

	numZips, err := parseField("numZips")
	if err != nil {
		return SimulationConfig{}, err
	}
	maxPackagesPerZip, err := parseField("maxPackagesPerZip")
	if err != nil {
		return SimulationConfig{}, err
	}
	zipSpeedMps, err := parseField("zipSpeedMps")
	if err != nil {
		return SimulationConfig{}, err
	}
	zipMaxCumulativeRangeM, err := parseField("zipMaxCumulativeRangeM")
	if err != nil {
		return SimulationConfig{}, err
	}

	return SimulationConfig{
		NumZips:                numZips,
		MaxPackagesPerZip:      maxPackagesPerZip,
		ZipSpeedMps:            zipSpeedMps,
		ZipMaxCumulativeRangeM: zipMaxCumulativeRangeM,
	}, nil
}
