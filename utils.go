package firevault

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// asInt returns the parameter as a int64, or error if it can't convert
func asInt(param string) (int64, error) {
	i, err := strconv.ParseInt(param, 0, 64)
	if err != nil {
		return 0, errors.New("firevault: " + err.Error())
	}

	return i, nil
}

// asUint returns the parameter as a uint64, or error if it can't convert
func asUint(param string) (uint64, error) {
	u, err := strconv.ParseUint(param, 0, 64)
	if err != nil {
		return 0, errors.New("firevault: " + err.Error())
	}

	return u, nil
}

// asFloat returns the parameter as a float64, or error if it can't convert
func asFloat(param string) (float64, error) {
	f, err := strconv.ParseFloat(param, 32)
	if err != nil {
		return 0, errors.New("firevault: " + err.Error())
	}

	return f, nil
}

// asTime returns the parameter as a time.Time, or error if it can't convert
func asTime(param string) (time.Time, error) {
	params := strings.Split(param, "|")
	if len(params) < 2 {
		return time.Time{}, errors.New("firevault: time param must be in the format 'layout|value'")
	}

	t, err := time.Parse(params[0], params[1])
	if err != nil {
		return time.Time{}, errors.New("firevault: " + err.Error())
	}

	return t, nil
}
