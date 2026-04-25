package profile

import "testing"

func TestProfileValidate(t *testing.T) {
	cases := []struct {
		name    string
		p       Profile
		wantErr bool
	}{
		{"valid", Profile{Name: "x", Width: 120, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"}, false},
		{"empty name", Profile{Width: 120, Height: 240, FPS: 20, Quality: 9, Scaler: "lanczos"}, true},
		{"zero width", Profile{Name: "x", Width: 0, Height: 240, FPS: 20, Quality: 9, Scaler: "lanczos"}, true},
		{"negative height", Profile{Name: "x", Width: 120, Height: -1, FPS: 20, Quality: 9, Scaler: "lanczos"}, true},
		{"fps zero", Profile{Name: "x", Width: 120, Height: 240, FPS: 0, Quality: 9, Scaler: "lanczos"}, true},
		{"quality below range", Profile{Name: "x", Width: 120, Height: 240, FPS: 20, Quality: 0, Scaler: "lanczos"}, true},
		{"quality above range", Profile{Name: "x", Width: 120, Height: 240, FPS: 20, Quality: 32, Scaler: "lanczos"}, true},
		{"empty scaler", Profile{Name: "x", Width: 120, Height: 240, FPS: 20, Quality: 9, Scaler: ""}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.p.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
