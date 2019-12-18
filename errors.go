package darksky

import "errors"

var (
	// ErrBadRequest is returned when the API server returns a
	// status code that is >= 400
	ErrBadRequest = errors.New("Bad HTTP Request")

	// ErrUnableToLoadTimezone is returned when the timezone information
	// cannot be loaded or parsed from the Forecast
	ErrUnableToLoadTimezone = errors.New("Unable to Load Timezone Data")
)

// type Error struct {
// 	Code string `json:"code"`
// }

// func (err Error) Error() string {
// 	return fmt.Sprintf("%s.", err.Code)
// }

// func (err Error) MarshalJSON() ([]byte, error) {
// 	var tmp struct {
// 		Error struct {
// 			Code string `json:"code"`
// 		} `json:"error"`
// 	}
// 	tmp.Error = err
// 	return json.Marshal(tmp)
// }

// func (err *Error) UnmarshalJSON(data []byte) error {
// 	var tmp struct {
// 		Error struct {
// 			Code string `json:"code"`
// 		} `json:"error"`
// 	}
// 	if err := json.Unmarshal(data, &tmp); err != nil {
// 		return err
// 	}
// 	*err = tmp.Error
// 	return nil
// }
