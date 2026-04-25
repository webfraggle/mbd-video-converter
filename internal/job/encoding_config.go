package job

import (
	"fmt"
	"path/filepath"
	"strings"
)

type EncodingConfig struct {
	ProfileNameForPattern string  // for {profile}
	Width, Height         int
	FPS                   int
	Quality               int
	Saturation, Gamma     float64
	Scaler                string
	AdvancedArgs          string
	OutputDir             string // "" → input dir
	FilenamePattern       string
	OnExist               string // "overwrite" | "suffix" | "fail"
}

// ResolveOutputPath resolves {name}/{profile}/{fps}/{w}/{h} placeholders, sets ".mjpeg" extension.
func (c EncodingConfig) ResolveOutputPath(inputPath string) string {
	dir := c.OutputDir
	if dir == "" {
		dir = filepath.Dir(inputPath)
	}
	base := filepath.Base(inputPath)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	r := strings.NewReplacer(
		"{name}", name,
		"{profile}", c.ProfileNameForPattern,
		"{fps}", fmt.Sprintf("%d", c.FPS),
		"{w}", fmt.Sprintf("%d", c.Width),
		"{h}", fmt.Sprintf("%d", c.Height),
	)
	resolved := r.Replace(c.FilenamePattern)
	return filepath.Join(dir, resolved+".mjpeg")
}

// applySuffixForCollision returns a path with " (N)" appended before the extension
// where N is the smallest positive integer that does not collide. exists tells us
// whether a candidate path already exists (injectable for tests).
func applySuffixForCollision(path string, exists func(string) bool) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	for i := 1; i < 10000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", stem, i, ext))
		if !exists(candidate) {
			return candidate
		}
	}
	return path
}
