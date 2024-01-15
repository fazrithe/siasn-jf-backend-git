package store

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"time"

	. "github.com/if-itb/siasn-jf-backend/errnum"
	"github.com/if-itb/siasn-jf-backend/store/models"
	"github.com/if-itb/siasn-jf-backend/store/object"
	"github.com/if-itb/siasn-libs-backend/auth"
	"github.com/if-itb/siasn-libs-backend/ec"
	"github.com/if-itb/siasn-libs-backend/httputil"
)

const (
	TimeoutActivityAdmissionSubmit             = TimeoutDefault
	TimeoutActivityAdmissionSupportDocUpload   = TimeoutDefault
	TimeoutActivityAdmissionSupportDocPreview  = TimeoutDefault
	TimeoutActivityAdmissionSupportDocDownload = TimeoutDefault
	TimeoutActivityAdmissionAsnGet             = TimeoutDefault
	// Deprecated: use TimeoutActivityAdmissionAsnGet.
	TimeoutActivityAdmissionAsnGetNew = TimeoutDefault
	// Deprecated: use TimeoutActivityAdmissionAsnGet.
	TimeoutActivityAdmissionAsnGetOld           = TimeoutDefault
	TimeoutActivityStatusCsrSet                 = TimeoutDefault
	TimeoutActivityCertGenSubmit                = TimeoutDefault
	TimeoutActivityCertGenDocUpload             = TimeoutDefault
	TimeoutActivityCertGenDocPreview            = TimeoutDefault
	TimeoutActivityCertGenDocDownload           = TimeoutDefault
	TimeoutActivityAdmissionSearch              = TimeoutDefault
	TimeoutActivityAdmissionVerification        = TimeoutDefault
	TimeoutActivityAdmissionDetail              = TimeoutDefault
	TimeoutActivityRecommendationLetterUpload   = TimeoutDefault
	TimeoutActivityRecommendationLetterPreview  = TimeoutDefault
	TimeoutActivityRecommendationLetterDownload = TimeoutDefault
	TimeoutActivityRecommendationLetterSubmit   = TimeoutDefault
	TimeoutGetActivityStatusStatistic           = TimeoutDefault
)

type SchemaActivityId struct {
	ActivityId string `schema:"kegiatan_id" json:"kegiatan_id"`
}

// HandleActivityAdmissionSubmit handles a new admission request.
// Activity admission is created when an agency want to hold an event for their civil servants. A new activity
// will be created in the database with status set to `newly admitted`.
func (c *Client) HandleActivityAdmissionSubmit(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)

	ar := &models.ActivityAdmission{}
	err := c.decodeRequestJson(writer, request, ar)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityAdmissionSubmit)
	defer cancel()

	ar.SubmitterAsnId = user.AsnId
	ar.AgencyId = user.WorkAgencyId
	ar.AdmissionTimestamp = models.EpochTime(time.Now())

	activityId, err := c.InsertActivityAdmissionCtx(ctx, ar)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]string{
		"kegiatan_id": activityId,
	})
}

// HandleActivityAdmissionSupportDocUpload handles a request to upload an admission support document.
// We do not redirect the request to object storage signed URL. We instead return a JSON containing a document name
// which will have to be saved by the frontend and a signed URL which the frontend has to request with PUT method
// together with the file.
func (c *Client) HandleActivityAdmissionSupportDocUpload(writer http.ResponseWriter, request *http.Request) {
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

	filename, err := c.ActivityStorage.GenerateActivityFilename(ur.MimeType)
	if err != nil {
		if err == object.ErrFileTypeUnsupported {
			err = ec.NewError(ErrCodeMimeTypeNotSupported, Errs[ErrCodeMimeTypeNotSupported], err)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityAdmissionSupportDocUpload)
	defer cancel()

	url, err := c.ActivityStorage.GenerateActivityDocPutSign(ctx, path.Join(ActivitySupportDocSubdir, filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
		"filename": filename,
		"url":      url.String(),
	}, false)
}

// HandleActivityAdmissionSupportDocPreview handles a request to download a support document currently in temporary location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandleActivityAdmissionSupportDocPreview(writer http.ResponseWriter, request *http.Request) {
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

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityAdmissionSupportDocPreview)
	defer cancel()

	url, err := c.ActivityStorage.GenerateActivityDocGetSignTemp(ctx, path.Join(ActivitySupportDocSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleActivityAdmissionSupportDocDownload handles a request to download a support document from permanent location.
// Requires `filename` query parameter, the filename retrieved from uploading the file.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandleActivityAdmissionSupportDocDownload(writer http.ResponseWriter, request *http.Request) {
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

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityAdmissionSupportDocDownload)
	defer cancel()

	url, err := c.ActivityStorage.GenerateActivityDocGetSign(ctx, path.Join(ActivitySupportDocSubdir, s.Filename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
	// _ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
	// 	"Message": "Success",
	// 	"URL":     url.String(),
	// }, false)
}

// HandleActivityAdmissionAsnGet handles a request to get exact ASN data by ASN NIP, which is treated as new and old NIP.
// When there are two matches found, one by new and one for old NIP match, the one that match the new NIP is chosen.
// Agency ID is needed, retrieved from authentication token.
func (c *Client) HandleActivityAdmissionAsnGet(writer http.ResponseWriter, request *http.Request) {
	type schemaNip struct {
		Nip string `schema:"nip"`
	}

	user := auth.AssertReqGetUserDetail(request)

	s := &schemaNip{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityAdmissionAsnGet)
	defer cancel()

	asn, err := c.GetUserDetailByNipWorkAgencyId(ctx, s.Nip, user.WorkAgencyId)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if asn == nil {
		c.httpError(writer, ErrEntryNotFound)
		return
	}

	_ = httputil.WriteObj200(writer, asn)
}

// HandleActivityStatusCsrSubmit handles a request to set the activity admission status to certificate request.
func (c *Client) HandleActivityStatusCsrSubmit(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)

	s := &models.ActivityCsrRequest{}
	err := c.decodeRequestJson(writer, request, s)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityStatusCsrSet)
	defer cancel()

	s.SubmitterAsnId = user.AsnId
	s.AgencyId = user.WorkAgencyId

	modifiedAt, err := c.SetActivityStatusCsrCtx(ctx, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]interface{}{
		"kegiatan_id": s.ActivityId,
		"modified_at": modifiedAt.Unix(),
	})
}

// HandleActivityVerificationSet handles a request to set the activity admission status to accepted.
func (c *Client) HandleActivityVerificationSet(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)

	av := &models.ActivityVerificationRequest{}
	err := c.decodeRequestJson(writer, request, av)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityAdmissionVerification)
	defer cancel()

	av.SubmitterAsnId = user.AsnId
	av.AgencyId = user.WorkAgencyId

	modifiedAt, err := c.SetActivityStatusAcceptedCtx(ctx, av)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]interface{}{
		"kegiatan_id": av.ActivityId,
		"modified_at": modifiedAt.Unix(),
	})
}

// Deprecated: it is now generated.
// HandleActivityCertGenDocUpload handles request to upload a certificate.
// The returned upload URL must be called with PUT request.
func (c *Client) HandleActivityCertGenDocUpload(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)

	ur := &models.ActivityCertGenUploadRequest{AgencyId: user.WorkAgencyId}
	err := c.decodeRequestJson(writer, request, ur)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityCertGenDocUpload)
	defer cancel()

	filename, err := c.VerifyActivityCertCtx(ctx, ur)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	url, err := c.ActivityStorage.GenerateActivityDocPutSign(ctx, filename)
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
		"url": url.String(),
	}, false)
}

// Deprecated: it is now generated.
// HandleActivityCertGenDocPreview handles a request to download a certificate currently in temporary location.
// Requires `kegiatan_id`, `peserta_user_id`, and `jenis` query parameters. They are used to determine the filename.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandleActivityCertGenDocPreview(writer http.ResponseWriter, request *http.Request) {
	s := &models.ActivityCertGenUploadRequest{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.ActivityId == "" || s.AttendeeAsnId == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityCertGenDocPreview)
	defer cancel()

	filename, err := c.CreateCertFilename(s.ActivityId, s.AttendeeAsnId, s.Type)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	url, err := c.ActivityStorage.GenerateActivityDocGetSignTemp(ctx, filename)
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleActivityCertGenDocDownload handles a request to download a certificate from permanent location.
// Requires `kegiatan_id`, `peserta_user_id`, and `jenis` query parameters. They are used to determine the filename.
// This handler redirects the request. It returns 302 to a signed URL to download the document.
func (c *Client) HandleActivityCertGenDocDownload(writer http.ResponseWriter, request *http.Request) {
	s := &models.ActivityCertGenDownloadRequest{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.ActivityId == "" || s.AttendeeAsnId == "" {
		c.httpError(writer, ErrEntryNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityCertGenDocDownload)
	defer cancel()

	basename, err := c.GenerateActivityCertificateCtx(ctx, s.ActivityId, s.AttendeeAsnId, s.ForceRegenerate)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	url, err := c.ActivityStorage.GenerateActivityDocGetSign(ctx, path.Join(ActivityCertSubdir, basename))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// HandleActivityCertGenSubmit handles certificate/PAK submission.
// This may return some common errors for example 404 when the activity cannot be found. The submitted certificates/PAKs
// must have already existed in object storage temporary location.
//
// PAK document type is only supported for activity with type uji kompetensi perpindahan jabatan (models.ActivityTypeMutationExam).
func (c *Client) HandleActivityCertGenSubmit(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)

	acg := &models.ActivityCertGenRequest{SubmitterAsnId: user.AsnId, AgencyId: user.WorkAgencyId}
	err := c.decodeRequestJson(writer, request, acg)
	if err != nil {
		return
	}

	// jsonData := `{
	// 			"name": "` + acg.Signing.JabtanInstansiPengusul + `",
	// 			"jabatan_instansi_pengusul": "` + acg.Signing.JabtanInstansiPengusul + `",
	// 			"nama_pejabat_instansi_pengusul": "` + acg.Signing.NamaPejabatInstansiPengusul + `",
	// 			"nip_pejabat_instansi_pengusul": "` + acg.Signing.NipPejabatInstansiPnegusul + `",
	// 			"jabatan_instansi_penyelenggara": "` + acg.Signing.JabatanInstansiPenyelenggara + `",
	// 			"nama_pejabat_instansi_penyelenggara": "` + acg.Signing.NamaPejabatInstansiPenyelenggara + `",
	// 			"nip_pejabat_instansi_penyelenggara": "` + acg.Signing.NipPejabatInstansiPenyelanggara + `",
	// 			"jabatan_pemateri": "` + acg.Signing.JabatanPemateri + `",
	// 			"nama_pemateri": "` + acg.Signing.NamaPemateri + `",
	// 			"nip_pemateri": "` + acg.Signing.NipPemateri + `"
	// 			}`

	// recordID := acg.ActivityId
	// templateID := acg.TemplateId
	// query := "UPDATE kegiatan SET penandatangan = $1, template_id = $2 WHERE kegiatan_id = $3"

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityAdmissionSubmit)
	defer cancel()

	// mtx, err := c.createMtxDb(ctx, c.Db)
	// if err != nil {
	// 	return
	// }
	// defer func() {
	// 	c.completeMtx(mtx, err)
	// }()

	// result, err := mtx.Exec(query, jsonData, templateID, recordID)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	modifiedAt, err := c.InsertActivityCertCtx(ctx, acg)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]interface{}{
		"kegiatan_id": acg.ActivityId,
		// "penandatangan": result,
		"modified_at": modifiedAt.Unix(),
	})
}

// Deprecated: use the paginated version.
// HandleActivityAdmissionSearch handles a request to get admission list of a work agency.
// Agency ID will be retrieved from authentication token.
func (c *Client) HandleActivityAdmissionSearch(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)

	type schemaAdmissionSearch struct {
		// AdmissionDate is a date with a format of YYYY-MM-DD (e.g. 2006-12-31)
		AdmissionDate   string `schema:"tgl_usulan"`
		AdmissionStatus int    `schema:"status"`
		AdmissionType   int    `schema:"jenis_kegiatan"`
	}

	query := &schemaAdmissionSearch{}
	err := c.decodeRequestSchema(writer, request, query)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityAdmissionSearch)
	defer cancel()

	if _, ok := models.ActivityAdmissionStatuses[query.AdmissionStatus]; query.AdmissionStatus != 0 && !ok {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeFilterStatusInvalid, Errs[ErrCodeFilterStatusInvalid]))
		return
	}

	if _, ok := models.ActivityTypes[query.AdmissionType]; query.AdmissionType != 0 && !ok {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeFilterInvalidType, Errs[ErrCodeFilterInvalidType]))
		return
	}

	var admissionDate models.Iso8601Date
	if query.AdmissionDate != "" {
		admissionDate, err = models.ParseIso8601Date(query.AdmissionDate)
		if err != nil {
			c.httpError(writer, ec.NewError(ErrCodeFilterInvalidDate, Errs[ErrCodeFilterInvalidDate], err))
			return
		}
	}

	searchFilter := &ActivityAdmissionSearchFilter{
		WorkAgencyId:    user.WorkAgencyId,
		AdmissionDate:   admissionDate,
		AdmissionStatus: query.AdmissionStatus,
		AdmissionType:   query.AdmissionType,
	}

	admissions, err := c.SearchActivityAdmissionsCtx(ctx, searchFilter)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, admissions)
}

// HandleActivityAdmissionSearchPaginated handles a request to get admission list of a work agency.
// Agency ID will be retrieved from authentication token.
func (c *Client) HandleActivityAdmissionSearchPaginated(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityAdmissionSearch)
	defer cancel()

	user := auth.AssertReqGetUserDetail(request)

	type schemaAdmissionSearch struct {
		// AdmissionDate is a date with a format of YYYY-MM-DD (e.g. 2006-12-31)
		AdmissionDate   string `schema:"tgl_usulan"`
		AdmissionStatus int    `schema:"status"`
		AdmissionType   int    `schema:"jenis_kegiatan"`
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

	if _, ok := models.ActivityAdmissionStatuses[query.AdmissionStatus]; query.AdmissionStatus != 0 && !ok {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeFilterStatusInvalid, Errs[ErrCodeFilterStatusInvalid]))
		return
	}

	if _, ok := models.ActivityTypes[query.AdmissionType]; query.AdmissionType != 0 && !ok {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeFilterInvalidType, Errs[ErrCodeFilterInvalidType]))
		return
	}

	var admissionDate models.Iso8601Date
	if query.AdmissionDate != "" {
		admissionDate, err = models.ParseIso8601Date(query.AdmissionDate)
		if err != nil {
			c.httpError(writer, ec.NewError(ErrCodeFilterInvalidDate, Errs[ErrCodeFilterInvalidDate], err))
			return
		}
	}

	searchFilter := &ActivityAdmissionSearchFilter{
		WorkAgencyId:    user.WorkAgencyId,
		AdmissionDate:   admissionDate,
		AdmissionStatus: query.AdmissionStatus,
		AdmissionType:   query.AdmissionType,
		CountPerPage:    countPerPage,
		PageNumber:      pageNumber,
	}

	admissions, err := c.SearchActivityAdmissionsPaginatedCtx(ctx, searchFilter)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, (*models.IdPaginatedList)(admissions))
}

// HandleActivityAdmissionSearchPembina handles a request to get admission list of a work agency.
// This handler will only process request from 'pembina'.
func (c *Client) HandleActivityAdmissionSearchPembina(writer http.ResponseWriter, request *http.Request) {
	// TODO: pembina can see multiple agencies?
	type schemaAdmissionSearch struct {
		AgencyId string `schema:"instansi_id"`
		// AdmissionDate is a date with a format of YYYY-MM-DD (e.g. 2006-12-31)
		AdmissionDate   string `schema:"tgl_usulan"`
		AdmissionStatus int    `schema:"status"`
		AdmissionType   int    `schema:"jenis_kegiatan"`
	}

	// TODO: Check if the user is a 'pembina'

	query := &schemaAdmissionSearch{}
	err := c.decodeRequestSchema(writer, request, query)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityAdmissionSearch)
	defer cancel()

	if query.AgencyId == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeFilterInvalidAgencyId, Errs[ErrCodeFilterInvalidAgencyId]))
		return
	}

	if _, ok := models.ActivityAdmissionStatuses[query.AdmissionStatus]; query.AdmissionStatus != 0 && !ok {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeFilterStatusInvalid, Errs[ErrCodeFilterStatusInvalid]))
		return
	}

	if _, ok := models.ActivityTypes[query.AdmissionType]; query.AdmissionType != 0 && !ok {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeFilterInvalidType, Errs[ErrCodeFilterInvalidType]))
		return
	}

	var admissionDate models.Iso8601Date
	if query.AdmissionDate != "" {
		admissionDate, err = models.ParseIso8601Date(query.AdmissionDate)
		if err != nil {
			c.httpError(writer, ec.NewError(ErrCodeFilterInvalidDate, Errs[ErrCodeFilterInvalidDate], err))
			return
		}
	}

	searchFilter := &ActivityAdmissionSearchFilter{
		WorkAgencyId:    query.AgencyId,
		AdmissionDate:   admissionDate,
		AdmissionStatus: query.AdmissionStatus,
		AdmissionType:   query.AdmissionType,
	}

	admissions, err := c.SearchActivityAdmissionsCtx(ctx, searchFilter)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, admissions)
}

// HandleActivityAdmissionDetail handles a request to get an activity admission's detail.
// Agency ID will be retrieved from authentication token. If it's a 'pembina' the they can also specify
func (c *Client) HandleActivityAdmissionDetail(writer http.ResponseWriter, request *http.Request) {
	type schemaAdmissionDetail struct {
		ActivityId string `schema:"kegiatan_id"`
	}

	query := &schemaAdmissionDetail{}
	err := c.decodeRequestSchema(writer, request, query)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityAdmissionDetail)
	defer cancel()

	// TODO: retrieve agency ID from auth token. If the user is pembina, pass empty string to the agencyId

	activityAdmissionDetail, err := c.GetActivityAdmissionDetailCtx(ctx, query.ActivityId, "") // TODO: Change this
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, activityAdmissionDetail)
}

// Deprecated: no longer needed.
// HandleActivityRecommendationLetterUpload handles a request to upload a recommendation letter.
// This will return the generated recommendation letter filename and a signed URL that can be used to upload the file with
// PUT method. You do not need to submit the mime type as the only allowed type is application/pdf.
func (c *Client) HandleActivityRecommendationLetterUpload(writer http.ResponseWriter, request *http.Request) {
	type schemaActivityId struct {
		ActivityId string `json:"kegiatan_id"`
	}

	s := &schemaActivityId{}
	err := c.decodeRequestJson(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	user := auth.AssertReqGetUserDetail(request)

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityRecommendationLetterUpload)
	defer cancel()

	err = c.VerifyActivityRecommendationLetterCtx(ctx, s.ActivityId, user.WorkAgencyId)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	url, err := c.ActivityStorage.GenerateActivityDocPutSign(ctx, path.Join(ActivityRecommendationLetterSubdir, fmt.Sprintf("%s.pdf", s.ActivityId)))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
		"url": url.String(),
	}, false)
}

// Deprecated: no longer needed.
// HandleActivityRecommendationLetterPreview handles a request to preview a recommendation letter in temporary location.
// This will redirect your request to the storage signed URL.
func (c *Client) HandleActivityRecommendationLetterPreview(writer http.ResponseWriter, request *http.Request) {
	type schemaActivityId struct {
		ActivityId string `schema:"kegiatan_id"`
	}

	s := &schemaActivityId{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.ActivityId == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityRecommendationLetterPreview)
	defer cancel()

	url, err := c.ActivityStorage.GenerateActivityDocGetSignTemp(ctx, path.Join(ActivityRecommendationLetterSubdir, fmt.Sprintf("%s.pdf", s.ActivityId)))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// Deprecated: no longer needed.
// HandleActivityRecommendationLetterDownload handles a request to preview a recommendation letter in permanent location.
// This will redirect your request to the storage signed URL.
func (c *Client) HandleActivityRecommendationLetterDownload(writer http.ResponseWriter, request *http.Request) {
	type schemaActivityId struct {
		ActivityId string `schema:"kegiatan_id"`
	}

	s := &schemaActivityId{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.ActivityId == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeStorageFileNotFound, Errs[ErrCodeStorageFileNotFound]))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityRecommendationLetterDownload)
	defer cancel()

	url, err := c.ActivityStorage.GenerateActivityDocGetSign(ctx, path.Join(ActivityRecommendationLetterSubdir, fmt.Sprintf("%s.pdf", s.ActivityId)))
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, url.String(), http.StatusFound)
}

// Deprecated: no longer needed.
// HandleActivityRecommendationLetterSubmit handles request to submit recommendation letter.
func (c *Client) HandleActivityRecommendationLetterSubmit(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityRecommendationLetterSubmit)
	defer cancel()

	user := auth.AssertReqGetUserDetail(request)

	admission := &models.ActivityRecommendationLetterAdmission{}
	err := c.decodeRequestJson(writer, request, admission)
	if err != nil {
		return
	}

	if admission.RecommendationLetter == nil {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeActivityNoRecommendationLetter, Errs[ErrCodeActivityNoRecommendationLetter]))
		return
	}

	admission.SubmitterAsnId = user.AsnId
	admission.AgencyId = user.WorkAgencyId

	err = c.SubmitActivityRecommendationLetterCtx(ctx, admission)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]interface{}{
		"kegiatan_id": admission.ActivityId,
	})
}

// HandleGetActivityStatusStatistic returns the number of activity items for each status.
func (c *Client) HandleGetActivityStatusStatistic(writer http.ResponseWriter, _ *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutGetActivityStatusStatistic)
	defer cancel()

	statistic, err := c.GetActivityStatusStatisticCtx(ctx)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, statistic)
}
