package errorx

import "testing"

func TestShouldAlertHTTPError(t *testing.T) {
	if ShouldAlertHTTPError(ErrParamsError) {
		t.Fatal("bare params should not alert")
	}
	if !ShouldAlertHTTPError(ErrParamsError.WithDetail("Error 1406")) {
		t.Fatal("params with detail should alert")
	}
	if ShouldAlertHTTPError(ErrTopicStepInvalid) {
		t.Fatal("topic step invalid should not alert")
	}
	if !ShouldAlertHTTPError(ErrUnknown) {
		t.Fatal("unknown should alert")
	}
}
