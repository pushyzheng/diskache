package diskache

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	TMP_DIR     = "tmp"
)

func randStringBytes(n int) (string, []byte) {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b), b
}

func cleanDir() {
	os.RemoveAll(TMP_DIR)
}

// Test Set-ting and Get-ting a value in cache
func TestSetGet(t *testing.T) {
	// Cleanup
	defer cleanDir()

	// Create an instance
	opts := &Opts{
		Directory: TMP_DIR,
	}

	dc, err := New(opts)
	if err != nil {
		t.Error("Expected to create a new instance of Diskache")
	}

	// Set a value in cache
	key, value := randStringBytes(16)
	err = dc.Set(key, value)
	if err != nil {
		t.Error("Expected to be able to set value in cache")
	}

	// Read from cache
	cached, inCache := dc.Get(key)
	if inCache {
		if comp := bytes.Compare(value, cached); comp != 0 {
			t.Error("Expected to get the same value that was set in cache")
		}
	} else {
		t.Error("Expected to get value from cache")
	}
}

func TestDiskache_Delete(t *testing.T) {
	opts := &Opts{
		Directory: TMP_DIR,
	}
	cache, err := New(opts)
	if err != nil {
		t.Error(err)
	}
	err = cache.SetStr("foo", "Hello World")
	if err != nil {
		t.Error(err)
	}
	ok := cache.Delete("foo")
	if !ok {
		t.Error("expected: true, actual: false")
	}
	_, exists := cache.GetStr("foo")
	if exists {
		t.Error("expected: false, actual: true")
	}
}

func TestGetExpiredKey(t *testing.T) {
	opts := &Opts{
		Directory: TMP_DIR,
	}
	dc, err := New(opts)
	if err != nil {
		t.Error(err)
	}
	key := "TestGetExpiredKey_1"
	err = dc.SetExpired(key, []byte("String"), 1000)
	if err != nil {
		t.Error(err)
	}
	_, exists := dc.Get(key)
	if !exists {
		t.Error("expected: true, actual: false")
	}
	log.Println("waiting 2 seconds...")
	time.Sleep(time.Second * 2)
	_, exists = dc.Get(key)
	if exists {
		t.Error("expected: false, actual: true")
	}
	ttl, err := dc.getExpiredTime(key)
	if err != nil {
		t.Error(err)
	}
	if ttl != 0 {
		t.Errorf("expected: 0, actual: %d", ttl)
	}
}

// Test concurrent Set by multiple go routines
func TestConcurrent(t *testing.T) {
	// Cleanup
	defer cleanDir()

	// Create an instance
	opts := &Opts{
		Directory: TMP_DIR,
	}

	dc, err := New(opts)
	if err != nil {
		t.Error("Expected to create a new instance of Diskache")
	}

	// Set multiple times the same value
	key, value := randStringBytes(16)
	for i := 0; i < 1000; i++ {
		go func() {
			if err := dc.Set(key, value); err != nil {
				t.Error("Expected Diskache to handle concurrency")
			}
		}()
	}

	// Read from cache
	cached, inCache := dc.Get(key)
	if inCache {
		if comp := bytes.Compare(value, cached); comp != 0 {
			t.Error("Expected to get the same value that was set in cache")
		}
	} else {
		t.Error("Expected to get value from cache")
	}
}

func TestDiskache_IsExpired(t *testing.T) {
	opts := &Opts{
		Directory: TMP_DIR,
	}
	dc, err := New(opts)
	if err != nil {
		t.Error("create instance error", err)
	}
	k := "TestDiskache_IsExpired"
	err = dc.SetExpired(k, []byte("Hello"), 1000)
	if err != nil {
		t.Error(err)
	}
	expired, err := dc.IsExpired(k)
	if err != nil {
		t.Error(err)
	}
	if expired {
		t.Error("expected: false, actual: true")
	}
	log.Println("wait 2 seconds...")
	time.Sleep(time.Second * 2)
	expired, err = dc.IsExpired(k)
	if err != nil {
		t.Error(err)
	}
	if !expired {
		t.Error("expected: true, actual: false")
	}
}

func TestTTL(t *testing.T) {
	opts := &Opts{
		Directory: TMP_DIR,
	}
	dc, err := New(opts)
	ttl, err := dc.getExpiredTime(time.Now().String())
	if err != nil {
		t.Error(err)
	}
	if ttl != expiredTimestamp {
		t.Errorf("expected: 1, actual: %d", ttl)
	}
	k := "test-key"
	err = dc.setExpiredTime(k, getTimestamp())
	if err != nil {
		t.Error(err)
	}
	ttl, err = dc.getExpiredTime(k)
	if err != nil {
		t.Error(err)
	}
	if ttl == -1 {
		t.Error("getExpiredTime must > 0")
	}
}

func TestSetExpiredTime(t *testing.T) {
	opts := &Opts{
		Directory: TMP_DIR,
	}
	dc, err := New(opts)

	k := "test-key-2"
	err = dc.SetExpired(k, []byte("value"), time.Minute.Milliseconds())
	if err != nil {
		t.Error(err)
	}
}

func TestNoExpiredKey(t *testing.T) {
	opts := &Opts{
		Directory: TMP_DIR,
	}
	dc, err := New(opts)
	if err != nil {
		t.Error(err)
	}
	err = dc.SetExpired("TestNoExpiredKey", []byte("value"), 1000)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 2)
	fmt.Println(dc.getExpiredTime("TestNoExpiredKey"))
}

// Benchmark Set operations
func BenchmarkSet(b *testing.B) {
	// Cleanup
	defer cleanDir()

	// Create an instance
	opts := &Opts{
		Directory: TMP_DIR,
	}

	dc, err := New(opts)
	if err != nil {
		b.Error("Expected to create a new instance of Diskache")
	}

	// Set multiple values in cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key, value := randStringBytes(16)
		err = dc.Set(key, value)
		if err != nil {
			b.Error("Expected to be able to set value in cache")
		}
	}
}

// Benchmark Get operations
func BenchmarkGet(b *testing.B) {
	// Cleanup
	defer cleanDir()

	// Create an instance
	opts := &Opts{
		Directory: TMP_DIR,
	}

	dc, err := New(opts)
	if err != nil {
		b.Error("Expected to create a new instance of Diskache")
	}

	// Set a value in cache
	key, value := randStringBytes(16)
	err = dc.Set(key, value)
	if err != nil {
		b.Error("Expected to be able to set value in cache")
	}

	// Read from cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dc.Get(key)
	}
}
