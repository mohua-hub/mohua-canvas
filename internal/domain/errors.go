package domain

import "errors"

type Fault struct { Code string }
func (f *Fault) Error() string { return f.Code }
func Fail(code string) error { return &Fault{Code: code} }
func Code(err error) string {
	var fault *Fault
	if errors.As(err, &fault) { return fault.Code }
	return "INTERNAL"
}
