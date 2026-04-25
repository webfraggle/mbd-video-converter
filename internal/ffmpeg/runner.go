package ffmpeg

import (
	"bufio"
	"context"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// ProgressUpdate is delivered while ffmpeg is encoding.
type ProgressUpdate struct {
	Ratio float64 // 0..1; 0 if total duration unknown
	Frame int     // most recent frame count (-1 if unknown)
}

// RunResult is delivered exactly once after the process exits.
type RunResult struct {
	Err      error
	Stderr   string // tail of stderr lines (last ~20)
	ExitCode int
}

// Run starts ffmpeg with args, parses progress, reports updates on updates.
// totalDurationMicros is the source video's duration in microseconds (use 0 if unknown).
// Returns when the process ends or ctx is cancelled.
func Run(ctx context.Context, ffmpegPath string, args []string, totalDurationMicros int64, updates chan<- ProgressUpdate) RunResult {
	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	applyHideWindow(cmd) // platform-specific (windows: hide console)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return RunResult{Err: err, ExitCode: -1}
	}

	if err := cmd.Start(); err != nil {
		return RunResult{Err: err, ExitCode: -1}
	}

	tail := newRingTail(20)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := bufio.NewScanner(stderr)
		s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		var frame int = -1
		for s.Scan() {
			line := s.Text()
			tail.push(line)
			k, v, ok := parseProgressLine(line)
			if !ok {
				continue
			}
			switch k {
			case "frame":
				if n, err := strconv.Atoi(v); err == nil {
					frame = n
				}
			case "out_time_ms":
				if cur, err := strconv.ParseInt(v, 10, 64); err == nil {
					select {
					case updates <- ProgressUpdate{Ratio: progressRatio(cur, totalDurationMicros), Frame: frame}:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	waitErr := cmd.Wait()
	wg.Wait()
	close(updates)

	exitCode := 0
	if waitErr != nil {
		exitCode = -1
		if ee, ok := waitErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		}
	}

	return RunResult{
		Err:      waitErr,
		Stderr:   strings.Join(tail.snapshot(), "\n"),
		ExitCode: exitCode,
	}
}

func parseProgressLine(line string) (string, string, bool) {
	if line == "" {
		return "", "", false
	}
	i := strings.IndexByte(line, '=')
	if i <= 0 {
		return "", "", false
	}
	return line[:i], line[i+1:], true
}

func progressRatio(currentMicros, totalMicros int64) float64 {
	if totalMicros <= 0 {
		return 0
	}
	r := float64(currentMicros) / float64(totalMicros)
	if r > 1 {
		return 1
	}
	if r < 0 {
		return 0
	}
	return r
}

type ringTail struct {
	buf  []string
	max  int
	next int
	full bool
}

func newRingTail(n int) *ringTail { return &ringTail{buf: make([]string, n), max: n} }

func (r *ringTail) push(s string) {
	r.buf[r.next] = s
	r.next = (r.next + 1) % r.max
	if r.next == 0 {
		r.full = true
	}
}

func (r *ringTail) snapshot() []string {
	if !r.full {
		return append([]string(nil), r.buf[:r.next]...)
	}
	out := make([]string, 0, r.max)
	out = append(out, r.buf[r.next:]...)
	out = append(out, r.buf[:r.next]...)
	return out
}

var durationRE = regexp.MustCompile(`Duration:\s+(\d+):(\d+):(\d+)\.(\d+)`)

// Probe runs `ffmpeg -i <input>` and parses the Duration line from stderr.
// Returns 0 if the duration cannot be determined.
func Probe(ctx context.Context, ffmpegPath, inputPath string) (int64, error) {
	cmd := exec.CommandContext(ctx, ffmpegPath, "-i", inputPath)
	applyHideWindow(cmd)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 0, err
	}
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	var dur int64
	s := bufio.NewScanner(stderr)
	for s.Scan() {
		if d, ok := parseDurationLine(s.Text()); ok {
			dur = d
		}
	}
	// ffmpeg returns non-zero when no output is specified — that is expected.
	_ = cmd.Wait()
	return dur, nil
}

func parseDurationLine(line string) (int64, bool) {
	m := durationRE.FindStringSubmatch(line)
	if m == nil {
		return 0, false
	}
	h, _ := strconv.Atoi(m[1])
	mi, _ := strconv.Atoi(m[2])
	s, _ := strconv.Atoi(m[3])
	frac, _ := strconv.Atoi(m[4])
	// Normalize frac to microseconds: ffmpeg emits 2-digit hundredths typically; pad.
	micros := frac
	for i := len(m[4]); i < 6; i++ {
		micros *= 10
	}
	for i := len(m[4]); i > 6; i-- {
		micros /= 10
	}
	total := int64(h*3600+mi*60+s)*1_000_000 + int64(micros)
	return total, true
}
