package captcha

import (
	"bytes"
)

type CaptchaKind int

const (
	CaptchaNone   CaptchaKind = iota
	CaptchaSlider             // Shopee slider/puzzle
	CaptchaPuzzle
	CaptchaVerifyPage
)

// Detect nhận diện CAPTCHA từ response (markers HTML / status / body).
func Detect(status int, contentType string, body []byte) CaptchaKind {
	switch {
	case bytes.Contains(body, []byte("slider")) && bytes.Contains(body, []byte("captcha")):
		return CaptchaSlider
	case bytes.Contains(body, []byte("puzzle-verify")):
		return CaptchaPuzzle
	case status == 403 && bytes.Contains(body, []byte("verify")):
		return CaptchaVerifyPage
	default:
		return CaptchaNone
	}
}
