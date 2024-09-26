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
