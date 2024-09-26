package logging

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGCPLogging(t *testing.T) {
	n := &myIOWriter{}
	ctx := context.Background()
	logger := NewLoggerWithWriter(n)
	ctx = SetLogger(ctx, logger)

	Reportf(ctx, "test")

	assert.NotEmpty(t, n.errorMessage.Message)
	assert.NotEmpty(t, n.errorMessage.ErrReport)
	assert.NotEmpty(t, n.errorMessage.StackTrace)
}

type errorMessage struct {
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	App        string `json:"app"`
	FlowID     string `json:"flow-id"`
	RootTaskId string `json:"root-task-id"`
	StackTrace string `json:"stack_trace"`
	ErrReport  string `json:"@type"`
}

type myIOWriter struct {
	errorMessage
}

func (m *myIOWriter) Write(p []byte) (n int, err error) {
	json.Unmarshal(p, &m)

	return len(p), nil
}
