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
	TimeoutPromotionAdmissionSubmit                               = TimeoutDefault
	TimeoutPromotionAdmissionPakLetterUpload                      = TimeoutDefault
	TimeoutPromotionAdmissionPakLetterPreview                     = TimeoutDefault
	TimeoutPromotionPakLetterDownload                             = TimeoutDefault
	TimeoutPromotionAdmissionPakLetterTemplateDownload            = TimeoutDefault
	TimeoutPromotionAdmissionRecommendationLetterUpload           = TimeoutDefault
	TimeoutPromotionAdmissionRecommendationLetterPreview          = TimeoutDefault
	TimeoutPromotionRecommendationLetterDownload                  = TimeoutDefault
	TimeoutPromotionAdmissionRecommendationLetterTemplateDownload = TimeoutDefault
	TimeoutPromotionAdmissionPromotionLetterUpload                = TimeoutDefault
	TimeoutPromotionAdmissionPromotionLetterPreview               = TimeoutDefault
	TimeoutPromotionAdmissionPromotionLetterDownload              = TimeoutDefault
	TimeoutPromotionAdmissionTestCertificateUpload                = TimeoutDefault
	TimeoutPromotionAdmissionTestCertificatePreview               = TimeoutDefault
	TimeoutPromotionAdmissionTestCertificateDownload              = TimeoutDefault
	TimeoutPromotionAdmissionAccept                               = TimeoutDefault
	TimeoutPromotionAdmissionReject                               = TimeoutDefault
	TimeoutPromotionAdmissionSearch                               = TimeoutDefault
	TimeoutGetPromotionStatusStatistic                            = TimeoutDefault
)

// HandlePromotionAdmissionSubmit handles a new admission request.
// Dismissal admission is created for Modul Pengangkatan. A new promotion request entry
// will be created in the database with status set to `newly admitted`.
func (c *Client) HandlePromotionAdmissionSubmit(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionAdmissionSubmit)
	defer cancel()

	user := auth.AssertReqGetUserDetail(request)

	pa := &models.PromotionAdmission{}
	err := c.decodeRequestJson(writer, request, pa)
	if err != nil {
		return
	}

	pa.SubmitterAsnId = user.AsnId
	pa.AgencyId = user.WorkAgencyId

	promotionId, err := c.InsertPromotionAdmissionCtx(ctx, pa)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]string{
		"pengangkatan_id": promotionId,
	})
}

// HandlePromotionAdmissionPakLetterUpload handles a request to upload an admission PAK letter.
// We do not redirect the request to object storage signed URL. We instead return a JSON containing a document name
// which will have to be saved by the frontend and a signed URL which the frontend has to request with PUT method
// together with the file.
func (c *Client) HandlePromotionAdmissionPakLetterUpload(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionAdmissionPakLetterUpload)
	defer cancel()

	defer request.Body.Close()

	filename, err := c.PromotionStorage.GeneratePromotionDocName("application/pdf")
	if err != nil {
		if err == object.ErrFileTypeUnsupported {
			err = ec.NewError(ErrCodeMimeTypeNotSupported, Errs[ErrCodeMimeTypeNotSupported], err)
		}
		return
	}

	url, err := c.PromotionStorage.GeneratePromotionDocPutSign(ctx, path.Join(PromotionPakLetterSubdir, filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
		"filename": filename,
		"url":      url.String(),
	}, false)
}

// HandlePromotionAdmissionPakLetterPreview handles a request to download a PAK letter currently in temporary location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandlePromotionAdmissionPakLetterPreview(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionAdmissionPakLetterPreview)
	defer cancel()

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

	url, err := c.PromotionStorage.GeneratePromotionDocGetSignTemp(ctx, path.Join(PromotionPakLetterSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandlePromotionPakLetterDownload handles a request to download a PAK letter document from permanent location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandlePromotionPakLetterDownload(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionPakLetterDownload)
	defer cancel()

	type schemaPromotionId struct {
		PromotionId string `schema:"pengangkatan_id"`
	}

	s := &schemaPromotionId{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.PromotionId == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	found, err := c.IsPromotionExistCtx(ctx, s.PromotionId)
	if err != nil {
		c.httpError(writer, err)
		return
	}
	if !found {
		c.httpError(writer, ErrEntryNotFound)
		return
	}

	url, err := c.PromotionStorage.GeneratePromotionDocGetSign(ctx, path.Join(PromotionPakLetterSubdir, fmt.Sprintf("%s.pdf", s.PromotionId)))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandlePromotionAdmissionPakLetterTemplateDownload handles a request to download admission PAK letter template.
func (c *Client) HandlePromotionAdmissionPakLetterTemplateDownload(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionAdmissionPakLetterTemplateDownload)
	defer cancel()

	url, err := c.PromotionStorage.GeneratePromotionTemplatePakLetterGetSign(ctx)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandlePromotionAdmissionRecommendationLetterUpload handles a request to upload an admission Recommendation letter.
// We do not redirect the request to object storage signed URL. We instead return a JSON containing a document name
// which will have to be saved by the frontend and a signed URL which the frontend has to request with PUT method
// together with the file.
func (c *Client) HandlePromotionAdmissionRecommendationLetterUpload(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionAdmissionRecommendationLetterUpload)
	defer cancel()

	defer request.Body.Close()

	filename, err := c.PromotionStorage.GeneratePromotionDocName("application/pdf")
	if err != nil {
		if err == object.ErrFileTypeUnsupported {
			err = ec.NewError(ErrCodeMimeTypeNotSupported, Errs[ErrCodeMimeTypeNotSupported], err)
		}
		return
	}

	url, err := c.PromotionStorage.GeneratePromotionDocPutSign(ctx, path.Join(PromotionRecommendationLetterSubdir, filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
		"filename": filename,
		"url":      url.String(),
	}, false)
}

// HandlePromotionAdmissionRecommendationLetterPreview handles a request to download a Recommendation letter currently in temporary location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandlePromotionAdmissionRecommendationLetterPreview(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionAdmissionRecommendationLetterPreview)
	defer cancel()

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

	url, err := c.PromotionStorage.GeneratePromotionDocGetSignTemp(ctx, path.Join(PromotionRecommendationLetterSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandlePromotionRecommendationLetterDownload handles a request to download a Recommendation letter document from permanent location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandlePromotionRecommendationLetterDownload(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionRecommendationLetterDownload)
	defer cancel()

	type schemaPromotionId struct {
		PromotionId string `schema:"pengangkatan_id"`
	}

	s := &schemaPromotionId{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.PromotionId == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	found, err := c.IsPromotionExistCtx(ctx, s.PromotionId)
	if err != nil {
		c.httpError(writer, err)
		return
	}
	if !found {
		c.httpError(writer, ErrEntryNotFound)
		return
	}

	url, err := c.PromotionStorage.GeneratePromotionDocGetSign(ctx, path.Join(PromotionRecommendationLetterSubdir, fmt.Sprintf("%s.pdf", s.PromotionId)))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandlePromotionAdmissionRecommendationLetterTemplateDownload handles a request to download admission Recommendation letter template.
func (c *Client) HandlePromotionAdmissionRecommendationLetterTemplateDownload(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionAdmissionRecommendationLetterTemplateDownload)
	defer cancel()

	url, err := c.PromotionStorage.GeneratePromotionTemplateRecommendationLetterGetSign(ctx)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// Deprecated: it is now generated.
// HandlePromotionAdmissionPromotionLetterUpload handles a request to upload a promotion's PromotionAdmission letter.
// We do not redirect the request to object storage signed URL. We instead return a JSON containing a document name
// which will have to be saved by the frontend and a signed URL which the frontend has to request with PUT method
// together with the file.
func (c *Client) HandlePromotionAdmissionPromotionLetterUpload(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionAdmissionPromotionLetterUpload)
	defer cancel()

	defer request.Body.Close()

	filename, err := c.PromotionStorage.GeneratePromotionDocName("application/pdf")
	if err != nil {
		if err == object.ErrFileTypeUnsupported {
			err = ec.NewError(ErrCodeMimeTypeNotSupported, Errs[ErrCodeMimeTypeNotSupported], err)
		}
		return
	}

	url, err := c.PromotionStorage.GeneratePromotionDocPutSign(ctx, path.Join(PromotionPromotionLetterSubdir, filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
		"filename": filename,
		"url":      url.String(),
	}, false)
}

// Deprecated: it is now generated.
// HandlePromotionAdmissionPromotionLetterPreview handles a request to download a PromotionAdmission letter currently in temporary location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandlePromotionAdmissionPromotionLetterPreview(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionAdmissionPromotionLetterPreview)
	defer cancel()

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

	url, err := c.PromotionStorage.GeneratePromotionDocGetSignTemp(ctx, path.Join(PromotionPromotionLetterSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandlePromotionAdmissionPromotionLetterDownload handles a request to download a PromotionAdmission letter document from permanent location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandlePromotionAdmissionPromotionLetterDownload(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionAdmissionPromotionLetterDownload)
	defer cancel()

	type schemaPromotionId struct {
		PromotionId     string `schema:"pengangkatan_id"`
		ForceRegenerate bool   `schema:"force"`
	}

	s := &schemaPromotionId{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	filename, err := c.GeneratePromotionLetterCtx(ctx, s.PromotionId, s.ForceRegenerate)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	url, err := c.PromotionStorage.GeneratePromotionDocGetSign(ctx, path.Join(PromotionPromotionLetterSubdir, filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandlePromotionAdmissionTestCertificateUpload handles a request to upload a promotion's test certificate.
// We do not redirect the request to object storage signed URL. We instead return a JSON containing a document name
// which will have to be saved by the frontend and a signed URL which the frontend has to request with PUT method
// together with the file.
func (c *Client) HandlePromotionAdmissionTestCertificateUpload(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionAdmissionTestCertificateUpload)
	defer cancel()

	defer request.Body.Close()

	filename, err := c.PromotionStorage.GeneratePromotionDocName("application/pdf")
	if err != nil {
		if err == object.ErrFileTypeUnsupported {
			err = ec.NewError(ErrCodeMimeTypeNotSupported, Errs[ErrCodeMimeTypeNotSupported], err)
		}
		return
	}

	url, err := c.PromotionStorage.GeneratePromotionDocPutSign(ctx, path.Join(PromotionTestCertificateSubdir, filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
		"filename": filename,
		"url":      url.String(),
	}, false)
}

// HandlePromotionAdmissionTestCertificatePreview handles a request to download a test certificate currently in temporary location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandlePromotionAdmissionTestCertificatePreview(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionAdmissionTestCertificatePreview)
	defer cancel()

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

	url, err := c.PromotionStorage.GeneratePromotionDocGetSignTemp(ctx, path.Join(PromotionTestCertificateSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandlePromotionAdmissionTestCertificateDownload handles a request to download a test certificate document from permanent location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandlePromotionAdmissionTestCertificateDownload(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionAdmissionTestCertificateDownload)
	defer cancel()

	type schemaPromotionId struct {
		PromotionId string `schema:"pengangkatan_id"`
	}

	s := &schemaPromotionId{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.PromotionId == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	found, err := c.IsPromotionExistCtx(ctx, s.PromotionId)
	if err != nil {
		c.httpError(writer, err)
		return
	}
	if !found {
		c.httpError(writer, ErrEntryNotFound)
		return
	}

	url, err := c.PromotionStorage.GeneratePromotionDocGetSign(ctx, path.Join(PromotionTestCertificateSubdir, fmt.Sprintf("%s.pdf", s.PromotionId)))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandlePromotionAdmissionAccept handles a promotion accept request.
func (c *Client) HandlePromotionAdmissionAccept(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionAdmissionAccept)
	defer cancel()

	user := auth.AssertReqGetUserDetail(request)

	p := &models.PromotionAdmission{}
	err := c.decodeRequestJson(writer, request, p)
	if err != nil {
		return
	}

	p.SubmitterAsnId = user.AsnId

	modifiedAt, err := c.SetPromotionStatusAcceptedCtx(ctx, p)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]interface{}{
		"pengangkatan_id": p.PromotionId,
		"modified_at":     modifiedAt.Unix(),
	})
}

// HandlePromotionAdmissionReject handles a promotion reject request.
func (c *Client) HandlePromotionAdmissionReject(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionAdmissionReject)
	defer cancel()

	user := auth.AssertReqGetUserDetail(request)

	p := &models.PromotionReject{}
	err := c.decodeRequestJson(writer, request, p)
	if err != nil {
		return
	}

	p.SubmitterAsnId = user.AsnId

	modifiedAt, err := c.SetPromotionStatusRejectCtx(ctx, p)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]interface{}{
		"pengangkatan_id": p.PromotionId,
		"modified_at":     modifiedAt.Unix(),
	})
}

// HandlePromotionAdmissionSearchPaginated handles a request to get promotion admission list.
func (c *Client) HandlePromotionAdmissionSearchPaginated(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionAdmissionSearch)
	defer cancel()

	type schemaAdmissionSearch struct {
		// AdmissionDate is a date with a format of YYYY-MM-DD (e.g. 2006-12-31)
		AdmissionDate   string `schema:"tgl_doc_surat_rekomendasi"`
		AdmissionStatus int    `schema:"status"`
		AdmissionType   int    `schema:"jenis_pengangkatan"`
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

	if _, ok := models.PromotionAdmissionStatuses[query.AdmissionStatus]; query.AdmissionStatus != 0 && !ok {
		c.httpError(writer, ErrPromotionFilterInvalidStatus)
		return
	}

	if _, ok := models.PromotionTypes[query.AdmissionType]; query.AdmissionType != 0 && !ok {
		c.httpError(writer, ErrPromotionFilterInvalidType)
		return
	}

	var admissionDate models.Iso8601Date
	if query.AdmissionDate != "" {
		admissionDate, err = models.ParseIso8601Date(query.AdmissionDate)
		if err != nil {
			c.httpError(writer, ErrPromotionFilterInvalidDate)
			return
		}
	}

	searchFilter := &PromotionAdmissionSearchFilter{
		AdmissionDate:   admissionDate,
		AdmissionStatus: query.AdmissionStatus,
		AdmissionType:   query.AdmissionType,
		CountPerPage:    countPerPage,
		PageNumber:      pageNumber,
	}

	admissions, err := c.SearchPromotionAdmissionsPaginatedCtx(ctx, searchFilter)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, (*models.IdPaginatedList)(admissions))
}

// HandleGetPromotionStatusStatistic returns the number of promotion items for each status.
func (c *Client) HandleGetPromotionStatusStatistic(writer http.ResponseWriter, _ *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutGetPromotionStatusStatistic)
	defer cancel()

	statistic, err := c.GetPromotionStatusStatisticCtx(ctx)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, statistic)
}
