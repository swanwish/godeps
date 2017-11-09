package common

import "errors"

var (
	// Database Related Error
	ErrNoConnectionProvider  = errors.New(ErrorMessageNoConnectionProvider)
	ErrNoTransactionFunction = errors.New(ErrorMessageNoTransactionFunction)

	ErrNotExist            = errors.New(ErrorMessageNotExist)
	ErrAlreadyExist        = errors.New(ErrorMessageAlreadyExist)
	ErrInvalidParameter    = errors.New(ErrorMessageInvalidParameter)
	ErrInvalidStatus       = errors.New(ErrorMessageInvalidStatus)
	ErrInternalError       = errors.New(ErrorMessageInternalError)
	ErrNoPrivileges        = errors.New(ErrorMessageNoPrivileges)
	ErrRefreshTokenExpired = errors.New(ErrorMessageRefreshTokenExpired)
	ErrTokenExpired        = errors.New(ErrorMessageTokenExpired)
	ErrNoDeviceInformation = errors.New(ErrorMessageNoDeviceInformation)
	ErrLimitExceed         = errors.New(ErrorMessageLimitExceed)
)
