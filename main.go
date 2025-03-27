package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Data Structure Definition
type VisorState struct {
	InitialHeight   float64  `json:"initial_height"`
	Height          float64  `json:"height"`
	ScheduledFreeze *float64 `json:"scheduled_freeze_height"`
	ConsensusTime   string   `json:"consensus_time"`
}

// Prometheus Metric definition
var (
	initialHeight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "visor_initial_height",
		Help: "Initial height of the visor state",
	})
	height = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "visor_current_height",
		Help: "Current height of the visor state",
	})
	scheduledFreezeHeight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "visor_scheduled_freeze_height",
		Help: "Scheduled freeze height of the visor state (null if not scheduled)",
	})
	consensusTimestamp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "visor_consensus_timestamp_unix",
		Help: "Consensus timestamp in Unix format",
	})
	mutex sync.Mutex
)

// normalizeTime normalizes the time string to the RFC3339Nano format.
// 1. If the number of decimal places is 10 or more, it is reduced to 9.
// 2. If there is no time zone information (Z or ±HH:MM) at the end of the string, “Z” is added.
func normalizeTime(timeStr string) string {
	// Regular expression processing that reduces the number of digits after the decimal point to 9 if it is 10 or more
	reFraction := regexp.MustCompile(`\.(\d{10,})`)
	timeStr = reFraction.ReplaceAllStringFunc(timeStr, func(match string) string {
		// match is “.”-containing numbers. Cut to a maximum of 10 characters (1+9 digits including a dot).
		if len(match) > 10 {
			return match[:10]
		}
		return match
	})

	// Checks if there is time zone information at the end (Z or +HH:MM or -HH:MM)
	reTZ := regexp.MustCompile(`(Z|[+-]\d{2}:\d{2})$`)
	if !reTZ.MatchString(timeStr) {
		timeStr += "Z"
	}
	return timeStr
}

func updateMetrics(filePath string) {
	mutex.Lock()
	defer mutex.Unlock()

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading JSON file: %v", err)
		return
	}

	var state VisorState
	if err := json.Unmarshal(data, &state); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		return
	}

	// Metric update
	initialHeight.Set(state.InitialHeight)
	height.Set(state.Height)
	if state.ScheduledFreeze != nil {
		scheduledFreezeHeight.Set(*state.ScheduledFreeze)
	} else {
		scheduledFreezeHeight.Set(0)
	}

	// consensus_time Normalization and parsing
	normalizedTime := normalizeTime(state.ConsensusTime)
	parsedTime, err := time.Parse(time.RFC3339Nano, normalizedTime)
	if err != nil {
		log.Printf("Error parsing consensus time: %v (normalized: %s)", err, normalizedTime)
	} else {
		consensusTimestamp.Set(float64(parsedTime.Unix()))
	}

	log.Println("Metrics updated successfully.")
}

func main() {
	// Receives the file path and port as arguments
	filePath := flag.String("file", "visor_abci_state.json", "Path to the visor JSON file")
	port := flag.Int("port", 8080, "Port number for the exporter HTTP server")
	flag.Parse()

	prometheus.MustRegister(initialHeight)
	prometheus.MustRegister(height)
	prometheus.MustRegister(scheduledFreezeHeight)
	prometheus.MustRegister(consensusTimestamp)

	// Read and update JSON file every 10 seconds
	go func() {
		for {
			updateMetrics(*filePath)
			time.Sleep(10 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())

	serverAddr := fmt.Sprintf(":%d", *port)
	log.Printf("Prometheus Exporter is running on %s, reading file: %s\n", serverAddr, *filePath)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
