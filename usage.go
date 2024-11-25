package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type cpuTime struct {
	User uint64 `json:"user"`
	Nice uint64 `json:"nice"`
	Sys  uint64 `json:"sys"`
	Idle uint64 `json:"idle"`
}

type memoryStats struct {
	TotalMemory uint64  `json:"total_memory_mb"`
	FreeMemory  uint64  `json:"free_memory_mb"`
	UsedMemory  uint64  `json:"used_memory_mb"`
	PercentUsed float64 `json:"percent_used"`
}

type systemStats struct {
	CPUCores int         `json:"cpu_cores"`
	CPUUsage []float64   `json:"cpu_usage"`
	Memory   memoryStats `json:"memory"`
}

func getSystemStats(c *gin.Context) {
	cpuCores := runtime.NumCPU()
	cpuUsage := getCPUUsage()
	memory := getMemoryUsage()

	stats := systemStats{
		CPUCores: cpuCores,
		CPUUsage: cpuUsage,
		Memory:   memory,
	}

	c.JSON(http.StatusOK, stats)
}

func getCPUTimes() ([]cpuTime, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cpuTimes []cpuTime
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		if strings.HasPrefix(fields[0], "cpu") {
			var cpuTime cpuTime
			cpuTime.User, _ = strconv.ParseUint(fields[1], 10, 64)
			cpuTime.Nice, _ = strconv.ParseUint(fields[2], 10, 64)
			cpuTime.Sys, _ = strconv.ParseUint(fields[3], 10, 64)
			cpuTime.Idle, _ = strconv.ParseUint(fields[4], 10, 64)
			cpuTimes = append(cpuTimes, cpuTime)
		}
	}

	return cpuTimes, nil
}

func getCPUUsage() []float64 {
	cpuTimes1, err := getCPUTimes()
	if err != nil {
		return nil
	}
	time.Sleep(time.Second) // Wait for a second to measure again
	cpuTimes2, err := getCPUTimes()
	if err != nil {
		return nil
	}

	var cpuUsage []float64
	for i := 0; i < len(cpuTimes1); i++ {
		totalBusy1 := cpuTimes1[i].User + cpuTimes1[i].Nice + cpuTimes1[i].Sys
		totalBusy2 := cpuTimes2[i].User + cpuTimes2[i].Nice + cpuTimes2[i].Sys
		totalIdle1 := cpuTimes1[i].Idle
		totalIdle2 := cpuTimes2[i].Idle

		totalTime1 := totalBusy1 + totalIdle1
		totalTime2 := totalBusy2 + totalIdle2

		// Calculate CPU usage as a percentage
		cpuUsage = append(cpuUsage, float64(totalBusy2-totalBusy1)/float64(totalTime2-totalTime1)*100)
	}
	return cpuUsage
}

func getMemoryUsage() memoryStats {
	memInfo, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		fmt.Println("Ошибка получения информации о памяти:", err)
		return memoryStats{}
	}

	var totalMemory, freeMemory, availableMemory, buffers, cachedMemory uint64

	// Парсинг данных из meminfo
	for _, line := range strings.Split(string(memInfo), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fmt.Sscanf(line, "MemTotal: %d kB", &totalMemory)
		} else if strings.HasPrefix(line, "MemFree:") {
			fmt.Sscanf(line, "MemFree: %d kB", &freeMemory)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			fmt.Sscanf(line, "MemAvailable: %d kB", &availableMemory)
		} else if strings.HasPrefix(line, "Buffers:") {
			fmt.Sscanf(line, "Buffers: %d kB", &buffers)
		} else if strings.HasPrefix(line, "Cached:") {
			fmt.Sscanf(line, "Cached: %d kB", &cachedMemory)
		}
	}

	// Используемая память = Общая память - Свободная память - Буферы - Кэшированная память
	usedMemory := totalMemory - freeMemory - buffers - cachedMemory
	percentUsed := (float64(usedMemory) / float64(totalMemory)) * 100

	return memoryStats{
		TotalMemory: KbToMb(totalMemory),
		FreeMemory:  KbToMb(freeMemory),
		UsedMemory:  KbToMb(usedMemory),
		PercentUsed: percentUsed,
	}
}

func KbToMb(kb uint64) uint64 {
	return kb / 1024
}
