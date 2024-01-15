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
	TimeoutDismissalAdmissionSubmit                  = TimeoutDefault
	TimeoutDismissalAdmissionGet                     = TimeoutDefault
	TimeoutDismissalAdmissionsSearch                 = TimeoutDefault
	TimeoutDismissalAdmissionSupportDocUpload        = TimeoutDefault
	TimeoutDismissalAdmissionSupportDocPreview       = TimeoutDefault
	TimeoutDismissalAdmissionSupportDocDownload      = TimeoutDefault
	TimeoutDismissalAcceptSet                        = TimeoutDefault
	TimeoutDismissalAcceptanceLetterUpload           = TimeoutDefault
	TimeoutDismissalAcceptanceLetterPreview          = TimeoutDefault
	TimeoutDismissalAcceptanceLetterDownload         = TimeoutDefault
	TimeoutDismissalAcceptanceLetterTemplateDownload = TimeoutDefault
	TimeoutDismissalDenySet                          = TimeoutDefault
	TimeoutDismissalDenySupportDocUpload             = TimeoutDefault
	TimeoutDismissalDenySupportDocPreview            = TimeoutDefault
	TimeoutDismissalDenySupportDocDownload           = TimeoutDefault
	TimeoutGetDismissalStatusStatistic               = TimeoutDefault
)

// HandleDismissalAdmissionSubmit handles a new admission request.
// Dismissal admission is created for Modul Pemberhentian. A new agency dismissal request entry
// will be created in the database with status set to `newly admitted`.
func (c *Client) HandleDismissalAdmissionSubmit(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)

	da := &models.DismissalAdmission{}
	err := c.decodeRequestJson(writer, request, da)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDismissalAdmissionSubmit)
	defer cancel()

	da.SubmitterAsnId = user.AsnId
	da.AgencyId = user.WorkAgencyId

	dismissalId, err := c.InsertDismissalAdmissionCtx(ctx, da)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]string{
		"pemberhentian_id": dismissalId,
	})
}

// HandleDismissalAdmissionGet handles getting a dismissal admission detail.
func (c *Client) HandleDismissalAdmissionGet(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)

	type schemaDismissalId struct {
		DismissalId string `schema:"pemberhentian_id"`
	}
	di := &schemaDismissalId{}
	err := c.decodeRequestSchema(writer, request, di)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDismissalAdmissionGet)
	defer cancel()

	dismissal, err := c.GetDismissalDetailCtx(ctx, di.DismissalId, user.WorkAgencyId)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, dismissal)
}

// Deprecated: use the paginated version.
// HandleDismissalAdmissionsSearch handles searching for dismissal admissions.
// Only a subset of dismissal data fields are returned.
func (c *Client) HandleDismissalAdmissionsSearch(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDismissalAdmissionsSearch)
	defer cancel()

	type schemaAdmissionSearch struct {
		// AdmissionDate is a date with a format of YYYY-MM-DD (e.g. 2006-12-31)
		AdmissionDate string `schema:"tgl_usulan"`
		Status        int    `schema:"status"`
	}
	query := &schemaAdmissionSearch{}
	err := c.decodeRequestSchema(writer, request, query)
	if err != nil {
		return
	}

	if _, ok := models.DismissalAdmissionStatuses[query.Status]; query.Status != 0 && !ok {
		c.httpError(writer, ErrDismissalSearchStatusInvalid)
		return
	}

	var admissionDate models.Iso8601Date
	if query.AdmissionDate != "" {
		admissionDate, err = models.ParseIso8601Date(query.AdmissionDate)
		if err != nil {
			c.httpError(writer, ErrDismissalSearchInvalidDate)
			return
		}
	}

	dismissals, err := c.SearchDismissalAdmissionsCtx(ctx, admissionDate, query.Status, user.WorkAgencyId)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, dismissals)
}

// HandleDismissalAdmissionsSearchPaginated handles searching for dismissal admissions.
// Only a subset of dismissal data fields are returned.
func (c *Client) HandleDismissalAdmissionsSearchPaginated(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDismissalAdmissionsSearch)
	defer cancel()

	type schemaAdmissionSearch struct {
		// AdmissionDate is a date with a format of YYYY-MM-DD (e.g. 2006-12-31)
		AdmissionDate string `schema:"tgl_usulan"`
		Status        int    `schema:"status"`
		PageNumber    int    `schema:"halaman"`
		CountPerPage  int    `schema:"jumlah_per_halaman"`
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

	if _, ok := models.DismissalAdmissionStatuses[query.Status]; query.Status != 0 && !ok {
		c.httpError(writer, ErrDismissalSearchStatusInvalid)
		return
	}

	var admissionDate models.Iso8601Date
	if query.AdmissionDate != "" {
		admissionDate, err = models.ParseIso8601Date(query.AdmissionDate)
		if err != nil {
			c.httpError(writer, ErrDismissalSearchInvalidDate)
			return
		}
	}

	dismissals, err := c.SearchDismissalAdmissionsPaginatedCtx(ctx, admissionDate, query.Status, user.WorkAgencyId, pageNumber, countPerPage)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, dismissals)
}

// HandleDismissalAdmissionSupportDocUpload handles a request to upload an admission support document.
// We do not redirect the request to object storage signed URL. We instead return a JSON containing a document name
// which will have to be saved by the frontend and a signed URL which the frontend has to request with PUT method
// together with the file.
func (c *Client) HandleDismissalAdmissionSupportDocUpload(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	filename, err := c.DismissalStorage.GenerateDismissalDocName("application/pdf")
	if err != nil {
		if err == object.ErrFileTypeUnsupported {
			err = ec.NewError(ErrCodeMimeTypeNotSupported, Errs[ErrCodeMimeTypeNotSupported], err)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDismissalAdmissionSupportDocUpload)
	defer cancel()

	url, err := c.DismissalStorage.GenerateDismissalDocPutSign(ctx, path.Join(DismissalSupportDocSubdir, filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
		"filename": filename,
		"url":      url.String(),
	}, false)
}

// HandleDismissalAdmissionSupportDocPreview handles a request to download a support document currently in temporary location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandleDismissalAdmissionSupportDocPreview(writer http.ResponseWriter, request *http.Request) {
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

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDismissalAdmissionSupportDocPreview)
	defer cancel()

	url, err := c.DismissalStorage.GenerateDismissalDocGetSignTemp(ctx, path.Join(DismissalSupportDocSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleDismissalAdmissionSupportDocDownload handles a request to download a support document from permanent location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandleDismissalAdmissionSupportDocDownload(writer http.ResponseWriter, request *http.Request) {
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

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDismissalAdmissionSupportDocDownload)
	defer cancel()

	url, err := c.DismissalStorage.GenerateDismissalDocGetSign(ctx, path.Join(DismissalSupportDocSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleDismissalAcceptSet handles a request to set the dismissal admission status to accepted.
func (c *Client) HandleDismissalAcceptSet(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)

	da := &models.DismissalAcceptanceRequest{}
	err := c.decodeRequestJson(writer, request, da)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDismissalAcceptSet)
	defer cancel()

	da.SubmitterAsnId = user.AsnId
	da.AgencyId = user.WorkAgencyId

	modifiedAt, err := c.SetDismissalStatusAcceptedCtx(ctx, da)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]interface{}{
		"pemberhentian_id": da.DismissalId,
		"modified_at":      modifiedAt.Unix(),
	})
}

// Deprecated: no longer uploaded.
// HandleDismissalAcceptanceLetterUpload handles a request to upload an dismissal acceptance letter.
// We do not redirect the request to object storage signed URL. We instead return a JSON containing a document name
// which will have to be saved by the frontend and a signed URL which the frontend has to request with PUT method
// together with the file.
func (c *Client) HandleDismissalAcceptanceLetterUpload(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	filename, err := c.DismissalStorage.GenerateDismissalDocName("application/pdf")
	if err != nil {
		if err == object.ErrFileTypeUnsupported {
			err = ec.NewError(ErrCodeMimeTypeNotSupported, Errs[ErrCodeMimeTypeNotSupported], err)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDismissalAcceptanceLetterUpload)
	defer cancel()

	url, err := c.DismissalStorage.GenerateDismissalDocPutSign(ctx, path.Join(DismissalAcceptanceLetterSubdir, filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
		"filename": filename,
		"url":      url.String(),
	}, false)
}

// Deprecated: no longer uploaded.
// HandleDismissalAcceptanceLetterPreview handles a request to download a support document currently in temporary location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandleDismissalAcceptanceLetterPreview(writer http.ResponseWriter, request *http.Request) {
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

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDismissalAcceptanceLetterPreview)
	defer cancel()

	url, err := c.DismissalStorage.GenerateDismissalDocGetSignTemp(ctx, path.Join(DismissalAcceptanceLetterSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleDismissalAcceptanceLetterDownload handles a request to download a support document from permanent location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandleDismissalAcceptanceLetterDownload(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDismissalAcceptanceLetterDownload)
	defer cancel()

	type schemaDismissalId struct {
		DismissalId string `schema:"pemberhentian_id"`
		NoRedirect  bool   `schema:"no_redirect"`
	}

	s := &schemaDismissalId{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.DismissalId == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	found, err := c.IsDismissalExistCtx(ctx, s.DismissalId)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if !found {
		c.httpError(writer, ErrEntryNotFound)
		return
	}

	url, err := c.DismissalStorage.GenerateDismissalDocGetSign(ctx, path.Join(DismissalAcceptanceLetterSubdir, fmt.Sprintf("%s.pdf", s.DismissalId)))
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

// HandleDismissalAcceptanceLetterTemplateDownload handles a request to download dismissal acceptance letter template.
func (c *Client) HandleDismissalAcceptanceLetterTemplateDownload(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDismissalAcceptanceLetterTemplateDownload)
	defer cancel()

	url, err := c.DismissalStorage.GenerateDismissalTemplateAcceptanceLetterGetSign(ctx)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleDismissalDenySet handles a request to deny dismissal request.
// Support documents can be uploaded first, but this step is optional.
func (c *Client) HandleDismissalDenySet(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)

	dd := &models.DismissalDenyRequest{}
	err := c.decodeRequestJson(writer, request, dd)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDismissalDenySet)
	defer cancel()

	dd.SubmitterAsnId = user.AsnId
	dd.AgencyId = user.WorkAgencyId

	modifiedAt, err := c.SetDismissalStatusDeniedCtx(ctx, dd)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]interface{}{
		"pemberhentian_id": dd.DismissalId,
		"modified_at":      modifiedAt.Unix(),
	})
}

// HandleDismissalDenySupportDocUpload handles a request to upload an deny support document.
// We do not redirect the request to object storage signed URL. We instead return a JSON containing a document name
// which will have to be saved by the frontend and a signed URL which the frontend has to request with PUT method
// together with the file.
func (c *Client) HandleDismissalDenySupportDocUpload(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	filename, err := c.DismissalStorage.GenerateDismissalDocName("application/pdf")
	if err != nil {
		if err == object.ErrFileTypeUnsupported {
			err = ec.NewError(ErrCodeMimeTypeNotSupported, Errs[ErrCodeMimeTypeNotSupported], err)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDismissalDenySupportDocUpload)
	defer cancel()

	url, err := c.DismissalStorage.GenerateDismissalDocPutSign(ctx, path.Join(DismissalDenySupportDocSubdir, filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
		"filename": filename,
		"url":      url.String(),
	}, false)
}

// HandleDismissalDenySupportDocPreview handles a request to download a support document currently in temporary location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandleDismissalDenySupportDocPreview(writer http.ResponseWriter, request *http.Request) {
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

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDismissalDenySupportDocPreview)
	defer cancel()

	url, err := c.DismissalStorage.GenerateDismissalDocGetSignTemp(ctx, path.Join(DismissalDenySupportDocSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleDismissalDenySupportDocDownload handles a request to download a support document from permanent location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandleDismissalDenySupportDocDownload(writer http.ResponseWriter, request *http.Request) {
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

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDismissalDenySupportDocDownload)
	defer cancel()

	url, err := c.DismissalStorage.GenerateDismissalDocGetSign(ctx, path.Join(DismissalDenySupportDocSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleGetDismissalStatusStatistic returns the number of dismissal items for each status.
func (c *Client) HandleGetDismissalStatusStatistic(writer http.ResponseWriter, _ *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutGetDismissalStatusStatistic)
	defer cancel()

	statistic, err := c.GetDismissalStatusStatisticCtx(ctx)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, statistic)
}
