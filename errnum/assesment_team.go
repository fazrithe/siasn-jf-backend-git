package errnum

import "github.com/if-itb/siasn-libs-backend/ec"

const (
	// ErrCodeAssessmentTeamAdmissionAssessorCountEven - 18401: the number of assessors must be odd.
	ErrCodeAssessmentTeamAdmissionAssessorCountEven = iota + ServiceErrorCode*10000 + 8*1000 + 401
	// ErrCodeAssessmentTeamAdmissionAssessorCountInvalid - 18402: at least three assessors must be supplied.
	ErrCodeAssessmentTeamAdmissionAssessorCountInvalid
	// ErrCodeAssessmentTeamAdmissionNumberInvalid - 18403: admission number is invalid.
	ErrCodeAssessmentTeamAdmissionNumberInvalid
	// ErrCodeAssessmentTeamAdmissionFunctionalPositionIdInvalid - 18404: functional position id is invalid.
	ErrCodeAssessmentTeamAdmissionFunctionalPositionIdInvalid
	// ErrCodeAssessmentTeamAssessorRoleInvalid - 18405: assessor role is invalid.
	ErrCodeAssessmentTeamAssessorRoleInvalid
	// ErrCodeAssessmentTeamAdmissionFilterInvalidStatus - 18406: admission status supplied contains value outside the valid range (1-2).
	ErrCodeAssessmentTeamAdmissionFilterInvalidStatus
	// ErrCodeAssessmentTeamAdmissionFilterInvalidDate - 18407: the date format supplied does not conform to the date format required.
	ErrCodeAssessmentTeamAdmissionFilterInvalidDate
	// ErrCodeAssessmentTeamStatusNotCreated - 18408: assessment team status is not created.
	ErrCodeAssessmentTeamStatusNotCreated
	// ErrCodeAssessmentTeamAssessorStatusInvalid - 18409: assessor status supplied contains value outside the valid range (1-2).
	ErrCodeAssessmentTeamAssessorStatusInvalid
)

var (
	ErrAssessmentTeamAdmissionAssessorCountEven           *ec.Error
	ErrAssessmentTeamAdmissionAssessorCountInvalid        *ec.Error
	ErrAssessmentTeamAdmissionNumberInvalid               *ec.Error
	ErrAssessmentTeamAdmissionFunctionalPositionIdInvalid *ec.Error
	ErrAssessmentTeamAssessorRoleInvalid                  *ec.Error
	ErrAssessmentTeamAdmissionFilterInvalidStatus         *ec.Error
	ErrAssessmentTeamAdmissionFilterInvalidDate           *ec.Error
	ErrAssessmentTeamAssessorStatusInvalid                *ec.Error
)

func init() {
	Errs[ErrCodeAssessmentTeamAdmissionAssessorCountEven] = "the number of assessors must be odd"
	Errs[ErrCodeAssessmentTeamAdmissionAssessorCountInvalid] = "at least three assessors must be supplied"
	Errs[ErrCodeAssessmentTeamAdmissionNumberInvalid] = "admission number is invalid"
	Errs[ErrCodeAssessmentTeamAdmissionFunctionalPositionIdInvalid] = "functional position id is invalid"
	Errs[ErrCodeAssessmentTeamAssessorRoleInvalid] = "assessor role is invalid"
	Errs[ErrCodeAssessmentTeamAdmissionFilterInvalidStatus] = "admission status must be >= 1 and <= 2"
	Errs[ErrCodeAssessmentTeamAdmissionFilterInvalidDate] = "date must be in the format of YYYY-MM-DD (e.g. 2006-12-31)"
	Errs[ErrCodeAssessmentTeamStatusNotCreated] = "assessment team status is not created"
	Errs[ErrCodeAssessmentTeamAssessorStatusInvalid] = "assessor status supplied contains value outside the valid range (1-2)"

	ErrAssessmentTeamAdmissionAssessorCountEven = ec.NewErrorBasic(ErrCodeAssessmentTeamAdmissionAssessorCountEven, Errs[ErrCodeAssessmentTeamAdmissionAssessorCountEven])
	ErrAssessmentTeamAdmissionAssessorCountInvalid = ec.NewErrorBasic(ErrCodeAssessmentTeamAdmissionAssessorCountInvalid, Errs[ErrCodeAssessmentTeamAdmissionAssessorCountInvalid])
	ErrAssessmentTeamAdmissionNumberInvalid = ec.NewErrorBasic(ErrCodeAssessmentTeamAdmissionNumberInvalid, Errs[ErrCodeAssessmentTeamAdmissionNumberInvalid])
	ErrAssessmentTeamAdmissionFunctionalPositionIdInvalid = ec.NewErrorBasic(ErrCodeAssessmentTeamAdmissionFunctionalPositionIdInvalid, Errs[ErrCodeAssessmentTeamAdmissionFunctionalPositionIdInvalid])
	ErrAssessmentTeamAssessorRoleInvalid = ec.NewErrorBasic(ErrCodeAssessmentTeamAssessorRoleInvalid, Errs[ErrCodeAssessmentTeamAssessorRoleInvalid])
	ErrAssessmentTeamAdmissionFilterInvalidStatus = ec.NewErrorBasic(ErrCodeAssessmentTeamAdmissionFilterInvalidStatus, Errs[ErrCodeAssessmentTeamAdmissionFilterInvalidStatus])
	ErrAssessmentTeamAdmissionFilterInvalidDate = ec.NewErrorBasic(ErrCodeAssessmentTeamAdmissionFilterInvalidDate, Errs[ErrCodeAssessmentTeamAdmissionFilterInvalidDate])
	ErrAssessmentTeamAssessorStatusInvalid = ec.NewErrorBasic(ErrCodeAssessmentTeamAssessorStatusInvalid, Errs[ErrCodeAssessmentTeamAssessorStatusInvalid])

}
