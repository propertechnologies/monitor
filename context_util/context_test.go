package context_util

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFlowID(t *testing.T) {
	c := context.Background()
	c = context.WithValue(c, "FlowID", "test-flow")

	assert.Equal(t, "test-flow", GetFlowID(c))
}

func TestGetRequestId(t *testing.T) {
	c := context.Background()
	c = context.WithValue(c, "RequestId", "asd-123")

	assert.Equal(t, "asd-123", GetRequestID(c))
}

func TestGetEnv(t *testing.T) {
	c := context.Background()
	c = context.WithValue(c, "env", "local")

	assert.Equal(t, "local", GetEnv(c))
}

func TestGetdebug(t *testing.T) {
	c := context.Background()
	c = context.WithValue(c, "debug", "true")

	assert.Equal(t, "true", GetDebug(c))
}

func TestGetServiceName(t *testing.T) {
	c := context.Background()
	c = context.WithValue(c, "ServiceName", "ledgerlord")

	assert.Equal(t, "ledgerlord", GetServiceName(c))
}

func TestThatIsProdReturnsTrueIfEnvIsProd(t *testing.T) {
	c := context.Background()
	c = context.WithValue(c, "env", "prod")

	assert.True(t, IsProd(c))
}

func TestThatIsProdReturnsFalseIfEnvIsNotProd(t *testing.T) {
	c := context.Background()
	c = context.WithValue(c, "env", "dev")

	assert.False(t, IsProd(c))
}

func TestThatIsDebugReturnsTrueIfDebugIsOn(t *testing.T) {
	c := context.Background()
	c = SetDebugOn(c)

	assert.True(t, IsDebugOn(c))
}

func TestThatIsDebugReturnsFalseIfDebugIsNotPresentOrFalse(t *testing.T) {
	c := context.Background()
	c = context.WithValue(c, "debug", "false")

	assert.False(t, IsDebugOn(c))

	c = context.Background()
	assert.False(t, IsDebugOn(c))
}

func TestThatIsDebugOnReturnsTrueIfEnvironmentIsLocal(t *testing.T) {
	c := context.Background()
	c = context.WithValue(c, "env", "local")

	assert.True(t, IsDebugOn(c))
}

func TestAWSTaskIDIsReadFromContext(t *testing.T) {
	c := context.Background()
	c = context.WithValue(c, "awsTaskId", "11111")

	assert.Equal(t, "11111", GetAwsTaskID(c))
}

type logMock struct {
}

func (l logMock) Infof(s string, i ...interface{}) {
}

func (l logMock) Warnf(s string, i ...interface{}) {
}

func (l logMock) Errorf(s string, i ...interface{}) {
}

func (l logMock) Error(i ...interface{}) {
}

func (l logMock) Warn(i ...interface{}) {
}

func (l logMock) Info(i ...interface{}) {
}

func (l logMock) AddField(key string, value interface{}) {
}
