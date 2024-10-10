package properrors

import (
	"errors"
	"testing"
)

func TestErrorPrintsItsWrappedError(t *testing.T) {
	err := ErrLaunchingBot.Wrap(errors.New("foo"))

	if err.Error() != "0004: Error while launching bot desc:https://ledgerlord.proper.ai/errors/0004 w:foo" {
		t.Errorf("Error message is not as expected")
	}
}

func TestCreateErrorsAndCheckMessagesValues(t *testing.T) {
	var cases = []struct {
		name     string
		input    error
		expected string
	}{
		{
			name:     "Test creating error without subtype",
			input:    ErrFailedToLogin,
			expected: "0001: Failed to login desc:https://ledgerlord.proper.ai/errors/0001",
		},
		{
			name:     "Test creating error with subtype",
			input:    ErrFailedToLoginByExpiredCredentials,
			expected: "0001-01: Failed to login:expired credentials desc:https://ledgerlord.proper.ai/errors/0001-01",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.input.Error() != c.expected {
				t.Errorf("results don´t match: expected %v, got %v", c.expected, c.input)
			}
		})
	}
}

func TestThatIsReturnsTrueIfErrorIsTheSame(t *testing.T) {
	err := ErrFailedToLoginByExpiredCredentials
	if errors.Is(err, ErrFailedToLoginByExpiredCredentials) == false {
		t.Errorf("Is() should return true")
	}
}
