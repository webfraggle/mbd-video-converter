//go:build integration

package ffmpeg

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestRunner_RealFFmpeg_Roundtrip(t *testing.T) {
	bin, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not in PATH")
	}
	dir := t.TempDir()
	in := filepath.Join(dir, "in.mp4")
	// Generate a 1-second test pattern.
	gen := exec.Command(bin, "-y", "-f", "lavfi", "-i", "testsrc=duration=1:size=120x240:rate=20", in)
	gen.Stderr = os.Stderr
	if err := gen.Run(); err != nil {
		t.Fatalf("setup: %v", err)
	}

	out := filepath.Join(dir, "out.mjpeg")
	args := BuildArgs(EncodingInput{
		InputPath: in, OutputPath: out,
		Width: 120, Height: 240, FPS: 20, Quality: 9,
		Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos",
	})
	updates := make(chan ProgressUpdate, 32)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res := Run(ctx, bin, args, 1_000_000, updates)
	if res.Err != nil {
		t.Fatalf("run: %v\nstderr:\n%s", res.Err, res.Stderr)
	}
	st, err := os.Stat(out)
	if err != nil {
		t.Fatalf("output: %v", err)
	}
	if st.Size() < 1000 {
		t.Errorf("output suspiciously small: %d bytes", st.Size())
	}
}
