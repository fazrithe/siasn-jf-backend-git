package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mime"
	"path"
	"time"

	"github.com/google/uuid"
	. "github.com/if-itb/siasn-jf-backend/errnum"
	"github.com/if-itb/siasn-jf-backend/store/models"
	"github.com/if-itb/siasn-jf-backend/store/object"
	"github.com/if-itb/siasn-libs-backend/docx"
	"github.com/if-itb/siasn-libs-backend/ec"
	"github.com/if-itb/siasn-libs-backend/metricutil"
	"github.com/if-itb/siasn-libs-backend/search"
	"github.com/lib/pq"
)

// CheckPromotionAdmissionInsert verifies that promotion request is valid.
func (c *Client) CheckPromotionAdmissionInsert(ctx context.Context, request *models.PromotionAdmission) (err error) {
	if request.AdmissionNumber == "" {
		return ec.NewError(ErrCodePromotionAdmissionFieldEmpty, Errs[ErrCodePromotionAdmissionFieldEmpty], errors.New("nomor_usulan cannot be empty"))
	}

	_, err = models.ParseIso8601Date(string(request.AdmissionDate))
	if err != nil {
		return ec.NewError(ErrCodePromotionInvalidDate, Errs[ErrCodePromotionInvalidDate], errors.New("tanggal_usulan invalid"))
	}

	if _, ok := models.PromotionTypes[request.PromotionType]; !ok {
		return ErrPromotionAdmissionInvalidPromotionType
	}

	switch request.PromotionType {
	case models.PromotionTypeTransfer:
		if request.RecommendationLetter == nil || request.RecommendationLetter.Filename == "" {
			return ErrPromotionAdmissionRecommendationLetterEmpty
		}

		_, err = models.ParseIso8601Date(string(request.RecommendationLetter.DocumentDate))
		if err != nil {
			return ec.NewError(ErrCodePromotionInvalidDate, Errs[ErrCodePromotionInvalidDate], fmt.Errorf("surat_rekomendasi's tgl_dokumen invalid: %w", err))
		}

		if request.PakLetter == nil || request.PakLetter.Filename == "" {
			return ErrPromotionAdmissionPakLetterEmpty
		}

		_, err = models.ParseIso8601Date(string(request.PakLetter.DocumentDate))
		if err != nil {
			return ec.NewError(ErrCodePromotionInvalidDate, Errs[ErrCodePromotionInvalidDate], fmt.Errorf("surat_pak's tgl_dokumen invalid: %w", err))
		}

	case models.PromotionTypePromotion:
		if request.TestCertificate == nil || request.TestCertificate.Filename == "" {
			return ErrPromotionAdmissionTestCertificateEmpty
		}

		_, err = models.ParseIso8601Date(string(request.TestCertificate.DocumentDate))
		if err != nil {
			return ec.NewError(ErrCodePromotionInvalidDate, Errs[ErrCodePromotionInvalidDate], fmt.Errorf("sertifikat_uji_kompetensi's tgl_dokumen invalid: %w", err))
		}

		if request.TestStatus == 0 {
			return ec.NewError(
				ErrCodePromotionAdmissionFieldEmpty,
				Errs[ErrCodePromotionAdmissionFieldEmpty],
				fmt.Errorf("test_status cannot be empty, must be 1 (pass) or 2 (fail)"),
			)
		}

		if _, ok := models.PromotionCompetencyTestStatuses[request.TestStatus]; !ok {
			return ErrPromotionAdmissionInvalidCompetencyTestStatus
		}
	}

	return nil
}

// InsertPromotionAdmissionCtx inserts a new promotion admission request (pengajuan pengangkatan).
// It also moves PAK and recommendation letter from temporary storage to permanent storage in object storage.
func (c *Client) InsertPromotionAdmissionCtx(ctx context.Context, request *models.PromotionAdmission) (promotionId string, err error) {
	if err = c.CheckPromotionAdmissionInsert(ctx, request); err != nil {
		return "", err
	}

	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return "", err
	}

	defer func() {
		c.completeMtx(mtx, err)
	}()
	var (
		pakLetterDocumentName              sql.NullString
		pakLetterDocumentNumber            sql.NullString
		pakLetterDocumentDate              sql.NullString
		recommendationLetterDocumentName   sql.NullString
		recommendationLetterDocumentNumber sql.NullString
		recommendationLetterDocumentDate   sql.NullString
		testCertificateDocumentName        sql.NullString
		testCertificateDocumentNumber      sql.NullString
		testCertificateDocumentDate        sql.NullString
	)
	if request.PakLetter != nil {
		pakLetterDocumentName = sql.NullString{String: request.PakLetter.Filename, Valid: request.PakLetter.Filename != ""}
		pakLetterDocumentNumber = sql.NullString{String: request.PakLetter.DocumentNumber, Valid: request.PakLetter.DocumentNumber != ""}
		pakLetterDocumentDate = sql.NullString{String: string(request.PakLetter.DocumentDate), Valid: request.PakLetter.DocumentDate != ""}
	}
	if request.RecommendationLetter != nil {
		recommendationLetterDocumentName = sql.NullString{String: request.RecommendationLetter.Filename, Valid: request.RecommendationLetter.Filename != ""}
		recommendationLetterDocumentNumber = sql.NullString{String: request.RecommendationLetter.DocumentNumber, Valid: request.RecommendationLetter.DocumentNumber != ""}
		recommendationLetterDocumentDate = sql.NullString{String: string(request.RecommendationLetter.DocumentDate), Valid: request.RecommendationLetter.DocumentDate != ""}
	}
	if request.TestCertificate != nil {
		testCertificateDocumentName = sql.NullString{String: request.TestCertificate.Filename, Valid: request.TestCertificate.Filename != ""}
		testCertificateDocumentNumber = sql.NullString{String: request.TestCertificate.DocumentNumber, Valid: request.TestCertificate.DocumentNumber != ""}
		testCertificateDocumentDate = sql.NullString{String: string(request.TestCertificate.DocumentDate), Valid: request.TestCertificate.DocumentDate != ""}
	}

	// Create a new promotion entry
	promotionId = uuid.New().String()
	_, err = mtx.ExecContext(ctx,
		`insert into pengangkatan (
                          uuid_pengangkatan,
                          asn_id,
                          no_usulan,
                          tgl_usulan,
                          jenis_pengangkatan,
                          jabatan_fungsional_tujuan_id,
                          status,
                          status_ts,
                          status_by,
                          test_status,
                          test_nilai,
                          nama_doc_pak,
                          no_doc_pak,
                          tgl_doc_pak,
                          nama_doc_surat_rekomendasi,
                          no_doc_surat_rekomendasi,
                          tgl_doc_surat_rekomendasi,
                          nama_doc_sertifikat_uji_kompetensi,
                          no_doc_sertifikat_uji_kompetensi,
                          tgl_doc_sertifikat_uji_kompetensi)
			values ($1, $2, $3, $4::date, $5, $6, $7, now(), $8, $9, $10, $11, $12, $13::date, $14, $15, $16::date, $17, $18, $19::date)`,
		promotionId,
		request.AsnId,
		request.AdmissionNumber,
		request.AdmissionDate,
		request.PromotionType,
		request.PromotionPositionId,
		models.PromotionAdmissionStatusCreated,
		request.SubmitterAsnId,
		request.TestStatus,
		float64ToNullFloat64(request.TestScore),
		pakLetterDocumentName,
		pakLetterDocumentNumber,
		pakLetterDocumentDate,
		recommendationLetterDocumentName,
		recommendationLetterDocumentNumber,
		recommendationLetterDocumentDate,
		testCertificateDocumentName,
		testCertificateDocumentNumber,
		testCertificateDocumentDate,
	)
	if err != nil {
		return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to pengangkatan: %w", err))
	}

	if request.PakLetter != nil {
		// Save PAK letter
		_, err = c.PromotionStorage.SavePromotionFile(ctx,
			path.Join(PromotionPakLetterSubdir, request.PakLetter.Filename),
			path.Join(PromotionPakLetterSubdir, promotionId),
		)
		if err != nil {
			if err == object.ErrTempFileNotFound {
				return "", ErrStorageFileNotFound
			}
			return "", ec.NewError(ErrCodeStorageCopyFail, Errs[ErrCodeStorageCopyFail], err)
		}
	}

	if request.RecommendationLetter != nil {
		// Save recommendation letter
		_, err = c.PromotionStorage.SavePromotionFile(ctx,
			path.Join(PromotionRecommendationLetterSubdir, request.RecommendationLetter.Filename),
			path.Join(PromotionRecommendationLetterSubdir, promotionId),
		)
		if err != nil {
			if err == object.ErrTempFileNotFound {
				return "", ErrStorageFileNotFound
			}
			return "", ec.NewError(ErrCodeStorageCopyFail, Errs[ErrCodeStorageCopyFail], err)
		}
	}

	if request.TestCertificate != nil {
		// Save test certificate
		_, err = c.PromotionStorage.SavePromotionFile(ctx,
			path.Join(PromotionTestCertificateSubdir, request.TestCertificate.Filename),
			path.Join(PromotionTestCertificateSubdir, promotionId),
		)
		if err != nil {
			if err == object.ErrTempFileNotFound {
				return "", ErrStorageFileNotFound
			}
			return "", ec.NewError(ErrCodeStorageCopyFail, Errs[ErrCodeStorageCopyFail], err)
		}
	}

	return promotionId, nil
}

// IsPromotionExistCtx checks if promotion entry exists.
func (c *Client) IsPromotionExistCtx(ctx context.Context, promotionId string) (isExist bool, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)

	d := 0
	err = mdb.QueryRowContext(ctx, "select 1 from pengangkatan where uuid_pengangkatan = $1", promotionId).Scan(&d)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	return true, nil
}

// SetPromotionStatusAcceptedCtx sets a promotion status as models.PromotionAdmissionStatusAccepted.
// If the status of the promotion is not models.PromotionAdmissionStatusCreated, this will return error code errnum.ErrCodePromotionAdmissionStatusProcessedFurther.
// If the status is already models.PromotionAdmissionStatusAccepted, this will return error code errnum.ErrCodePromotionAdmissionStatusAlreadyAccepted.
func (c *Client) SetPromotionStatusAcceptedCtx(ctx context.Context, promotion *models.PromotionAdmission) (modifiedAt time.Time, err error) {
	promotionId, err := uuid.Parse(promotion.PromotionId)
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodeUuidInvalid, Errs[ErrCodeUuidInvalid], err)
	}

	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return time.Time{}, err
	}

	defer func() {
		c.completeMtx(mtx, err)
	}()

	currentStatus := 0
	err = mtx.QueryRowContext(ctx, "select status from pengangkatan where uuid_pengangkatan = $1 for update", promotionId).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, ErrEntryNotFound
		}

		return time.Time{}, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	if currentStatus == models.PromotionAdmissionStatusAccepted {
		return time.Time{}, ErrPromotionAdmissionStatusAlreadyAccepted
	}

	if currentStatus != models.PromotionAdmissionStatusCreated {
		return time.Time{}, ErrPromotionAdmissionStatusProcessedFurther
	}

	modifiedAt = time.Now()
	_, err = mtx.ExecContext(
		ctx,
		"update pengangkatan set status = $1, status_ts = $2, status_by = $3 where uuid_pengangkatan = $4",
		models.PromotionAdmissionStatusAccepted,
		modifiedAt,
		promotion.SubmitterAsnId,
		promotionId.String(),
	)
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], err)
	}

	return modifiedAt, nil
}

// SetPromotionStatusRejectCtx sets a promotion status as models.PromotionAdmissionStatusRejected.
// If the status of the promotion is not models.PromotionAdmissionStatusCreated, this will return error code errnum.ErrCodePromotionAdmissionStatusProcessedFurther.
// If the status is already models.PromotionAdmissionStatusRejected, this will return error code errnum.ErrCodePromotionAdmissionStatusAlreadyRejected.
func (c *Client) SetPromotionStatusRejectCtx(ctx context.Context, promotion *models.PromotionReject) (modifiedAt time.Time, err error) {
	promotionId, err := uuid.Parse(promotion.PromotionId)
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodeUuidInvalid, Errs[ErrCodeUuidInvalid], err)
	}

	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return time.Time{}, err
	}

	defer func() {
		c.completeMtx(mtx, err)
	}()

	currentStatus := 0
	err = mtx.QueryRowContext(ctx, "select status from pengangkatan where uuid_pengangkatan = $1 for update", promotionId).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, ErrEntryNotFound
		}

		return time.Time{}, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	if currentStatus == models.PromotionAdmissionStatusRejected {
		return time.Time{}, ErrPromotionAdmissionStatusAlreadyRejected
	}

	if currentStatus != models.PromotionAdmissionStatusCreated {
		return time.Time{}, ErrPromotionAdmissionStatusProcessedFurther
	}

	modifiedAt = time.Now()
	_, err = mtx.ExecContext(
		ctx,
		"update pengangkatan set status = $1, status_ts = $2, status_by = $3, alasan_tidak_diangkat = $4 where uuid_pengangkatan = $5",
		models.PromotionAdmissionStatusRejected,
		modifiedAt,
		promotion.SubmitterAsnId,
		promotion.RejectReason,
		promotionId.String(),
	)
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], err)
	}

	return modifiedAt, nil
}

// SearchPromotionAdmissionsPaginatedCtx searches for the list of promotion admissions.
// It will return empty slice if no admissions are found.
func (c *Client) SearchPromotionAdmissionsPaginatedCtx(ctx context.Context, filter *PromotionAdmissionSearchFilter) (result *search.PaginatedList, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)
	profileMdb := metricutil.NewDB(c.ProfileDb, c.SqlMetrics)

	c.Logger.Debugf("%v", (filter.PageNumber-1)*filter.CountPerPage)
	admissionRows, err := mdb.QueryContext(
		ctx,
		"select uuid_pengangkatan, asn_id, status, tgl_doc_surat_rekomendasi, jenis_pengangkatan from pengangkatan where ($1 <= 0 or status = $1) and ($2 <= 0 or jenis_pengangkatan = $2) and ($3::date is null or tgl_doc_surat_rekomendasi::date = $3) order by tgl_doc_surat_rekomendasi desc limit $4 offset $5",
		filter.AdmissionStatus,
		filter.AdmissionType,
		sql.NullString{Valid: string(filter.AdmissionDate) != "", String: string(filter.AdmissionDate)},
		filter.CountPerPage+1,
		(filter.PageNumber-1)*filter.CountPerPage,
	)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer admissionRows.Close()

	admissions := make([]*models.PromotionItem, 0)
	asnIds := make([]string, 0)
	for admissionRows.Next() {
		admission := &models.PromotionItem{}
		err = admissionRows.Scan(
			&admission.PromotionId,
			&admission.AsnId,
			&admission.Status,
			&admission.RecommendationLetterDate,
			&admission.PromotionType,
		)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
		admissions = append(admissions, admission)
		asnIds = append(asnIds, admission.AsnId)
	}
	if err = admissionRows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	rows, err := profileMdb.QueryContext(ctx, `select	id,	nama from orang	where id = any ($1)`, pq.Array(asnIds))
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer rows.Close()

	asns := make(map[string]string, 0)
	for rows.Next() {
		s := &models.Asn{}
		err = rows.Scan(
			&s.AsnId,
			&s.Name,
		)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
		asns[s.AsnId] = s.Name
	}
	if err = rows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	for _, admission := range admissions {
		admission.Name = asns[admission.AsnId]
	}

	hasNext := false
	if len(admissions) > filter.CountPerPage {
		hasNext = true
		admissions = admissions[:filter.CountPerPage]
	}

	return &search.PaginatedList{
		Data:     admissions,
		Metadata: search.CreatePaginatedListMetadataNoTotalNext(filter.PageNumber, len(admissions), hasNext),
	}, nil
}

// GetPromotionStatusStatisticCtx returns the number of promotion items for each status.
func (c *Client) GetPromotionStatusStatisticCtx(ctx context.Context) (statistics []*models.StatisticStatus, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)
	rows, err := mdb.QueryContext(ctx, "select status, jumlah from pengangkatan_status_statistik")
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer rows.Close()

	statistics = make([]*models.StatisticStatus, 0)
	statisticMap := make(map[int]*models.StatisticStatus)
	for status := range models.PromotionAdmissionStatuses {
		statusStatistic := &models.StatisticStatus{
			Status: status,
		}
		statistics = append(statistics, statusStatistic)
		statisticMap[status] = statusStatistic
	}
	for rows.Next() {
		var temp models.StatisticStatus
		err = rows.Scan(&temp.Status, &temp.Count)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
		statusStatistic, ok := statisticMap[temp.Status]
		if ok {
			statusStatistic.Count = temp.Count
		}
	}
	if err = rows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	return statistics, nil
}

// GeneratePromotionLetterCtx generates a promotion letter and store it in the storage.
// Filename is generated by adding a pdf extension to promotionId. The base filename is returned,
// it can be accessed in the promotion letter subdir in permanent bucket.
//
// To help reduce performance load, it is only generated once, unless forceRegenerate is set to true.
func (c *Client) GeneratePromotionLetterCtx(ctx context.Context, promotionId string, forceRegenerate bool) (filename string, err error) {
	filename = fmt.Sprintf("%s.pdf", promotionId)
	fullPath := path.Join(PromotionPromotionLetterSubdir, filename)

	if !forceRegenerate {
		fileFound := true
		meta, err := c.PromotionStorage.GetPromotionFileMetadata(ctx, fullPath)
		if err != nil && !errors.Is(err, object.ErrFileNotFound) {
			return "", ec.NewError(ErrCodeStorageGetMetadataFail, Errs[ErrCodeStorageGetMetadataFail], err)
		}

		if errors.Is(err, object.ErrFileNotFound) {
			fileFound = false
		} else {
			if ct, _, err := mime.ParseMediaType(meta.ContentType); meta.ContentLength <= 0 || err != nil || ct != "application/pdf" {
				fileFound = false
			}
		}

		if fileFound {
			return filename, nil
		}
	}

	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)
	referenceMtx, err := c.createMtxDb(ctx, c.ReferenceDb)
	if err != nil {
		return "", err
	}
	profileMtx, err := c.createMtxDb(ctx, c.ProfileDb)
	if err != nil {
		return "", err
	}
	defer func() {
		c.completeMtx(profileMtx, err)
	}()
	defer func() {
		c.completeMtx(referenceMtx, err)
	}()

	status := 0
	functionalPositionId := ""
	asnId := ""
	data := &PromotionLetterTemplate{}
	err = mdb.QueryRowContext(
		ctx,
		"select status, no_usulan, asn_id, to_char(tgl_usulan, 'YYYY-MM-DD'), jabatan_fungsional_tujuan_id from pengangkatan where uuid_pengangkatan = $1",
		promotionId,
	).Scan(
		&status,
		&data.AdmissionNumber,
		&asnId,
		&data.AdmissionDate,
		&functionalPositionId,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrEntryNotFound
		}
		return "", ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	if status != models.PromotionAdmissionStatusAccepted {
		return "", ErrPromotionAdmissionStatusNotAccepted
	}

	positions, err := c.getFunctionalPositionNames(ctx, referenceMtx, []string{functionalPositionId})
	if err != nil {
		return "", err
	}

	detail, err := c.getUserDetailByAsnId(ctx, profileMtx, referenceMtx, asnId, "")
	if err != nil {
		return "", err
	}

	if detail == nil {
		return "", ec.NewError(ErrCodeEntryNotFound, Errs[ErrCodeEntryNotFound], errors.New("ASN cannot be found in profile database"))
	}

	data.PromotionPositionName = positions[functionalPositionId]
	data.Name = detail.Name

	err = c.generatePromotionLetterCtx(ctx, fullPath, data)
	if err != nil {
		if errors.Is(err, docx.ErrSiasnRendererBadTemplate) {
			return "", ec.NewError(ErrCodeDocumentGenerateBadTemplate, Errs[ErrCodeDocumentGenerateBadTemplate], err)
		}
		return "", ec.NewError(ErrCodeDocumentGenerate, Errs[ErrCodeDocumentGenerate], err)
	}

	return filename, nil
}
