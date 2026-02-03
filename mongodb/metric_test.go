package mongodb

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestMetric_Init(t *testing.T) {
	m := &metric{}
	methods := []methodMetric{"method1", "method2"}

	m.init(methods...)

	// Check if date is set correctly
	expectedDate := time.Now().Format("2006-01-02")
	if m.date != expectedDate {
		t.Errorf("Expected date %s, got %s", expectedDate, m.date)
	}

	// Check if methods are initialized
	for _, method := range methods {
		if op, exists := m.summary[method]; !exists {
			t.Errorf("Method %s not initialized", method)
		} else if op == nil {
			t.Errorf("Operation for method %s is nil", method)
		}
	}
}

func TestMetric_Print(t *testing.T) {
	m := &metric{}
	method := methodMetric("testMethod")

	// Test empty method
	result := m.sprint(time.Now(), methodMetric("nonexistentMethod"))
	if result != "" {
		t.Errorf("Expected empty string for nonexistent method, got %s", result)
	}

	// Test method with data
	m.init(method)
	m.summary[method].readCount = 5
	m.summary[method].readFailed = 1
	m.summary[method].writeCount = 10
	m.summary[method].writeFailed = 2
	m.summary[method].lastIssue = errors.New("test error")

	expected := string(method) + " (" + time.Now().Format("2006-01-02") + ") - read: 5 (failed: 1), write: 10 (failed: 2), issue: test error"
	result = m.sprint(time.Now(), method)
	if !strings.HasSuffix(result, expected) {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestMetric_IncRead(t *testing.T) {
	m := &metric{}
	method := methodMetric("testMethod")

	// Test successful read
	m.incRead(method, nil)
	if m.summary[method].readCount != 1 {
		t.Errorf("Expected read count 1, got %d", m.summary[method].readCount)
	}
	if m.summary[method].readFailed != 0 {
		t.Errorf("Expected read failed 0, got %d", m.summary[method].readFailed)
	}

	// Test failed read
	err := errors.New("read error")
	m.incRead(method, err)
	if m.summary[method].readCount != 2 {
		t.Errorf("Expected read count 2, got %d", m.summary[method].readCount)
	}
	if m.summary[method].readFailed != 1 {
		t.Errorf("Expected read failed 1, got %d", m.summary[method].readFailed)
	}
	if m.summary[method].lastIssue != err {
		t.Errorf("Expected error %v, got %v", err, m.summary[method].lastIssue)
	}
}

func TestMetric_IncWrite(t *testing.T) {
	m := &metric{}
	method := methodMetric("testMethod")

	// Test successful write
	m.incWrite(method, nil)
	if m.summary[method].writeCount != 1 {
		t.Errorf("Expected write count 1, got %d", m.summary[method].writeCount)
	}
	if m.summary[method].writeFailed != 0 {
		t.Errorf("Expected write failed 0, got %d", m.summary[method].writeFailed)
	}

	// Test failed write
	err := errors.New("write error")
	m.incWrite(method, err)
	if m.summary[method].writeCount != 2 {
		t.Errorf("Expected write count 2, got %d", m.summary[method].writeCount)
	}
	if m.summary[method].writeFailed != 1 {
		t.Errorf("Expected write failed 1, got %d", m.summary[method].writeFailed)
	}
	if m.summary[method].lastIssue != err {
		t.Errorf("Expected error %v, got %v", err, m.summary[method].lastIssue)
	}
}

func TestMetric_ConcurrentAccess(t *testing.T) {
	m := &metric{}
	method := methodMetric("testMethod")
	iterations := 1000

	// Test concurrent reads
	done := make(chan bool)
	for i := 0; i < iterations; i++ {
		go func() {
			m.incRead(method, nil)
			done <- true
		}()
	}

	// Wait for all goroutines to finish
	for i := 0; i < iterations; i++ {
		<-done
	}

	if m.summary[method].readCount != int64(iterations) {
		t.Errorf("Expected read count %d, got %d", iterations, m.summary[method].readCount)
	}
}
