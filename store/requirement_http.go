package store

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"time"

	. "github.com/fazrithe/siasn-jf-backend-git/errnum"
	"github.com/fazrithe/siasn-jf-backend-git/libs/auth"
	"github.com/fazrithe/siasn-jf-backend-git/libs/ec"
	"github.com/fazrithe/siasn-jf-backend-git/libs/httputil"
	"github.com/fazrithe/siasn-jf-backend-git/store/models"
	"github.com/fazrithe/siasn-jf-backend-git/store/object"
	"github.com/google/uuid"
)

const (
	TimeoutRequirementAdmissionSubmit                                  = TimeoutDefault
	TimeoutRequirementAdmissionEdit                                    = TimeoutDefault
	TimeoutRequirementAdmissionCoverLetterUpload                       = TimeoutDefault
	TimeoutRequirementAdmissionEstimationDocUpload                     = TimeoutDefault
	TimeoutRequirementAdmissionCoverLetterPreview                      = TimeoutDefault
	TimeoutRequirementAdmissionEstimationDocPreview                    = TimeoutDefault
	TimeoutRequirementAdmissionCoverLetterDownload                     = TimeoutDefault
	TimeoutRequirementAdmissionEstimationDocDownload                   = TimeoutDefault
	TimeoutRequirementAdmissionSearch                                  = TimeoutDefault
	TimeoutRequirementAdmissionDetailGet                               = TimeoutDefault
	TimeoutRequirementAdmissionCoverLetterTemplateDownload             = TimeoutDefault
	TimeoutRequirementVerificationRecommendationLetterUpload           = TimeoutDefault
	TimeoutRequirementVerificationRecommendationLetterPreview          = TimeoutDefault
	TimeoutRequirementVerificationRecommendationLetterDownloadUnsigned = TimeoutDefault
	TimeoutRequirementVerificationRecommendationLetterDownloadSigned   = TimeoutDefault
	TimeoutRequirementAdmissionVerification                            = TimeoutDefault
	TimeoutRequirementVerifierGet                                      = TimeoutDefault
	TimeoutRequirementVerificationBulkSubmitRecommendationLetter       = TimeoutDefault
	TimeoutRequirementVerificationSignRecommendationLetter             = TimeoutDefault
	TimeoutGetRequirementStatusStatistic                               = TimeoutDefault
)

// HandleRequirementAdmissionSubmit handles a new admission request.
// Requirement admission is created for Modul Perhitungan. A new agency requirement entry
// will be created in the database with status set to `newly admitted`.
func (c *Client) HandleRequirementAdmissionSubmit(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)

	rr := &models.RequirementAdmission{}
	err := c.decodeRequestJson(writer, request, rr)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementAdmissionSubmit)
	defer cancel()

	rr.SubmitterAsnId = user.AsnId
	rr.AgencyId = user.WorkAgencyId
	rr.AdmissionTimestamp = models.EpochTime(time.Now())

	requirementId, err := c.InsertRequirementAdmissionCtx(ctx, rr)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]string{
		"kebutuhan_id": requirementId,
	})
}

// HandleRequirementAdmissionEdit handles an edit admission request.
// Requirement admission is created for Modul Perhitungan. Any fields provided will overwrite the previous data except
// estimation documents which instead add to the existing estimation documents.
func (c *Client) HandleRequirementAdmissionEdit(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementAdmissionEdit)
	defer cancel()

	r := &models.RequirementAdmission{}
	err := c.decodeRequestJson(writer, request, r)
	if err != nil {
		return
	}

	r.SubmitterAsnId = user.AsnId
	r.AgencyId = user.WorkAgencyId

	err = c.EditRequirementAdmissionCtx(ctx, r)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]string{
		"kebutuhan_id": r.RequirementId,
	})
}

// HandleRequirementAdmissionCoverLetterUpload handles a request to upload a cover letter.
// This will return the generated cover letter filename and a signed URL that can be used to upload the file with
// PUT method. You do not need to submit the mime type as the only allowed type is application/pdf.
func (c *Client) HandleRequirementAdmissionCoverLetterUpload(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementAdmissionCoverLetterUpload)
	defer cancel()

	filename, _ := c.RequirementStorage.GenerateRequirementFilename("application/pdf")

	url, err := c.RequirementStorage.GenerateRequirementDocPutSign(ctx, path.Join(RequirementCoverLetterSubdir, filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
		"filename": filename,
		"url":      url.String(),
	}, false)
}

// HandleRequirementAdmissionEstimationDocUpload handles a request to upload an estimation document.
// This will return the generated cover letter filename and a signed URL that can be used to upload the file with
// PUT method. You have to submit the mime type of the file you want to upload.
func (c *Client) HandleRequirementAdmissionEstimationDocUpload(writer http.ResponseWriter, request *http.Request) {
	type uploadRequest struct {
		// MimeType is the mime type of the document you want to upload.
		MimeType string `json:"mime_type"`
	}

	ur := &uploadRequest{}
	err := c.decodeRequestJson(writer, request, ur)
	if err != nil {
		return
	}

	filename, err := c.RequirementStorage.GenerateRequirementFilename(ur.MimeType)
	if err != nil {
		if err == object.ErrFileTypeUnsupported {
			err = ec.NewError(ErrCodeMimeTypeNotSupported, Errs[ErrCodeMimeTypeNotSupported], err)
			c.httpError(writer, ec.NewError(ErrCodeMimeTypeNotSupported, Errs[ErrCodeMimeTypeNotSupported], err))
			return
		}
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementAdmissionEstimationDocUpload)
	defer cancel()

	url, err := c.RequirementStorage.GenerateRequirementDocPutSign(ctx, path.Join(RequirementEstimationDocSubdir, filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
		"filename": filename,
		"url":      url.String(),
	}, false)
}

// HandleRequirementAdmissionCoverLetterPreview handles a request to preview a cover letter in temporary location.
// This will redirect your request to the storage signed URL.
func (c *Client) HandleRequirementAdmissionCoverLetterPreview(writer http.ResponseWriter, request *http.Request) {
	type schemaFilename struct {
		Filename string `schema:"filename"`
	}

	s := &schemaFilename{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.Filename == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementAdmissionCoverLetterPreview)
	defer cancel()

	url, err := c.RequirementStorage.GenerateRequirementDocGetSignTemp(ctx, path.Join(RequirementCoverLetterSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleRequirementAdmissionEstimationDocPreview handles a request to preview an estimation doc in temporary location.
// This will redirect your request to the storage signed URL.
func (c *Client) HandleRequirementAdmissionEstimationDocPreview(writer http.ResponseWriter, request *http.Request) {
	type schemaFilename struct {
		Filename string `schema:"filename"`
	}

	s := &schemaFilename{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.Filename == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementAdmissionEstimationDocPreview)
	defer cancel()

	url, err := c.RequirementStorage.GenerateRequirementDocGetSignTemp(ctx, path.Join(RequirementEstimationDocSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleRequirementAdmissionCoverLetterDownload handles a request to download a cover letter in permanent location.
// This will redirect your request to the storage signed URL.
func (c *Client) HandleRequirementAdmissionCoverLetterDownload(writer http.ResponseWriter, request *http.Request) {
	type schemaRequirementId struct {
		RequirementId string `schema:"kebutuhan_id"`
	}

	s := &schemaRequirementId{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.RequirementId == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementAdmissionCoverLetterDownload)
	defer cancel()

	url, err := c.RequirementStorage.GenerateRequirementDocGetSign(
		ctx,
		path.Join(RequirementCoverLetterSubdir, fmt.Sprintf("%s.%s", s.RequirementId, "pdf")),
	)
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleRequirementAdmissionEstimationDocDownload handles a request to download an estimation doc in permanent location.
// This will redirect your request to the storage signed URL.
func (c *Client) HandleRequirementAdmissionEstimationDocDownload(writer http.ResponseWriter, request *http.Request) {
	type schemaFilename struct {
		Filename string `schema:"filename"`
	}

	s := &schemaFilename{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.Filename == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementAdmissionEstimationDocDownload)
	defer cancel()

	url, err := c.RequirementStorage.GenerateRequirementDocGetSign(ctx, path.Join(RequirementEstimationDocSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// Deprecated: use the paginated version.
// HandleRequirementAdmissionSearch handles a request to get admission list of a work agency.
// Agency ID will be retrieved from authentication token.
func (c *Client) HandleRequirementAdmissionSearch(writer http.ResponseWriter, request *http.Request) {
	type schemaAdmissionSearch struct {
		// AdmissionDate is a date with a format of YYYY-MM-DD (e.g. 2006-12-31)
		AdmissionDate   string `schema:"tgl_usulan"`
		AdmissionStatus int    `schema:"status"`
	}

	user := auth.AssertReqGetUserDetail(request)

	query := &schemaAdmissionSearch{}
	err := c.decodeRequestSchema(writer, request, query)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementAdmissionSearch)
	defer cancel()

	if _, ok := models.RequirementAdmissionStatuses[query.AdmissionStatus]; query.AdmissionStatus != 0 && !ok {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeRequirementFilterStatusInvalid, Errs[ErrCodeRequirementFilterStatusInvalid]))
		return
	}

	var admissionDate models.Iso8601Date
	if query.AdmissionDate != "" {
		admissionDate, err = models.ParseIso8601Date(query.AdmissionDate)
		if err != nil {
			c.httpError(writer, ec.NewError(ErrCodeRequirementFilterInvalidDate, Errs[ErrCodeRequirementFilterInvalidDate], err))
			return
		}
	}

	searchFilter := &RequirementAdmissionSearchFilter{
		AgencyId:        user.WorkAgencyId,
		AdmissionDate:   admissionDate,
		AdmissionStatus: query.AdmissionStatus,
	}

	admissions, err := c.SearchRequirementAdmissionsCtx(ctx, searchFilter)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, admissions)
}

// HandleRequirementAdmissionSearchPaginated handles a request to get admission list of a work agency.
// Agency ID will be retrieved from authentication token.
func (c *Client) HandleRequirementAdmissionSearchPaginated(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementAdmissionSearch)
	defer cancel()

	type schemaAdmissionSearch struct {
		// AdmissionDate is a date with a format of YYYY-MM-DD (e.g. 2006-12-31)
		AdmissionDate   string `schema:"tgl_usulan"`
		AdmissionStatus int    `schema:"status"`
		PageNumber      int    `schema:"halaman"`
		CountPerPage    int    `schema:"jumlah_per_halaman"`
	}

	user := auth.AssertReqGetUserDetail(request)

	query := &schemaAdmissionSearch{}
	err := c.decodeRequestSchema(writer, request, query)
	if err != nil {
		return
	}

	countPerPage := 10
	if query.CountPerPage != 0 {
		countPerPage = query.CountPerPage
	}

	pageNumber := 1
	if query.PageNumber != 0 {
		pageNumber = query.PageNumber
	}

	err = c.httpErrorVerifyListMeta(writer, pageNumber, countPerPage)
	if err != nil {
		return
	}

	if _, ok := models.RequirementAdmissionStatuses[query.AdmissionStatus]; query.AdmissionStatus != 0 && !ok {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeRequirementFilterStatusInvalid, Errs[ErrCodeRequirementFilterStatusInvalid]))
		return
	}

	var admissionDate models.Iso8601Date
	if query.AdmissionDate != "" {
		admissionDate, err = models.ParseIso8601Date(query.AdmissionDate)
		if err != nil {
			c.httpError(writer, ec.NewError(ErrCodeRequirementFilterInvalidDate, Errs[ErrCodeRequirementFilterInvalidDate], err))
			return
		}
	}

	searchFilter := &RequirementAdmissionSearchFilter{
		AgencyId:        user.WorkAgencyId,
		AdmissionDate:   admissionDate,
		AdmissionStatus: query.AdmissionStatus,
		CountPerPage:    countPerPage,
		PageNumber:      pageNumber,
	}

	admissions, err := c.SearchRequirementAdmissionsPaginatedCtx(ctx, searchFilter)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, admissions)
}

// HandleRequirementAdmissionDetailGet handles a request to get admission detail with matched id.
func (c *Client) HandleRequirementAdmissionDetailGet(writer http.ResponseWriter, request *http.Request) {
	type schemaAdmissionDetail struct {
		RequirementId string `schema:"kebutuhan_id"`
	}

	user := auth.AssertReqGetUserDetail(request)

	query := &schemaAdmissionDetail{}
	err := c.decodeRequestSchema(writer, request, query)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementAdmissionDetailGet)
	defer cancel()

	if _, err = uuid.Parse(query.RequirementId); err != nil {
		c.httpError(writer, ErrEntryNotFound)
		return
	}

	admission, err := c.GetRequirementAdmissionDetailCtx(ctx, query.RequirementId, user.WorkAgencyId)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, admission)
}

// HandleRequirementAdmissionCoverLetterTemplateDownload handles a request to download cover letter template.
func (c *Client) HandleRequirementAdmissionCoverLetterTemplateDownload(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementAdmissionCoverLetterTemplateDownload)
	defer cancel()

	url, err := c.RequirementStorage.GenerateRequirementTemplateCoverLetterGetSign(ctx)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// Deprecated: now generated and cannot be uploaded.
// HandleRequirementVerificationRecommendationLetterUpload handles a request to upload a recommendation letter.
// This will return the generated recommendation letter filename and a signed URL that can be used to upload the file with
// PUT method. You do not need to submit the mime type as the only allowed type is application/pdf.
func (c *Client) HandleRequirementVerificationRecommendationLetterUpload(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementVerificationRecommendationLetterUpload)
	defer cancel()

	filename, _ := c.RequirementStorage.GenerateRequirementFilename("application/pdf")
	url, err := c.RequirementStorage.GenerateRequirementDocPutSign(ctx, path.Join(RequirementRecommendationLetterSubdir, filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{"url": url.String(), "filename": filename}, false)
}

// Deprecated: now generated and cannot be uploaded.
// HandleRequirementVerificationRecommendationLetterPreview handles a request to preview a recommendation letter in temporary location.
// This will redirect your request to the storage signed URL.
func (c *Client) HandleRequirementVerificationRecommendationLetterPreview(writer http.ResponseWriter, request *http.Request) {
	type schemaFilename struct {
		Filename string `schema:"filename"`
	}

	s := &schemaFilename{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		return
	}

	if s.Filename == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementVerificationRecommendationLetterPreview)
	defer cancel()

	url, err := c.RequirementStorage.GenerateRequirementDocGetSignTemp(ctx, path.Join(RequirementRecommendationLetterSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// Deprecated: now generated and cannot be uploaded.
// HandleRequirementVerificationRecommendationLetterDownload handles a request to preview a recommendation letter in permanent location.
// This will redirect your request to the storage signed URL.
func (c *Client) HandleRequirementVerificationRecommendationLetterDownload(writer http.ResponseWriter, request *http.Request) {
	type schemaFilename struct {
		Filename string `schema:"filename"`
	}

	s := &schemaFilename{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		return
	}

	if s.Filename == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementVerificationRecommendationLetterDownloadUnsigned)
	defer cancel()

	url, err := c.RequirementStorage.GenerateRequirementDocGetSign(ctx, path.Join(RequirementRecommendationLetterSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleRequirementVerificationRecommendationLetterDownloadUnsigned handles a request to download unsigned recommendation letter in permanent location.
// This will redirect your request to the storage signed URL.
func (c *Client) HandleRequirementVerificationRecommendationLetterDownloadUnsigned(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementVerificationRecommendationLetterDownloadUnsigned)
	defer cancel()

	s := &schemaFilename{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		return
	}

	if s.Filename == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	url, err := c.RequirementStorage.GenerateRequirementDocGetSign(ctx, path.Join(RequirementRecommendationLetterSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	if s.NoRedirect {
		_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
			"url": url.String(),
		}, false)
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// Deprecated: signed letter can be downloaded from Teken Digital.
// HandleRequirementVerificationRecommendationLetterDownloadSigned handles a request to download signed recommendation letter in permanent location.
// This will redirect your request to the storage signed URL.
func (c *Client) HandleRequirementVerificationRecommendationLetterDownloadSigned(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementVerificationRecommendationLetterDownloadSigned)
	defer cancel()

	type schemaFilename struct {
		Filename string `schema:"filename"`
	}

	s := &schemaFilename{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		return
	}

	if s.Filename == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	url, err := c.RequirementStorage.GenerateRequirementDocGetSign(ctx, path.Join(RequirementRecommendationLetterSubdir, fmt.Sprintf("signed-%s", s.Filename)))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleRequirementVerificationSet handles a request to set the requirement admission status to accepted.
func (c *Client) HandleRequirementVerificationSet(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)

	rv := &models.RequirementVerificationRequest{}
	err := c.decodeRequestJson(writer, request, rv)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementAdmissionVerification)
	defer cancel()

	rv.SubmitterAsnId = user.AsnId
	rv.AgencyId = user.WorkAgencyId

	modifiedAt, err := c.SetRequirementStatusAcceptedCtx(ctx, rv)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]interface{}{
		"kebutuhan_id": rv.RequirementId,
		"modified_at":  modifiedAt.Unix(),
	})
}

// HandleRequirementDenySet handles a request to set the requirement admission status to denied.
func (c *Client) HandleRequirementDenySet(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)

	rd := &models.RequirementRevisionRequest{}
	err := c.decodeRequestJson(writer, request, rd)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementAdmissionVerification)
	defer cancel()

	rd.SubmitterAsnId = user.AsnId
	rd.AgencyId = user.WorkAgencyId
	rd.DenyTimestamp = models.EpochTime(time.Now())

	modifiedAt, err := c.SetRequirementStatusRevisionCtx(ctx, rd)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]interface{}{
		"kebutuhan_id": rd.RequirementId,
		"modified_at":  modifiedAt.Unix(),
	})
}

// HandleRequirementVerifiersGet handles a request to get verifier list of a work agency.
// Agency ID will be retrieved from authentication token.
func (c *Client) HandleRequirementVerifiersGet(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementVerifierGet)
	defer cancel()

	positions, err := c.GetRequirementVerifiersCtx(ctx, user.WorkAgencyId)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, positions)
}

// HandleRequirementVerificationBulkSubmitRecommendationLetter handles submitting a recommendation letter to update multiple
// requirement at once.
func (c *Client) HandleRequirementVerificationBulkSubmitRecommendationLetter(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementVerificationBulkSubmitRecommendationLetter)
	defer cancel()

	user := auth.AssertReqGetUserDetail(request)

	type submitReq struct {
		RequirementIds       []string         `json:"kebutuhan_id"`
		RecommendationLetter *models.Document `json:"surat_rekomendasi"`
	}
	s := &submitReq{}
	err := c.decodeRequestJson(writer, request, s)
	if err != nil {
		return
	}

	if s.RequirementIds == nil || len(s.RequirementIds) == 0 {
		c.httpError(writer, ErrEntryNotFound)
		return
	}

	filename, err := c.BulkSubmitRequirementRecommendationLetterCtx(ctx, s.RequirementIds, s.RecommendationLetter, user.AsnId)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]interface{}{
		"kebutuhan_id": s.RequirementIds,
		"filename":     filename,
	})
}

// HandleRequirementVerificationSignRecommendationLetter handles a request to async sign the recommendation letter.
func (c *Client) HandleRequirementVerificationSignRecommendationLetter(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutRequirementVerificationSignRecommendationLetter)
	defer cancel()

	type schemaFilename struct {
		Filename string `json:"filename"`
	}
	s := &schemaFilename{}
	err := c.decodeRequestJson(writer, request, s)
	if err != nil {
		return
	}

	if s.Filename == "" {
		c.httpError(writer, ErrEntryNotFound)
		return
	}

	err = c.SignRequirementRecommendationLetterCtx(ctx, s.Filename)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, s)
}

// HandleGetRequirementStatusStatistic returns the number of requirement items for each status.
func (c *Client) HandleGetRequirementStatusStatistic(writer http.ResponseWriter, _ *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutGetRequirementStatusStatistic)
	defer cancel()

	statistic, err := c.GetRequirementStatusStatisticCtx(ctx)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, statistic)
}
