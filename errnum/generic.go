package errnum

import "github.com/if-itb/siasn-libs-backend/ec"

// IsClientError checks whether an instance of ec.Error is a client side error.
func IsClientError(err *ec.Error) bool {
	return err.Code%1000/100 == 4
}

const ServiceErrorCode = 1

const (
	// ErrCodeRequestJsonDecode - 10410: request JSON cannot be decoded.
	ErrCodeRequestJsonDecode = iota + ServiceErrorCode*10000 + 0*1000 + 410
	// ErrCodeRequestBodyNil - 10411: request body is nil.
	ErrCodeRequestBodyNil
	// ErrCodeRequestQueryParamParse - 10412: request query parameter cannot be parsed.
	ErrCodeRequestQueryParamParse
	// ErrCodeListCountPerPage - 10413: count must be >= 1 and <= 100.
	ErrCodeListCountPerPage
	// ErrCodeListPageNumber - 10414: page number must be >= 1.
	ErrCodeListPageNumber
	// ErrCodeMimeTypeNotSupported - 10415: mime type of the document is not supported/allowed.
	ErrCodeMimeTypeNotSupported
	// ErrCodeStorageFileNotFound - 10416: the literal file cannot be found in literally the object storage.
	ErrCodeStorageFileNotFound
	// ErrCodeEntryNotFound - 10417: a generic error telling the requester that the entry cannot be found.
	ErrCodeEntryNotFound
	// ErrCodeUuidInvalid - 10418: not a valid UUID.
	ErrCodeUuidInvalid
	// ErrCodeNoIdToken - 10419: no ID token present in request.
	ErrCodeNoIdToken
	// ErrCodeInvalidUser - 10420: user detail cannot be found.
	ErrCodeInvalidUser
	// ErrCodeNoOAuth2ExchangeCode - 10421: returned when the user does not supply OAuth2 exchange code in the request.
	ErrCodeNoOAuth2ExchangeCode
	// ErrCodeNoWorkAgencyId - 10422: This particular user does not have agency ID.
	ErrCodeNoWorkAgencyId
	// ErrCodeEndpointGone - 10423: The endpoint is no longer available.
	ErrCodeEndpointGone
	// ErrCodeDocumentGenerateBadTemplate - 10424: unable to generate document from docx template, some placeholders
	// have incorrect syntax, must not contain spaces between two words (e.g. `{{ nama instansi }}` is not allowed, must be
	// `{{ nama_instansi }}`).
	ErrCodeDocumentGenerateBadTemplate
)

const (
	// ErrCodeTxStart - 10501: Failed in creating a transaction.
	ErrCodeTxStart = iota + ServiceErrorCode*10000 + 0*1000 + 501
	// ErrCodeTxCommit - 10502: Failed in committing a transaction.
	ErrCodeTxCommit
	// ErrCodeTableCreate - 10503: Failed in creating table.
	ErrCodeTableCreate
	// ErrCodeResponseParseFail - 10504: Cannot read response from backend services.
	ErrCodeResponseParseFail
	// ErrCodePrepareFail - 10505: Cannot prepare SQL statement.
	ErrCodePrepareFail
	// ErrCodeExecFail - 10506: General error, SQL execution failed.
	// Can be used whenever Exec code fails. You can also create a new code
	// just to be more specific for example for the case when execution fails just when inserting new employee entry.
	ErrCodeExecFail
	// ErrCodeQueryFail - 10507: General error, SQL query failed.
	// Can be used whenever Query/QueryRow fails, just like ErrCodeExecFail for Exec and Prepare.
	ErrCodeQueryFail
	// ErrCodeStorageCopyFail - 10508: Copying (or moving + deleting) files in object storage failed.
	ErrCodeStorageCopyFail
	// ErrCodeStorageSignFail - 10509: Failed to create signed URL (for any purposes).
	ErrCodeStorageSignFail
	// ErrCodeCannotVerifyIdToken - 10510: Server error, cannot verify ID token just after it was issued, so this is our fault.
	ErrCodeCannotVerifyIdToken
	// ErrCodeOAuth2ExchangeFailed - 10511: Code exchange failed.
	ErrCodeOAuth2ExchangeFailed
	// ErrCodeRoleUnauthorized - 10512: User is unauthorized to do the action.
	ErrCodeRoleUnauthorized
	// ErrCodeDocumentGenerate - 10513: unable to generate document from docx template.
	ErrCodeDocumentGenerate
	// ErrCodeStoragePutFail - 10508: Copying (or moving + deleting) files in object storage failed.
	ErrCodeStoragePutFail
	// ErrCodeStorageGetMetadataFail - 10509: Getting metadata for a file in object storage failed.
	ErrCodeStorageGetMetadataFail
)

// Errs map ensures that there are no duplicate error codes in this service.
// This map contains the default error message for each error code, you don't have to use it, but each error
// code must be registered in this map.
var Errs = map[int]string{
	ErrCodeRequestJsonDecode:           "cannot decode request body as JSON",
	ErrCodeRequestBodyNil:              "request body is nil",
	ErrCodeRequestQueryParamParse:      "cannot parse query parameters",
	ErrCodeListCountPerPage:            "count must be >= 1 and <= 100",
	ErrCodeListPageNumber:              "page number must be >= 1",
	ErrCodeMimeTypeNotSupported:        "mime type of the document is not supported/allowed",
	ErrCodeStorageFileNotFound:         "document not found in storage",
	ErrCodeEntryNotFound:               "entry cannot be found",
	ErrCodeUuidInvalid:                 "not a valid UUID string",
	ErrCodeNoIdToken:                   "no ID token present in request",
	ErrCodeInvalidUser:                 "user detail cannot be found",
	ErrCodeNoOAuth2ExchangeCode:        "no OAuth2 exchange code",
	ErrCodeEndpointGone:                "the endpoint is no longer available",
	ErrCodeRoleUnauthorized:            "user is unauthorized to do the action",
	ErrCodeDocumentGenerateBadTemplate: "unable to generate document from docx template, some placeholders have incorrect syntax, e.g. must not contain spaces between two words (`{{ nama instansi }}` is not allowed, must be `{{ nama_instansi }}`)",

	ErrCodeResponseParseFail:      "cannot read response from backend services",
	ErrCodePrepareFail:            "cannot prepare SQL statement",
	ErrCodeExecFail:               "SQL statement execution failed",
	ErrCodeQueryFail:              "SQL query failed",
	ErrCodeStorageCopyFail:        "copying (or moving + deleting) files in object storage failed",
	ErrCodeStorageSignFail:        "failed to create signed URL for this operation",
	ErrCodeCannotVerifyIdToken:    "cannot verify newly issued ID token",
	ErrCodeOAuth2ExchangeFailed:   "code exchange failed",
	ErrCodeNoWorkAgencyId:         "user has no work agency",
	ErrCodeDocumentGenerate:       "unable to generate document from docx template",
	ErrCodeStoragePutFail:         "unable to put file into object storage",
	ErrCodeStorageGetMetadataFail: "unable to retrieve file metadata from storage",
}

// ErrsToHttp is a map of error codes to HTTP status codes. Those that do not exist in this map
// should be treated as 500 instead. Because writing http. constants are long and ugly, we just write the numbers
// so you can just read directly.
//
// Status codes must not be WebDAV status codes, or codes that are not defined by standard.
var ErrsToHttp = map[int]int{
	ErrCodeRequestJsonDecode:           400,
	ErrCodeRequestBodyNil:              400,
	ErrCodeRequestQueryParamParse:      400,
	ErrCodeListCountPerPage:            400,
	ErrCodeListPageNumber:              400,
	ErrCodeMimeTypeNotSupported:        400,
	ErrCodeStorageFileNotFound:         404,
	ErrCodeEntryNotFound:               404,
	ErrCodeUuidInvalid:                 400,
	ErrCodeNoIdToken:                   400,
	ErrCodeInvalidUser:                 403,
	ErrCodeNoOAuth2ExchangeCode:        403,
	ErrCodeNoWorkAgencyId:              403,
	ErrCodeEndpointGone:                451,
	ErrCodeRoleUnauthorized:            403,
	ErrCodeDocumentGenerateBadTemplate: 400,
}

var (
	ErrRequestJsonDecode      = ec.NewErrorBasic(ErrCodeRequestJsonDecode, Errs[ErrCodeRequestJsonDecode])
	ErrRequestBodyNil         = ec.NewErrorBasic(ErrCodeRequestBodyNil, Errs[ErrCodeRequestBodyNil])
	ErrRequestQueryParamParse = ec.NewErrorBasic(ErrCodeRequestQueryParamParse, Errs[ErrCodeRequestQueryParamParse])
	ErrListCountPerPage       = ec.NewErrorBasic(ErrCodeListCountPerPage, Errs[ErrCodeListCountPerPage])
	ErrListPageNumber         = ec.NewErrorBasic(ErrCodeListPageNumber, Errs[ErrCodeListPageNumber])
	ErrStorageFileNotFound    = ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound])
	ErrEntryNotFound          = ec.NewErrorBasic(ErrCodeEntryNotFound, Errs[ErrCodeEntryNotFound])
	ErrUuidInvalid            = ec.NewErrorBasic(ErrCodeUuidInvalid, Errs[ErrCodeUuidInvalid])
	ErrInvalidUser            = ec.NewErrorBasic(ErrCodeInvalidUser, Errs[ErrCodeInvalidUser])
	ErrNoWorkAgencyId         = ec.NewErrorBasic(ErrCodeNoWorkAgencyId, Errs[ErrCodeNoWorkAgencyId])
)
