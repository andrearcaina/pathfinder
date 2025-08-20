package runner

import (
	"github.com/andrearcaina/pathfinder/internal/metrics"
	"github.com/andrearcaina/pathfinder/internal/ui"
)

func Run(flags metrics.Flags) error {
	report, err := metrics.Analyze(flags)
	if err != nil {
		return err
	}

	ui.PrintReport(report)

	return nil
}
