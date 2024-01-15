package errnum

import "github.com/if-itb/siasn-libs-backend/ec"

const (
	// ErrCodePromotionAdmissionGeneric - 13401: generic promotion admission error
	ErrCodePromotionAdmissionGeneric = iota + ServiceErrorCode*10000 + 3*1000 + 401
	// ErrCodePromotionAdmissionFieldEmpty - 13402: one or more of required request field(s) is empty.
	ErrCodePromotionAdmissionFieldEmpty
	// ErrCodePromotionAdmissionInvalidPromotionType - 13403: promotion type can only be one of (1-perpindahan jabatan, 2-promosi, 3-inpassing).
	ErrCodePromotionAdmissionInvalidPromotionType
	// ErrCodePromotionAdmissionAsnNotFound - 13404: ASN for promotion not found.
	ErrCodePromotionAdmissionAsnNotFound
	// ErrCodePromotionAdmissionInvalidCompetencyTestStatus - 13405: competency test status can only be one of (1-Lulus, 2-Tidak Lulus).
	ErrCodePromotionAdmissionInvalidCompetencyTestStatus
	// ErrCodePromotionAdmissionPakLetterEmpty - 13406: PAK letter field is empty.
	ErrCodePromotionAdmissionPakLetterEmpty
	// ErrCodePromotionAdmissionRecommendationLetterEmpty - 13407: recommendation letter field is empty.
	ErrCodePromotionAdmissionRecommendationLetterEmpty
	// ErrCodePromotionAdmissionPositionNotValid - 13408: promotion position is invalid.
	ErrCodePromotionAdmissionPositionNotValid
	// ErrCodePromotionAdmissionStatusProcessedFurther - 13409: Cannot set promotion admission status to accepted, promotion already processed further.
	ErrCodePromotionAdmissionStatusProcessedFurther
	// ErrCodePromotionAdmissionStatusAlreadyAccepted - 13410: Promotion admission already accepted.
	ErrCodePromotionAdmissionStatusAlreadyAccepted
	// ErrCodePromotionAdmissionStatusAlreadyRejected - 13411: Promotion admission already rejected.
	ErrCodePromotionAdmissionStatusAlreadyRejected
	// ErrCodePromotionFilterInvalidStatus - 13412: admission status supplied contains value outside the valid range (1-3).
	ErrCodePromotionFilterInvalidStatus
	// ErrCodePromotionFilterInvalidDate - 13413: the date format supplied does not conform to the date format required.
	ErrCodePromotionFilterInvalidDate
	// ErrCodePromotionFilterInvalidType - 13414: admission type supplied contains value outside the valid range (1-3).
	ErrCodePromotionFilterInvalidType
	// ErrCodePromotionTestCertificateEmpty - 13415: test certificate field is empty.
	ErrCodePromotionTestCertificateEmpty
	// ErrCodePromotionInvalidDate - 13416: the date format supplied does not conform to the date format required.
	ErrCodePromotionInvalidDate
	// ErrCodePromotionAdmissionStatusNotAccepted - 13417: Promotion admission is not accepted.
	ErrCodePromotionAdmissionStatusNotAccepted
)

var (
	ErrPromotionAdmissionInvalidPromotionType        *ec.Error
	ErrPromotionAdmissionAsnNotFound                 *ec.Error
	ErrPromotionAdmissionInvalidCompetencyTestStatus *ec.Error
	ErrPromotionAdmissionPakLetterEmpty              *ec.Error
	ErrPromotionAdmissionRecommendationLetterEmpty   *ec.Error
	ErrPromotionAdmissionTestCertificateEmpty        *ec.Error
	ErrPromotionAdmissionPositionNotValid            *ec.Error
	ErrPromotionAdmissionStatusProcessedFurther      *ec.Error
	ErrPromotionAdmissionStatusAlreadyAccepted       *ec.Error
	ErrPromotionAdmissionStatusAlreadyRejected       *ec.Error
	ErrPromotionAdmissionStatusNotAccepted           *ec.Error
	ErrPromotionFilterInvalidStatus                  *ec.Error
	ErrPromotionFilterInvalidDate                    *ec.Error
	ErrPromotionFilterInvalidType                    *ec.Error
)

func init() {
	Errs[ErrCodePromotionAdmissionGeneric] = "error creating new promotion admission"
	Errs[ErrCodePromotionAdmissionFieldEmpty] = "one or more of required request field(s) is empty"
	Errs[ErrCodePromotionAdmissionInvalidPromotionType] = "promotion type can only be one of (1-perpindahan jabatan, 2-promosi, 3-inpassing)"
	Errs[ErrCodePromotionAdmissionAsnNotFound] = "ASN for promotion not found"
	Errs[ErrCodePromotionAdmissionInvalidCompetencyTestStatus] = "competency test status can only be one of (1-Lulus, 2-Tidak Lulus)"
	Errs[ErrCodePromotionAdmissionPakLetterEmpty] = "surat_pak cannot be empty"
	Errs[ErrCodePromotionAdmissionRecommendationLetterEmpty] = "surat_rekomendasi cannot be empty"
	Errs[ErrCodePromotionTestCertificateEmpty] = "sertifikat_uji_kompetensi cannot be empty"
	Errs[ErrCodePromotionAdmissionPositionNotValid] = "promotion position is invalid"
	Errs[ErrCodePromotionAdmissionStatusProcessedFurther] = "cannot set promotion admission status to accepted, promotion already processed further"
	Errs[ErrCodePromotionAdmissionStatusAlreadyAccepted] = "promotion admission already accepted"
	Errs[ErrCodePromotionAdmissionStatusAlreadyRejected] = "promotion admission already rejected"
	Errs[ErrCodePromotionFilterInvalidStatus] = "admission status must be >= 1 and <= 3"
	Errs[ErrCodePromotionFilterInvalidDate] = "date must be in the format of YYYY-MM-DD (e.g. 2006-12-31)"
	Errs[ErrCodePromotionFilterInvalidType] = "type (jenis_pengangkatan) must be >= 1 and <= 3"
	Errs[ErrCodePromotionInvalidDate] = "date must be in the format of YYYY-MM-DD (e.g. 2006-12-31)"
	Errs[ErrCodePromotionAdmissionStatusNotAccepted] = "promotion admission is not accepted"

	ErrsToHttp[ErrCodePromotionAdmissionFieldEmpty] = 400
	ErrsToHttp[ErrCodePromotionAdmissionInvalidPromotionType] = 400
	ErrsToHttp[ErrCodePromotionAdmissionAsnNotFound] = 400
	ErrsToHttp[ErrCodePromotionAdmissionInvalidCompetencyTestStatus] = 400
	ErrsToHttp[ErrCodePromotionAdmissionPakLetterEmpty] = 400
	ErrsToHttp[ErrCodePromotionAdmissionRecommendationLetterEmpty] = 400
	ErrsToHttp[ErrCodePromotionTestCertificateEmpty] = 400
	ErrsToHttp[ErrCodePromotionAdmissionPositionNotValid] = 400
	ErrsToHttp[ErrCodePromotionAdmissionStatusProcessedFurther] = 400
	ErrsToHttp[ErrCodePromotionAdmissionStatusAlreadyAccepted] = 400
	ErrsToHttp[ErrCodePromotionAdmissionStatusAlreadyRejected] = 400
	ErrsToHttp[ErrCodePromotionFilterInvalidStatus] = 400
	ErrsToHttp[ErrCodePromotionFilterInvalidDate] = 400
	ErrsToHttp[ErrCodePromotionFilterInvalidType] = 400
	ErrsToHttp[ErrCodePromotionInvalidDate] = 400
	ErrsToHttp[ErrCodePromotionAdmissionStatusNotAccepted] = 400

	ErrPromotionAdmissionInvalidPromotionType = ec.NewErrorBasic(ErrCodePromotionAdmissionInvalidPromotionType, Errs[ErrCodePromotionAdmissionInvalidPromotionType])
	ErrPromotionAdmissionAsnNotFound = ec.NewErrorBasic(ErrCodePromotionAdmissionAsnNotFound, Errs[ErrCodePromotionAdmissionAsnNotFound])
	ErrPromotionAdmissionInvalidCompetencyTestStatus = ec.NewErrorBasic(ErrCodePromotionAdmissionInvalidCompetencyTestStatus, Errs[ErrCodePromotionAdmissionInvalidCompetencyTestStatus])
	ErrPromotionAdmissionPakLetterEmpty = ec.NewErrorBasic(ErrCodePromotionAdmissionPakLetterEmpty, Errs[ErrCodePromotionAdmissionPakLetterEmpty])
	ErrPromotionAdmissionRecommendationLetterEmpty = ec.NewErrorBasic(ErrCodePromotionAdmissionRecommendationLetterEmpty, Errs[ErrCodePromotionAdmissionRecommendationLetterEmpty])
	ErrPromotionAdmissionTestCertificateEmpty = ec.NewErrorBasic(ErrCodePromotionTestCertificateEmpty, Errs[ErrCodePromotionTestCertificateEmpty])
	ErrPromotionAdmissionPositionNotValid = ec.NewErrorBasic(ErrCodePromotionAdmissionPositionNotValid, Errs[ErrCodePromotionAdmissionPositionNotValid])
	ErrPromotionAdmissionStatusProcessedFurther = ec.NewErrorBasic(ErrCodePromotionAdmissionStatusProcessedFurther, Errs[ErrCodePromotionAdmissionStatusProcessedFurther])
	ErrPromotionAdmissionStatusAlreadyAccepted = ec.NewErrorBasic(ErrCodePromotionAdmissionStatusAlreadyAccepted, Errs[ErrCodePromotionAdmissionStatusAlreadyAccepted])
	ErrPromotionAdmissionStatusAlreadyRejected = ec.NewErrorBasic(ErrCodePromotionAdmissionStatusAlreadyRejected, Errs[ErrCodePromotionAdmissionStatusAlreadyRejected])
	ErrPromotionFilterInvalidStatus = ec.NewErrorBasic(ErrCodePromotionFilterInvalidStatus, Errs[ErrCodePromotionFilterInvalidStatus])
	ErrPromotionFilterInvalidDate = ec.NewErrorBasic(ErrCodePromotionFilterInvalidDate, Errs[ErrCodePromotionFilterInvalidDate])
	ErrPromotionFilterInvalidType = ec.NewErrorBasic(ErrCodePromotionFilterInvalidType, Errs[ErrCodePromotionFilterInvalidType])
	ErrPromotionAdmissionStatusNotAccepted = ec.NewErrorBasic(ErrCodePromotionAdmissionStatusNotAccepted, Errs[ErrCodePromotionAdmissionStatusNotAccepted])
}
