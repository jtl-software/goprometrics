package store

import (
	"errors"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCanAppendNew(t *testing.T) {

	store := FakeStore{store: map[string]float64{}}

	opts := MetricOpts{
		Ns:   "test",
		Name: "testname",
	}
	n, err := Append(store, opts)
	if err != nil {
		t.Errorf("No Error expected when append new Item to Store - receive %s", err)
	}

	if n != true && len(store.store) == 1 {
		t.Errorf("Expect to append a new item to store")
	}
}

func TestDoNotAppendWhenAlreadyExist(t *testing.T) {

	store := FakeStore{store: map[string]float64{}}

	opts := MetricOpts{
		Ns:   "test",
		Name: "testname",
	}
	n, err := Append(store, opts)
	if err != nil || n != true {
		t.Errorf("No Error expected when append new Item to Store - receive %s", err)
	}

	n, err = Append(store, opts)
	if n != false {
		t.Errorf("Item with key %s should already exists - expect that new new item was append to store", opts.Key())
	}
}

func TestAppendWontCrashOnPanic(t *testing.T) {

	store := FakePanicStore{}
	opts := MetricOpts{
		Ns:   "test",
		Name: "testname",
	}

	_, err := Append(store, opts)
	if err == nil {
		t.Errorf("No Error expected when append new Item to Store - receive %s", err)
	}
}

func TestAppendCounter_incremented_on_first_append(t *testing.T) {
	before := testutil.ToFloat64(AppendCounter)
	s := FakeStore{store: map[string]float64{}}
	opts := MetricOpts{Ns: "mgr", Name: "appendctr_first"}
	Append(s, opts)
	after := testutil.ToFloat64(AppendCounter)
	if after-before != 1.0 {
		t.Errorf("AppendCounter should increase by 1 on first append; before=%.0f after=%.0f", before, after)
	}
}

func TestAppendCounter_not_incremented_on_second_append(t *testing.T) {
	s := FakeStore{store: map[string]float64{}}
	opts := MetricOpts{Ns: "mgr", Name: "appendctr_second"}
	Append(s, opts) // first — counted
	before := testutil.ToFloat64(AppendCounter)
	Append(s, opts) // second — must not count again
	after := testutil.ToFloat64(AppendCounter)
	if after != before {
		t.Errorf("AppendCounter should not increase on repeat append; before=%.0f after=%.0f", before, after)
	}
}

func TestErrorCounter_incremented_on_panic(t *testing.T) {
	before := testutil.ToFloat64(ErrorCounter)
	s := FakePanicStore{}
	opts := MetricOpts{Ns: "mgr", Name: "errctr_panic"}
	Append(s, opts)
	after := testutil.ToFloat64(ErrorCounter)
	if after-before != 1.0 {
		t.Errorf("ErrorCounter should increase by 1 on recovered panic; before=%.0f after=%.0f", before, after)
	}
}

func TestAppend_concurrent_safety(t *testing.T) {
	s := &SafeFakeStore{store: map[string]float64{}}
	opts := MetricOpts{Ns: "mgr", Name: "concurrent"}

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			Append(s, opts)
		}()
	}
	wg.Wait()

	// Regardless of how many goroutines raced, the metric should be registered exactly once.
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.store) != 1 {
		t.Errorf("concurrent Append should register metric exactly once; store size = %d", len(s.store))
	}
}

type FakeStore struct {
	store map[string]float64
}

func (f FakeStore) Append(opts MetricOpts) {
	f.store[opts.Key()] = 0
}

func (f FakeStore) Inc(opts MetricOpts, value float64) {
	f.store[opts.Key()] += value
}

func (f FakeStore) Has(opts MetricOpts) bool {
	_, has := f.store[opts.Key()]
	return has
}

type FakePanicStore struct{}

func (f FakePanicStore) Append(opts MetricOpts) {
	panic(errors.New("should painc"))
}

func (f FakePanicStore) Inc(opts MetricOpts, value float64) {
	// not required
}

func (f FakePanicStore) Has(opts MetricOpts) bool {
	return false
}

// SafeFakeStore is a thread-safe fake used for concurrency tests.
type SafeFakeStore struct {
	mu    sync.Mutex
	store map[string]float64
}

func (f *SafeFakeStore) Append(opts MetricOpts) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.store[opts.Key()] = 0
}

func (f *SafeFakeStore) Inc(opts MetricOpts, value float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.store[opts.Key()] += value
}

func (f *SafeFakeStore) Has(opts MetricOpts) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, has := f.store[opts.Key()]
	return has
}
