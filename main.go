package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	url          = "http://srv.msk01.gigacorp.local/_stats"
	interval     = 30 * time.Second
	maxLoadAvg   = 30.0
	maxMemUsage  = 0.8
	maxDiskUsage = 0.9
	maxNetUsage  = 0.9
	maxErrors    = 3
)

func main() {
	errorCount := 0

	for {
		resp, err := http.Get(url)
		if err != nil || resp.StatusCode != http.StatusOK {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic.")
				return
			}
			time.Sleep(interval)
			continue
		}
		errorCount = 0

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Println("Error reading response body:", err)
			time.Sleep(interval)
			continue
		}

		data := strings.Split(string(body), ",")
		if len(data) != 7 {
			fmt.Println("Invalid data format.")
			time.Sleep(interval)
			continue
		}

		loadAvg, err := strconv.ParseFloat(data[0], 64)
		if err != nil {
			fmt.Println("Error parsing load average:", err)
			time.Sleep(interval)
			continue
		}
		totalMem, err := strconv.ParseInt(data[1], 10, 64)
		if err != nil {
			fmt.Println("Error parsing total memory:", err)
			time.Sleep(interval)
			continue
		}
		usedMem, err := strconv.ParseInt(data[2], 10, 64)
		if err != nil {
			fmt.Println("Error parsing used memory:", err)
			time.Sleep(interval)
			continue
		}
		totalDisk, err := strconv.ParseInt(data[3], 10, 64)
		if err != nil {
			fmt.Println("Error parsing total disk:", err)
			time.Sleep(interval)
			continue
		}
		usedDisk, err := strconv.ParseInt(data[4], 10, 64)
		if err != nil {
			fmt.Println("Error parsing used disk:", err)
			time.Sleep(interval)
			continue
		}
		totalNet, err := strconv.ParseInt(data[5], 10, 64)
		if err != nil {
			fmt.Println("Error parsing total network bandwidth:", err)
			time.Sleep(interval)
			continue
		}
		usedNet, err := strconv.ParseInt(data[6], 10, 64)
		if err != nil {
			fmt.Println("Error parsing used network bandwidth:", err)
			time.Sleep(interval)
			continue
		}

		if loadAvg > maxLoadAvg {
			fmt.Printf("Load Average is too high: %.2f\n", loadAvg)
		}

		memUsage := float64(usedMem) / float64(totalMem)
		if memUsage > maxMemUsage {
			fmt.Printf("Memory usage too high: %.2f%%\n", memUsage*100)
		}

		freeDiskSpaceMB := (totalDisk - usedDisk) / (1024 * 1024)
		diskUsage := float64(usedDisk) / float64(totalDisk)
		if diskUsage > maxDiskUsage {
			fmt.Printf("Free disk space is too low: %d Mb left\n", freeDiskSpaceMB)
		}

		netUsage := float64(usedNet) / float64(totalNet)
		if netUsage > maxNetUsage {
			freeNetMb := float64(totalNet-usedNet) / (1024 * 1024) * 8 // Convert to Mbit/s
			fmt.Printf("Network bandwidth usage high: %.2f Mbit/s available\n", freeNetMb)
		}

		time.Sleep(interval)
	}
}
