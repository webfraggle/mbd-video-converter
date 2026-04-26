package job

import "testing"

func TestResolveOutputPath(t *testing.T) {
	cfg := EncodingConfig{
		ProfileNameForPattern: "1.05",
		Width: 120, Height: 240, FPS: 20,
		FilenamePattern: "{name}_{profile}_{fps}fps",
		OutputDir:       "/out",
	}
	got := cfg.ResolveOutputPath("/in/My Video.mp4")
	want := "/out/My Video_1.05_20fps.mjpeg"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestResolveOutputPath_DefaultDirIsInputDir(t *testing.T) {
	cfg := EncodingConfig{
		ProfileNameForPattern: "1.90",
		Width: 120, Height: 240, FPS: 20,
		FilenamePattern: "{name}_{profile}",
		OutputDir:       "",
	}
	got := cfg.ResolveOutputPath("/in/clip.mp4")
	want := "/in/clip_1.90.mjpeg"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestResolveOutputPath_AllPlaceholders(t *testing.T) {
	cfg := EncodingConfig{
		ProfileNameForPattern: "X",
		Width: 80, Height: 160, FPS: 25,
		FilenamePattern: "{name}-{w}x{h}-{fps}-{profile}",
	}
	got := cfg.ResolveOutputPath("/foo/bar.mov")
	want := "/foo/bar-80x160-25-X.mjpeg"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestApplyOnExist_Suffix(t *testing.T) {
	got := applySuffixForCollision("/x/foo.mjpeg", func(p string) bool {
		return p == "/x/foo.mjpeg" || p == "/x/foo (1).mjpeg"
	})
	if got != "/x/foo (2).mjpeg" {
		t.Errorf("got %q want /x/foo (2).mjpeg", got)
	}
}
