package properrors

import "fmt"

var (
	ErrFailedToLogin    = NewProperError("0001", "Failed to login")
	ErrAccountNotFound  = NewProperError("0002", "Account not found")
	ErrSecondFactorAuth = NewProperError("0003", "Error en login por second factor")
)

type (
	ProperError struct {
		ID   string `json:"id"`
		Err  string `json:"error"`
		Desc string `json:"description"`
	}
)

func NewProperError(id string, err string) *ProperError {
	return &ProperError{
		ID:   id,
		Desc: "https://proper.com/errors/" + id,
		Err:  err,
	}
}

func (p *ProperError) Error() string {
	return fmt.Sprintf("%s: %s desc:%s", p.ID, p.Err, p.Desc)
}
