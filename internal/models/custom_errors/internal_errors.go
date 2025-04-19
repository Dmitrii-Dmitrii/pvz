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

var (
	ErrStartServer    = &InternalError{Message: "failed to start server"}
	ErrShutdownServer = &InternalError{Message: "failed to shutdown server"}
	ErrEnvLoading     = &InternalError{Message: "failed to load .env file loading"}

	ErrCreatePool        = &InternalError{Message: "failed to create connection pool"}
	ErrBeginTransaction  = &InternalError{Message: "failed to begin transaction"}
	ErrCommitTransaction = &InternalError{Message: "failed to commit transaction"}
	ErrScanRow           = &InternalError{Message: "failed to scan row"}

	ErrInvalidUuid          = &InternalError{Message: "invalid UUID"}
	ErrConvertUuidToOpenapi = &InternalError{Message: "failed to convert uuid to openapi types"}

	ErrCreatePvz   = &InternalError{Message: "failed to create pvz"}
	ErrGetPvz      = &InternalError{Message: "failed to get pvz"}
	ErrPvzNotFound = &InternalError{Message: "pvz not found"}

	ErrCreateReception        = &InternalError{Message: "failed to create reception"}
	ErrGetReceptionInProgress = &InternalError{Message: "failed to get reception in progress"}
	ErrGetReception           = &InternalError{Message: "failed to get reception"}
	ErrGetLastReceptionStatus = &InternalError{Message: "failed to get last reception status"}
	ErrCloseReception         = &InternalError{Message: "failed to close reception"}

	ErrCreateProduct = &InternalError{Message: "failed to create product"}
	ErrDeleteProduct = &InternalError{Message: "failed to delete product"}

	ErrCreateUser       = &InternalError{Message: "failed to create user"}
	ErrGetUserByEmail   = &InternalError{Message: "failed to get user by email"}
	ErrGetUserById      = &InternalError{Message: "failed to get user by id"}
	ErrGenerateJWTToken = &InternalError{Message: "failed to generate jwt token"}
	ErrHashPassword     = &InternalError{Message: "failed to hash password"}
)
