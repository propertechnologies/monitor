package properrors

import "fmt"

var (
	ErrFailedToLogin    = New("0001", "Failed to login")
	ErrAccountNotFound  = New("0002", "Account not found")
	ErrSecondFactorAuth = New("0003", "Error during second factor authentication")
	ErrLaunchingBot     = New("0004", "Error while launching bot")
)

type (
	Error struct {
		ID     string `json:"id"`
		Err    string `json:"error"`
		Desc   string `json:"description"`
		WError error  `json:"info"`
	}
)

func New(id string, err string) *Error {
	return &Error{
		ID:   id,
		Desc: "https://ledgerlord.proper.ai/errors/" + id,
		Err:  err,
	}
}

func (p *Error) Error() string {
	format := "%s: %s desc:%s"
	if p.WError != nil {
		format += " w:%s"
		return fmt.Sprintf(format, p.ID, p.Err, p.Desc, p.WError)
	}

	return fmt.Sprintf(format, p.ID, p.Err, p.Desc)
}

func (p *Error) Wrap(e error) *Error {
	p.WError = e
	return p
}
