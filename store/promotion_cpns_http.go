package store

import (
	"context"
	"fmt"
	"net/http"
	"path"

	. "github.com/fazrithe/siasn-jf-backend-git/errnum"
	"github.com/fazrithe/siasn-jf-backend-git/store/models"
	"github.com/if-itb/siasn-libs-backend/auth"
	"github.com/if-itb/siasn-libs-backend/ec"
	"github.com/if-itb/siasn-libs-backend/httputil"
)

const (
	TimeoutPromotionCpnsAdmissionGet                     = TimeoutDefault
	TimeoutPromotionCpnsAdmissionSearchPaginated         = TimeoutDefault
	TimeoutPromotionCpnsAdmissionPakLetterUpload         = TimeoutDefault
	TimeoutPromotionCpnsAdmissionPakLetterPreview        = TimeoutDefault
	TimeoutPromotionCpnsAdmissionPakLetterDownload       = TimeoutDefault
	TimeoutPromotionCpnsAdmissionPromotionLetterUpload   = TimeoutDefault
	TimeoutPromotionCpnsAdmissionPromotionLetterPreview  = TimeoutDefault
	TimeoutPromotionCpnsAdmissionPromotionLetterDownload = TimeoutDefault
	TimeoutPromotionCpnsAdmissionSubmit                  = TimeoutDefault
	TimeoutGetPromotionCpnsStatusStatistic               = TimeoutDefault
)

// HandlePromotionCpnsAdmissionGet handles retrieving promotion CPNS admission detail.
func (c *Client) HandlePromotionCpnsAdmissionGet(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionCpnsAdmissionGet)
	defer cancel()

	pa := &struct {
		PromotionCpnsId string `schema:"pengangkatan_cpns_id"`
	}{}
	err := c.decodeRequestSchema(writer, request, pa)
	if err != nil {
		return
	}

	promotion, err := c.GetPromotionCpnsAdmissionDetailCtx(ctx, pa.PromotionCpnsId)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, promotion)
}

// HandlePromotionCpnsAdmissionSearchPaginated handles a request to get CPNS promotion admission list.
func (c *Client) HandlePromotionCpnsAdmissionSearchPaginated(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionCpnsAdmissionSearchPaginated)
	defer cancel()

	query := &PromotionCpnsAdmissionSearchFilter{}
	err := c.decodeRequestSchema(writer, request, query)
	if err != nil {
		return
	}

	if query.CountPerPage == 0 {
		query.CountPerPage = 10
	}

	if query.PageNumber == 0 {
		query.PageNumber = 1
	}

	err = c.httpErrorVerifyListMeta(writer, query.PageNumber, query.CountPerPage)
	if err != nil {
		return
	}

	result, err := c.SearchPromotionCpnsAdmissionsPaginatedCtx(ctx, query)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, (*models.IdPaginatedList)(result))
}

// HandlePromotionCpnsAdmissionPakLetterUpload handles a request to upload a cpns promotion admission PAK letter.
// We do not redirect the request to object storage signed URL. We instead return a JSON containing a document name
// which will have to be saved by the frontend and a signed URL which the frontend has to request with PUT method
// together with the file.
func (c *Client) HandlePromotionCpnsAdmissionPakLetterUpload(writer http.ResponseWriter, _ *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionCpnsAdmissionPakLetterUpload)
	defer cancel()

	filename, err := c.PromotionCpnsStorage.GeneratePromotionCpnsDocName("application/pdf")
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodePromotionCpnsAdmissionGeneric, Errs[ErrCodePromotionCpnsAdmissionGeneric], err))
		return
	}

	url, err := c.PromotionCpnsStorage.GeneratePromotionCpnsDocPutSign(ctx, path.Join(PromotionCpnsPakLetterSubdir, filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
		"filename": filename,
		"url":      url.String(),
	}, false)
}

// HandlePromotionCpnsAdmissionPakLetterPreview handles a request to download a PAK letter currently in temporary location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandlePromotionCpnsAdmissionPakLetterPreview(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionCpnsAdmissionPakLetterPreview)
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

	url, err := c.PromotionCpnsStorage.GeneratePromotionCpnsDocGetSignTemp(ctx, path.Join(PromotionCpnsPakLetterSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandlePromotionCpnsAdmissionPakLetterDownload handles a request to download a PAK letter document from permanent location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandlePromotionCpnsAdmissionPakLetterDownload(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionCpnsAdmissionPakLetterDownload)
	defer cancel()

	type schemaPromotionCpnsId struct {
		PromotionCpnsId string `schema:"pengangkatan_cpns_id"`
	}

	s := &schemaPromotionCpnsId{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.PromotionCpnsId == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	_, err = c.getPromotionCpnsCtx(ctx, s.PromotionCpnsId)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	url, err := c.PromotionStorage.GeneratePromotionDocGetSign(ctx, path.Join(PromotionCpnsPakLetterSubdir, fmt.Sprintf("%s.pdf", s.PromotionCpnsId)))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandlePromotionCpnsAdmissionPromotionLetterUpload handles a request to upload a cpns promotion admission promotion letter.
// We do not redirect the request to object storage signed URL. We instead return a JSON containing a document name
// which will have to be saved by the frontend and a signed URL which the frontend has to request with PUT method
// together with the file.
func (c *Client) HandlePromotionCpnsAdmissionPromotionLetterUpload(writer http.ResponseWriter, _ *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionCpnsAdmissionPromotionLetterUpload)
	defer cancel()

	filename, err := c.PromotionCpnsStorage.GeneratePromotionCpnsDocName("application/pdf")
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodePromotionCpnsAdmissionGeneric, Errs[ErrCodePromotionCpnsAdmissionGeneric], err))
		return
	}

	url, err := c.PromotionCpnsStorage.GeneratePromotionCpnsDocPutSign(ctx, path.Join(PromotionCpnsPromotionLetterSubdir, filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
		"filename": filename,
		"url":      url.String(),
	}, false)
}

// HandlePromotionCpnsAdmissionPromotionLetterPreview handles a request to download a cpns promotion admission
// promotion currently in temporary location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandlePromotionCpnsAdmissionPromotionLetterPreview(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionCpnsAdmissionPromotionLetterPreview)
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

	url, err := c.PromotionCpnsStorage.GeneratePromotionCpnsDocGetSignTemp(ctx, path.Join(PromotionCpnsPromotionLetterSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandlePromotionCpnsAdmissionPromotionLetterDownload handles a request to download a cpns promotion admission letter
// document from permanent location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandlePromotionCpnsAdmissionPromotionLetterDownload(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionCpnsAdmissionPromotionLetterDownload)
	defer cancel()

	type schemaPromotionCpnsId struct {
		PromotionCpnsId string `schema:"pengangkatan_cpns_id"`
	}

	s := &schemaPromotionCpnsId{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.PromotionCpnsId == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	_, err = c.getPromotionCpnsCtx(ctx, s.PromotionCpnsId)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	url, err := c.PromotionStorage.GeneratePromotionDocGetSign(ctx, path.Join(PromotionCpnsPromotionLetterSubdir, fmt.Sprintf("%s.pdf", s.PromotionCpnsId)))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandlePromotionCpnsAdmissionSubmit handles a new CPNS promotion admission request.
// Dismissal admission is created for Modul Pengangkatan CPNS. A new CPNS promotion request entry
// will be created in the database with status set to `newly admitted`.
func (c *Client) HandlePromotionCpnsAdmissionSubmit(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPromotionCpnsAdmissionSubmit)
	defer cancel()

	user := auth.AssertReqGetUserDetail(request)

	pca := &models.PromotionCpnsAdmission{}
	err := c.decodeRequestJson(writer, request, pca)
	if err != nil {
		return
	}

	pca.SubmitterAsnId = user.AsnId

	promotionCpnsId, err := c.SubmitPromotionCpnsAdmissionCtx(ctx, pca)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]string{
		"pengangkatan_cpns_id": promotionCpnsId,
	})
}

// HandleGetPromotionCpnsStatusStatistic returns the number of cpns promotion items for each status.
func (c *Client) HandleGetPromotionCpnsStatusStatistic(writer http.ResponseWriter, _ *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutGetPromotionCpnsStatusStatistic)
	defer cancel()

	statistic, err := c.GetPromotionCpnsStatusStatisticCtx(ctx)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, statistic)
}
