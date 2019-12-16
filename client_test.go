package darksky_test

import (
	"flag"
	"testing"

	darksky "github.com/sophiaehlen/darksky-client"
)

var (
	apiKey string
)

func init() {
	flag.StringVar(&apiKey, "key", "", "Your TEST secret key for the Dark Sky API. If present, integration tests will be run using this key.")
}

func TestClient_Forecast(t *testing.T) {
	if apiKey == "" {
		t.Skip("No API key provided")
	}

	type checkFn func(*testing.T, *darksky.Forecast, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoErr := func() checkFn {
		return func(t *testing.T, fc *darksky.Forecast, err error) {
			if err != nil {
				t.Fatalf("err = %v; want nil", err)
			}
		}
	}
	hasLatitude := func(lat float64) checkFn {
		return func(t *testing.T, fc *darksky.Forecast, err error) {
			if fc.Latitude != lat {
				t.Errorf("Latitude = %f; want %f", fc.Latitude, lat)
			}
		}
	}
	hasLongitude := func(long float64) checkFn {
		return func(t *testing.T, fc *darksky.Forecast, err error) {
			if fc.Longitude != long {
				t.Errorf("Longitude = %f; want %f", fc.Longitude, long)
			}
		}
	}
	hasTemperature := func() checkFn {
		return func(t *testing.T, fc *darksky.Forecast, err error) {
			if fc.Currently.Temperature == 0.0 {
				t.Errorf("Currently.Temperature = nil; want non-nil")
			}
		}
	}

	c := darksky.Client{
		Key: apiKey,
	}
	tests := map[string]struct {
		lat    float64
		long   float64
		checks []checkFn
	}{
		"valid forecast with correct coords": {
			lat:    32.589720,
			long:   -116.466988,
			checks: check(hasNoErr(), hasLatitude(32.589720), hasLongitude(-116.466988), hasTemperature()),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fc, err := c.Forecast(tc.lat, tc.long)
			for _, check := range tc.checks {
				check(t, fc, err)
			}
		})
	}
}
