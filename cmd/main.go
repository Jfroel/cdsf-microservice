package main

import (
	"flag"
	"log"
	"math"
	"os"
	"runtime"

	"github.com/Jfroel/cdsf-microservice/services"
)

func levelToSize(levels int) int {
	return int(math.Pow(2.0, float64(levels))) - 1
}

type server interface {
	Run() error
}

func main() {
	// Define the flags to specify port numbers and addresses
	var (
		proxyPort      = flag.Int("proxyport", 9090, "proxy server port")
		filterPort     = flag.Int("filterport", 9091, "filter service port")
		filterAddr     = flag.String("filteraddr", "filter:9091", "filter service address")
		filterCapacity = flag.Int("filter_capacity", levelToSize(18), "maximum number of items allowed in the filter service")
		filterType     = flag.String("filter_type", "subtree", "locking style for the filter: coarseRW or subtree")
		cpus           = flag.Int("cpus", 8, "number of cpus the filter can use")
	)

	// Parse the flags
	flag.Parse()

	var srv server
	var cmd = os.Args[1]

	runtime.GOMAXPROCS(*cpus)

	// Switch statement to create the correct service based on the command
	switch cmd {
	case "proxy":
		// Create a new frontend service with the specified ports and addresses
		srv = services.NewProxy(
			*proxyPort,
			*filterAddr,
			"1",
		)
	case "filter":
		srv = services.NewFilter(
			"filter",
			*filterPort,
			*filterType,
			*filterCapacity,
		)
	default:
		// If an unknown command is provided, log an error and exit
		log.Fatalf("unknown cmd: %s", cmd)
	}

	// Start the server and log any errors that occur
	if err := srv.Run(); err != nil {
		log.Fatalf("run %s error: %v", cmd, err)
	}
}
