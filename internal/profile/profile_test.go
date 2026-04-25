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

func TestFactoryProfiles(t *testing.T) {
	want := []Profile{
		{ID: "factory:0.96", Name: "0.96\" Display", Factory: true, Width: 80, Height: 160, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
		{ID: "factory:1.05", Name: "1.05\" Display", Factory: true, Width: 120, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
		{ID: "factory:1.14", Name: "1.14\" Display", Factory: true, Width: 135, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
		{ID: "factory:1.90", Name: "1.90\" Display", Factory: true, Width: 120, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
	}
	got := Factory()
	if len(got) != len(want) {
		t.Fatalf("got %d profiles, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("profile %d:\n got=%+v\nwant=%+v", i, got[i], want[i])
		}
	}
	for _, p := range got {
		if err := p.Validate(); err != nil {
			t.Errorf("factory profile %s invalid: %v", p.ID, err)
		}
	}
}
