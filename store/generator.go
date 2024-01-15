package store

import (
	"context"
	"io"
	"os"
	"path"

	"github.com/fazrithe/siasn-jf-backend-git/libs/docx"
	"github.com/google/uuid"
)

const (
	TemplateFilenameActivityCertificate             = "template/activity/certificate.docx"
	TemplateFilenameRequirementRecommendationLetter = "template/requirement/recommendation-letter.docx"
	TemplateFilenamePromotionLetter                 = "template/promotion/promotion-letter.docx"
	TemplateFilenameDismissalAcceptanceLetter       = "template/dismissal/acceptance-letter.docx"
)

var activityTemplatePaths = map[string]struct{}{
	TemplateFilenameActivityCertificate: {},
}

var requirementTemplatePaths = map[string]struct{}{
	TemplateFilenameRequirementRecommendationLetter: {},
}

var promotionTemplatePaths = map[string]struct{}{
	TemplateFilenamePromotionLetter: {},
}

var dismissalTemplatePaths = map[string]struct{}{
	TemplateFilenameDismissalAcceptanceLetter: {},
}

// loadActivityTemplate loads template from activity object storage to a local path.
// Does not return ec.Error.
func (c *Client) loadActivityTemplate(ctx context.Context, templateFilename string, localOutputPath string) (err error) {
	out, err := c.ActivityStorage.GetActivityFile(ctx, templateFilename)
	if err != nil {
		return err
	}

	f, err := os.Create(localOutputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	buffer := make([]byte, 100*1024) // 100KB buffer
	_, err = io.CopyBuffer(f, out, buffer)
	if err != nil {
		return err
	}

	return nil
}

// loadRequirementTemplate loads template from requirement object storage to a local path.
// Does not return ec.Error.
func (c *Client) loadRequirementTemplate(ctx context.Context, templateFilename string, localOutputPath string) (err error) {
	out, err := c.RequirementStorage.GetRequirementFile(ctx, templateFilename)
	if err != nil {
		return err
	}

	f, err := os.Create(localOutputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	buffer := make([]byte, 100*1024) // 100KB buffer
	_, err = io.CopyBuffer(f, out, buffer)
	if err != nil {
		return err
	}

	return nil
}

// loadPromotionTemplate loads template from promotion object storage to a local path.
// Does not return ec.Error.
func (c *Client) loadPromotionTemplate(ctx context.Context, templateFilename string, localOutputPath string) (err error) {
	out, err := c.PromotionStorage.GetPromotionFile(ctx, templateFilename)
	if err != nil {
		return err
	}

	f, err := os.Create(localOutputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	buffer := make([]byte, 100*1024) // 100KB buffer
	_, err = io.CopyBuffer(f, out, buffer)
	if err != nil {
		return err
	}

	return nil
}

// loadDismissalTemplate loads template from promotion object storage to a local path.
// Does not return ec.Error.
func (c *Client) loadDismissalTemplate(ctx context.Context, templateFilename string, localOutputPath string) (err error) {
	out, err := c.DismissalStorage.GetDismissalFile(ctx, templateFilename)
	if err != nil {
		return err
	}

	f, err := os.Create(localOutputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	buffer := make([]byte, 100*1024) // 100KB buffer
	_, err = io.CopyBuffer(f, out, buffer)
	if err != nil {
		return err
	}

	return nil
}

// renderDocxTemplateCtx renders a template pointed in localTemplatePath (in local disk) and store it in object storage with filename as filename.
// The generated file is a pdf, so filename should have a .pdf extension, although this is not mandatory. Content-Type
// is automatically set to "application/pdf".
//
// data must be JSON marshal-able. The placeholders in template are matched with the JSON field names from data.
func (c *Client) renderDocxTemplateCtx(ctx context.Context, localTemplatePath string, filename string, data interface{}, putFunc func(ctx context.Context, filename string, contentType string, file io.ReadSeeker) (err error)) (err error) {
	localOutputPath := path.Join(os.TempDir(), uuid.NewString())

	err = c.DocxRenderer.RenderAsPdfCtx(ctx, data, localTemplatePath, localOutputPath)
	if err != nil {
		return err
	}
	defer os.Remove(localOutputPath) // Delete result after saving to object storage.

	f, err := os.Open(localOutputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	err = putFunc(ctx, filename, "application/pdf", f)
	if err != nil {
		return err
	}

	return nil
}

type ActivityCertificateTemplate struct {
	AttendeeName       string            `json:"nama"`
	AttendeeNip        string            `json:"nip"`
	AttendeeBirthday   string            `json:"tempat_lahir"`
	AttendeeBirthplace string            `json:"tgl_lahir"`
	AttendeePicture    *docx.InlineImage `json:"foto"`
	FunctionalPosition string            `json:"jabatan_fungsional"`
	Agency             string            `json:"instansi"`
	OrganizerAgency    string            `json:"instansi_penyelenggara"`
	Qualification      string            `json:"kualifikasi"`
	ActivityName       string            `json:"kegiatan"`
	// Duration, in hour.
	Duration        string `json:"durasi"`
	AdmissionNumber string `json:"no_usulan"`
	StartDate       string `json:"tgl_mulai"`
	EndDate         string `json:"tgl_selesai"`
	Description     string `json:"deskripsi"`
	DocumentNumber  string `json:"no_dokumen"`
	DocumentDate    string `json:"tgl_dokumen"`
}

// generateActivityCertificateCtx generates activity certificate and store it in object storage with filename as key.
// The generated file is a pdf, so filename should have a .pdf extension, although this is not mandatory. Content-Type
// is automatically set to "application/pdf".
func (c *Client) generateActivityCertificateCtx(ctx context.Context, filename string, data *ActivityCertificateTemplate) (err error) {
	localTemplatePath := path.Join(os.TempDir(), uuid.NewString())
	err = c.loadActivityTemplate(ctx, TemplateFilenameActivityCertificate, localTemplatePath)
	if err != nil {
		return err
	}
	defer os.Remove(localTemplatePath) // Delete template after rendering, load template again later, on and on, because template can be changed on runtime anytime.

	return c.renderDocxTemplateCtx(ctx, localTemplatePath, filename, data, c.ActivityStorage.PutActivityFile)
}

type RequirementRecommendationLetterTemplate struct {
	DocumentNumber   string `json:"no_dokumen"`
	DocumentDate     string `json:"tgl_dokumen"`
	Position         string `json:"jabatan_fungsional"`
	Agency           string `json:"instansi"`
	CalculationCount int    `json:"total_perhitungan"`
	RequirementCount int    `json:"total_kebutuhan"`

	Requirements []*RequirementRecommendationLetterTemplateEntry `json:"kebutuhan"`
}

type RequirementRecommendationLetterTemplateEntry struct {
	FunctionalPosition     string                                              `json:"jabatan_jenjang_nama"`
	SubtotalBezetting      int                                                 `json:"subtotal_bezetting"`
	SubtotalCalculation    int                                                 `json:"subtotal_perhitungan"`
	SubtotalRequirement    int                                                 `json:"subtotal_kebutuhan"`
	SubtotalRecommendation int                                                 `json:"subtotal_rekomendasi"`
	OrganizationUnits      []*RequirementRecommendationLetterTemplateEntryUnor `json:"unor"`
}

type RequirementRecommendationLetterTemplateEntryUnor struct {
	OrganizationUnit string `json:"unit_organisasi"`
	Bezetting        int    `json:"bezetting"`
	Calculation      int    `json:"perhitungan"`
	Requirement      int    `json:"kebutuhan"`
	Recommendation   int    `json:"rekomendasi"`
}

// generateRequirementRecommendationLetterCtx generates requirement recommendation letter and store it in object storage with filename as key.
// The generated file is a pdf, so filename should have a .pdf extension, although this is not mandatory. Content-Type
// is automatically set to "application/pdf".
func (c *Client) generateRequirementRecommendationLetterCtx(ctx context.Context, filename string, data *RequirementRecommendationLetterTemplate) (err error) {
	localTemplatePath := path.Join(os.TempDir(), uuid.NewString())
	err = c.loadRequirementTemplate(ctx, TemplateFilenameRequirementRecommendationLetter, localTemplatePath)
	if err != nil {
		return err
	}
	defer os.Remove(localTemplatePath) // Delete template after rendering, load template again later, on and on, because template can be changed on runtime anytime.

	return c.renderDocxTemplateCtx(ctx, localTemplatePath, filename, data, c.RequirementStorage.PutRequirementFile)
}

type PromotionLetterTemplate struct {
	AdmissionNumber       string `json:"nomor_usulan"`
	AdmissionDate         string `json:"tanggal_usulan"`
	Name                  string `json:"nama"`
	PromotionPositionName string `json:"nama_jf"`
	SignedDate            string `json:"tanggal_ttd"`
}

// generatePromotionLetterCtx generates promotion letter and store it in object storage with filename as key.
// The generated file is a pdf, so filename should have a .pdf extension, although this is not mandatory. Content-Type
// is automatically set to "application/pdf".
func (c *Client) generatePromotionLetterCtx(ctx context.Context, filename string, data *PromotionLetterTemplate) (err error) {
	localTemplatePath := path.Join(os.TempDir(), uuid.NewString())
	err = c.loadPromotionTemplate(ctx, TemplateFilenamePromotionLetter, localTemplatePath)
	if err != nil {
		return err
	}
	defer os.Remove(localTemplatePath) // Delete template after rendering, load template again later, on and on, because template can be changed on runtime anytime.

	return c.renderDocxTemplateCtx(ctx, localTemplatePath, filename, data, c.PromotionStorage.PutPromotionFile)
}

type DismissalAcceptanceTemplate struct {
	DocumentNumber   string `json:"no_dokumen"`
	DocumentDate     string `json:"tgl_dokumen"`
	DecreeNumber     string `json:"no_sk"`
	DecreeDate       string `json:"tgl_sk"`
	DismissalDate    string `json:"tgl_pemberhentian"`
	DismissalReason  string `json:"alasan_pemberhentian"`
	AsnName          string `json:"nama"`
	AsnNip           string `json:"nip"`
	AsnGrade         string `json:"pangkat"`
	Position         string `json:"jabatan_fungsional"`
	OrganizationUnit string `json:"unor"`
}

// generateDismissalAcceptanceLetterCtx generates dismissal acceptance letter and store it in object storage with filename as key.
// The generated file is a pdf, so filename should have a .pdf extension, although this is not mandatory. Content-Type
// is automatically set to "application/pdf".
func (c *Client) generateDismissalAcceptanceLetterCtx(ctx context.Context, filename string, data *DismissalAcceptanceTemplate) (err error) {
	localTemplatePath := path.Join(os.TempDir(), uuid.NewString())
	err = c.loadDismissalTemplate(ctx, TemplateFilenameDismissalAcceptanceLetter, localTemplatePath)
	if err != nil {
		return err
	}
	defer os.Remove(localTemplatePath) // Delete template after rendering, load template again later, on and on, because template can be changed on runtime anytime.

	return c.renderDocxTemplateCtx(ctx, localTemplatePath, filename, data, c.DismissalStorage.PutDismissalFile)
}
