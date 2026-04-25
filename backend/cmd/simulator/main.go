package main

import "travelingzip/core"

func main() {
	hospitalsPath, ordersPath := core.DefaultInputPaths()
	hospitals := core.LoadHospitals(hospitalsPath)
	orders := core.LoadOrders(ordersPath)
	runner := core.NewRunner(orders, hospitals, core.NewZipScheduler(hospitals, core.DefaultConfig()))
	runner.Run()
}
