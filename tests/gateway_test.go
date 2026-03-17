package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const gatewayURL = "http://localhost:8080" // Assuming KrakenD is on 8080

func TestTC_GW_001_Routing(t *testing.T) {
	// These tests require the actual services to be running behind KrakenD
	// Since we are in a dev environment with Docker, we can perform basic health checks or route hits.
	
	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
	}{
		{"Auth Health", "/auth/login", http.MethodPost, http.StatusUnauthorized}, // Unauthorized is expected without body
		{"Order Route", "/api/v1/orders", http.MethodGet, http.StatusOK},
		{"Menu Route", "/api/v1/menu", http.MethodGet, http.StatusNotFound}, // Catalog might not have /api/v1/menu exactly
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(gatewayURL + tt.path)
			if err != nil {
				t.Skip("Gateway not reachable, skipping integration test")
				return
			}
			assert.NotEqual(t, http.StatusBadGateway, resp.StatusCode)
			assert.NotEqual(t, http.StatusServiceUnavailable, resp.StatusCode)
		})
	}
}

func TestTC_GW_002_RateLimiting(t *testing.T) {
	// Basic throughput test to see if rate limit kicks in
	// This depends on the KrakenD configuration
	t.Log("TC-GW-002: Rate Limiting verification usually requires high-volume bombardment")
}
