package darksky_test

import (
	"flag"
	"fmt"
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
	c := darksky.Client{
		Key: apiKey,
	}

	lat := 32.589720
	long := -116.466988

	forecast, err := c.Forecast(lat, long)
	if err != nil {
		t.Errorf("Forecast() err = %v; want %v", err, nil)
	}
	if forecast == nil {
		t.Fatalf("Forecast() = nil; want non-nil value")
	}
	fmt.Println(forecast)
}
