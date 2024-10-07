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
	interval     = 5 * time.Second
	maxLoadAvg   = 30.0
	maxMemUsage  = 0.8
	maxDiskUsage = 0.9
	maxNetUsage  = 0.9
	maxErrors    = 3
)

// Структура для хранения статистики сервера
type ServerStats struct {
	LoadAvg   float64
	TotalMem  int64
	UsedMem   int64
	TotalDisk int64
	UsedDisk  int64
	TotalNet  int64
	UsedNet   int64
}

func main() {
	errorCount := 0

	for {
		stats, err := fetchServerStats(url)
		if err != nil {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistics after multiple attempts.")
				return
			}
			fmt.Printf("Error fetching server stats: %v. Retrying...\n", err)
			time.Sleep(interval)
			continue
		}

		// Сброс счётчика ошибок при успешном запросе
		errorCount = 0

		// Анализ статистики сервера
		analyzeServerStats(stats)

		time.Sleep(interval)
	}
}

// Функция для получения статистики с сервера
func fetchServerStats(url string) (ServerStats, error) {
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		if err == nil {
			err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return ServerStats{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ServerStats{}, fmt.Errorf("error reading response body: %w", err)
	}

	data := strings.Split(string(body), ",")
	if len(data) != 7 {
		return ServerStats{}, fmt.Errorf("invalid data format")
	}

	// Парсим данные из строки в структуру
	stats, err := parseServerStats(data)
	if err != nil {
		return ServerStats{}, fmt.Errorf("error parsing server stats: %w", err)
	}

	return stats, nil
}

// Функция для парсинга данных в структуру ServerStats
func parseServerStats(data []string) (ServerStats, error) {
	loadAvg, err := strconv.ParseFloat(data[0], 64)
	if err != nil {
		return ServerStats{}, fmt.Errorf("error parsing load average: %w", err)
	}
	totalMem, err := strconv.ParseInt(data[1], 10, 64)
	if err != nil {
		return ServerStats{}, fmt.Errorf("error parsing total memory: %w", err)
	}
	usedMem, err := strconv.ParseInt(data[2], 10, 64)
	if err != nil {
		return ServerStats{}, fmt.Errorf("error parsing used memory: %w", err)
	}
	totalDisk, err := strconv.ParseInt(data[3], 10, 64)
	if err != nil {
		return ServerStats{}, fmt.Errorf("error parsing total disk: %w", err)
	}
	usedDisk, err := strconv.ParseInt(data[4], 10, 64)
	if err != nil {
		return ServerStats{}, fmt.Errorf("error parsing used disk: %w", err)
	}
	totalNet, err := strconv.ParseInt(data[5], 10, 64)
	if err != nil {
		return ServerStats{}, fmt.Errorf("error parsing total network bandwidth: %w", err)
	}
	usedNet, err := strconv.ParseInt(data[6], 10, 64)
	if err != nil {
		return ServerStats{}, fmt.Errorf("error parsing used network bandwidth: %w", err)
	}

	return ServerStats{
		LoadAvg:   loadAvg,
		TotalMem:  totalMem,
		UsedMem:   usedMem,
		TotalDisk: totalDisk,
		UsedDisk:  usedDisk,
		TotalNet:  totalNet,
		UsedNet:   usedNet,
	}, nil
}

// Функция для анализа и вывода статистики
func analyzeServerStats(stats ServerStats) {
	// Проверка нагрузки процессора
	if stats.LoadAvg > maxLoadAvg {
		// Округляем значение Load Average до целого числа
		fmt.Printf("Load Average is too high: %d\n", int(stats.LoadAvg))
	}

	// Проверка использования памяти
	memUsage := float64(stats.UsedMem) / float64(stats.TotalMem)
	if memUsage > maxMemUsage {
		fmt.Printf("Memory usage too high: %.0f%%\n", memUsage*100)
	}

	// Проверка использования дискового пространства
	freeDiskSpaceMB := (stats.TotalDisk - stats.UsedDisk) / (1024 * 1024)
	diskUsage := float64(stats.UsedDisk) / float64(stats.TotalDisk)
	if diskUsage > maxDiskUsage {
		fmt.Printf("Free disk space is too low: %d Mb left\n", freeDiskSpaceMB)
	}

	// Проверка использования сети
	netUsage := float64(stats.UsedNet) / float64(stats.TotalNet)
	if netUsage > maxNetUsage {
		// Исправляем расчет свободной пропускной способности сети для правильного отображения
		freeNetMb := float64(stats.TotalNet-stats.UsedNet) / (1000 * 1000)
		fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", freeNetMb)
	}
}
