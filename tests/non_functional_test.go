package integration_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTC_SEC_001_SQLInjection(t *testing.T) {
	// Simple test to ensure malicious inputs don't crash the service
	maliciousInputs := []string{"' OR 1=1 --", "'; DROP TABLE orders; --"}
	
	for _, input := range maliciousInputs {
		resp, err := http.Get("http://localhost:8080/api/v1/orders/" + input)
		if err != nil {
			continue // Service down
		}
		// We expect 400 or 404, but NOT 500
		assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestTC_PERF_001_Throughput(t *testing.T) {
	// Baseline throughput check
	start := time.Now()
	count := 0
	for i := 0; i < 50; i++ {
		resp, err := http.Get("http://localhost:8080/api/v1/orders")
		if err == nil && resp.StatusCode == http.StatusOK {
			count++
		}
	}
	duration := time.Since(start)
	t.Logf("Completed %d requests in %v (%.2f RPS)", count, duration, float64(count)/duration.Seconds())
}
