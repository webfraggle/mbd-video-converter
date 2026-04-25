package job

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeRunner struct {
	mu     sync.Mutex
	calls  []string // input paths
	failOn map[string]error
	delay  time.Duration
}

func (f *fakeRunner) Run(ctx context.Context, j *Job, cfg EncodingConfig, onProgress func(float64)) error {
	f.mu.Lock()
	f.calls = append(f.calls, j.InputPath)
	wantErr, hasErr := f.failOn[j.InputPath]
	d := f.delay
	f.mu.Unlock()

	if d > 0 {
		select {
		case <-time.After(d):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if hasErr {
		return wantErr
	}
	onProgress(1.0)
	return nil
}

func TestQueueSequentialAndSkipOnError(t *testing.T) {
	r := &fakeRunner{failOn: map[string]error{"b.mp4": errors.New("boom")}}
	q := NewQueue(r)
	q.Add(&Job{InputPath: "a.mp4"})
	q.Add(&Job{InputPath: "b.mp4"})
	q.Add(&Job{InputPath: "c.mp4"})

	updates := make(chan QueueEvent, 32)
	q.SubscribeEvents(updates)

	done := make(chan struct{})
	go func() {
		q.Run(context.Background(), EncodingConfig{})
		close(done)
	}()
	<-done

	jobs := q.Snapshot()
	if jobs[0].Status != StatusDone {
		t.Errorf("a: status=%v err=%q", jobs[0].Status, jobs[0].Error)
	}
	if jobs[1].Status != StatusFailed || jobs[1].Error == "" {
		t.Errorf("b: status=%v err=%q", jobs[1].Status, jobs[1].Error)
	}
	if jobs[2].Status != StatusDone {
		t.Errorf("c: status=%v err=%q", jobs[2].Status, jobs[2].Error)
	}

	if len(r.calls) != 3 || r.calls[0] != "a.mp4" || r.calls[1] != "b.mp4" || r.calls[2] != "c.mp4" {
		t.Errorf("calls = %v want [a b c]", r.calls)
	}
}

func TestQueueCancelStopsAfterCurrent(t *testing.T) {
	r := &fakeRunner{delay: 200 * time.Millisecond}
	q := NewQueue(r)
	q.Add(&Job{InputPath: "a.mp4"})
	q.Add(&Job{InputPath: "b.mp4"})

	ctx, cancel := context.WithCancel(context.Background())
	go q.Run(ctx, EncodingConfig{})
	time.Sleep(50 * time.Millisecond)
	cancel()
	time.Sleep(300 * time.Millisecond)

	jobs := q.Snapshot()
	if jobs[0].Status != StatusCancelled && jobs[0].Status != StatusDone {
		t.Errorf("first job: status %v", jobs[0].Status)
	}
	if jobs[1].Status != StatusPending && jobs[1].Status != StatusCancelled {
		t.Errorf("second job: status %v (should not have run)", jobs[1].Status)
	}
}

func TestQueueCancelOne(t *testing.T) {
	r := &fakeRunner{delay: 200 * time.Millisecond}
	q := NewQueue(r)
	q.Add(&Job{InputPath: "a.mp4"})
	q.Add(&Job{InputPath: "b.mp4"})

	go q.Run(context.Background(), EncodingConfig{})
	time.Sleep(50 * time.Millisecond)
	q.CancelJob(q.Snapshot()[0].ID)
	time.Sleep(400 * time.Millisecond)

	jobs := q.Snapshot()
	if jobs[0].Status != StatusCancelled {
		t.Errorf("first job: status %v want cancelled", jobs[0].Status)
	}
	if jobs[1].Status != StatusDone {
		t.Errorf("second job: status %v want done (queue continues)", jobs[1].Status)
	}
}
