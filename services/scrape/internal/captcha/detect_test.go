package captcha

import (
	"testing"
)

func TestDetect_Kinds(t *testing.T) {
	if got := Detect(200, "text/html", []byte("<html>...slider...captcha...</html>")); got != CaptchaSlider {
		t.Errorf("Expected CaptchaSlider, got %v", got)
	}
	if got := Detect(200, "text/html", []byte(`<div id="puzzle-verify">`)); got != CaptchaPuzzle {
		t.Errorf("Expected CaptchaPuzzle, got %v", got)
	}
	if got := Detect(403, "text/html", []byte("please verify")); got != CaptchaVerifyPage {
		t.Errorf("Expected CaptchaVerifyPage, got %v", got)
	}
	if got := Detect(200, "application/json", []byte(`{"error":0}`)); got != CaptchaNone {
		t.Errorf("Expected CaptchaNone, got %v", got)
	}
}
