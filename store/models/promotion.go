package models

const (
	PromotionAdmissionStatusCreated = iota + 1
	PromotionAdmissionStatusAccepted
	PromotionAdmissionStatusRejected
)

var PromotionAdmissionStatuses = map[int]struct{}{
	PromotionAdmissionStatusCreated:  {},
	PromotionAdmissionStatusAccepted: {},
	PromotionAdmissionStatusRejected: {},
}

const (
	PromotionTypeTransfer = iota + 1
	PromotionTypePromotion
	PromotionTypeInPassing
)

var PromotionTypes = map[int]struct{}{
	PromotionTypeTransfer:  {},
	PromotionTypePromotion: {},
	PromotionTypeInPassing: {},
}

const (
	PromotionCompetencyTestStatusPass = iota + 1
	PromotionCompetencyTestStatusFail
)

var PromotionCompetencyTestStatuses = map[int]struct{}{
	PromotionCompetencyTestStatusPass: {},
	PromotionCompetencyTestStatusFail: {},
}

type PromotionAdmission struct {
	AdmissionNumber string      `json:"nomor_usulan"`
	AdmissionDate   Iso8601Date `json:"tanggal_usulan"`

	PromotionId string `json:"pengangkatan_id"`
	AsnId       string `json:"asn_id"`

	TestStatus int     `json:"test_status"`
	TestScore  float64 `json:"test_nilai,omitempty"`

	PromotionType       int    `json:"jenis_pengangkatan"`
	PromotionPositionId string `json:"jabatan_fungsional_tujuan_id"`

	PakLetter            *Document `json:"surat_pak,omitempty"`
	RecommendationLetter *Document `json:"surat_rekomendasi,omitempty"`
	TestCertificate      *Document `json:"sertifikat_uji_kompetensi,omitempty"`

	PromotionLetter *Document `json:"surat_pengangkatan,omitempty"`

	PromotionDenialReason string `json:"alasan_penolakan_pengangkatan,omitempty"`

	// SubmitterAsnId is the ASN ID of the submitter (the user), can be retrieved from ID token.
	SubmitterAsnId string `json:"-"`
	AgencyId       string `json:"-"`
}

type PromotionReject struct {
	PromotionId  string `json:"pengangkatan_id"`
	RejectReason string `json:"alasan_tidak_diangkat"`

	// SubmitterAsnId is the ASN ID of the submitter (the user), can be retrieved from ID token.
	SubmitterAsnId string `json:"-"`
}

type PromotionItem struct {
	PromotionId   string `json:"pengangkatan_id"`
	AsnId         string `json:"asn_id"`
	Name          string `json:"nama"`
	Status        int    `json:"status"`
	PromotionType int    `json:"jenis_pengangkatan"`

	// RecommendationLetterDate is when the recommendation letter published.
	RecommendationLetterDate Iso8601Date `json:"tgl_doc_surat_rekomendasi"`
}
