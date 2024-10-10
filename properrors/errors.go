package properrors

import (
	"fmt"
)

var (
	ErrFailedToLogin                     = New("0001", "Failed to login")
	ErrFailedToLoginByExpiredCredentials = ErrFailedToLogin.WithSubType("01", "expired credentials")
	ErrFailedToLoginByInvalidCredentials = ErrFailedToLogin.WithSubType("02", "invalid credentials")

	ErrAccountNotFound      = New("0002", "Account not found")
	ErrSecondFactorAuth     = New("0003", "Error during second factor authentication")
	ErrLaunchingBot         = New("0004", "Error while launching bot")
	ErrorCredentialNotFound = New("0005", "Error credentials not found")
	ErrorGettingCredentials = New("0006", "Error getting credentials")
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

func (p *Error) WithSubType(code string, errMsg string) *Error {
	t := New(p.ID, p.Err)
	t.ID = fmt.Sprintf("%s-%s", p.ID, code)
	t.Err = fmt.Sprintf("%s:%s", p.Err, errMsg)
	t.Desc = fmt.Sprintf("%s-%s", p.Desc, code)

	return t
}

func (p *Error) Is(err error) bool {
	if err == nil {
		return false
	}
	e, ok := err.(*Error)
	if !ok {
		return false
	}

	return (e.ID == p.ID) && (e.Err == p.Err) && (e.Desc == p.Desc)
}
