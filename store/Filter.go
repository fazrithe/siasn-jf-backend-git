package store

import (
	"github.com/if-itb/siasn-jf-backend/store/models"
)

// ActivityAdmissionSearchFilter represents the filters that can be applied when searching for admissions.
type ActivityAdmissionSearchFilter struct {
	// WorkAgencyId the ID of work agency (instansi kerja) an admission is tied to.
	WorkAgencyId string

	// AdmissionDate is the exact date to search the admissions.
	// The search will be performed between this value and one day after this value.
	AdmissionDate models.Iso8601Date

	// AdmissionStatus is one of the statuses in models.ActivityAdmissionStatuses
	AdmissionStatus int

	// AdmissionType is one of the types in models.ActivityTypes
	AdmissionType int

	PageNumber int `json:"halaman"`

	CountPerPage int `json:"jumlah_per_halaman"`
}

// RequirementAdmissionSearchFilter represents the filters that can be applied when searching for requirements.
type RequirementAdmissionSearchFilter struct {
	// AgencyId the ID of work agency (instansi kerja) an admission is tied to.
	AgencyId string

	// AdmissionDate is the exact date to search the admissions.
	// The search will be performed between this value and one day after this value.
	AdmissionDate models.Iso8601Date

	// AdmissionStatus is one of the statuses in models.RequirementAdmissionStatuses
	AdmissionStatus int

	PageNumber int `json:"halaman"`

	CountPerPage int `json:"jumlah_per_halaman"`
}

// PromotionAdmissionSearchFilter represents the filters that can be applied when searching for promotion.
type PromotionAdmissionSearchFilter struct {
	// AdmissionDate is the exact date to search the admissions.
	// The search will be performed between this value and one day after this value.
	AdmissionDate models.Iso8601Date

	// AdmissionStatus is one of the statuses in models.PromotionAdmissionStatuses
	AdmissionStatus int

	// AdmissionType is one of the types in models.PromotionTypes
	AdmissionType int

	PageNumber int `json:"halaman"`

	CountPerPage int `json:"jumlah_per_halaman"`
}

// PromotionCpnsAdmissionSearchFilter represents the filters that can be applied when searching for CPNS promotion.
type PromotionCpnsAdmissionSearchFilter struct {
	// AdmissionDate is the exact date to search the admissions.
	// The search will be performed between this value and one day after this value.
	AdmissionDate string `json:"tgl_usulan" schema:"tgl_usulan"`

	// AdmissionStatus is one of the statuses in models.PromotionAdmissionStatuses
	AdmissionStatus int `json:"status" schema:"status"`

	PageNumber int `json:"halaman" schema:"halaman"`

	CountPerPage int `json:"jumlah_per_halaman" schema:"jumlah_per_halaman"`
}

// AssessmentTeamAdmissionSearchFilter represents the filters that can be applied when searching for assessment team admissions.
type AssessmentTeamAdmissionSearchFilter struct {
	// AgencyId the ID of work agency (instansi kerja) an admission is tied to.
	AgencyId string

	// SubmitterAsnId the ASN ID of admission submitter.
	SubmitterAsnId string `json:"-"`

	// AdmissionDate is the exact date to search the admissions.
	// The search will be performed between this value and one day after this value.
	AdmissionDate models.Iso8601Date

	// AdmissionStatus is one of the statuses in models.ActivityAdmissionStatuses
	AdmissionStatus int

	PageNumber int `json:"halaman"`

	CountPerPage int `json:"jumlah_per_halaman"`
}
