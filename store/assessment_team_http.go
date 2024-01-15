package store

import (
	"context"
	"fmt"
	"net/http"
	"path"

	. "github.com/if-itb/siasn-jf-backend/errnum"
	"github.com/if-itb/siasn-jf-backend/store/models"
	"github.com/if-itb/siasn-jf-backend/store/object"
	"github.com/if-itb/siasn-libs-backend/auth"
	"github.com/if-itb/siasn-libs-backend/ec"
	"github.com/if-itb/siasn-libs-backend/httputil"
)

const (
	TimeoutAssessmentTeamAdmissionSupportDocUpload             = TimeoutDefault
	TimeoutAssessmentTeamAdmissionSupportDocPreview            = TimeoutDefault
	TimeoutAssessmentTeamAdmissionSupportDocDownload           = TimeoutDefault
	TimeoutAssessmentTeamAdmissionRecommendationLetterUpload   = TimeoutDefault
	TimeoutAssessmentTeamAdmissionRecommendationLetterPreview  = TimeoutDefault
	TimeoutAssessmentTeamAdmissionRecommendationLetterDownload = TimeoutDefault
	TimeoutAssessmentTeamAdmissionSubmit                       = TimeoutDefault
	TimeoutAssessmentTeamVerificationSubmit                    = TimeoutDefault
	TimeoutAssessmentTeamGet                                   = TimeoutDefault
	TimeoutAssessmentTeamSearch                                = TimeoutDefault
)

// HandleAssessmentTeamAdmissionSubmit handles a new assessment team admission request. Assessment team admission is
// created for Modul Tim Penilaian. A new assessment team admission entry will be created in the database with
// status set to `created`.
func (c *Client) HandleAssessmentTeamAdmissionSubmit(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutAssessmentTeamAdmissionSubmit)
	defer cancel()

	user := auth.AssertReqGetUserDetail(request)

	admission := &models.AssessmentTeamAdmission{}
	err := c.decodeRequestJson(writer, request, admission)
	if err != nil {
		return
	}

	admission.SubmitterAsnId = user.AsnId
	admission.AgencyId = user.WorkAgencyId

	admissionId, err := c.InsertAssessmentTeamAdmissionCtx(ctx, admission)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]string{
		"tim_penilaian_id": admissionId,
	})
}

// HandleAssessmentTeamAdmissionSupportDocUpload handles a request to upload an admission support document.
// We do not redirect the request to object storage signed URL. We instead return a JSON containing a document name
// which will have to be saved by the frontend and a signed URL which the frontend has to request with PUT method
// together with the file.
func (c *Client) HandleAssessmentTeamAdmissionSupportDocUpload(writer http.ResponseWriter, request *http.Request) {
	type uploadRequest struct {
		// MimeType is the mime type of the document you want to upload.
		MimeType string `json:"mime_type"`
	}

	ur := &uploadRequest{}
	err := c.decodeRequestJson(writer, request, ur)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	filename, err := c.AssessmentTeamStorage.GenerateAssessmentTeamFilename(ur.MimeType)
	if err != nil {
		if err == object.ErrFileTypeUnsupported {
			err = ec.NewError(ErrCodeMimeTypeNotSupported, Errs[ErrCodeMimeTypeNotSupported], err)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutAssessmentTeamAdmissionSupportDocUpload)
	defer cancel()

	url, err := c.AssessmentTeamStorage.GenerateAssessmentTeamDocPutSign(ctx, path.Join(AssessmentTeamSupportDocSubdir, filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200(writer, map[string]string{
		"filename": filename,
		"url":      url.String(),
	})
}

// HandleAssessmentTeamAdmissionSupportDocPreview handles a request to download a support document currently in temporary location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandleAssessmentTeamAdmissionSupportDocPreview(writer http.ResponseWriter, request *http.Request) {
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

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutAssessmentTeamAdmissionSupportDocPreview)
	defer cancel()

	url, err := c.AssessmentTeamStorage.GenerateAssessmentTeamDocGetSignTemp(ctx, path.Join(AssessmentTeamSupportDocSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleAssessmentTeamAdmissionSupportDocDownload handles a request to download a support document from permanent location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandleAssessmentTeamAdmissionSupportDocDownload(writer http.ResponseWriter, request *http.Request) {
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

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutAssessmentTeamAdmissionSupportDocDownload)
	defer cancel()

	url, err := c.AssessmentTeamStorage.GenerateAssessmentTeamDocGetSign(ctx, path.Join(AssessmentTeamSupportDocSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleAssessmentTeamGet handles retrieving assessment team detail.
func (c *Client) HandleAssessmentTeamGet(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutAssessmentTeamGet)
	defer cancel()

	s := &struct {
		AssessmentTeamId string `schema:"tim_penilaian_id"`
	}{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		return
	}

	admission, err := c.GetAssessmentTeamCtx(ctx, s.AssessmentTeamId)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, admission)
}

// HandleAssessmentTeamSearch handles a request to get assessment team list.
func (c *Client) HandleAssessmentTeamSearch(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutAssessmentTeamSearch)
	defer cancel()

	type schemaAdmissionSearch struct {
		// AdmissionDate is a date with a format of YYYY-MM-DD (e.g. 2006-12-31)
		AdmissionDate   string `schema:"tgl_usulan"`
		AdmissionStatus int    `schema:"status"`
		PageNumber      int    `schema:"halaman"`
		CountPerPage    int    `schema:"jumlah_per_halaman"`
	}

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

	if _, ok := models.AssessmentTeamStatuses[query.AdmissionStatus]; query.AdmissionStatus != 0 && !ok {
		c.httpError(writer, ErrAssessmentTeamAdmissionFilterInvalidStatus)
		return
	}

	var admissionDate models.Iso8601Date
	if query.AdmissionDate != "" {
		admissionDate, err = models.ParseIso8601Date(query.AdmissionDate)
		if err != nil {
			c.httpError(writer, ErrAssessmentTeamAdmissionFilterInvalidDate)
			return
		}
	}

	searchFilter := &AssessmentTeamAdmissionSearchFilter{
		AdmissionDate:   admissionDate,
		AdmissionStatus: query.AdmissionStatus,
		PageNumber:      pageNumber,
		CountPerPage:    countPerPage,
	}

	admissions, err := c.SearchAssessmentTeamsCtx(ctx, searchFilter)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, (*models.IdPaginatedList)(admissions))
}

// HandleAssessmentTeamVerificationRecommendationLetterUpload handles a request to upload a verification recommendation letter.
// We do not redirect the request to object storage signed URL. We instead return a JSON containing a document name
// which will have to be saved by the frontend and a signed URL which the frontend has to request with PUT method
// together with the file. You do not need to submit the mime type as the only allowed type is application/pdf.
func (c *Client) HandleAssessmentTeamVerificationRecommendationLetterUpload(writer http.ResponseWriter, request *http.Request) {
	type schemaAssessmentTeamId struct {
		AssessmentTeamId string `json:"tim_penilaian_id"`
	}

	s := &schemaAssessmentTeamId{}
	err := c.decodeRequestJson(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutAssessmentTeamAdmissionRecommendationLetterUpload)
	defer cancel()

	url, err := c.AssessmentTeamStorage.GenerateAssessmentTeamDocPutSign(ctx, path.Join(AssessmentTeamRecommendationLetterSubdir, fmt.Sprintf("%s.pdf", s.AssessmentTeamId)))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
		"url": url.String(),
	}, false)
}

// HandleAssessmentTeamVerificationRecommendationLetterPreview handles a request to download a recommendation letter currently in temporary location.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandleAssessmentTeamVerificationRecommendationLetterPreview(writer http.ResponseWriter, request *http.Request) {
	type schemaAssessmentTeamId struct {
		AssessmentTeamId string `schema:"tim_penilaian_id"`
	}

	s := &schemaAssessmentTeamId{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.AssessmentTeamId == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutAssessmentTeamAdmissionRecommendationLetterPreview)
	defer cancel()

	url, err := c.AssessmentTeamStorage.GenerateAssessmentTeamDocGetSignTemp(ctx, path.Join(AssessmentTeamRecommendationLetterSubdir, fmt.Sprintf("%s.pdf", s.AssessmentTeamId)))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleAssessmentTeamVerificationRecommendationLetterDownload handles a request to download a recommendation document from permanent location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandleAssessmentTeamVerificationRecommendationLetterDownload(writer http.ResponseWriter, request *http.Request) {
	type schemaAssessmentTeamId struct {
		AssessmentTeamId string `schema:"tim_penilaian_id"`
	}

	s := &schemaAssessmentTeamId{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.AssessmentTeamId == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutAssessmentTeamAdmissionRecommendationLetterDownload)
	defer cancel()

	url, err := c.AssessmentTeamStorage.GenerateAssessmentTeamDocGetSign(ctx, path.Join(AssessmentTeamRecommendationLetterSubdir, fmt.Sprintf("%s.pdf", s.AssessmentTeamId)))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleAssessmentTeamVerificationSubmit handles a new assessment team verification request. Assessment team status
// in the database will be set to `verified`.
func (c *Client) HandleAssessmentTeamVerificationSubmit(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutAssessmentTeamVerificationSubmit)
	defer cancel()

	user := auth.AssertReqGetUserDetail(request)

	verification := &models.AssessmentTeamVerification{}
	err := c.decodeRequestJson(writer, request, verification)
	if err != nil {
		return
	}

	verification.SubmitterAsnId = user.AsnId

	updatedAt, err := c.SetAssessmentTeamVerificationCtx(ctx, verification)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]interface{}{
		"tim_penilaian_id": verification.AssessmentTeamId,
		"updated_at":       updatedAt.Unix(),
	})
}
