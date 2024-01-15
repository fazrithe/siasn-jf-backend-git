package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path"

	"github.com/google/uuid"
	. "github.com/if-itb/siasn-jf-backend/errnum"
	"github.com/if-itb/siasn-jf-backend/store/models"
	"github.com/if-itb/siasn-jf-backend/store/object"
	"github.com/if-itb/siasn-libs-backend/ec"
	"github.com/if-itb/siasn-libs-backend/metricutil"
	"github.com/if-itb/siasn-libs-backend/search"
)

// getPromotionCpnsCtx returns the status of a CPNS promotion admission.
// If the admission is not found, errnum.ErrEntryNotFound is returned.
func (c *Client) getPromotionCpnsCtx(ctx context.Context, promotionCpnsId string) (status int, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)
	return c.getPromotionCpnsDhCtx(ctx, mdb, promotionCpnsId)
}

// getPromotionCpnsDhCtx returns the status of a CPNS promotion admission.
// If the admission is not found, errnum.ErrEntryNotFound is returned.
func (c *Client) getPromotionCpnsDhCtx(ctx context.Context, dh metricutil.DbHandler, promotionCpnsId string) (status int, err error) {
	err = dh.QueryRowContext(ctx, "select status from pengangkatan_cpns where pengangkatan_cpns_id = $1", promotionCpnsId).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrEntryNotFound
		}
		return 0, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	return status, nil
}

// GetPromotionCpnsAdmissionDetailCtx retrieves detail about CPNS promotion admission detail.
func (c *Client) GetPromotionCpnsAdmissionDetailCtx(ctx context.Context, promotionCpnsId string) (admission *models.PromotionCpnsAdmission, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)
	profileMdb := metricutil.NewDB(c.ProfileDb, c.SqlMetrics)
	referenceMdb := metricutil.NewDB(c.ReferenceDb, c.SqlMetrics)

	admission = &models.PromotionCpnsAdmission{PromotionCpnsId: promotionCpnsId, PakLetter: &models.Document{}, PromotionLetter: &models.Document{}}
	err = mdb.QueryRowContext(ctx, `
select
	asn_id,
	jabatan_fungsional_tujuan_id,
	angka_kredit_pertama,
	unor_id,
	tgl_usulan,
	no_usulan,
	nama_doc_pak,
	no_doc_pak,
	tgl_doc_pak,
	nama_doc_surat_pengangkatan,
	no_doc_surat_pengangkatan,
	tgl_doc_surat_pengangkatan,
	status,
    status_by
from pengangkatan_cpns where pengangkatan_cpns_id = $1
`, promotionCpnsId).Scan(
		&admission.AsnId,
		&admission.PromotionPositionId,
		&admission.FirstCreditNumber,
		&admission.OrganizationUnitId,
		&admission.AdmissionDate,
		&admission.AdmissionNumber,
		&admission.PakLetter.DocumentName,
		&admission.PakLetter.DocumentNumber,
		&admission.PakLetter.DocumentDate,
		&admission.PromotionLetter.DocumentName,
		&admission.PromotionLetter.DocumentNumber,
		&admission.PromotionLetter.DocumentDate,
		&admission.Status,
		&admission.SubmitterAsnId,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrEntryNotFound
		}
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	asns, err := c.getAsnNipNames(ctx, profileMdb, []string{admission.AsnId})
	if err != nil {
		return nil, err
	}

	unors, err := c.getOrganizationUnitNames(ctx, referenceMdb, []string{admission.OrganizationUnitId})
	if err != nil {
		return nil, err
	}

	positions, err := c.getFunctionalPositionNames(ctx, referenceMdb, []string{admission.PromotionPositionId})
	if err != nil {
		return nil, err
	}

	admission.AsnNip = asns[admission.AsnId].Nip
	admission.AsnName = asns[admission.AsnId].AsnName
	admission.OrganizationUnit = unors[admission.OrganizationUnitId]
	admission.PromotionPosition = positions[admission.PromotionPositionId]

	return admission, nil
}

// SearchPromotionCpnsAdmissionsPaginatedCtx searches for the list of CPNS promotion admissions.
// It will return empty slice if no admissions are found.
func (c *Client) SearchPromotionCpnsAdmissionsPaginatedCtx(ctx context.Context, filter *PromotionCpnsAdmissionSearchFilter) (result *search.PaginatedList, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)
	profileMdb := metricutil.NewDB(c.ProfileDb, c.SqlMetrics)
	referenceMtx, err := c.createMtxDb(ctx, c.ReferenceDb)
	if err != nil {
		return nil, err
	}
	defer func() {
		c.completeMtx(referenceMtx, err)
	}()

	admissionRows, err := mdb.QueryContext(
		ctx,
		`
select
    pengangkatan_cpns_id,
	asn_id,
	jabatan_fungsional_tujuan_id,
	angka_kredit_pertama,
	unor_id,
	tgl_usulan,
	no_usulan,
	status
from pengangkatan_cpns where ($1 <= 0 or status = $1) and ($2::date is null or tgl_usulan::date = $2) order by tgl_usulan desc limit $3 offset $4
`,
		filter.AdmissionStatus,
		sql.NullString{Valid: filter.AdmissionDate != "", String: filter.AdmissionDate},
		filter.CountPerPage+1,
		(filter.PageNumber-1)*filter.CountPerPage,
	)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer admissionRows.Close()

	admissions := make([]*models.PromotionCpnsItem, 0)
	asnIds := make([]string, 0)
	positionIds := make([]string, 0)
	unorIds := make([]string, 0)
	i := 0
	hasNext := false
	for admissionRows.Next() {
		i++
		if i > filter.CountPerPage {
			// Stop scanning and mark hasNext as true as there are countPerPage+1 entries found.
			// We don't need the last row, just want to know if the last row exist at all.
			hasNext = true
			break
		}
		admission := &models.PromotionCpnsItem{}
		err = admissionRows.Scan(
			&admission.PromotionCpnsId,
			&admission.AsnId,
			&admission.PromotionPositionId,
			&admission.FirstCreditNumber,
			&admission.OrganizationUnitId,
			&admission.AdmissionDate,
			&admission.AdmissionNumber,
			&admission.Status,
		)
		if err != nil {
			_ = admissionRows.Close()
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
		admissions = append(admissions, admission)
		asnIds = append(asnIds, admission.AsnId)
		positionIds = append(positionIds, admission.PromotionPositionId)
		unorIds = append(unorIds, admission.OrganizationUnitId)
	}
	_ = admissionRows.Close()
	if err = admissionRows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	asns, err := c.getAsnNipNames(ctx, profileMdb, asnIds)
	if err != nil {
		return nil, err
	}

	unors, err := c.getOrganizationUnitNames(ctx, referenceMtx, unorIds)
	if err != nil {
		return nil, err
	}

	positions, err := c.getFunctionalPositionNames(ctx, referenceMtx, positionIds)
	if err != nil {
		return nil, err
	}

	for _, a := range admissions {
		a.AsnName = asns[a.AsnId].AsnName
		a.AsnNip = asns[a.AsnId].Nip
		a.PromotionPosition = positions[a.PromotionPositionId]
		a.OrganizationUnit = unors[a.OrganizationUnitId]
	}

	return &search.PaginatedList{
		Data:     admissions,
		Metadata: search.CreatePaginatedListMetadataNoTotalNext(filter.PageNumber, len(admissions), hasNext),
	}, nil
}

// checkPromotionCpnsAdmissionSubmitRequest checks whether the request to create a new CPNS promotion is valid.
func (c *Client) checkPromotionCpnsAdmissionSubmitRequest(request *models.PromotionCpnsAdmission) (err error) {
	if request.AsnId == "" {
		return ec.NewError(ErrCodePromotionCpnsAdmissionFieldInvalid, Errs[ErrCodePromotionCpnsAdmissionFieldInvalid], errors.New("asn_id is required"))
	}

	if request.AdmissionNumber == "" {
		return ec.NewError(ErrCodePromotionCpnsAdmissionFieldInvalid, Errs[ErrCodePromotionCpnsAdmissionFieldInvalid], errors.New("nomor_usulan is required"))
	}

	if request.PromotionPositionId == "" {
		return ec.NewError(ErrCodePromotionCpnsAdmissionFieldInvalid, Errs[ErrCodePromotionCpnsAdmissionFieldInvalid], errors.New("jabatan_fungsional_tujuan_id is required"))
	}

	if request.OrganizationUnitId == "" {
		return ec.NewError(ErrCodePromotionCpnsAdmissionFieldInvalid, Errs[ErrCodePromotionCpnsAdmissionFieldInvalid], errors.New("unor_id is required"))
	}

	_, err = request.AdmissionDate.Time()
	if err != nil {
		return ec.NewError(ErrCodePromotionCpnsAdmissionFieldInvalid, Errs[ErrCodePromotionCpnsAdmissionFieldInvalid], fmt.Errorf("tanggal_usulan cannot be parsed: %w", err))
	}

	if request.PakLetter == nil {
		return ec.NewError(ErrCodePromotionCpnsAdmissionFieldInvalid, Errs[ErrCodePromotionCpnsAdmissionFieldInvalid], errors.New("surat_pak is required"))
	}

	if request.PakLetter.Filename == "" {
		return ec.NewError(ErrCodePromotionCpnsAdmissionFieldInvalid, Errs[ErrCodePromotionCpnsAdmissionFieldInvalid], errors.New("nama_file in surat_pak is required"))
	}

	if request.PakLetter.DocumentNumber == "" {
		return ec.NewError(ErrCodePromotionCpnsAdmissionFieldInvalid, Errs[ErrCodePromotionCpnsAdmissionFieldInvalid], errors.New("no_dokumen in surat_pak is required"))
	}

	_, err = request.PakLetter.DocumentDate.Time()
	if err != nil {
		return ec.NewError(ErrCodePromotionCpnsAdmissionFieldInvalid, Errs[ErrCodePromotionCpnsAdmissionFieldInvalid], fmt.Errorf("tgl_dokumen in surat_pak cannot be parsed: %w", err))
	}

	if request.PromotionLetter == nil {
		return ec.NewError(ErrCodePromotionCpnsAdmissionFieldInvalid, Errs[ErrCodePromotionCpnsAdmissionFieldInvalid], errors.New("surat_pengangkatan is required"))
	}

	if request.PromotionLetter.DocumentNumber == "" {
		return ec.NewError(ErrCodePromotionCpnsAdmissionFieldInvalid, Errs[ErrCodePromotionCpnsAdmissionFieldInvalid], errors.New("no_dokumen in surat_pengangkatan is required"))
	}

	_, err = request.PromotionLetter.DocumentDate.Time()
	if err != nil {
		return ec.NewError(ErrCodePromotionCpnsAdmissionFieldInvalid, Errs[ErrCodePromotionCpnsAdmissionFieldInvalid], fmt.Errorf("tgl_dokumen in surat_pengangkatan cannot be parsed: %w", err))
	}

	return nil
}

// SubmitPromotionCpnsAdmissionCtx creates a new CPNS promotion admission.
func (c *Client) SubmitPromotionCpnsAdmissionCtx(ctx context.Context, request *models.PromotionCpnsAdmission) (admissionId string, err error) {
	if err = c.checkPromotionCpnsAdmissionSubmitRequest(request); err != nil {
		return "", err
	}

	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return "", err
	}

	defer func() {
		c.completeMtx(mtx, err)
	}()

	admissionId = uuid.NewString()
	_, err = mtx.ExecContext(ctx,
		`insert into pengangkatan_cpns (
                               pengangkatan_cpns_id,
                               asn_id,
                               jabatan_fungsional_tujuan_id,
                               angka_kredit_pertama,
                               unor_id,
                               tgl_usulan,
                               no_usulan,
                               nama_doc_pak,
                               no_doc_pak,
                               tgl_doc_pak,
                               nama_doc_surat_pengangkatan,
                               no_doc_surat_pengangkatan,
                               tgl_doc_surat_pengangkatan,
                               status,
                               status_ts,
                               status_by)
                               values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, current_timestamp, $15)`,
		admissionId,
		request.AsnId,
		request.PromotionPositionId,
		request.FirstCreditNumber,
		request.OrganizationUnitId,
		request.AdmissionDate,
		request.AdmissionNumber,
		request.PakLetter.Filename,
		request.PakLetter.DocumentNumber,
		request.PakLetter.DocumentDate,
		request.PromotionLetter.Filename,
		request.PromotionLetter.DocumentNumber,
		request.PromotionLetter.DocumentDate,
		models.PromotionCpnsAdmissionStatusCreated,
		request.SubmitterAsnId,
	)
	if err != nil {
		return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], err)
	}

	tempFilenamePak := path.Join(PromotionCpnsPakLetterSubdir, request.PakLetter.Filename)
	tempFilenamePromotionLetter := path.Join(PromotionCpnsPromotionLetterSubdir, request.PromotionLetter.Filename)
	filenamePak := path.Join(PromotionCpnsPakLetterSubdir, fmt.Sprintf("%s.pdf", admissionId))
	filenamePromotionLetter := path.Join(PromotionCpnsPromotionLetterSubdir, fmt.Sprintf("%s.pdf", admissionId))

	_, err = c.PromotionCpnsStorage.SavePromotionCpnsFile(ctx, tempFilenamePak, filenamePak, false)
	if err != nil {
		if err == object.ErrFileNotFound {
			return "", ErrStorageFileNotFound
		}
		return "", ec.NewError(ErrCodeStorageCopyFail, Errs[ErrCodeStorageCopyFail], err)
	}

	_, err = c.PromotionCpnsStorage.SavePromotionCpnsFile(ctx, tempFilenamePromotionLetter, filenamePromotionLetter, false)
	if err != nil {
		if err == object.ErrFileNotFound {
			return "", ErrStorageFileNotFound
		}
		return "", ec.NewError(ErrCodeStorageCopyFail, Errs[ErrCodeStorageCopyFail], err)
	}

	_ = c.PromotionCpnsStorage.DeletePromotionCpnsFilesTemp(ctx, []string{tempFilenamePak, tempFilenamePromotionLetter})

	return admissionId, nil
}

// GetPromotionCpnsStatusStatisticCtx returns the number of cpns promotion items for each status.
func (c *Client) GetPromotionCpnsStatusStatisticCtx(ctx context.Context) (statistics []*models.StatisticStatus, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)
	rows, err := mdb.QueryContext(ctx, "select status, jumlah from pengangkatan_cpns_status_statistik")
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer rows.Close()

	statistics = make([]*models.StatisticStatus, 0)
	statisticMap := make(map[int]*models.StatisticStatus)
	for status := range models.PromotionCpnsAdmissionStatuses {
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
