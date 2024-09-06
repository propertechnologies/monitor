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
		ID   string `json:"id"`
		Err  string `json:"error"`
		Desc string `json:"description"`
		Info string `json:"info"`
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
	return fmt.Sprintf("%s: %s desc:%s", p.ID, p.Err, p.Desc)
}

func (p *Error) WithExtraInfo(info string) *Error {
	p.Info = info
	return p
}
