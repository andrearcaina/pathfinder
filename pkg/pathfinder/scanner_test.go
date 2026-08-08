package pathfinder

import (
	"testing"
	"time"
)

func TestBuildCodebaseReportPerformanceMetrics(t *testing.T) {
	tests := []struct {
		name               string
		dependencies       bool
		dependencyWorkers  int
		resultConsumers    int
		pipelineGoroutines int
		totalWorkers       int
	}{
		{
			name:               "file scanning only",
			dependencyWorkers:  0,
			resultConsumers:    1,
			pipelineGoroutines: 5,
			totalWorkers:       3,
		},
		{
			name:               "with dependency scanning",
			dependencies:       true,
			dependencyWorkers:  3,
			resultConsumers:    2,
			pipelineGoroutines: 9,
			totalWorkers:       6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := Config{
				WorkerFlag:     3,
				DependencyFlag: tt.dependencies,
				ThroughputFlag: true,
			}
			aggregation := newScanAggregation()
			aggregation.codebaseStats.TotalFiles = 10

			report := buildCodebaseReport(flags, time.Now().Add(-time.Second), nil, aggregation)
			metrics := report.PerformanceMetrics

			if metrics.FileWorkers != 3 {
				t.Fatalf("FileWorkers = %d, want 3", metrics.FileWorkers)
			}
			if metrics.DependencyWorkers != tt.dependencyWorkers {
				t.Fatalf("DependencyWorkers = %d, want %d", metrics.DependencyWorkers, tt.dependencyWorkers)
			}
			if metrics.TotalWorkers != tt.totalWorkers {
				t.Fatalf("TotalWorkers = %d, want %d", metrics.TotalWorkers, tt.totalWorkers)
			}
			if metrics.ResultConsumers != tt.resultConsumers {
				t.Fatalf("ResultConsumers = %d, want %d", metrics.ResultConsumers, tt.resultConsumers)
			}
			if metrics.PipelineGoroutines != tt.pipelineGoroutines {
				t.Fatalf("PipelineGoroutines = %d, want %d", metrics.PipelineGoroutines, tt.pipelineGoroutines)
			}
			if metrics.LogicalCPUs < 1 || metrics.GOMAXPROCS < 1 || metrics.OSThreadsCreated < 1 {
				t.Fatalf("invalid runtime metrics: %+v", metrics)
			}
		})
	}
}
