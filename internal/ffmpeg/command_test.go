package ffmpeg

import (
	"reflect"
	"testing"
)

func TestBuildArgs_Factory105(t *testing.T) {
	in := EncodingInput{
		InputPath:  "video.mp4",
		OutputPath: "video_1.05_20fps.mjpeg",
		Width:      120, Height: 240, FPS: 20, Quality: 9,
		Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos",
	}
	want := []string{
		"-y",
		"-i", "video.mp4",
		"-vf", "fps=20,scale=120:240:flags=lanczos,eq=saturation=2.5:gamma=0.8",
		"-q:v", "9",
		"-progress", "pipe:2",
		"-nostats",
		"-loglevel", "error",
		"video_1.05_20fps.mjpeg",
	}
	got := BuildArgs(in)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestBuildArgs_AdvancedOverridesFilterBlock(t *testing.T) {
	in := EncodingInput{
		InputPath:    "in.mp4",
		OutputPath:   "out.mjpeg",
		AdvancedArgs: `-vf "fps=15,scale=80:160" -q:v 5`,
	}
	want := []string{
		"-y",
		"-i", "in.mp4",
		"-vf", "fps=15,scale=80:160",
		"-q:v", "5",
		"-progress", "pipe:2",
		"-nostats",
		"-loglevel", "error",
		"out.mjpeg",
	}
	got := BuildArgs(in)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestBuildArgs_FactoryAllFour(t *testing.T) {
	cases := []struct {
		name string
		in   EncodingInput
		vf   string
		q    string
	}{
		{"0.96", EncodingInput{Width: 80, Height: 160, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
			"fps=20,scale=80:160:flags=lanczos,eq=saturation=2.5:gamma=0.8", "9"},
		{"1.05", EncodingInput{Width: 120, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
			"fps=20,scale=120:240:flags=lanczos,eq=saturation=2.5:gamma=0.8", "9"},
		{"1.14", EncodingInput{Width: 135, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
			"fps=20,scale=135:240:flags=lanczos,eq=saturation=2.5:gamma=0.8", "9"},
		{"1.90", EncodingInput{Width: 120, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
			"fps=20,scale=120:240:flags=lanczos,eq=saturation=2.5:gamma=0.8", "9"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.in.InputPath = "in.mp4"
			tc.in.OutputPath = "out.mjpeg"
			args := BuildArgs(tc.in)
			vf := argAfter(args, "-vf")
			q := argAfter(args, "-q:v")
			if vf != tc.vf {
				t.Errorf("vf got %q want %q", vf, tc.vf)
			}
			if q != tc.q {
				t.Errorf("q got %q want %q", q, tc.q)
			}
		})
	}
}

func argAfter(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}
