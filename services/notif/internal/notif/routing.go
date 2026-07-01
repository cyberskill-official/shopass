package notif

var channelRank = map[string]int{
	"push":  0,
	"email": 1,
	"sms":   2,
}

type UserChannels struct {
	Push  bool
	Email bool
	SMS   bool
}

func ResolveChannel(desired []string, caps UserChannels, highValue bool) (string, bool) {
	avail := func(c string) bool {
		switch c {
		case "push":
			return caps.Push
		case "email":
			return caps.Email
		case "sms":
			return caps.SMS && (highValue || (!caps.Push && !caps.Email))
		}
		return false
	}

	best, found := "", false
	for _, c := range desired {
		if !avail(c) {
			continue
		}
		if !found || channelRank[c] < channelRank[best] {
			best, found = c, true
		}
	}
	return best, found
}
