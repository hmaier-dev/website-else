package events_test

import (
	"testing"
	"codeberg.org/schnelles_geraet/website-else/backend/internal/events"
)

func TestFetchCoordinates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Use a known location
		coord, err := events.FetchCoordinates("Moritzbastei", "Leipzig")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		t.Logf("%s,%s", coord.Lat, coord.Lon)
		if coord.Lat == "" || coord.Lon == "" {
			t.Error("Expected non-empty coordinates")
		}
	})

	t.Run("No Results", func(t *testing.T) {
		// Use a non-existent location
		_, err := events.FetchCoordinates("Unknown Place", "Nowhere")
		if err == nil || err.Error() != "no results found" {
			t.Error("Expected 'no results found' error")
		}
	})
}

