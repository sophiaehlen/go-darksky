package darksky

import (
	"time"
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
	Minutely struct {
		Summary string `json:"summary"`
		Icon    string `json:"icon"`
	} `json:"minutely"`
	Hourly struct {
		Summary string `json:"summary"`
		Icon    string `json:"icon"`
		Data    []struct {
			Time                int     `json:"time"`
			Summary             string  `json:"summary"`
			Icon                string  `json:"icon"`
			PrecipIntensity     float64 `json:"precipIntensity"`
			PrecipProbability   float64 `json:"precipProbability"`
			PrecipType          string  `json:"precipType"`
			Temperature         float64 `json:"temperature"`
			ApparentTemperature float64 `json:"apparentTemperature"`
			DewPoint            float64 `json:"dewPoint"`
			Humidity            float64 `json:"humidity"`
			Pressure            float64 `json:"pressure"`
			WindSpeed           float64 `json:"windSpeed"`
			WindGust            float64 `json:"windGust"`
			WindBearing         int     `json:"windBearing"`
			CloudCover          float64 `json:"cloudCover"`
			UvIndex             int     `json:"uvIndex"`
			Visibility          float64 `json:"visibility"`
			Ozone               float64 `json:"ozone"`
		} `json:"data"`
	} `json:"hourly"`
	Daily struct {
		Summary string `json:"summary"`
		Icon    string `json:"icon"`
		Data    []struct {
			Time                        int     `json:"time"`
			Summary                     string  `json:"summary"`
			Icon                        string  `json:"icon"`
			SunriseTime                 int     `json:"sunriseTime"`
			SunsetTime                  int     `json:"sunsetTime"`
			MoonPhase                   float64 `json:"moonPhase"`
			PrecipProbability           float64 `json:"precipProbability"`
			PrecipType                  string  `json:"precipType"`
			TemperatureHigh             float64 `json:"temperatureHigh"`
			TemperatureHighTime         int     `json:"temperatureHighTime"`
			TemperatureLow              float64 `json:"temperatureLow"`
			TemperatureLowTime          int     `json:"temperatureLowTime"`
			ApparentTemperatureHigh     float64 `json:"apparentTemperatureHigh"`
			ApparentTemperatureHighTime int     `json:"apparentTemperatureHighTime"`
			ApparentTemperatureLow      float64 `json:"apparentTemperatureLow"`
			ApparentTemperatureLowTime  int     `json:"apparentTemperatureLowTime"`
			DewPoint                    float64 `json:"dewPoint"`
			Humidity                    float64 `json:"humidity"`
			Pressure                    float64 `json:"pressure"`
			WindSpeed                   float64 `json:"windSpeed"`
			WindGust                    float64 `json:"windGust"`
			CloudCover                  float64 `json:"cloudCover"`
			UvIndex                     int     `json:"uvIndex"`
			TemperatureMin              float64 `json:"temperatureMin"`
			TemperatureMinTime          int     `json:"temperatureMinTime"`
			TemperatureMax              float64 `json:"temperatureMax"`
			TemperatureMaxTime          int     `json:"temperatureMaxTime"`
			ApparentTemperatureMin      float64 `json:"apparentTemperatureMin"`
			ApparentTemperatureMinTime  int     `json:"apparentTemperatureMinTime"`
			ApparentTemperatureMax      float64 `json:"apparentTemperatureMax"`
			ApparentTemperatureMaxTime  int     `json:"apparentTemperatureMaxTime"`
		} `json:"data"`
	} `json:"daily"`
	Alerts []struct {
		Title       string `json:"title"`
		Time        int    `json:"time"`
		Expires     int    `json:"expires"`
		Description string `json:"description"`
		URI         string `json:"uri"`
	} `json:"alerts"`
}

type ForecastService struct {
	client *Client
}

// CurrentTemperature will return the current temperature
// in Fahrenheit of the forecast
func (f *Forecast) CurrentTemperature() float64 {
	return f.Currently.Temperature
}

// LocalTime will return the time of the forecast with
// the proper timezone localization based on the GPS
// coordinates.
// It will be in the yyyy-mm-dd hh:mm:ss +-0000 ZONE
func (f *Forecast) LocalTime() (time.Time, error) {
	loc, err := time.LoadLocation(f.Timezone)
	if err != nil {
		return time.Time{}, ErrUnableToLoadTimezone // default value for time.Time is 0001-01-01 00:00:00 +0000 UTC
	}
	tm := time.Unix(int64(f.Currently.Time), 0)
	timeInUTC := time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour(), tm.Minute(), tm.Second(), tm.Nanosecond(), time.UTC)
	localTime := timeInUTC.In(loc)
	return localTime, nil
}

// func NewForecastService(dsKey string) (*ForecastService, error) {
// 	key := os.Getenv(dsKey)
// 	if len(key) == 0 {
// 		err := errors.New(fmt.Sprintf("Dark Sky Env Key (%v) Not Found\n", dsKey))
// 		return nil, err
// 	}
// 	client := NewClient(nil, key)
// 	return &ForecastService{
// 		client,
// 	}, nil
// }
