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
	hasErr := func() checkFn {
		return func(t *testing.T, fc *darksky.Forecast, err error) {
			if err == nil {
				t.Fatalf("err = nil; want non-nil")
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
		"invalid latitude": {
			lat:    132.0,
			long:   -116.466988,
			checks: check(hasErr()),
		},
		"invalid longitude": {
			lat:    32.0,
			long:   -181.0,
			checks: check(hasErr()),
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
