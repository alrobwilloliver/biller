package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"
)

type fakeBackgroundTask struct {
	runCallSpy int
	// pidSpy     int
	fakeRun func(f *fakeBackgroundTask, ctx context.Context) error
}

func (f *fakeBackgroundTask) Run(ctx context.Context) error {
	err := f.fakeRun(f, ctx)

	return err
}

// These tests use `time.Sleep()`, so they shouldn't run in short mode
func isShort(t *testing.T) {
	t.Helper()

	if testing.Short() {
		t.SkipNow()
	}
}

// Test we are calling run the correct amount of times in a set period
func Test_Background_RunCount(t *testing.T) {
	isShort(t)

	const testInterval = time.Millisecond * 15
	const waitTime = time.Millisecond * 100

	backgroundTask := &fakeBackgroundTask{fakeRun: func(f *fakeBackgroundTask, ctx context.Context) error {
		f.runCallSpy++
		return nil
	}}

	svc := NewBackground(
		BackgroundServiceConfig{
			Environment: "local",
			Interval:    testInterval,
			Name:        "testBackgroundTask",
			Task:        backgroundTask,
		},
		prometheus.NewRegistry(),
		zaptest.NewLogger(t),
	)

	ctx, cancel := context.WithCancel(context.Background())
	errg, ctx := errgroup.WithContext(ctx)
	errg.Go(func() error {
		return svc.Run(ctx)
	})
	// wait for (*Runner).Run() to be called x amount of times
	time.Sleep(waitTime)
	cancel()
	err := errg.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected error running ticker: %s", err)
	}

	want := int(waitTime)/int(testInterval) + 1
	got := backgroundTask.runCallSpy

	if got != want {
		t.Errorf(".Run is not being called enough times, want %d, got %d", want, got)
	}
}

// Test our context has a deadline set with the correct value
func Test_Background_RunContextDeadline(t *testing.T) {
	isShort(t)

	const testInterval = time.Millisecond * 20
	const waitTime = time.Millisecond * 25
	const acceptableDifference = time.Millisecond * 5

	backgroundTask := &fakeBackgroundTask{
		fakeRun: func(f *fakeBackgroundTask, ctx context.Context) error {
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Error("This context should have a deadline")
			}

			difference := time.Until(deadline)

			if !(difference >= testInterval-acceptableDifference && difference <= testInterval+acceptableDifference) {
				t.Errorf("The deadline should be ~%s±%s, got %s", testInterval, acceptableDifference, difference)
			}

			return nil
		}}

	svc := NewBackground(
		BackgroundServiceConfig{
			Environment: "local",
			Interval:    testInterval,
			Name:        "testBackgroundTask",
			Task:        backgroundTask,
		},
		prometheus.NewRegistry(),
		zaptest.NewLogger(t),
	)

	ctx, cancel := context.WithCancel(context.Background())
	errg, ctx := errgroup.WithContext(ctx)
	errg.Go(func() error {
		return svc.Run(ctx)
	})
	time.Sleep(waitTime)
	cancel()
	err := errg.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected error running task: %s", err)
	}
}

// Test the context we have expires when we take too long to run
func Test_Background_RunTimeout(t *testing.T) {
	isShort(t)

	const testInterval = time.Millisecond * 50
	const waitTime = time.Millisecond * 60

	backgroundTask := &fakeBackgroundTask{
		fakeRun: func(f *fakeBackgroundTask, ctx context.Context) error {
			// in this test we only care about the first call, as in during the following calls
			// we cancel the context by stopping the service, so it will return a different error,
			// `context.Cancelled`, instead of `context.DeadlineExceeded`
			if f.runCallSpy >= 1 {
				return nil
			}

			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				t.Fatalf("context should be valid, got '%v'", ctx.Err())
			}

			// this should cause our context deadline to be exceeded
			time.Sleep(time.Millisecond * 55)

			if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
				t.Fatalf("context should be expired, got '%v'", ctx.Err())
			}

			f.runCallSpy++

			return ctx.Err()
		}}

	svc := NewBackground(
		BackgroundServiceConfig{
			Environment: "local",
			Interval:    testInterval,
			Name:        "testBackgroundTask",
			Task:        backgroundTask,
		},
		prometheus.NewRegistry(),
		zaptest.NewLogger(t),
	)

	ctx, cancel := context.WithCancel(context.Background())
	errg, ctx := errgroup.WithContext(ctx)
	errg.Go(func() error {
		return svc.Run(ctx)
	})

	time.Sleep(waitTime)
	cancel()
}

// Test when the task returns an error, the task should be run again/retried
func Test_Background_RunError(t *testing.T) {
	isShort(t)

	const testInterval = time.Millisecond * 20
	const waitTime = time.Millisecond * 45

	backgroundTask := &fakeBackgroundTask{
		fakeRun: func(f *fakeBackgroundTask, ctx context.Context) error {
			f.runCallSpy++
			return fmt.Errorf("an error happened")
		}}

	svc := NewBackground(
		BackgroundServiceConfig{
			Environment: "local",
			Interval:    testInterval,
			Name:        "testBackgroundTask",
			Task:        backgroundTask,
		},
		prometheus.NewRegistry(),
		zaptest.NewLogger(t),
	)

	ctx, cancel := context.WithCancel(context.Background())
	errg, ctx := errgroup.WithContext(ctx)
	errg.Go(func() error {
		return svc.Run(ctx)
	})
	time.Sleep(waitTime)
	cancel()

	want := 3
	got := backgroundTask.runCallSpy

	if got != want {
		t.Errorf("Want %d (*Runner).Run() calls, got %d", want, got)
	}
}

// Test when the parent context is cancelled
func Test_Background_ParentCancel(t *testing.T) {
	isShort(t)

	const testInterval = time.Millisecond * 15
	const waitTime = time.Millisecond * 50

	backgroundTask := &fakeBackgroundTask{
		fakeRun: func(f *fakeBackgroundTask, ctx context.Context) error {
			f.runCallSpy++
			return nil
		}}

	ctx, cancel := context.WithCancel(context.Background())

	svc := NewBackground(
		BackgroundServiceConfig{
			Environment: "local",
			Interval:    testInterval,
			Name:        "testBackgroundTask",
			Task:        backgroundTask,
		},
		prometheus.NewRegistry(),
		zaptest.NewLogger(t),
	)

	cancel()

	errg, ctx := errgroup.WithContext(ctx)
	errg.Go(func() error {
		return svc.Run(ctx)
	})
	time.Sleep(waitTime)
	cancel()

	if backgroundTask.runCallSpy != 1 {
		t.Error("If the parent context is cancelled the task should only be called once")
	}
}
