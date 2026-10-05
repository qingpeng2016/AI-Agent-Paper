package errorx

import "testing"

func TestRespErr_LocalizedMessage(t *testing.T) {
	err := NewRespErr(100100, "参数错误", "參數錯誤", "Params error.")

	cases := []struct {
		locale string
		want   string
	}{
		{"zh-Hans", "参数错误"},
		{"zh-CN", "参数错误"},
		{"zh-Hant", "參數錯誤"},
		{"zh-TW", "參數錯誤"},
		{"en", "Params error."},
		{"en-US", "Params error."},
		{"", "参数错误"},
	}
	for _, tc := range cases {
		if got := err.LocalizedMessage(tc.locale); got != tc.want {
			t.Fatalf("locale=%q got=%q want=%q", tc.locale, got, tc.want)
		}
	}
}

func TestNormalizeLocale(t *testing.T) {
	if got := NormalizeLocale("zh-HK"); got != LocaleHant {
		t.Fatalf("got %s", got)
	}
	if got := NormalizeLocale("en-GB"); got != LocaleEn {
		t.Fatalf("got %s", got)
	}
}
