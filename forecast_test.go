package darksky_test

import (
	"testing"

	darksky "github.com/sophiaehlen/darksky-client"
)

var (
	stLat  = 32.589720
	stLong = -116.466988
)

func TestForecast_CurrentTemperature(t *testing.T) {
	if apiKey == "" {
		t.Log("No API key provided. Running unit tests using recorded responses. Be sure to run against the real API before commiting.")
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
	hasCurrTemperature := func() checkFn {
		return func(t *testing.T, fc *darksky.Forecast, err error) {
			if fc.Currently.Temperature == 0.0 {
				t.Errorf("Currently.Temperature = nil; want non-nil")
			}
		}
	}
	tests := map[string]struct {
		lat    float64
		long   float64
		checks []checkFn
	}{
		"current temperature is present": {
			lat:  stLat,
			long: stLong,
			checks: check(
				hasNoErr(),
				hasCurrTemperature(),
			),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c, teardown := darkskyClient(t)
			defer teardown()
			fc, err := c.Forecast(tc.lat, tc.long)
			for _, check := range tc.checks {
				check(t, fc, err)
			}
		})
	}
}

func TestForecast_LocalTimeOfForecast(t *testing.T) {
	if apiKey == "" {
		t.Log("No API key provided. Running unit tests using recorded responses. Be sure to run against the real API before commiting.")
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
	hasTimezone := func() checkFn {
		return func(t *testing.T, fc *darksky.Forecast, err error) {
			if fc.Timezone == "" {
				t.Errorf("Timezone = nil; want non-nil")
			}
		}
	}
	canConvertTime := func() checkFn {
		return func(t *testing.T, fc *darksky.Forecast, err error) {
			_, err = fc.LocalTime()
			if err != nil {
				t.Fatalf("err = %v; want nil", err)
			}
			// TODO: add a check to make sure the time isn't zero
			// if  {
			// 	t.Fatalf("LocalTime = %v; want %v")
			// }
		}
	}
	// convertedTimeNotZero := func(unixTimeStr string) checkFn {
	// 	return func(t *testing.T, fc *darksky.Forecast, err error) {
	// 		lt, err = fc.LocalTime()
	// 		if lt == new(time.Time) {
	// 			t.Fatalf("LocalTime = %v; want %v")
	// 		}
	// 	}
	// }

	// timeIsNotZerovalue
	tests := map[string]struct {
		lat    float64
		long   float64
		want   string
		checks []checkFn
	}{
		"timezone is present": {
			lat:  stLat,
			long: stLong,
			checks: check(
				hasNoErr(),
				hasTimezone(),
			),
		},
		"time can be converted": {
			lat:  stLat,
			long: stLong,
			checks: check(
				canConvertTime(),
			),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c, teardown := darkskyClient(t)
			defer teardown()
			fc, err := c.Forecast(tc.lat, tc.long)
			for _, check := range tc.checks {
				check(t, fc, err)
			}
		})
	}
}
