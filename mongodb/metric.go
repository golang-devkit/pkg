package mongodb

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/golang-devkit/pkg/logger/log"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	MethodRead          methodMetric = "Read"
	MethodReadPrimary   methodMetric = "ReadPrimary"
	MethodReadSecondary methodMetric = "ReadSecondary"
	MethodWrite         methodMetric = "Write"
	MethodUploadFile    methodMetric = "UploadFile"
	MethodDownloadFile  methodMetric = "DownloadFile"
)

type methodMetric string

type metric struct {
	date    string
	summary map[methodMetric]*operation
	sync.Mutex
}

type operation struct {
	readCount   int64
	readFailed  int64
	writeCount  int64
	writeFailed int64
	lastIssue   error

	sync.Mutex
}

func (m *metric) init(method ...methodMetric) {
	m.Lock()
	defer m.Unlock()
	// reset daily
	if m.date == "" || m.summary == nil || m.date != time.Now().Format("2006-01-02") {
		// set current date
		m.date = time.Now().Format("2006-01-02")
		m.summary = make(map[methodMetric]*operation)
	}
	// init summary map
	for _, mtd := range method {
		if _, ok := m.summary[mtd]; !ok {
			m.summary[mtd] = &operation{}
		}
	}
}

func (m *metric) sprint(t time.Time, method methodMetric) string {
	m.init()
	op, ok := m.summary[method]
	if !ok {
		return ""
	}

	// logger := log.NewFile()
	// defer logger.SyncFile(fmt.Sprintf("metric/mongodb_%s.log", m.date), log.SyncOption{UseFilenameOriginal: true})
	// fmt.Fprintf(logger, "%s\n", log.NewLoggerWith("mongodb-metric", "info", map[string]any{
	// 	"type":         method,
	// 	"date":         m.date,
	// 	"read":         op.readCount,
	// 	"read_failed":  op.readFailed,
	// 	"write":        op.writeCount,
	// 	"write_failed": op.writeFailed,
	// 	"issue":        fmt.Sprintf("%v", op.lastIssue),
	// }).JsonEncode())

	return fmt.Sprintf("%s | %s (%s) - read: %d (failed: %d), write: %d (failed: %d), issue: %v",
		time.Since(t).String(), method, m.date,
		op.readCount, op.readFailed,
		op.writeCount, op.writeFailed,
		op.lastIssue)
}

func (m *metric) print(t time.Time, method methodMetric) {
	m.init()
	op, ok := m.summary[method]
	if !ok {
		return
	}
	mt := map[string]any{
		"type":         method,
		"date":         m.date,
		"read":         op.readCount,
		"read_failed":  op.readFailed,
		"write":        op.writeCount,
		"write_failed": op.writeFailed,
		"issue":        fmt.Sprintf("%v", op.lastIssue),
		"duration":     time.Since(t).String(),
	}
	fmt.Fprintf(os.Stdout, "%s\n", log.NewLoggerWith("mongodb-metric", "info", mt).JsonEncode())
}

func (m *metric) incRead(method methodMetric, issue error) {
	m.init(method)
	op, ok := m.summary[method]
	if !ok {
		return
	}
	op.Lock()
	defer op.Unlock()
	op.readCount++
	// not found is not considered as a failed read
	if errors.Is(issue, mongo.ErrNoDocuments) {
		return
	}
	// other errors are considered as failed read
	if issue != nil {
		op.readFailed++
		op.lastIssue = issue
	}
}

func (m *metric) incWrite(method methodMetric, issue error) {
	m.init(method)
	op, ok := m.summary[method]
	if !ok {
		return
	}
	op.Lock()
	defer op.Unlock()
	op.writeCount++
	// not found is not considered as a failed write
	if errors.Is(issue, mongo.ErrNoDocuments) {
		return
	}
	// other errors are considered as failed write
	if issue != nil {
		op.writeFailed++
		op.lastIssue = issue
	}
}
