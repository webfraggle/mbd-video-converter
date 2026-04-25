package job

import (
	"context"
	"sync"
)

// Runner abstracts ffmpeg invocation so the queue can be tested with a fake.
type Runner interface {
	Run(ctx context.Context, j *Job, cfg EncodingConfig, onProgress func(float64)) error
}

type QueueEvent struct {
	JobID    string
	Status   JobStatus
	Progress float64
	Error    string
}

type Queue struct {
	mu     sync.Mutex
	jobs   []*Job
	runner Runner
	events []chan QueueEvent
}

func NewQueue(r Runner) *Queue { return &Queue{runner: r} }

func (q *Queue) Add(j *Job) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs = append(q.jobs, j)
}

func (q *Queue) Snapshot() []Job {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]Job, len(q.jobs))
	for i, j := range q.jobs {
		out[i] = Job{ID: j.ID, InputPath: j.InputPath, OutputPath: j.OutputPath, Status: j.Status, Progress: j.Progress, Error: j.Error}
	}
	return out
}

func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs = nil
}

func (q *Queue) SubscribeEvents(ch chan QueueEvent) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.events = append(q.events, ch)
}

func (q *Queue) emit(j *Job) {
	ev := QueueEvent{JobID: j.ID, Status: j.Status, Progress: j.Progress, Error: j.Error}
	q.mu.Lock()
	subs := append([]chan QueueEvent(nil), q.events...)
	q.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- ev:
		default:
			// drop if full
		}
	}
}

func (q *Queue) CancelJob(id string) {
	q.mu.Lock()
	for _, j := range q.jobs {
		if j.ID == id {
			q.mu.Unlock()
			j.Cancel()
			return
		}
	}
	q.mu.Unlock()
}

// Run processes pending jobs sequentially.
// Skips on error. Honors ctx cancellation between (and during) jobs.
func (q *Queue) Run(ctx context.Context, cfg EncodingConfig) {
	for {
		q.mu.Lock()
		var next *Job
		for _, j := range q.jobs {
			if j.Status == StatusPending {
				next = j
				break
			}
		}
		q.mu.Unlock()

		if next == nil {
			return
		}
		if ctx.Err() != nil {
			next.Status = StatusCancelled
			q.emit(next)
			return
		}

		jobCtx, jobCancel := context.WithCancel(ctx)
		next.setCancel(jobCancel)
		next.Status = StatusRunning
		q.emit(next)

		err := q.runner.Run(jobCtx, next, cfg, func(p float64) {
			next.Progress = p
			q.emit(next)
		})
		// Capture whether the job context was cancelled before we clean it up.
		// This distinguishes "job was cancelled by user" from "job returned an error".
		jobCtxCancelled := jobCtx.Err() != nil
		jobCancel()

		switch {
		case err == nil:
			next.Status = StatusDone
			next.Progress = 1
		case ctx.Err() != nil || jobCtxCancelled:
			next.Status = StatusCancelled
		default:
			next.Status = StatusFailed
			next.Error = err.Error()
		}
		q.emit(next)
	}
}
