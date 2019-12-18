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

type Client struct {
	Key        string
	BaseURL    string
	HttpClient interface {
		Do(*http.Request) (*http.Response, error)
	}

	// Services for different endpoints of the Dark Sky API
	ForecastS *ForecastService
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

// func NewClient(key string) *Client {

// 	c := &Client{
// 		Key:     key,
// 		BaseURL: DefaultBaseURL,
// 		Client:  http.DefaultClient,
// 	}
// 	c.ForecastS = &ForecastService{client: c}
// 	return c
// }

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
