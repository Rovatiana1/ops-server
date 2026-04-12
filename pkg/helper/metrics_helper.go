package helper

import (
	"strconv"
	"strings"
)

type MetricsSnapshot struct {
	Goroutines int

	Memory struct {
		AllocMB    float64
		HeapMB     float64
		ResidentMB float64
	}

	GC struct {
		Count float64
		Sum   float64
	}

	CPU float64
	FDS int
}

func ParsePrometheusMetrics(input string) MetricsSnapshot {
	lines := strings.Split(input, "\n")

	var snap MetricsSnapshot

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := parts[0]
		valStr := parts[len(parts)-1]

		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			continue
		}

		switch key {

		case "go_goroutines":
			snap.Goroutines = int(val)

		case "go_memstats_alloc_bytes":
			snap.Memory.AllocMB = val / 1024 / 1024

		case "go_memstats_heap_alloc_bytes":
			snap.Memory.HeapMB = val / 1024 / 1024

		case "process_resident_memory_bytes":
			snap.Memory.ResidentMB = val / 1024 / 1024

		case "go_gc_duration_seconds_count":
			snap.GC.Count = val

		case "go_gc_duration_seconds_sum":
			snap.GC.Sum = val

		case "process_cpu_seconds_total":
			snap.CPU = val

		case "process_open_fds":
			snap.FDS = int(val)
		}
	}

	return snap
}

type HealthStatus struct {
	Status string   `json:"status"`
	Score  int      `json:"score"`
	Issues []string `json:"issues"`
}

func EvaluateHealth(m MetricsSnapshot) HealthStatus {
	score := 100
	var issues []string

	if m.Goroutines > 200 {
		score -= 30
		issues = append(issues, "High goroutine count")
	}

	if m.Memory.ResidentMB > 200 {
		score -= 25
		issues = append(issues, "High memory usage")
	}

	if m.FDS > 500 {
		score -= 25
		issues = append(issues, "Too many file descriptors")
	}

	if m.GC.Count > 100 && m.GC.Sum/m.GC.Count > 0.01 {
		score -= 20
		issues = append(issues, "GC overhead high")
	}

	status := "green"
	if score < 70 {
		status = "yellow"
	}
	if score < 40 {
		status = "red"
	}

	return HealthStatus{
		Status: status,
		Score:  score,
		Issues: issues,
	}
}
