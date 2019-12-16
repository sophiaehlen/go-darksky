package darksky

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	DefaultBaseURL = "https://api.darksky.net"
)

type Forecast struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Currently struct {
		Time                 int     `json:"time"`
		Summary              string  `json:"summary"`
		Icon                 string  `json:"icon"`
		NearestStormDistance int     `json:"nearestStormDistance"`
		PrecipIntensity      float64 `json:"precipIntensity"`
		PrecipIntensityError float64 `json:"precipIntensityError"`
		PrecipProbability    float64 `json:"precipProbability"`
		PrecipType           string  `json:"precipType"`
		Temperature          float64 `json:"temperature"`
		ApparentTemperature  float64 `json:"apparentTemperature"` // "feels like temp in Fahrenheit"
		DewPoint             float64 `json:"dewPoint"`
		Humidity             float64 `json:"humidity"`
		Pressure             float64 `json:"pressure"`
		WindSpeed            float64 `json:"windSpeed"`
		WindGust             float64 `json:"windGust"`
		WindBearing          int     `json:"windBearing"`
		CloudCover           float64 `json:"cloudCover"`
	} `json:"currently"`
	Alerts []struct {
		Title       string `json:"title"`
		Time        int    `json:"time"`
		Expires     int    `json:"expires"`
		Description string `json:"description"`
		URI         string `json:"uri"`
	} `json:"alerts"`
}

type Client struct {
	Key        string
	BaseURL    string
	HttpClient interface {
		Do(*http.Request) (*http.Response, error)
	}
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	httpClient := c.HttpClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	if req.Method != http.MethodGet {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.SetBasicAuth(c.Key, "")
	return httpClient.Do(req)
}

func (c *Client) url(path string) string {
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
	return fmt.Sprintf("%s%s", c.BaseURL, path)
}

func (c *Client) latlong(lat, long float64) string {
	return fmt.Sprintf("%f,%f", lat, long)
}

func (c *Client) Forecast(lat, long float64) (*Forecast, error) {
	endpoint := c.url("/forecast")

	endpoint = endpoint + "/" + c.Key + "/" + c.latlong(lat, long)

	req, err := http.NewRequest(http.MethodGet, endpoint, strings.NewReader(endpoint))

	res, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, ErrBadRequest
	}
	var forecast Forecast
	err = json.Unmarshal(body, &forecast)
	if err != nil {
		return nil, err
	}
	return &forecast, nil
}

// func parseError(data []byte) error {
// 	var se Error
// 	err := json.Unmarshal(data, &se)
// 	if err != nil {
// 		return err
// 	}
// 	return se
// }
