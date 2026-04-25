package job

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/webfraggle/mbd-video-converter/internal/ffmpeg"
)

type FFmpegRunner struct {
	BinaryPath string
}

func (r *FFmpegRunner) Run(ctx context.Context, j *Job, cfg EncodingConfig, onProgress func(float64)) error {
	out := cfg.ResolveOutputPath(j.InputPath)

	switch cfg.OnExist {
	case "fail":
		if _, err := os.Stat(out); err == nil {
			return fmt.Errorf("output already exists: %s", out)
		}
	case "suffix":
		out = applySuffixForCollision(out, func(p string) bool {
			_, err := os.Stat(p)
			return err == nil
		})
	default: // "overwrite" or unset
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	j.OutputPath = out

	dur, _ := ffmpeg.Probe(ctx, r.BinaryPath, j.InputPath)

	args := ffmpeg.BuildArgs(ffmpeg.EncodingInput{
		InputPath: j.InputPath, OutputPath: out,
		Width: cfg.Width, Height: cfg.Height, FPS: cfg.FPS, Quality: cfg.Quality,
		Saturation: cfg.Saturation, Gamma: cfg.Gamma, Scaler: cfg.Scaler,
		AdvancedArgs: cfg.AdvancedArgs,
	})

	updates := make(chan ffmpeg.ProgressUpdate, 8)
	go func() {
		for u := range updates {
			onProgress(u.Ratio)
		}
	}()

	log.Printf("ffmpeg: %s %v", r.BinaryPath, args)
	res := ffmpeg.Run(ctx, r.BinaryPath, args, dur, updates)
	if res.Err != nil {
		log.Printf("ffmpeg: exit=%d err=%v\n%s", res.ExitCode, res.Err, res.Stderr)
		// Clean up partial output on failure / cancellation.
		_ = os.Remove(out)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("ffmpeg exit %d: %s", res.ExitCode, lastLine(res.Stderr))
	}
	return nil
}

func lastLine(s string) string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '\n' {
			return s[i+1:]
		}
	}
	return s
}
