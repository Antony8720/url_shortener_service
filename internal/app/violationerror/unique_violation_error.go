package violationerror

import (
	"github.com/google/uuid"
)

type UniqueViolationError struct {
	Err error
	UserID uuid.UUID
	Short string
	Long string
}

func(uve *UniqueViolationError) Error() string{
	return uve.Err.Error()
}