package log

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestConcurrentBaseAndChildLoggers(t *testing.T) {
	var out bytes.Buffer

	l, err := New(Config{
		Level:  LevelDebug,
		Format: FormatJSON,
		Writer: &out,
		now:    time.Now,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	child := l.With(String("scope", "child")).WithGroup("jobs")

	const workers = 32

	const perWorker = 40

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		workerID := i

		go func() {
			defer wg.Done()

			for j := 0; j < perWorker; j++ {
				l.Info("base", Int("worker", workerID), Int("n", j))
				child.Info("child", Int("worker", workerID), Int("n", j))
			}
		}()
	}

	wg.Wait()

	lines := splitNonEmptyLines(out.String())
	want := workers * perWorker * 2

	if len(lines) != want {
		t.Fatalf("line count mismatch: got=%d want=%d", len(lines), want)
	}

	for _, line := range lines {
		obj := parseJSONLine(t, line)

		msg, _ := obj[keyMsg].(string)
		switch msg {
		case "base":
			if _, ok := obj[keyGroup]; ok {
				t.Fatalf("base logs should not have group: %#v", obj)
			}
		case "child":
			checks := []boolCheckCase{
				{name: "group-jobs", ok: obj[keyGroup] == "jobs"},
				{name: "scope-child", ok: obj["scope"] == "child"},
			}
			for _, check := range checks {
				if !check.ok {
					t.Fatalf("child check failed (%s): %#v", check.name, obj)
				}
			}
		default:
			t.Fatalf("unexpected message value: %q", msg)
		}
	}
}

func TestConcurrentWithDerivation(t *testing.T) {
	var out bytes.Buffer

	l, err := New(Config{Level: LevelDebug, Format: FormatJSON, Writer: &out, now: time.Now})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	const workers = 24

	const perWorker = 35

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		id := i

		go func() {
			defer wg.Done()

			derived := l.With(String("worker_id", fmt.Sprintf("w-%d", id))).WithGroup("derive")
			for j := 0; j < perWorker; j++ {
				derived.Info(
					"derived",
					String("dup", "v1"),
					Int("n", j),
					String("dup", "v2"),
				)
			}
		}()
	}

	wg.Wait()

	lines := splitNonEmptyLines(out.String())
	want := workers * perWorker

	if len(lines) != want {
		t.Fatalf("line count mismatch: got=%d want=%d", len(lines), want)
	}

	for _, line := range lines {
		obj := parseJSONLine(t, line)

		checks := []boolCheckCase{
			{name: "group-derive", ok: obj[keyGroup] == "derive"},
			{name: "dup-v2", ok: obj["dup"] == "v2"},
		}
		for _, check := range checks {
			if !check.ok {
				t.Fatalf("derive check failed (%s): %#v", check.name, obj)
			}
		}
	}
}

func TestStressLogging100Goroutines(t *testing.T) {
	var out bytes.Buffer

	l, err := New(Config{Level: LevelDebug, Format: FormatJSON, Writer: &out, now: time.Now})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	const workers = 100

	const perWorker = 20

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		id := i

		go func() {
			defer wg.Done()

			ctx := WithRequestID(context.Background(), fmt.Sprintf("req-%d", id))
			ctx = WithRequestMetadata(ctx, String("worker", fmt.Sprintf("w-%d", id)))

			for j := 0; j < perWorker; j++ {
				l.InfoCtx(ctx, "stress", Int("n", j))
			}
		}()
	}

	wg.Wait()

	lines := splitNonEmptyLines(out.String())
	want := workers * perWorker

	if len(lines) != want {
		t.Fatalf("line count mismatch: got=%d want=%d", len(lines), want)
	}

	for _, line := range lines {
		obj := parseJSONLine(t, line)

		checks := []boolCheckCase{
			{name: "msg-stress", ok: obj[keyMsg] == "stress"},
			{name: "ctx-metadata", ok: obj[keyRequestID] != nil && obj["worker"] != nil},
		}
		for _, check := range checks {
			if !check.ok {
				t.Fatalf("stress check failed (%s): %#v", check.name, obj)
			}
		}
	}
}

func TestConcurrentContextInjectionRetrieval(t *testing.T) {
	var out bytes.Buffer

	l, err := New(Config{Format: FormatJSON, Writer: &out, now: time.Now})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	const workers = 100

	errCh := make(chan error, workers)

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		id := i

		go func() {
			defer wg.Done()

			ctx := WithLogger(context.Background(), l)
			ctx = WithRequestID(ctx, fmt.Sprintf("rid-%d", id))
			ctx = WithRequestMetadata(ctx, String("k", fmt.Sprintf("v-%d", id)))

			if got := FromContext(ctx, nil); got != l {
				errCh <- fmt.Errorf("logger mismatch for worker %d", id)

				return
			}

			rid, ok := RequestID(ctx)
			if !ok || rid != fmt.Sprintf("rid-%d", id) {
				errCh <- fmt.Errorf("request_id mismatch for worker %d: got=%q ok=%v", id, rid, ok)

				return
			}

			md := RequestMetadata(ctx)
			if len(md) != 1 || md[0].toAttr().Value.String() != fmt.Sprintf("v-%d", id) {
				errCh <- fmt.Errorf("metadata mismatch for worker %d", id)

				return
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatal(err)
	}
}
