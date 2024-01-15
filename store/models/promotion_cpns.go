package models

const (
	PromotionCpnsAdmissionStatusCreated = iota + 1
	PromotionCpnsAdmissionStatusAccepted
	PromotionCpnsAdmissionStatusRejected
)

var PromotionCpnsAdmissionStatuses = map[int]struct{}{
	PromotionCpnsAdmissionStatusCreated:  {},
	PromotionCpnsAdmissionStatusAccepted: {},
	PromotionCpnsAdmissionStatusRejected: {},
}

type PromotionCpnsAdmission struct {
	AdmissionNumber string      `json:"nomor_usulan"`
	AdmissionDate   Iso8601Date `json:"tanggal_usulan"`

	PromotionCpnsId string `json:"pengangkatan_cpns_id"`
	AsnId           string `json:"asn_id"`
	AsnName         string `json:"nama"`
	AsnNip          string `json:"nip"`

	Status              int    `json:"status"`
	PromotionPositionId string `json:"jabatan_fungsional_tujuan_id"`
	PromotionPosition   string `json:"jabatan_fungsional_tujuan"`
	FirstCreditNumber   int    `json:"angka_kredit_pertama"`
	OrganizationUnitId  string `json:"unor_id"`
	OrganizationUnit    string `json:"unor"`

	PakLetter       *Document `json:"surat_pak,omitempty"`
	PromotionLetter *Document `json:"surat_pengangkatan,omitempty"`

	// SubmitterAsnId is the ASN ID of the submitter (the user), can be retrieved from ID token.
	SubmitterAsnId string `json:"-"`
}

type PromotionCpnsItem struct {
	PromotionCpnsId     string      `json:"pengangkatan_cpns_id"`
	AsnId               string      `json:"asn_id"`
	AsnNip              string      `json:"nip"`
	AsnName             string      `json:"nama"`
	Status              int         `json:"status"`
	PromotionPositionId string      `json:"jabatan_fungsional_tujuan_id"`
	PromotionPosition   string      `json:"jabatan_fungsional_tujuan"`
	FirstCreditNumber   int         `json:"angka_kredit_pertama"`
	OrganizationUnitId  string      `json:"unor_id"`
	OrganizationUnit    string      `json:"unor"`
	AdmissionNumber     string      `json:"nomor_usulan"`
	AdmissionDate       Iso8601Date `json:"tanggal_usulan"`
}
