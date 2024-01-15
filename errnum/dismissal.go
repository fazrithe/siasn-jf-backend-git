package errnum

import "github.com/if-itb/siasn-libs-backend/ec"

const (
	// ErrCodeDismissalAdmissionReasonEmpty - 19401: alasan_pemberhentian is empty.
	ErrCodeDismissalAdmissionReasonEmpty = iota + ServiceErrorCode*10000 + 9*1000 + 401
	// ErrCodeDismissalAdmissionNoSupportDocs - 19402: No support documents saved, must supply at least one filename.
	ErrCodeDismissalAdmissionNoSupportDocs
	// ErrCodeDismissalAdmissionAsnNotFound - 19403: ASN not found.
	ErrCodeDismissalAdmissionAsnNotFound
	// ErrCodeDismissalAcceptanceSignerNotFound - 19404: Signer user ID not found.
	ErrCodeDismissalAcceptanceSignerNotFound
	// ErrCodeDismissalAcceptanceNoLetter - 19405: Valid acceptance letter number and date must be supplied for dismissal acceptance.
	ErrCodeDismissalAcceptanceNoLetter
	// ErrCodeDismissalDenialNoReason - 19406: Dismissal is denied for no reason.
	ErrCodeDismissalDenialNoReason
	// ErrCodeDismissalSearchStatusInvalid - 19407: admission status supplied contains value outside the valid range (1-5).
	ErrCodeDismissalSearchStatusInvalid
	// ErrCodeDismissalSearchInvalidDate - 19408: the date format supplied does not conform to the date format required.
	ErrCodeDismissalSearchInvalidDate
	// ErrCodeDismissalDecreeDataEmpty - 19409: decree date, number, and reason detail must not be empty if dismissal reason is 2 - 5.
	ErrCodeDismissalDecreeDataEmpty
	// ErrCodeDismissalAdmissionNumberInvalid - 19410: admission number is invalid.
	ErrCodeDismissalAdmissionNumberInvalid
)

var (
	ErrDismissalAdmissionReasonEmpty     *ec.Error
	ErrDismissalAdmissionNoSupportDocs   *ec.Error
	ErrDismissalAdmissionAsnNotFound     *ec.Error
	ErrDismissalAcceptanceSignerNotFound *ec.Error
	ErrDismissalAcceptanceNoLetter       *ec.Error
	ErrDismissalDenialNoReason           *ec.Error
	ErrDismissalSearchStatusInvalid      *ec.Error
	ErrDismissalSearchInvalidDate        *ec.Error
	ErrDismissalDecreeDataEmpty          *ec.Error
	ErrDismissalAdmissionNumberInvalid   *ec.Error
)

func init() {
	Errs[ErrCodeDismissalAdmissionReasonEmpty] = "alasan_pemberhentian is empty"
	Errs[ErrCodeDismissalAdmissionNoSupportDocs] = "no support documents saved, must supply at least one filename"
	Errs[ErrCodeDismissalAdmissionAsnNotFound] = "ASN not found"
	Errs[ErrCodeDismissalAcceptanceSignerNotFound] = "signer user ID not found"
	Errs[ErrCodeDismissalAcceptanceNoLetter] = "valid acceptance letter number and date must be supplied for dismissal acceptance"
	Errs[ErrCodeDismissalDenialNoReason] = "dismissal is denied for no reason"
	Errs[ErrCodeDismissalSearchStatusInvalid] = "admission status supplied contains value outside the valid range"
	Errs[ErrCodeDismissalSearchInvalidDate] = "the date format supplied does not conform to the date format required"
	Errs[ErrCodeDismissalDecreeDataEmpty] = "decree date, number, and reason detail must not be empty if dismissal reason is 2 - 5"
	Errs[ErrCodeDismissalAdmissionNumberInvalid] = "admission number is invalid"

	ErrsToHttp[ErrCodeDismissalAdmissionReasonEmpty] = 400
	ErrsToHttp[ErrCodeDismissalAdmissionNoSupportDocs] = 400
	ErrsToHttp[ErrCodeDismissalAdmissionAsnNotFound] = 400
	ErrsToHttp[ErrCodeDismissalAcceptanceSignerNotFound] = 400
	ErrsToHttp[ErrCodeDismissalAcceptanceNoLetter] = 400
	ErrsToHttp[ErrCodeDismissalDenialNoReason] = 400
	ErrsToHttp[ErrCodeDismissalSearchStatusInvalid] = 400
	ErrsToHttp[ErrCodeDismissalSearchInvalidDate] = 400
	ErrsToHttp[ErrCodeDismissalDecreeDataEmpty] = 400
	ErrsToHttp[ErrCodeDismissalAdmissionNumberInvalid] = 400

	ErrDismissalAdmissionReasonEmpty = ec.NewErrorBasic(ErrCodeDismissalAdmissionReasonEmpty, Errs[ErrCodeDismissalAdmissionReasonEmpty])
	ErrDismissalAdmissionNoSupportDocs = ec.NewErrorBasic(ErrCodeDismissalAdmissionNoSupportDocs, Errs[ErrCodeDismissalAdmissionNoSupportDocs])
	ErrDismissalAdmissionAsnNotFound = ec.NewErrorBasic(ErrCodeDismissalAdmissionAsnNotFound, Errs[ErrCodeDismissalAdmissionAsnNotFound])
	ErrDismissalAcceptanceSignerNotFound = ec.NewErrorBasic(ErrCodeDismissalAcceptanceSignerNotFound, Errs[ErrCodeDismissalAcceptanceSignerNotFound])
	ErrDismissalAcceptanceNoLetter = ec.NewErrorBasic(ErrCodeDismissalAcceptanceNoLetter, Errs[ErrCodeDismissalAcceptanceNoLetter])
	ErrDismissalDenialNoReason = ec.NewErrorBasic(ErrCodeDismissalDenialNoReason, Errs[ErrCodeDismissalDenialNoReason])
	ErrDismissalSearchStatusInvalid = ec.NewErrorBasic(ErrCodeDismissalSearchStatusInvalid, Errs[ErrCodeDismissalSearchStatusInvalid])
	ErrDismissalSearchInvalidDate = ec.NewErrorBasic(ErrCodeDismissalSearchInvalidDate, Errs[ErrCodeDismissalSearchInvalidDate])
	ErrDismissalDecreeDataEmpty = ec.NewErrorBasic(ErrCodeDismissalDecreeDataEmpty, Errs[ErrCodeDismissalDecreeDataEmpty])
	ErrDismissalAdmissionNumberInvalid = ec.NewErrorBasic(ErrCodeDismissalAdmissionNumberInvalid, Errs[ErrCodeDismissalAdmissionNumberInvalid])
}
