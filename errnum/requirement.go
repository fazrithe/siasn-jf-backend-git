package errnum

const (
	// ErrCodeRequirementAdmissionPositionGradeEmpty - 16401: jabatan_jenjang is empty.
	ErrCodeRequirementAdmissionPositionGradeEmpty = iota + ServiceErrorCode*10000 + 6*1000 + 401
	// ErrCodeRequirementAdmissionRequirementCountInvalid - 16402: Jumlah kebutuhan must be > 0.
	ErrCodeRequirementAdmissionRequirementCountInvalid
	// ErrCodeRequirementAdmissionNoEstimationDocs - 16403: No estimation docs saved, must supply at least one filename.
	ErrCodeRequirementAdmissionNoEstimationDocs
	// ErrCodeRequirementAdmissionCoverLetterInvalid - 16404: Supplied cover letter is invalid.
	ErrCodeRequirementAdmissionCoverLetterInvalid
	// ErrCodeRequirementAdmissionNoFiscalYear - 16405: Fiscal year required.
	ErrCodeRequirementAdmissionNoFiscalYear
	// ErrCodeRequirementAdmissionNoAdmissionNumber - 16406: Admission number is needed.
	ErrCodeRequirementAdmissionNoAdmissionNumber
)

const (
	// ErrCodeRequirementAdmissionBezetting - 16501: Failed to calculate the bezetting of a requirement admission.
	ErrCodeRequirementAdmissionBezetting = iota + ServiceErrorCode*10000 + 6*1000 + 501
)

const (
	// ErrCodeRequirementFilterStatusInvalid - 17041: admission status supplied contains value outside the valid range (1-5).
	ErrCodeRequirementFilterStatusInvalid = iota + ServiceErrorCode*10000 + 7*1000 + 401
	// ErrCodeRequirementFilterInvalidDate - 17042: the date format supplied does not conform to the date format required.
	ErrCodeRequirementFilterInvalidDate
)

const (
	// ErrCodeRequirementVerificationStatusProcessedFurther - 18401: Cannot set requirement status to accepted or uploading a recommendation letter, requirement already processed further.
	ErrCodeRequirementVerificationStatusProcessedFurther = iota + ServiceErrorCode*10000 + 8*1000 + 401
	// ErrCodeRequirementVerificationStatusAlreadyAccepted - 18402: Requirement already accepted.
	ErrCodeRequirementVerificationStatusAlreadyAccepted
	// ErrCodeRequirementVerificationNoRecommendationLetter - 18403: Requirement has to have a signed recommendation letter first.
	ErrCodeRequirementVerificationNoRecommendationLetter
	// ErrCodeRequirementVerificationRecommendationLetterInvalid - 18404: Invalid recommendation letter for verification.
	ErrCodeRequirementVerificationRecommendationLetterInvalid
	// ErrCodeRequirementVerificationStatusNotAccepted - 18405: Requirement has not been accepted.
	ErrCodeRequirementVerificationStatusNotAccepted
	// ErrCodeRequirementVerificationGenerateMustBeFromSameAgency - 18406: requirement recommendation letter generation: all requirement IDs must be from the same agency.
	ErrCodeRequirementVerificationGenerateMustBeFromSameAgency
)

func init() {
	Errs[ErrCodeRequirementAdmissionPositionGradeEmpty] = "jabatan_jenjang is required"
	Errs[ErrCodeRequirementAdmissionRequirementCountInvalid] = "jumlah kebutuhan must be > 0"
	Errs[ErrCodeRequirementAdmissionNoEstimationDocs] = "no estimation docs (dokumen perhitungan) saved, must supply at least one filename"
	Errs[ErrCodeRequirementAdmissionCoverLetterInvalid] = "cover letter is invalid"
	Errs[ErrCodeRequirementAdmissionBezetting] = "failed to calculate the bezetting of a requirement admission"
	Errs[ErrCodeRequirementAdmissionNoFiscalYear] = "fiscal year is required"
	Errs[ErrCodeRequirementAdmissionNoAdmissionNumber] = "admission number is required"

	Errs[ErrCodeRequirementVerificationStatusProcessedFurther] = "cannot set requirement status to accepted or uploading recommendation letter, requirement already processed further"
	Errs[ErrCodeRequirementVerificationStatusAlreadyAccepted] = "this requirement is already accepted"
	Errs[ErrCodeRequirementVerificationNoRecommendationLetter] = "requirement has to have a signed recommendation letter first"
	Errs[ErrCodeRequirementVerificationRecommendationLetterInvalid] = "invalid recommendation letter"
	Errs[ErrCodeRequirementVerificationStatusNotAccepted] = "requirement has not been accepted"
	Errs[ErrCodeRequirementVerificationGenerateMustBeFromSameAgency] = "requirement recommendation letter generation: all requirement IDs must be from the same agency"

	ErrsToHttp[ErrCodeRequirementAdmissionPositionGradeEmpty] = 400
	ErrsToHttp[ErrCodeRequirementAdmissionRequirementCountInvalid] = 400
	ErrsToHttp[ErrCodeRequirementAdmissionNoEstimationDocs] = 400
	ErrsToHttp[ErrCodeRequirementAdmissionCoverLetterInvalid] = 400

	ErrsToHttp[ErrCodeRequirementVerificationStatusProcessedFurther] = 400
	ErrsToHttp[ErrCodeRequirementVerificationStatusAlreadyAccepted] = 400
	ErrsToHttp[ErrCodeRequirementVerificationNoRecommendationLetter] = 400
	ErrsToHttp[ErrCodeRequirementVerificationRecommendationLetterInvalid] = 400
	ErrsToHttp[ErrCodeRequirementVerificationStatusNotAccepted] = 400
	ErrsToHttp[ErrCodeRequirementVerificationGenerateMustBeFromSameAgency] = 400
}
