package regime

import "errors"

var (
	ErrCountryNotOpen     = errors.New("regime: country not open")
	ErrUnsupportedRegime  = errors.New("regime: unsupported")
)

type Profile struct {
	Code        string
	Notes       []string
	BreachHours int
	DSARDays    int
}

type Adapter interface {
	Code() string
	Profile() Profile
}
