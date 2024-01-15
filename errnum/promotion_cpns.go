package errnum

import "github.com/if-itb/siasn-libs-backend/ec"

const (
	// ErrCodePromotionCpnsAdmissionGeneric - 14401: generic promotion admission error
	ErrCodePromotionCpnsAdmissionGeneric = iota + ServiceErrorCode*10000 + 4*1000 + 401
	// ErrCodePromotionCpnsAdmissionStatusNotCreated - 14402: admission status not created
	ErrCodePromotionCpnsAdmissionStatusNotCreated
	// ErrCodePromotionCpnsAdmissionFieldInvalid - 14403: one or more of required request field(s) is empty.
	ErrCodePromotionCpnsAdmissionFieldInvalid
)

var (
	ErrPromotionCpnsAdmissionStatusNotCreated *ec.Error
)

func init() {
	Errs[ErrCodePromotionCpnsAdmissionStatusNotCreated] = "CPNS promotion admission status not created (1)"
	Errs[ErrCodePromotionCpnsAdmissionFieldInvalid] = "one or more of required request field(s) is empty"

	ErrPromotionCpnsAdmissionStatusNotCreated = ec.NewErrorBasic(ErrCodePromotionCpnsAdmissionStatusNotCreated, Errs[ErrCodePromotionCpnsAdmissionStatusNotCreated])
}
