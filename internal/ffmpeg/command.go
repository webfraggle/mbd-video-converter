package ffmpeg

import (
	"fmt"

	"github.com/google/shlex"
)

type EncodingInput struct {
	InputPath    string
	OutputPath   string
	Width        int
	Height       int
	FPS          int
	Quality      int
	Saturation   float64
	Gamma        float64
	Scaler       string
	AdvancedArgs string // optional, replaces the filter+quality block
}

// BuildArgs returns the ffmpeg argument list (excluding the binary name itself).
func BuildArgs(in EncodingInput) []string {
	args := []string{"-y", "-i", in.InputPath}

	if in.AdvancedArgs != "" {
		mid, err := shlex.Split(in.AdvancedArgs)
		if err != nil {
			// Unparseable input: fall back to generated filter block.
			args = append(args, generatedFilterBlock(in)...)
		} else {
			args = append(args, mid...)
		}
	} else {
		args = append(args, generatedFilterBlock(in)...)
	}

	args = append(args, "-progress", "pipe:2", "-nostats", "-loglevel", "error", in.OutputPath)
	return args
}

func generatedFilterBlock(in EncodingInput) []string {
	vf := fmt.Sprintf("fps=%d,scale=%d:%d:flags=%s,eq=saturation=%s:gamma=%s",
		in.FPS, in.Width, in.Height, in.Scaler,
		formatFloat(in.Saturation), formatFloat(in.Gamma))
	return []string{"-vf", vf, "-q:v", fmt.Sprintf("%d", in.Quality)}
}

func formatFloat(f float64) string {
	// Trim trailing zeros: 2.5 → "2.5", 2.0 → "2", 0.8 → "0.8".
	s := fmt.Sprintf("%g", f)
	return s
}
