package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-units"
)

var (
	stats types.StatsJSON
)

func main() {
	// -- Flags --

	// Obtain the interval from the command line
	intervalString := flag.String("interval", "5s", "Interval between measurements")
	// Obtain the output file from the command line
	outputFile := flag.String("output", "measurements.csv", "Output file")

	// Parse the command line arguments
	flag.Parse()

	// -- Parsing --
	// Parse the interval string into a time.Duration
	interval, err := time.ParseDuration(*intervalString)
	if err != nil {
		log.Panicf("Failed to parse interval: %v\n", err)
	}

	// -- Docker Client --
	// Create a docker client
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Panicf("Failed to create docker client: %v\n", err)
	}

	// List all containers
	containers, err := docker.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Panicf("Failed to list containers: %v\n", err)
	}

	// Find the container with the name "kwok-kwok-kube-scheduler"
	var containerID string
	for _, container := range containers {
		if container.Names[0] == "/kwok-kwok-kube-scheduler" {
			containerID = container.ID
			break
		}
	}

	if containerID == "" {
		log.Panicf("Failed to find container with name \"kwok-kwok-kube-scheduler\"")
	}

	// -- Measurements File --
	// Open a file to write the measurements to
	file, err := os.Create(*outputFile)
	if err != nil {
		log.Panicf("Failed to open file: %v\n", err)
	}
	defer file.Close()

	// Write the header to the file
	_, err = file.WriteString("time;cpu_percentage;memory\n")
	if err != nil {
		log.Panicf("Failed to write header: %v\n", err)
	}

	// -- Interrupt --

	// Create a channel that traps the interrupt signal
	// and notifies the program to exit
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// -- Measurements --
	timeCounter := 0
	start := time.Now()

	go collectContainerStats(interrupt, docker, containerID)

	// Query the docker socket for container stats every interval
	for {
		select {
		case <-time.After(interval):
			stats := stats

			cpuPercent, memory := extractFromStats(stats)
			if err != nil {
				log.Printf("Failed to get container stats: %v\n", err)
				return
			}

			timeCounter += int(time.Since(start).Seconds())

			// Print to stdout
			log.Printf("Time: %v - CPU: %v - Memory: %v\n", timeCounter, cpuPercent, memory)

			// Write the stats to the file
			_, err = file.WriteString(fmt.Sprintf("%v;%v;%v\n", timeCounter, cpuPercent, memory))
			if err != nil {
				log.Panicf("Failed to write stats: %v\n", err)
			}

			start = time.Now()

		case <-interrupt:
			return
		}
	}
}

func extractFromStats(stats types.StatsJSON) (float64, string) {
	memory := units.BytesSize(float64(stats.MemoryStats.Usage))

	if len(memory) > 3 {
		memory = memory[:len(memory)-3]
	}

	var (
		cpuPercent = 0.0

		cpuDelta = float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)

		systemDelta = float64(stats.CPUStats.SystemUsage) - float64(stats.PreCPUStats.SystemUsage)
		onlineCPUs  = float64(stats.CPUStats.OnlineCPUs)
	)

	if onlineCPUs == 0.0 {
		onlineCPUs = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
	}

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
	}

	return cpuPercent, memory
}

func collectContainerStats(interrupt chan os.Signal, docker *client.Client, containerID string) {
	containerStats, err := docker.ContainerStats(context.Background(), containerID, true)
	if err != nil {
		log.Panicf("Failed to get container stats: %v\n", err)
	}
	defer containerStats.Body.Close()

	decoder := json.NewDecoder(containerStats.Body)

	for {
		select {
		case <-interrupt:
			return
		default:
			if err := decoder.Decode(&stats); err == io.EOF {
				go collectContainerStats(interrupt, docker, containerID)
				return
			} else if err != nil {
				log.Panicf("Failed to decode container stats: %v\n", err)
			}
		}
	}
}
