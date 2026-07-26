package sms

// Guard rejects non–high-value / non-OTP SMS (cost model §3.6).
func Guard(msg Message) error {
	if msg.HighValue || msg.OTP {
		return nil
	}
	return errNotHighValue
}

type guardError string

func (e guardError) Error() string { return string(e) }

const errNotHighValue = guardError("sms: only high_value or otp allowed")
