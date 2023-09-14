package agent

import (
	"testing"
)

func TestConfig_setDefault(t *testing.T) {
	config := Config{}
	config.setDefault()
	t.Run("Set default values", func(t *testing.T) {
		if config.Port != defPort {
			t.Errorf("default port error. Want: %d, got: %d", defPort, config.Port)
		}
		if config.PollInterval != defPoolInterval {
			t.Errorf("default pollInterval error. Want: %d, got: %d", defPoolInterval, config.PollInterval)
		}
		if config.ReportInterval != defReportInterval {
			t.Errorf("default reportInterval error. Want: %d, got: %d", defReportInterval, config.ReportInterval)
		}
		if !config.GzipCompress {
			t.Error("default gzipCompress error. Want: true, got: false")
		}
		if config.RateLimit != defRateLimit {
			t.Errorf("default rateLimit error. Want: %d, got: %d", defRateLimit, config.RateLimit)
		}
	})
}
