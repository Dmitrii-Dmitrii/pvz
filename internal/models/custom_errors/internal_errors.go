package custom_errors

import (
	"fmt"
)

type InternalError struct {
	Err     error
	Message string
}

func (e *InternalError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *InternalError) Unwrap() error {
	return e.Err
}

var (
	ErrBeginTransaction  = &InternalError{Message: "failed to begin transaction"}
	ErrCommitTransaction = &InternalError{Message: "failed to commit transaction"}
	ErrScanRow           = &InternalError{Message: "failed to scan row"}

	ErrSetUuid              = &InternalError{Message: "failed to set uuid"}
	ErrConvertUuidToPgtype  = &InternalError{Message: "failed to convert uuid to pgtype"}
	ErrUuidNotPresent       = &InternalError{Message: "invalid UUID: not present"}
	ErrConvertUuidToOpenapi = &InternalError{Message: "failed to convert uuid to openapi types"}

	ErrCreatePvz = &InternalError{Message: "failed to create pvz"}
	ErrGetPvz    = &InternalError{Message: "failed to get pvz"}

	ErrCreateReception        = &InternalError{Message: "failed to create reception"}
	ErrGetReceptionInProgress = &InternalError{Message: "failed to get reception in progress"}
	ErrGetReception           = &InternalError{Message: "failed to get reception"}
	ErrGetLastReceptionStatus = &InternalError{Message: "failed to get last reception status"}
	ErrCloseReception         = &InternalError{Message: "failed to close reception"}

	ErrCreateProduct = &InternalError{Message: "failed to create product"}
	ErrDeleteProduct = &InternalError{Message: "failed to delete product"}
)
