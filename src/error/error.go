package e

import "golang.org/x/xerrors"

var (
	ErrOTPNotMatch         = xerrors.New("ERROR.AUTH.OTP_NOT_MATCH")
	ErrUserIdInvalid       = xerrors.New("ERROR.AUTH.INVALID_USER_ID")
	ErrRoleInvalid         = xerrors.New("ERROR.AUTH.INVALID_ROLE")
	ErrOTPInvalid          = xerrors.New("ERROR.AUTH.INVALID_OTP")
	ErrUserNotVerify       = xerrors.New("ERROR.AUTH.USER_NOT_VERIFIED")
	ErrOTPNotChecked       = xerrors.New("ERROR.AUTH.OTP_NOT_CHECKED")
	ErrSignatureBodyFail   = xerrors.New("ERROR.AUTH.SIGNATURE_BODY_NOT_MATCH")
	ErrSignatureHeaderFail = xerrors.New("ERROR.AUTH.SIGNATURE_HEADER_NOT_MATCH")
)
