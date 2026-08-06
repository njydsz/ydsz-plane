package errs

import (
	"errors"
	"testing"
)

func TestWrapAndAs(t *testing.T) {
	cause := errors.New("db down")
	err := ErrInternal.Wrap(cause)

	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Fatal("errors.As must resolve *AppError")
	}
	if appErr.Code != "INTERNAL.ERROR" {
		t.Fatalf("code = %s", appErr.Code)
	}
	if !errors.Is(err, cause) {
		t.Fatal("wrapped cause must be discoverable")
	}
}

func TestWithDetailsDoesNotMutateOriginal(t *testing.T) {
	detailed := ErrValidation.WithDetails(FieldDetail{Field: "email", Reason: "invalid"})
	if len(ErrValidation.Details) != 0 {
		t.Fatal("registry error must stay pristine")
	}
	if len(detailed.Details) != 1 || detailed.Details[0].Field != "email" {
		t.Fatalf("details = %+v", detailed.Details)
	}
}
