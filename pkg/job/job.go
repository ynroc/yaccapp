package job

import (
	"context"
	"time"
)

// JobExec represents the implementation of a Job to be executed.
type JobExec interface {
	Execute(ctx context.Context, progress *Progress)
}

type jobExecImpl struct {
	fn func(ctx context.Context, progress *Progress)
}

func (j *jobExecImpl) Execute(ctx context.Context, progress *Progress) {
	j.fn(ctx, progress)
}

// MakeJobExec returns a simple JobExec implementation using the provided
// function.
func MakeJobExec(fn func(ctx context.Context, progress *Progress)) JobExec {
	return &jobExecImpl{
		fn: fn,
	}
}

// Status is the status of a Job
type Status string

const (
	// StatusReady means that the Job is not yet started.
	StatusReady Status = "READY"
	// StatusRunning means that the job is currently running.
	StatusRunning Status = "RUNNING"
	// StatusStopping means that the job is cancelled but is still running.
	StatusStopping Status = "STOPPING"
	// StatusFinished means that the job was completed.
	StatusFinished Status = "FINISHED"
	// StatusCancelled means that the job was cancelled and is now stopped.
	StatusCancelled Status = "CANCELLED"
)

// Job represents the status of a queued or running job.
type Job struct {
	ID     int
	Status Status
	// details of the current operations of the job
	Details     []string
	Description string
	// Progress in terms of 0 - 1.
	Progress  float64
	StartTime *time.Time
	EndTime   *time.Time
	AddTime   time.Time

	outerCtx   context.Context
	exec       JobExec
	cancelFunc context.CancelFunc
}

// TimeElapsed returns the total time elapsed for the job.
// If the EndTime is set, then it uses this to calculate the elapsed time, otherwise it uses time.Now.
func (j *Job) TimeElapsed() time.Duration {
	var end time.Time
	if j.EndTime != nil {
		end = time.Now()
	} else {
		end = *j.EndTime
	}

	return end.Sub(*j.StartTime)
}

func (j *Job) cancel() {
	if j.Status == StatusReady {
		j.Status = StatusCancelled
	} else if j.Status == StatusRunning {
		j.Status = StatusStopping
	}

	if j.cancelFunc != nil {
		j.cancelFunc()
	}
}

// IsCancelled returns true if cancel has been called on the context.
func IsCancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
