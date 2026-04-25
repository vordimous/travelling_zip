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
	numZips                int
	maxPackagesPerZip      int
	zipSpeedMps            int
	zipMaxCumulativeRangeM int
	unfulfilledOrders      []Order
}

func NewZipScheduler(
	hospitals map[string]Hospital,
	config SimulationConfig,
) *ZipScheduler {
	return &ZipScheduler{
		hospitals:              hospitals,
		numZips:                config.NumZips,
		maxPackagesPerZip:      config.MaxPackagesPerZip,
		zipSpeedMps:            config.ZipSpeedMps,
		zipMaxCumulativeRangeM: config.ZipMaxCumulativeRangeM,
		unfulfilledOrders:      []Order{},
	}
}

func (zipScheduler *ZipScheduler) QueueOrder(order Order) {
	zipScheduler.unfulfilledOrders = append(zipScheduler.unfulfilledOrders, order)
}

func (zipScheduler *ZipScheduler) LaunchFlights(currentTime int) []Flight {
	_ = currentTime
	return []Flight{}
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
