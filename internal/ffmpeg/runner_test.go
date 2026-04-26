package ffmpeg

import "testing"

func TestParseProgressLine(t *testing.T) {
	cases := []struct {
		in       string
		key, val string
		ok       bool
	}{
		{"out_time_ms=12345", "out_time_ms", "12345", true},
		{"frame=42", "frame", "42", true},
		{"progress=continue", "progress", "continue", true},
		{"weird line without equals", "", "", false},
		{"", "", "", false},
	}
	for _, tc := range cases {
		k, v, ok := parseProgressLine(tc.in)
		if k != tc.key || v != tc.val || ok != tc.ok {
			t.Errorf("parseProgressLine(%q) = (%q,%q,%v) want (%q,%q,%v)",
				tc.in, k, v, ok, tc.key, tc.val, tc.ok)
		}
	}
}

func TestProgressRatio(t *testing.T) {
	if r := progressRatio(5_000_000, 10_000_000); r < 0.49 || r > 0.51 {
		t.Errorf("ratio %v not ~0.5", r)
	}
	if r := progressRatio(0, 0); r != 0 {
		t.Errorf("ratio for zero duration = %v, want 0", r)
	}
	if r := progressRatio(20_000_000, 10_000_000); r != 1.0 {
		t.Errorf("ratio capped at 1, got %v", r)
	}
}

func TestParseDurationLine(t *testing.T) {
	cases := []struct {
		in   string
		want int64
		ok   bool
	}{
		{"  Duration: 00:01:23.45, start: 0.000000, bitrate: ...", int64((1*60+23)*1_000_000 + 450_000), true},
		{"  Duration: 00:00:01.00, ...", 1_000_000, true},
		{"random line", 0, false},
	}
	for _, tc := range cases {
		got, ok := parseDurationLine(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Errorf("parseDurationLine(%q) = (%d, %v); want (%d, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}
