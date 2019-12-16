package darksky_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	darksky "github.com/sophiaehlen/darksky-client"
)

var (
	apiKey string
	update bool
)

func init() {
	flag.StringVar(&apiKey, "key", "", "Your TEST secret key for the Dark Sky API. If present, integration tests will be run using this key.")
	flag.BoolVar(&update, "update", false, "Set this flag to update the responses used in local tests. This requires that the key flag is set so that we can interact with the Dark Sky API.")
}

func TestClient_Local(t *testing.T) {

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader((200))
		fmt.Fprintf(w, sample())
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	c := darksky.Client{
		Key:     "gibberish-key",
		BaseURL: server.URL,
	}
	_, err := c.Forecast(1.234, -1.234)
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
}

func darkskyClient(t *testing.T) (*darksky.Client, func()) {
	teardown := make([]func(), 0)
	c := darksky.Client{
		Key: apiKey,
	}
	if apiKey == "" {
		count := 0
		handler := func(w http.ResponseWriter, r *http.Request) {
			resp := readResponse(t, count)
			w.WriteHeader(resp.StatusCode)
			w.Write(resp.Body)
			count++
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		c.BaseURL = server.URL
		teardown = append(teardown, server.Close)
	}
	if update {
		rc := &recorderClient{}
		c.HttpClient = rc
		teardown = append(teardown, func() {
			t.Logf("len(responses) = %d", len(rc.responses))
			for i, res := range rc.responses {
				recordResponse(t, res, i)
			}
		})
	}
	return &c, func() {
		for _, fn := range teardown {
			fn()
		}
	}
}

func responsePath(t *testing.T, count int) string {
	return filepath.Join("testdata", filepath.FromSlash(fmt.Sprintf("%s.%d.json", t.Name(), count)))
}

func readResponse(t *testing.T, count int) response {
	var resp response
	path := responsePath(t, count)
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open the response file: %s. err = %v", path, err)
	}
	defer f.Close()
	jsonBytes, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatalf("failed to read the response file: %s. err = %v", path, err)
	}
	err = json.Unmarshal(jsonBytes, &resp)
	if err != nil {
		t.Fatalf("failed to json unmarshal the response file: %s. err = %v", path, err)
	}
	return resp
}

func recordResponse(t *testing.T, resp response, count int) {
	path := responsePath(t, count)
	err := os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		t.Fatalf("failed to create the response dir: %s. err = %v", filepath.Dir(path), err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create the response file: %s. err = %v", path, err)
	}
	defer f.Close()
	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal JSON for response file: %s. err = %v", path, err)
	}
	_, err = f.Write(jsonBytes)
	if err != nil {
		t.Fatalf("Failed to write json bytes for response file: %s. err = %v", path, err)
	}

}

func TestClient_Forecast(t *testing.T) {
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
			c, teardown := darkskyClient(t)
			defer teardown()
			fc, err := c.Forecast(tc.lat, tc.long)
			for _, check := range tc.checks {
				check(t, fc, err)
			}
		})
	}
}

func sample() string {
	return `{
    "latitude": 32.58972,
    "longitude": -116.466988,
    "timezone": "America/Los_Angeles",
    "currently": {
        "time": 1576521551,
        "summary": "Clear",
        "icon": "clear-day",
        "nearestStormDistance": 296,
        "nearestStormBearing": 39,
        "precipIntensity": 0,
        "precipProbability": 0,
        "temperature": 50.53,
        "apparentTemperature": 50.79,
        "dewPoint": 18.03,
        "humidity": 0.27,
        "pressure": 1024.3,
        "windSpeed": 21.65,
        "windGust": 31.5,
        "windBearing": 60,
        "cloudCover": 0.01,
        "uvIndex": 3,
        "visibility": 10,
        "ozone": 288.6
    },
    "minutely": {
        "summary": "Clear for the hour.",
        "icon": "clear-day",
        "data": [
            {
                "time": 1576521540,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576521600,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576521660,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576521720,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576521780,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576521840,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576521900,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576521960,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576522020,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576522080,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576522140,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576522200,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576522260,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576522320,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576522380,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576522440,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576522500,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576522560,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576522620,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576522680,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576522740,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576522800,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576522860,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576522920,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576522980,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576523040,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576523100,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576523160,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576523220,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576523280,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576523340,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576523400,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576523460,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576523520,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576523580,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576523640,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576523700,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576523760,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576523820,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576523880,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576523940,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576524000,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576524060,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576524120,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576524180,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576524240,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576524300,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576524360,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576524420,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576524480,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576524540,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576524600,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576524660,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576524720,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576524780,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576524840,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576524900,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576524960,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576525020,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576525080,
                "precipIntensity": 0,
                "precipProbability": 0
            },
            {
                "time": 1576525140,
                "precipIntensity": 0,
                "precipProbability": 0
            }
        ]
    },
    "hourly": {
        "summary": "Windy starting this evening, continuing until tomorrow morning.",
        "icon": "wind",
        "data": [
            {
                "time": 1576519200,
                "summary": "Clear",
                "icon": "clear-day",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 50.15,
                "apparentTemperature": 50.15,
                "dewPoint": 19.66,
                "humidity": 0.3,
                "pressure": 1024.5,
                "windSpeed": 20,
                "windGust": 29.54,
                "windBearing": 61,
                "cloudCover": 0,
                "uvIndex": 2,
                "visibility": 10,
                "ozone": 289.6
            },
            {
                "time": 1576522800,
                "summary": "Clear",
                "icon": "clear-day",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 50.61,
                "apparentTemperature": 50.61,
                "dewPoint": 16.94,
                "humidity": 0.26,
                "pressure": 1024,
                "windSpeed": 22.51,
                "windGust": 32.58,
                "windBearing": 60,
                "cloudCover": 0.02,
                "uvIndex": 3,
                "visibility": 10,
                "ozone": 287.8
            },
            {
                "time": 1576526400,
                "summary": "Clear",
                "icon": "clear-day",
                "precipIntensity": 0.0009,
                "precipProbability": 0.01,
                "precipType": "rain",
                "temperature": 50.83,
                "apparentTemperature": 50.83,
                "dewPoint": 13.85,
                "humidity": 0.23,
                "pressure": 1023.1,
                "windSpeed": 24.27,
                "windGust": 34.85,
                "windBearing": 59,
                "cloudCover": 0.05,
                "uvIndex": 3,
                "visibility": 10,
                "ozone": 285.1
            },
            {
                "time": 1576530000,
                "summary": "Clear",
                "icon": "clear-day",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 51.02,
                "apparentTemperature": 51.02,
                "dewPoint": 12.28,
                "humidity": 0.21,
                "pressure": 1023.2,
                "windSpeed": 24.77,
                "windGust": 35.87,
                "windBearing": 59,
                "cloudCover": 0.07,
                "uvIndex": 3,
                "visibility": 10,
                "ozone": 282.4
            },
            {
                "time": 1576533600,
                "summary": "Clear",
                "icon": "clear-day",
                "precipIntensity": 0.0005,
                "precipProbability": 0.01,
                "precipType": "rain",
                "temperature": 49.99,
                "apparentTemperature": 42.81,
                "dewPoint": 10.82,
                "humidity": 0.2,
                "pressure": 1023.2,
                "windSpeed": 24.52,
                "windGust": 36.57,
                "windBearing": 59,
                "cloudCover": 0.07,
                "uvIndex": 2,
                "visibility": 10,
                "ozone": 280.3
            },
            {
                "time": 1576537200,
                "summary": "Clear",
                "icon": "clear-day",
                "precipIntensity": 0.0014,
                "precipProbability": 0.01,
                "precipType": "rain",
                "temperature": 48.25,
                "apparentTemperature": 40.6,
                "dewPoint": 9.65,
                "humidity": 0.21,
                "pressure": 1023.4,
                "windSpeed": 23.89,
                "windGust": 37.26,
                "windBearing": 59,
                "cloudCover": 0.06,
                "uvIndex": 1,
                "visibility": 10,
                "ozone": 278.4
            },
            {
                "time": 1576540800,
                "summary": "Clear",
                "icon": "clear-day",
                "precipIntensity": 0.0012,
                "precipProbability": 0.01,
                "precipType": "rain",
                "temperature": 46.34,
                "apparentTemperature": 38.13,
                "dewPoint": 8.04,
                "humidity": 0.21,
                "pressure": 1023.8,
                "windSpeed": 23.48,
                "windGust": 38.31,
                "windBearing": 59,
                "cloudCover": 0.05,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 276.8
            },
            {
                "time": 1576544400,
                "summary": "Clear",
                "icon": "clear-night",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 44.94,
                "apparentTemperature": 36.46,
                "dewPoint": 5.01,
                "humidity": 0.19,
                "pressure": 1024.5,
                "windSpeed": 22.49,
                "windGust": 38.77,
                "windBearing": 59,
                "cloudCover": 0.05,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 275.4
            },
            {
                "time": 1576548000,
                "summary": "Clear",
                "icon": "clear-night",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 44.3,
                "apparentTemperature": 35.66,
                "dewPoint": 2.57,
                "humidity": 0.17,
                "pressure": 1024.8,
                "windSpeed": 22.2,
                "windGust": 38.56,
                "windBearing": 59,
                "cloudCover": 0.13,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 274.2
            },
            {
                "time": 1576551600,
                "summary": "Windy and Partly Cloudy",
                "icon": "wind",
                "precipIntensity": 0.0005,
                "precipProbability": 0.01,
                "precipType": "rain",
                "temperature": 43.82,
                "apparentTemperature": 34.96,
                "dewPoint": 0.88,
                "humidity": 0.16,
                "pressure": 1024.7,
                "windSpeed": 22.52,
                "windGust": 39.8,
                "windBearing": 59,
                "cloudCover": 0.37,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 273.5
            },
            {
                "time": 1576555200,
                "summary": "Windy and Partly Cloudy",
                "icon": "wind",
                "precipIntensity": 0.0007,
                "precipProbability": 0.01,
                "precipType": "rain",
                "temperature": 43.44,
                "apparentTemperature": 34.27,
                "dewPoint": 0.29,
                "humidity": 0.16,
                "pressure": 1025.1,
                "windSpeed": 23.45,
                "windGust": 40.11,
                "windBearing": 58,
                "cloudCover": 0.37,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 273.5
            },
            {
                "time": 1576558800,
                "summary": "Windy",
                "icon": "wind",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 43.08,
                "apparentTemperature": 33.6,
                "dewPoint": -0.13,
                "humidity": 0.16,
                "pressure": 1025.8,
                "windSpeed": 24.49,
                "windGust": 41.61,
                "windBearing": 57,
                "cloudCover": 0.27,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 273.9
            },
            {
                "time": 1576562400,
                "summary": "Windy",
                "icon": "wind",
                "precipIntensity": 0.0018,
                "precipProbability": 0.01,
                "precipType": "rain",
                "temperature": 42.76,
                "apparentTemperature": 32.97,
                "dewPoint": -1.55,
                "humidity": 0.15,
                "pressure": 1026.2,
                "windSpeed": 25.61,
                "windGust": 43.84,
                "windBearing": 56,
                "cloudCover": 0.2,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 274
            },
            {
                "time": 1576566000,
                "summary": "Windy",
                "icon": "wind",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 42.46,
                "apparentTemperature": 32.47,
                "dewPoint": -3.08,
                "humidity": 0.14,
                "pressure": 1026.3,
                "windSpeed": 26.1,
                "windGust": 44,
                "windBearing": 56,
                "cloudCover": 0.15,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 273.5
            },
            {
                "time": 1576569600,
                "summary": "Windy",
                "icon": "wind",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 41.11,
                "apparentTemperature": 30.83,
                "dewPoint": 0.22,
                "humidity": 0.18,
                "pressure": 1027.2,
                "windSpeed": 25.2,
                "windGust": 43.48,
                "windBearing": 56,
                "cloudCover": 0.25,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 272.7
            },
            {
                "time": 1576573200,
                "summary": "Windy",
                "icon": "wind",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 40.53,
                "apparentTemperature": 30.04,
                "dewPoint": 0.29,
                "humidity": 0.18,
                "pressure": 1027,
                "windSpeed": 25.27,
                "windGust": 43.04,
                "windBearing": 56,
                "cloudCover": 0.26,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 272.2
            },
            {
                "time": 1576576800,
                "summary": "Windy",
                "icon": "wind",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 40.08,
                "apparentTemperature": 29.45,
                "dewPoint": 0.43,
                "humidity": 0.18,
                "pressure": 1026.9,
                "windSpeed": 25.15,
                "windGust": 42.44,
                "windBearing": 57,
                "cloudCover": 0.23,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 272.4
            },
            {
                "time": 1576580400,
                "summary": "Windy",
                "icon": "wind",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 39.71,
                "apparentTemperature": 29.02,
                "dewPoint": 0.72,
                "humidity": 0.19,
                "pressure": 1026.7,
                "windSpeed": 24.91,
                "windGust": 41.75,
                "windBearing": 58,
                "cloudCover": 0.18,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 273.1
            },
            {
                "time": 1576584000,
                "summary": "Windy",
                "icon": "wind",
                "precipIntensity": 0.0002,
                "precipProbability": 0.01,
                "precipType": "rain",
                "temperature": 39.41,
                "apparentTemperature": 28.62,
                "dewPoint": 0.97,
                "humidity": 0.19,
                "pressure": 1026.6,
                "windSpeed": 24.89,
                "windGust": 41.68,
                "windBearing": 60,
                "cloudCover": 0.13,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 273.3
            },
            {
                "time": 1576587600,
                "summary": "Windy",
                "icon": "wind",
                "precipIntensity": 0.0002,
                "precipProbability": 0.01,
                "precipType": "rain",
                "temperature": 38.79,
                "apparentTemperature": 27.72,
                "dewPoint": 1.27,
                "humidity": 0.2,
                "pressure": 1026.9,
                "windSpeed": 25.26,
                "windGust": 42.86,
                "windBearing": 60,
                "cloudCover": 0.08,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 272.5
            },
            {
                "time": 1576591200,
                "summary": "Windy",
                "icon": "wind",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 38.19,
                "apparentTemperature": 26.79,
                "dewPoint": 1.22,
                "humidity": 0.21,
                "pressure": 1027.2,
                "windSpeed": 25.86,
                "windGust": 44.64,
                "windBearing": 61,
                "cloudCover": 0.04,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 271.4
            },
            {
                "time": 1576594800,
                "summary": "Windy",
                "icon": "wind",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 38.4,
                "apparentTemperature": 26.95,
                "dewPoint": 1.17,
                "humidity": 0.2,
                "pressure": 1027.5,
                "windSpeed": 26.46,
                "windGust": 45.81,
                "windBearing": 62,
                "cloudCover": 0,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 270.8
            },
            {
                "time": 1576598400,
                "summary": "Windy",
                "icon": "wind",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 39.86,
                "apparentTemperature": 28.76,
                "dewPoint": 0.85,
                "humidity": 0.19,
                "pressure": 1027.5,
                "windSpeed": 27.26,
                "windGust": 45.99,
                "windBearing": 62,
                "cloudCover": 0,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 271.5
            },
            {
                "time": 1576602000,
                "summary": "Windy",
                "icon": "wind",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 41.98,
                "apparentTemperature": 31.48,
                "dewPoint": -0.4,
                "humidity": 0.16,
                "pressure": 1027.1,
                "windSpeed": 28.07,
                "windGust": 45.56,
                "windBearing": 62,
                "cloudCover": 0,
                "uvIndex": 1,
                "visibility": 10,
                "ozone": 272.8
            },
            {
                "time": 1576605600,
                "summary": "Windy",
                "icon": "wind",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 43.53,
                "apparentTemperature": 33.51,
                "dewPoint": -1.56,
                "humidity": 0.15,
                "pressure": 1026.9,
                "windSpeed": 28.4,
                "windGust": 44.35,
                "windBearing": 62,
                "cloudCover": 0.02,
                "uvIndex": 3,
                "visibility": 10,
                "ozone": 273.6
            },
            {
                "time": 1576609200,
                "summary": "Windy",
                "icon": "wind",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 45.5,
                "apparentTemperature": 36.26,
                "dewPoint": -3.59,
                "humidity": 0.12,
                "pressure": 1026.2,
                "windSpeed": 27.84,
                "windGust": 42.03,
                "windBearing": 63,
                "cloudCover": 0.17,
                "uvIndex": 3,
                "visibility": 10,
                "ozone": 273.5
            },
            {
                "time": 1576612800,
                "summary": "Windy and Partly Cloudy",
                "icon": "wind",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 47.18,
                "apparentTemperature": 38.71,
                "dewPoint": -5.08,
                "humidity": 0.11,
                "pressure": 1025.7,
                "windSpeed": 26.73,
                "windGust": 38.97,
                "windBearing": 64,
                "cloudCover": 0.37,
                "uvIndex": 3,
                "visibility": 10,
                "ozone": 273
            },
            {
                "time": 1576616400,
                "summary": "Windy and Partly Cloudy",
                "icon": "wind",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 47.98,
                "apparentTemperature": 39.98,
                "dewPoint": -6.04,
                "humidity": 0.1,
                "pressure": 1024.9,
                "windSpeed": 25.49,
                "windGust": 36.28,
                "windBearing": 64,
                "cloudCover": 0.52,
                "uvIndex": 3,
                "visibility": 10,
                "ozone": 273
            },
            {
                "time": 1576620000,
                "summary": "Partly Cloudy",
                "icon": "partly-cloudy-day",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 47.02,
                "apparentTemperature": 38.9,
                "dewPoint": -5.59,
                "humidity": 0.11,
                "pressure": 1024.3,
                "windSpeed": 24.21,
                "windGust": 34.06,
                "windBearing": 64,
                "cloudCover": 0.57,
                "uvIndex": 2,
                "visibility": 10,
                "ozone": 274.1
            },
            {
                "time": 1576623600,
                "summary": "Partly Cloudy",
                "icon": "partly-cloudy-day",
                "precipIntensity": 0.0002,
                "precipProbability": 0.01,
                "precipType": "rain",
                "temperature": 45.44,
                "apparentTemperature": 37.05,
                "dewPoint": -4.43,
                "humidity": 0.12,
                "pressure": 1023.9,
                "windSpeed": 22.87,
                "windGust": 32.23,
                "windBearing": 63,
                "cloudCover": 0.57,
                "uvIndex": 1,
                "visibility": 10,
                "ozone": 275.8
            },
            {
                "time": 1576627200,
                "summary": "Partly Cloudy",
                "icon": "partly-cloudy-day",
                "precipIntensity": 0.0002,
                "precipProbability": 0.01,
                "precipType": "rain",
                "temperature": 43.94,
                "apparentTemperature": 35.29,
                "dewPoint": -3.72,
                "humidity": 0.13,
                "pressure": 1023.7,
                "windSpeed": 21.72,
                "windGust": 31.21,
                "windBearing": 63,
                "cloudCover": 0.57,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 277.4
            },
            {
                "time": 1576630800,
                "summary": "Partly Cloudy",
                "icon": "partly-cloudy-night",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 43.01,
                "apparentTemperature": 34.23,
                "dewPoint": -4.64,
                "humidity": 0.13,
                "pressure": 1023.6,
                "windSpeed": 20.91,
                "windGust": 31.63,
                "windBearing": 62,
                "cloudCover": 0.55,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 278.8
            },
            {
                "time": 1576634400,
                "summary": "Partly Cloudy",
                "icon": "partly-cloudy-night",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 41.93,
                "apparentTemperature": 32.96,
                "dewPoint": -6.46,
                "humidity": 0.12,
                "pressure": 1023.6,
                "windSpeed": 20.25,
                "windGust": 32.88,
                "windBearing": 62,
                "cloudCover": 0.52,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 280.3
            },
            {
                "time": 1576638000,
                "summary": "Partly Cloudy",
                "icon": "partly-cloudy-night",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 41.61,
                "apparentTemperature": 32.67,
                "dewPoint": -8.22,
                "humidity": 0.11,
                "pressure": 1023.1,
                "windSpeed": 19.66,
                "windGust": 33.44,
                "windBearing": 62,
                "cloudCover": 0.53,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 282.1
            },
            {
                "time": 1576641600,
                "summary": "Mostly Cloudy",
                "icon": "partly-cloudy-night",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 41.95,
                "apparentTemperature": 33.22,
                "dewPoint": -8.83,
                "humidity": 0.11,
                "pressure": 1022.9,
                "windSpeed": 19.22,
                "windGust": 32.86,
                "windBearing": 64,
                "cloudCover": 0.61,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 283.8
            },
            {
                "time": 1576645200,
                "summary": "Mostly Cloudy",
                "icon": "partly-cloudy-night",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 42.54,
                "apparentTemperature": 34.08,
                "dewPoint": -9.12,
                "humidity": 0.11,
                "pressure": 1023.2,
                "windSpeed": 18.86,
                "windGust": 31.57,
                "windBearing": 64,
                "cloudCover": 0.73,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 285.7
            },
            {
                "time": 1576648800,
                "summary": "Mostly Cloudy",
                "icon": "partly-cloudy-night",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 42.84,
                "apparentTemperature": 34.64,
                "dewPoint": -9.32,
                "humidity": 0.1,
                "pressure": 1023.1,
                "windSpeed": 18.15,
                "windGust": 29.45,
                "windBearing": 64,
                "cloudCover": 0.81,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 288.6
            },
            {
                "time": 1576652400,
                "summary": "Mostly Cloudy",
                "icon": "partly-cloudy-night",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 42.72,
                "apparentTemperature": 34.81,
                "dewPoint": -9.01,
                "humidity": 0.11,
                "pressure": 1022.7,
                "windSpeed": 16.83,
                "windGust": 26.02,
                "windBearing": 64,
                "cloudCover": 0.82,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 293.4
            },
            {
                "time": 1576656000,
                "summary": "Mostly Cloudy",
                "icon": "partly-cloudy-night",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 42.29,
                "apparentTemperature": 34.72,
                "dewPoint": -7.71,
                "humidity": 0.11,
                "pressure": 1022.1,
                "windSpeed": 15.18,
                "windGust": 21.76,
                "windBearing": 62,
                "cloudCover": 0.8,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 299.2
            },
            {
                "time": 1576659600,
                "summary": "Mostly Cloudy",
                "icon": "partly-cloudy-night",
                "precipIntensity": 0.0002,
                "precipProbability": 0.01,
                "precipType": "rain",
                "temperature": 41.5,
                "apparentTemperature": 34.14,
                "dewPoint": -5.74,
                "humidity": 0.13,
                "pressure": 1021.7,
                "windSpeed": 13.75,
                "windGust": 18.1,
                "windBearing": 62,
                "cloudCover": 0.75,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 303.5
            },
            {
                "time": 1576663200,
                "summary": "Mostly Cloudy",
                "icon": "partly-cloudy-night",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 40.37,
                "apparentTemperature": 33.03,
                "dewPoint": -4.32,
                "humidity": 0.15,
                "pressure": 1021.3,
                "windSpeed": 12.76,
                "windGust": 15.44,
                "windBearing": 60,
                "cloudCover": 0.66,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 304.9
            },
            {
                "time": 1576666800,
                "summary": "Partly Cloudy",
                "icon": "partly-cloudy-night",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 38.72,
                "apparentTemperature": 31.24,
                "dewPoint": -3.06,
                "humidity": 0.16,
                "pressure": 1021,
                "windSpeed": 11.98,
                "windGust": 13.33,
                "windBearing": 58,
                "cloudCover": 0.53,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 304.8
            },
            {
                "time": 1576670400,
                "summary": "Partly Cloudy",
                "icon": "partly-cloudy-night",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 37.42,
                "apparentTemperature": 29.89,
                "dewPoint": -1.95,
                "humidity": 0.18,
                "pressure": 1020.9,
                "windSpeed": 11.29,
                "windGust": 12.09,
                "windBearing": 56,
                "cloudCover": 0.4,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 305.4
            },
            {
                "time": 1576674000,
                "summary": "Clear",
                "icon": "clear-night",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 36.16,
                "apparentTemperature": 28.55,
                "dewPoint": -1.06,
                "humidity": 0.2,
                "pressure": 1020.9,
                "windSpeed": 10.75,
                "windGust": 11.49,
                "windBearing": 56,
                "cloudCover": 0.26,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 307.5
            },
            {
                "time": 1576677600,
                "summary": "Clear",
                "icon": "clear-night",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 34.96,
                "apparentTemperature": 27.25,
                "dewPoint": -0.4,
                "humidity": 0.22,
                "pressure": 1021,
                "windSpeed": 10.3,
                "windGust": 11.13,
                "windBearing": 56,
                "cloudCover": 0.1,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 310.2
            },
            {
                "time": 1576681200,
                "summary": "Clear",
                "icon": "clear-day",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 35.18,
                "apparentTemperature": 27.86,
                "dewPoint": 0.46,
                "humidity": 0.22,
                "pressure": 1020.4,
                "windSpeed": 9.59,
                "windGust": 10.41,
                "windBearing": 57,
                "cloudCover": 0,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 312.4
            },
            {
                "time": 1576684800,
                "summary": "Clear",
                "icon": "clear-day",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 38.79,
                "apparentTemperature": 33,
                "dewPoint": 0.74,
                "humidity": 0.2,
                "pressure": 1020.1,
                "windSpeed": 8.2,
                "windGust": 8.66,
                "windBearing": 60,
                "cloudCover": 0.02,
                "uvIndex": 0,
                "visibility": 10,
                "ozone": 313.5
            },
            {
                "time": 1576688400,
                "summary": "Clear",
                "icon": "clear-day",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 44.35,
                "apparentTemperature": 40.62,
                "dewPoint": -0.09,
                "humidity": 0.15,
                "pressure": 1020.2,
                "windSpeed": 6.56,
                "windGust": 6.62,
                "windBearing": 64,
                "cloudCover": 0.1,
                "uvIndex": 1,
                "visibility": 10,
                "ozone": 314.1
            },
            {
                "time": 1576692000,
                "summary": "Clear",
                "icon": "clear-day",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 49.03,
                "apparentTemperature": 46.85,
                "dewPoint": -1.14,
                "humidity": 0.12,
                "pressure": 1019.4,
                "windSpeed": 5.4,
                "windGust": 5.4,
                "windBearing": 71,
                "cloudCover": 0.2,
                "uvIndex": 2,
                "visibility": 10,
                "ozone": 314.6
            }
        ]
    },
    "daily": {
        "summary": "Possible light rain next Monday.",
        "icon": "rain",
        "data": [
            {
                "time": 1576483200,
                "summary": "Windy in the evening and overnight.",
                "icon": "wind",
                "sunriseTime": 1576507320,
                "sunsetTime": 1576543380,
                "moonPhase": 0.68,
                "precipIntensity": 0.0006,
                "precipIntensityMax": 0.0021,
                "precipIntensityMaxTime": 1576512000,
                "precipProbability": 0.03,
                "precipType": "rain",
                "temperatureHigh": 51.56,
                "temperatureHighTime": 1576529280,
                "temperatureLow": 37.66,
                "temperatureLowTime": 1576592400,
                "apparentTemperatureHigh": 51.58,
                "apparentTemperatureHighTime": 1576528860,
                "apparentTemperatureLow": 26.69,
                "apparentTemperatureLowTime": 1576592760,
                "dewPoint": 11.3,
                "humidity": 0.25,
                "pressure": 1023.6,
                "windSpeed": 20.41,
                "windGust": 44.09,
                "windGustTime": 1576564140,
                "windBearing": 58,
                "cloudCover": 0.09,
                "uvIndex": 3,
                "uvIndexTime": 1576525380,
                "visibility": 10,
                "ozone": 283.1,
                "temperatureMin": 40.62,
                "temperatureMinTime": 1576569600,
                "temperatureMax": 51.56,
                "temperatureMaxTime": 1576529280,
                "apparentTemperatureMin": 30.83,
                "apparentTemperatureMinTime": 1576569600,
                "apparentTemperatureMax": 51.58,
                "apparentTemperatureMaxTime": 1576528860
            },
            {
                "time": 1576569600,
                "summary": "Windy in the morning.",
                "icon": "wind",
                "sunriseTime": 1576593780,
                "sunsetTime": 1576629840,
                "moonPhase": 0.71,
                "precipIntensity": 0.0001,
                "precipIntensityMax": 0.0002,
                "precipIntensityMaxTime": 1576626720,
                "precipProbability": 0.02,
                "precipType": "rain",
                "temperatureHigh": 48.48,
                "temperatureHighTime": 1576616280,
                "temperatureLow": 34.28,
                "temperatureLowTime": 1576679460,
                "apparentTemperatureHigh": 39.98,
                "apparentTemperatureHighTime": 1576616460,
                "apparentTemperatureLow": 27.14,
                "apparentTemperatureLowTime": 1576679040,
                "dewPoint": -3.45,
                "humidity": 0.15,
                "pressure": 1025.2,
                "windSpeed": 23.52,
                "windGust": 46.02,
                "windGustTime": 1576597320,
                "windBearing": 62,
                "cloudCover": 0.37,
                "uvIndex": 3,
                "uvIndexTime": 1576610700,
                "visibility": 10,
                "ozone": 277,
                "temperatureMin": 37.66,
                "temperatureMinTime": 1576592400,
                "temperatureMax": 48.48,
                "temperatureMaxTime": 1576616280,
                "apparentTemperatureMin": 26.69,
                "apparentTemperatureMinTime": 1576592760,
                "apparentTemperatureMax": 39.98,
                "apparentTemperatureMaxTime": 1576616460
            },
            {
                "time": 1576656000,
                "summary": "Partly cloudy throughout the day.",
                "icon": "partly-cloudy-day",
                "sunriseTime": 1576680240,
                "sunsetTime": 1576716240,
                "moonPhase": 0.75,
                "precipIntensity": 0.0001,
                "precipIntensityMax": 0.0004,
                "precipIntensityMaxTime": 1576702800,
                "precipProbability": 0.02,
                "precipType": "rain",
                "temperatureHigh": 55.17,
                "temperatureHighTime": 1576703340,
                "temperatureLow": 33.71,
                "temperatureLowTime": 1576762020,
                "apparentTemperatureHigh": 54.67,
                "apparentTemperatureHighTime": 1576703340,
                "apparentTemperatureLow": 28.09,
                "apparentTemperatureLowTime": 1576763340,
                "dewPoint": 2.26,
                "humidity": 0.2,
                "pressure": 1019.1,
                "windSpeed": 7.02,
                "windGust": 21.76,
                "windGustTime": 1576656000,
                "windBearing": 47,
                "cloudCover": 0.32,
                "uvIndex": 3,
                "uvIndexTime": 1576696560,
                "visibility": 10,
                "ozone": 315.9,
                "temperatureMin": 34.28,
                "temperatureMinTime": 1576679460,
                "temperatureMax": 55.17,
                "temperatureMaxTime": 1576703340,
                "apparentTemperatureMin": 27.14,
                "apparentTemperatureMinTime": 1576679040,
                "apparentTemperatureMax": 54.67,
                "apparentTemperatureMaxTime": 1576703340
            },
            {
                "time": 1576742400,
                "summary": "Clear throughout the day.",
                "icon": "clear-day",
                "sunriseTime": 1576766640,
                "sunsetTime": 1576802700,
                "moonPhase": 0.79,
                "precipIntensity": 0.0001,
                "precipIntensityMax": 0.0003,
                "precipIntensityMaxTime": 1576757280,
                "precipProbability": 0.01,
                "precipType": "rain",
                "temperatureHigh": 60.25,
                "temperatureHighTime": 1576789440,
                "temperatureLow": 39.23,
                "temperatureLowTime": 1576850460,
                "apparentTemperatureHigh": 59.75,
                "apparentTemperatureHighTime": 1576789440,
                "apparentTemperatureLow": 34.03,
                "apparentTemperatureLowTime": 1576850820,
                "dewPoint": 8.77,
                "humidity": 0.24,
                "pressure": 1021.3,
                "windSpeed": 7.46,
                "windGust": 14.18,
                "windGustTime": 1576779480,
                "windBearing": 50,
                "cloudCover": 0.02,
                "uvIndex": 3,
                "uvIndexTime": 1576784760,
                "visibility": 10,
                "ozone": 310.4,
                "temperatureMin": 33.71,
                "temperatureMinTime": 1576762020,
                "temperatureMax": 60.25,
                "temperatureMaxTime": 1576789440,
                "apparentTemperatureMin": 28.09,
                "apparentTemperatureMinTime": 1576763340,
                "apparentTemperatureMax": 59.75,
                "apparentTemperatureMaxTime": 1576789440
            },
            {
                "time": 1576828800,
                "summary": "Mostly cloudy throughout the day.",
                "icon": "partly-cloudy-day",
                "sunriseTime": 1576853100,
                "sunsetTime": 1576889100,
                "moonPhase": 0.82,
                "precipIntensity": 0,
                "precipIntensityMax": 0,
                "precipIntensityMaxTime": 1576854000,
                "precipProbability": 0,
                "temperatureHigh": 63.94,
                "temperatureHighTime": 1576875420,
                "temperatureLow": 42.59,
                "temperatureLowTime": 1576936800,
                "apparentTemperatureHigh": 63.44,
                "apparentTemperatureHighTime": 1576875420,
                "apparentTemperatureLow": 39.53,
                "apparentTemperatureLowTime": 1576936680,
                "dewPoint": 2.92,
                "humidity": 0.15,
                "pressure": 1024.1,
                "windSpeed": 8.16,
                "windGust": 15.77,
                "windGustTime": 1576864920,
                "windBearing": 59,
                "cloudCover": 0.56,
                "uvIndex": 3,
                "uvIndexTime": 1576869720,
                "visibility": 10,
                "ozone": 291.1,
                "temperatureMin": 39.23,
                "temperatureMinTime": 1576850460,
                "temperatureMax": 63.94,
                "temperatureMaxTime": 1576875420,
                "apparentTemperatureMin": 34.03,
                "apparentTemperatureMinTime": 1576850820,
                "apparentTemperatureMax": 63.44,
                "apparentTemperatureMaxTime": 1576875420
            },
            {
                "time": 1576915200,
                "summary": "Overcast throughout the day.",
                "icon": "cloudy",
                "sunriseTime": 1576939500,
                "sunsetTime": 1576975500,
                "moonPhase": 0.86,
                "precipIntensity": 0.0001,
                "precipIntensityMax": 0.0003,
                "precipIntensityMaxTime": 1576962540,
                "precipProbability": 0.01,
                "precipType": "rain",
                "temperatureHigh": 69.02,
                "temperatureHighTime": 1576961760,
                "temperatureLow": 40.37,
                "temperatureLowTime": 1577023020,
                "apparentTemperatureHigh": 68.52,
                "apparentTemperatureHighTime": 1576961760,
                "apparentTemperatureLow": 39.15,
                "apparentTemperatureLowTime": 1577022660,
                "dewPoint": 1.84,
                "humidity": 0.14,
                "pressure": 1019.9,
                "windSpeed": 4.42,
                "windGust": 8.26,
                "windGustTime": 1576915200,
                "windBearing": 72,
                "cloudCover": 0.95,
                "uvIndex": 3,
                "uvIndexTime": 1576957260,
                "visibility": 10,
                "ozone": 288.2,
                "temperatureMin": 42.59,
                "temperatureMinTime": 1576936800,
                "temperatureMax": 69.02,
                "temperatureMaxTime": 1576961760,
                "apparentTemperatureMin": 39.53,
                "apparentTemperatureMinTime": 1576936680,
                "apparentTemperatureMax": 68.52,
                "apparentTemperatureMaxTime": 1576961760
            },
            {
                "time": 1577001600,
                "summary": "Overcast throughout the day.",
                "icon": "cloudy",
                "sunriseTime": 1577025960,
                "sunsetTime": 1577061960,
                "moonPhase": 0.89,
                "precipIntensity": 0.0001,
                "precipIntensityMax": 0.0003,
                "precipIntensityMaxTime": 1577026800,
                "precipProbability": 0.01,
                "precipType": "rain",
                "temperatureHigh": 63.58,
                "temperatureHighTime": 1577047800,
                "temperatureLow": 45.36,
                "temperatureLowTime": 1577102820,
                "apparentTemperatureHigh": 63.08,
                "apparentTemperatureHighTime": 1577047800,
                "apparentTemperatureLow": 43.73,
                "apparentTemperatureLowTime": 1577111460,
                "dewPoint": 8.74,
                "humidity": 0.19,
                "pressure": 1016.5,
                "windSpeed": 3.65,
                "windGust": 8.61,
                "windGustTime": 1577047800,
                "windBearing": 252,
                "cloudCover": 1,
                "uvIndex": 2,
                "uvIndexTime": 1577043780,
                "visibility": 10,
                "ozone": 298.1,
                "temperatureMin": 40.37,
                "temperatureMinTime": 1577023020,
                "temperatureMax": 63.58,
                "temperatureMaxTime": 1577047800,
                "apparentTemperatureMin": 39.15,
                "apparentTemperatureMinTime": 1577022660,
                "apparentTemperatureMax": 63.08,
                "apparentTemperatureMaxTime": 1577047800
            },
            {
                "time": 1577088000,
                "summary": "Possible light rain overnight.",
                "icon": "rain",
                "sunriseTime": 1577112360,
                "sunsetTime": 1577148360,
                "moonPhase": 0.93,
                "precipIntensity": 0.0091,
                "precipIntensityMax": 0.022,
                "precipIntensityMaxTime": 1577123700,
                "precipProbability": 0.41,
                "precipType": "rain",
                "temperatureHigh": 58.39,
                "temperatureHighTime": 1577137200,
                "temperatureLow": 48.58,
                "temperatureLowTime": 1577188680,
                "apparentTemperatureHigh": 57.89,
                "apparentTemperatureHighTime": 1577137200,
                "apparentTemperatureLow": 48.17,
                "apparentTemperatureLowTime": 1577187960,
                "dewPoint": 22.17,
                "humidity": 0.33,
                "pressure": 1013.7,
                "windSpeed": 4.38,
                "windGust": 23.85,
                "windGustTime": 1577113800,
                "windBearing": 105,
                "cloudCover": 1,
                "uvIndex": 2,
                "uvIndexTime": 1577130420,
                "visibility": 8.798,
                "ozone": 323.3,
                "temperatureMin": 45.36,
                "temperatureMinTime": 1577102820,
                "temperatureMax": 58.39,
                "temperatureMaxTime": 1577137200,
                "apparentTemperatureMin": 43.73,
                "apparentTemperatureMinTime": 1577111460,
                "apparentTemperatureMax": 57.89,
                "apparentTemperatureMaxTime": 1577137200
            }
        ]
    },
    "alerts": [
        {
            "title": "Wind Advisory",
            "regions": [
                "Orange County Inland",
                "Riverside County Mountains",
                "San Bernardino County Mountains",
                "San Bernardino and Riverside County Valleys-The Inland Empire",
                "San Diego County Inland Valleys",
                "San Diego County Mountains",
                "San Gorgonio Pass Near Banning",
                "Santa Ana Mountains and Foothills"
            ],
            "severity": "advisory",
            "time": 1576497600,
            "expires": 1576648800,
            "description": "...WIND ADVISORY REMAINS IN EFFECT UNTIL 10 PM PST TUESDAY... * WHAT...Areas of northeast winds 20 to 30 mph with gusts to 50 mph are expected. * WHERE...Mountains, valleys, and inland Orange County. The stronger gusts are expected in the Inland Empire below the Cajon and Banning Passes. near the coastal slopes of the Santa Ana Mountains, and near the ridge tops in the San Diego County mountains. * WHEN...Until 10 PM PST Tuesday. * IMPACTS...Gusty winds could blow around unsecured objects. Tree limbs could be blown down and a few power outages may result. * ADDITIONAL DETAILS...Winds will gradually increase today, peaking tonight into Tuesday morning.\n",
            "uri": "https://alerts.weather.gov/cap/wwacapget.php?x=CA125D22439AE0.WindAdvisory.125D22614AE0CA.SGXNPWSGX.36b83adb84b37e195e6eff16da67f679"
        }
    ],
    "flags": {
        "sources": [
            "nwspa",
            "cmc",
            "gfs",
            "hrrr",
            "icon",
            "isd",
            "madis",
            "nam",
            "sref",
            "darksky",
            "nearest-precip"
        ],
        "nearest-station": 0.307,
        "units": "us"
    },
    "offset": -8
}`
}
