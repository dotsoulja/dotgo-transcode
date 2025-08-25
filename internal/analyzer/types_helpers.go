package analyzer

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexibleFloat64 handles JSON values that may be either float64 or string.
type FlexibleFloat64 float64

func (f *FlexibleFloat64) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as float64
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		*f = FlexibleFloat64(num)
		return nil
	}

	// Try to unmarshal as a string and parse
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		parsed, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return fmt.Errorf("invalid float string: %s", str)
		}
		*f = FlexibleFloat64(parsed)
		return nil
	}

	return fmt.Errorf("unsupported JSON value for FlexibleFloat64: %s", string(data))
}
