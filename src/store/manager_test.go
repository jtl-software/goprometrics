package store

import (
	"errors"
	"testing"
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
