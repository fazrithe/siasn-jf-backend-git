package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mime"
	"path"
	"time"

	. "github.com/fazrithe/siasn-jf-backend-git/errnum"
	"github.com/fazrithe/siasn-jf-backend-git/libs/docx"
	"github.com/fazrithe/siasn-jf-backend-git/libs/ec"
	"github.com/fazrithe/siasn-jf-backend-git/libs/metricutil"
	"github.com/fazrithe/siasn-jf-backend-git/libs/search"
	"github.com/fazrithe/siasn-jf-backend-git/store/models"
	"github.com/fazrithe/siasn-jf-backend-git/store/object"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// CheckDismissalAdmissionInsert verifies that dismissal request is valid.
func (c *Client) CheckDismissalAdmissionInsert(request *models.DismissalAdmission) (err error) {
	if len(request.TempSupportDocuments) == 0 {
		return ErrDismissalAdmissionNoSupportDocs
	}

	if request.DismissalReason == "" {
		return ErrDismissalAdmissionReasonEmpty
	}

	mandatoryDecreeReasons := map[string]struct{}{
		"2": {},
		"3": {},
		"4": {},
		"5": {},
	}
	_, isMandatory := mandatoryDecreeReasons[request.DismissalReason]
	if isMandatory && (request.DecreeNumber == "" || request.DecreeDate == "" || request.ReasonDetail == "") {
		return ErrDismissalDecreeDataEmpty
	}

	if request.AdmissionNumber == "" {
		return ErrDismissalAdmissionNumberInvalid
	}

	return nil
}

// InsertDismissalAdmissionCtx inserts a new dismissal admission request (pengajuan pemberhentian).
// It also moves filenames defined request.TempSupportDocuments from temporary storage to permanent storage in
// object storage.
func (c *Client) InsertDismissalAdmissionCtx(ctx context.Context, request *models.DismissalAdmission) (dismissalId string, err error) {
	if err = c.CheckDismissalAdmissionInsert(request); err != nil {
		return "", err
	}

	mdb := metricutil.NewDB(c.ProfileDb, c.SqlMetrics)
	d := 0
	err = mdb.QueryRowContext(ctx, "select 1 from pns where id = $1", request.AsnId).Scan(&d)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrDismissalAdmissionAsnNotFound
		}
		return "", ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return "", err
	}

	defer func() {
		c.completeMtx(mtx, err)
	}()

	// A list of decree reason which will make decree number and date (and reason detail) mandatory.
	mandatoryDecreeReasons := map[string]struct{}{
		"2": {},
		"3": {},
		"4": {},
		"5": {},
	}
	_, isMandatory := mandatoryDecreeReasons[request.DismissalReason]

	dismissalId = uuid.New().String()
	query := "insert into pemberhentian(uuid_pemberhentian, asn_id, instansi_id, status, status_ts, status_by, alasan_pemberhentian, tgl_pemberhentian, nomor_sk, tgl_sk, detail_alasan, no_usulan) values($1, $2, $3, $4, current_timestamp, $5, $6, date(current_timestamp), $7, $8::date, $9, $10)"
	_, err = mtx.ExecContext(
		ctx,
		query,
		dismissalId,
		request.AsnId,
		request.AgencyId,
		models.DismissalAdmissionStatusCreated,
		request.SubmitterAsnId,
		request.DismissalReason,
		sql.NullString{Valid: isMandatory, String: request.DecreeNumber},
		sql.NullString{Valid: isMandatory, String: string(request.DecreeDate)},
		sql.NullString{Valid: isMandatory, String: request.ReasonDetail},
		request.AdmissionNumber,
	)
	if err != nil {
		return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to pemberhentian: %w", err))
	}

	docStmt, err := mtx.PrepareContext(ctx, "insert into pemberhentian_doc_pendukung(uuid_pemberhentian, filename, nama_doc) values($1, $2, $3)")
	if err != nil {
		return "", ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], fmt.Errorf("cannot prepare statement for pemberhentian_doc_pendukung table: %w", err))
	}

	filenameToDocName := make(map[string]string)
	var filenames []string
	for _, doc := range request.TempSupportDocuments {
		filenames = append(filenames, path.Join(DismissalSupportDocSubdir, doc.Filename))
		filenameToDocName[doc.Filename] = doc.DocumentName
	}
	results, err := c.DismissalStorage.SaveDismissalFiles(ctx, filenames)
	if err != nil {
		if err == object.ErrTempFileNotFound {
			return "", ErrStorageFileNotFound
		}
		return "", ec.NewError(ErrCodeStorageCopyFail, Errs[ErrCodeStorageCopyFail], err)
	}

	for _, result := range results {
		// Strip the directory from the filename and store only the basename
		basename := path.Base(result.Filename)
		docName := filenameToDocName[basename]
		_, err = docStmt.ExecContext(ctx, dismissalId, basename, docName)
		if err != nil {
			return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to pemberhentian_doc_pendukung: %w", err))
		}
	}

	return dismissalId, nil
}

// GetDismissalDetailCtx retrieves dismissal detail, including its support documents.
func (c *Client) GetDismissalDetailCtx(ctx context.Context, dismissalId string, agencyId string) (dismissal *models.DismissalAdmission, err error) {
	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return nil, err
	}

	defer func() {
		c.completeMtx(mtx, err)
	}()

	dismissal = &models.DismissalAdmission{DismissalId: dismissalId, AgencyId: agencyId}
	acceptanceLetterDocName := sql.NullString{}
	acceptanceLetterDocNumber := sql.NullString{}
	acceptanceLetterDocDate := sql.NullString{}
	decreeDate := sql.NullString{}
	err = mtx.QueryRowContext(
		ctx,
		"select asn_id, status, status_ts, status_by, coalesce(alasan_pemberhentian, ''), coalesce(alasan_tidak_diberhentikan, ''), nama_doc_surat_pemberhentian, nosurat_surat_pemberhentian, coalesce(ttd_user_id_surat_pemberhentian, ''), tgl_surat_pemberhentian, tgl_pemberhentian, coalesce(nomor_sk, ''), to_char(tgl_sk, 'YYYY-MM-DD'), coalesce(detail_alasan, ''), no_usulan from pemberhentian where uuid_pemberhentian = $1 and instansi_id = $2 for share",
		dismissalId,
		agencyId,
	).Scan(
		&dismissal.AsnId,
		&dismissal.Status,
		(*time.Time)(&dismissal.StatusTs),
		&dismissal.StatusBy,
		&dismissal.DismissalReason,
		&dismissal.DismissalDenyReason,
		&acceptanceLetterDocName,
		&acceptanceLetterDocNumber,
		&dismissal.DismissalLetterSignerAsnId,
		&acceptanceLetterDocDate,
		&dismissal.DismissalDate,
		&dismissal.DecreeNumber,
		&decreeDate,
		&dismissal.ReasonDetail,
		&dismissal.AdmissionNumber,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrEntryNotFound
		}
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot query pemberhentian: %w", err))
	}

	if dismissal.Status == models.DismissalAdmissionStatusAccepted && acceptanceLetterDocName.Valid && acceptanceLetterDocNumber.Valid && acceptanceLetterDocDate.Valid {
		docDate, _ := models.ParseIso8601Date(acceptanceLetterDocDate.String)
		dismissal.DismissalLetter = &models.Document{
			DocumentName:   acceptanceLetterDocName.String,
			DocumentNumber: acceptanceLetterDocNumber.String,
			DocumentDate:   docDate,
		}
	}
	dismissal.DecreeDate = models.Iso8601Date(decreeDate.String)

	supportDocRows, err := mtx.QueryContext(ctx, "select filename, nama_doc from pemberhentian_doc_pendukung where uuid_pemberhentian = $1", dismissalId)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot query pemberhentian_doc_pendukung: %w", err))
	}
	defer supportDocRows.Close()

	for supportDocRows.Next() {
		supportDoc := &models.Document{}
		err = supportDocRows.Scan(&supportDoc.Filename, &supportDoc.DocumentName)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot query pemberhentian_doc_pendukung: %w", err))
		}
		dismissal.SupportDocuments = append(dismissal.SupportDocuments, supportDoc)
	}
	if err = supportDocRows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot query pemberhentian_doc_pendukung: %w", err))
	}

	if dismissal.Status == models.DismissalAdmissionStatusRejected {
		denySupportDocsRows, err := mtx.QueryContext(ctx, "select filename from pemberhentian_doc_pendukung_penolakan where uuid_pemberhentian = $1", dismissalId)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot query pemberhentian_doc_pendukung_penolakan: %w", err))
		}
		defer denySupportDocsRows.Close()

		for denySupportDocsRows.Next() {
			supportDoc := &models.Document{}
			err = supportDocRows.Scan(&supportDoc.Filename)
			if err != nil {
				return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot query pemberhentian_doc_pendukung_penolakan: %w", err))
			}
			dismissal.DismissalDenySupportDocuments = append(dismissal.DismissalDenySupportDocuments, supportDoc)
		}
		if err = denySupportDocsRows.Err(); err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot query pemberhentian_doc_pendukung_penolakan: %w", err))
		}
	}

	profileMdb := metricutil.NewDB(c.ProfileDb, c.SqlMetrics)
	asns, err := c.getAsnNipNames(ctx, profileMdb, []string{dismissal.AsnId})
	if err != nil {
		return nil, err
	}

	dismissal.AsnNip = asns[dismissal.AsnId].Nip
	dismissal.AsnName = asns[dismissal.AsnId].AsnName

	return dismissal, nil
}

// Deprecated: use the paginated version.
// SearchDismissalAdmissionsCtx searches for dismissal admissions based on admission date and status in an agency.
func (c *Client) SearchDismissalAdmissionsCtx(ctx context.Context, admissionDate models.Iso8601Date, status int, agencyId string) (dismissalAdmissions []*models.DismissalAdmission, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)
	rows, err := mdb.QueryContext(
		ctx,
		"select uuid_pemberhentian, asn_id, status, status_ts, status_by, coalesce(alasan_pemberhentian, ''), coalesce(alasan_tidak_diberhentikan, ''), tgl_pemberhentian from pemberhentian where ($1 = '' or tgl_pemberhentian = $1::date) and ($2 = 0 or status = $2) and instansi_id = $3",
		string(admissionDate),
		status,
		agencyId,
	)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot query pemberhentian: %w", err))
	}
	defer rows.Close()

	dismissals := make([]*models.DismissalAdmission, 0)
	for rows.Next() {
		dismissal := &models.DismissalAdmission{AgencyId: agencyId}
		err = rows.Scan(&dismissal.DismissalId, &dismissal.AsnId, &dismissal.Status, &dismissal.StatusTs, &dismissal.StatusBy, &dismissal.DismissalReason, &dismissal.DismissalDenyReason, &dismissal.DismissalDate)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot query pemberhentian: %w", err))
		}
		dismissals = append(dismissals, dismissal)
	}
	if err = rows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot query pemberhentian: %w", err))
	}

	return dismissals, nil
}

// SearchDismissalAdmissionsPaginatedCtx searches for dismissal admissions based on admission date and status in an agency.
func (c *Client) SearchDismissalAdmissionsPaginatedCtx(ctx context.Context, admissionDate models.Iso8601Date, status int, agencyId string, pageNumber, countPerPage int) (result *search.PaginatedList, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)
	rows, err := mdb.QueryContext(
		ctx,
		"select uuid_pemberhentian, asn_id, status, status_ts, status_by, coalesce(alasan_pemberhentian, ''), coalesce(alasan_tidak_diberhentikan, ''), tgl_pemberhentian, coalesce(nomor_sk, ''), to_char(tgl_sk, 'YYYY-MM-DD'), coalesce(detail_alasan, ''), no_usulan from pemberhentian where ($1::date is null or tgl_pemberhentian = $1::date) and ($2 = 0 or status = $2) and instansi_id = $3 order by tgl_pemberhentian desc limit $4 offset $5",
		sql.NullString{Valid: string(admissionDate) != "", String: string(admissionDate)},
		status,
		agencyId,
		countPerPage+1,
		(pageNumber-1)*countPerPage,
	)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot query pemberhentian: %w", err))
	}
	defer rows.Close()

	dismissals := make([]*models.DismissalAdmission, 0)
	for rows.Next() {
		dismissal := &models.DismissalAdmission{AgencyId: agencyId}
		decreeDate := sql.NullString{}
		err = rows.Scan(
			&dismissal.DismissalId,
			&dismissal.AsnId,
			&dismissal.Status,
			&dismissal.StatusTs,
			&dismissal.StatusBy,
			&dismissal.DismissalReason,
			&dismissal.DismissalDenyReason,
			&dismissal.DismissalDate,
			&dismissal.DecreeNumber,
			&decreeDate,
			&dismissal.ReasonDetail,
			&dismissal.AdmissionNumber,
		)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot query pemberhentian: %w", err))
		}
		dismissal.DecreeDate = models.Iso8601Date(decreeDate.String)
		dismissals = append(dismissals, dismissal)
	}
	if err = rows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot query pemberhentian: %w", err))
	}

	hasNext := false
	if len(dismissals) > countPerPage {
		hasNext = true
		dismissals = dismissals[:countPerPage]
	}

	// We'll query about the attendees acceptance
	var asnIds []string
	for _, dismissal := range dismissals {
		asnIds = append(asnIds, dismissal.AsnId)
	}
	// Need to get the name of the attendees from the profile db
	// No need to query the names if no attendees found.
	asnNipMap := make(map[string]string)
	asnNameMap := make(map[string]string)
	if len(asnIds) > 0 {
		profileMdb := metricutil.NewDB(c.ProfileDb, c.SqlMetrics)
		profileRows, err := profileMdb.QueryContext(ctx, "select distinct on (orang.id) orang.id, nama, nip_baru from orang join (select id, nip_baru from pns where pns.id = ANY($1) union select id, nip_baru from pppk where pppk.id = ANY($1)) t on orang.id = t.id", pq.Array(asnIds))
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
		defer profileRows.Close()

		for profileRows.Next() {
			var tempId, tempName, tempNip string
			err = profileRows.Scan(&tempId, &tempName, &tempNip)
			if err != nil {
				return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
			}

			// We can guarantee the id exist in the map because we only search for the previously queried ids
			asnNipMap[tempId] = tempNip
			asnNameMap[tempId] = tempName
		}
		if err = profileRows.Err(); err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}

		// Ignore ASN that cannot be found
	}

	for _, d := range dismissals {
		d.AsnNip = asnNipMap[d.AsnId]
		d.AsnName = asnNameMap[d.AsnId]
	}

	return &search.PaginatedList{
		Data:     dismissals,
		Metadata: search.CreatePaginatedListMetadataNoTotalNext(pageNumber, len(dismissals), hasNext),
	}, nil
}

// IsDismissalExistCtx checks if dismissal entry exists.
func (c *Client) IsDismissalExistCtx(ctx context.Context, dismissalId string) (isExist bool, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)

	d := 0
	err = mdb.QueryRowContext(ctx, "select 1 from pemberhentian where uuid_pemberhentian = $1", dismissalId).Scan(&d)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	return true, nil
}

// SetDismissalStatusAcceptedCtx sets the dismissal status to accepted.
// It requires dismissal acceptance letter to be uploaded first.
func (c *Client) SetDismissalStatusAcceptedCtx(ctx context.Context, request *models.DismissalAcceptanceRequest) (modifiedAt time.Time, err error) {
	_, err = uuid.Parse(request.DismissalId)
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodeUuidInvalid, Errs[ErrCodeUuidInvalid], err)
	}

	if request.DismissalLetter == nil || request.DismissalLetter.DocumentNumber == "" || request.DismissalLetter.DocumentDate == "" {
		return time.Time{}, ErrDismissalAcceptanceNoLetter
	}

	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return time.Time{}, err
	}

	defer func() {
		c.completeMtx(mtx, err)
	}()

	d := 0
	err = mtx.QueryRowContext(ctx, "select 1 from pegawai where user_id = $1 and role_peg = $2 and instansi = $3", request.DismissalLetterSignerAsnId, models.StaffRoleSupervisor, request.AgencyId).Scan(&d)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, ErrDismissalAcceptanceSignerNotFound
		}

		return time.Time{}, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot query pegawai: %w", err))
	}

	err = mtx.QueryRowContext(
		ctx,
		"update pemberhentian set status = $1, status_ts = current_timestamp, status_by = $2, nama_doc_surat_pemberhentian = $3, nosurat_surat_pemberhentian = $4, ttd_user_id_surat_pemberhentian = $5, tgl_surat_pemberhentian = $6 where uuid_pemberhentian = $7 and status = $8 returning status_ts",
		models.DismissalAdmissionStatusAccepted,
		request.SubmitterAsnId,
		request.DismissalLetter.DocumentName,
		request.DismissalLetter.DocumentNumber,
		request.DismissalLetterSignerAsnId,
		string(request.DismissalLetter.DocumentDate),
		request.DismissalId,
		models.DismissalAdmissionStatusCreated,
	).Scan(&modifiedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, ErrEntryNotFound
		}

		return time.Time{}, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot query pemberhentian: %w", err))
	}

	profileMdb := metricutil.NewDB(c.ProfileDb, c.SqlMetrics)
	referenceMdb := metricutil.NewDB(c.ReferenceDb, c.SqlMetrics)
	_, err = c.retrieveAndGenerateDismissalAcceptanceLetterCtx(ctx, mtx, profileMdb, referenceMdb, request.DismissalId, true)
	if err != nil {
		return time.Time{}, err
	}

	return modifiedAt, nil
}

// SetDismissalStatusDeniedCtx sets the dismissal status to denied.
// Denial support documents can be uploaded optionally beforehand and supplied in the request. Only document filenames
// are stored.
func (c *Client) SetDismissalStatusDeniedCtx(ctx context.Context, request *models.DismissalDenyRequest) (modifiedAt time.Time, err error) {
	dismissalId, err := uuid.Parse(request.DismissalId)
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodeUuidInvalid, Errs[ErrCodeUuidInvalid], err)
	}

	if request.DismissalDenyReason == "" {
		return time.Time{}, ErrDismissalDenialNoReason
	}

	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return time.Time{}, err
	}

	defer func() {
		c.completeMtx(mtx, err)
	}()

	err = mtx.QueryRowContext(
		ctx,
		"update pemberhentian set status = $1, status_ts = current_timestamp, status_by = $2, alasan_tidak_diberhentikan = $3 where uuid_pemberhentian = $4 and status = $5 returning status_ts",
		models.DismissalAdmissionStatusRejected,
		request.SubmitterAsnId,
		request.DismissalDenyReason,
		request.DismissalId,
		models.DismissalAdmissionStatusCreated,
	).Scan(&modifiedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, ErrEntryNotFound
		}

		return time.Time{}, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot query pemberhentian: %w", err))
	}

	if request.TempDismissalDenySupportDocuments != nil && len(request.TempDismissalDenySupportDocuments) > 0 {
		docStmt, err := mtx.PrepareContext(ctx, "insert into pemberhentian_doc_pendukung_penolakan(uuid_pemberhentian, filename) values($1, $2)")
		if err != nil {
			return time.Time{}, ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], fmt.Errorf("cannot prepare statement for pemberhentian_doc_pendukung_penolakan table: %w", err))
		}

		var filenames []string
		for _, doc := range request.TempDismissalDenySupportDocuments {
			filenames = append(filenames, path.Join(DismissalDenySupportDocSubdir, doc.Filename))
		}

		results, err := c.DismissalStorage.SaveDismissalFiles(ctx, filenames)
		if err != nil {
			if err == object.ErrTempFileNotFound {
				return time.Time{}, ErrStorageFileNotFound
			}
			return time.Time{}, ec.NewError(ErrCodeStorageCopyFail, Errs[ErrCodeStorageCopyFail], err)
		}

		for _, result := range results {
			// Strip the directory from the filename and store only the basename
			basename := path.Base(result.Filename)
			_, err = docStmt.ExecContext(ctx, dismissalId, basename)
			if err != nil {
				return time.Time{}, ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to pemberhentian_doc_pendukung_penolakan: %w", err))
			}
		}
	}

	return modifiedAt, nil
}

// GetDismissalStatusStatisticCtx returns the number of dismissal items for each status.
func (c *Client) GetDismissalStatusStatisticCtx(ctx context.Context) (statistics []*models.StatisticStatus, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)
	rows, err := mdb.QueryContext(ctx, "select status, jumlah from pemberhentian_status_statistik")
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer rows.Close()

	statistics = make([]*models.StatisticStatus, 0)
	statisticMap := make(map[int]*models.StatisticStatus)
	for status := range models.DismissalAdmissionStatuses {
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

// retrieveAndGenerateDismissalAcceptanceLetterCtx generates a dismissal letter and store it in the storage.
// Filename is generated from dismissalId.pdf. The base filename is returned,
// it can be accessed in the dismissal acceptance subdir in permanent bucket.
//
// To help reduce performance load, it is only generated once, unless forceRegenerate is set to true.
func (c *Client) retrieveAndGenerateDismissalAcceptanceLetterCtx(ctx context.Context, dh metricutil.DbHandler, profileDh metricutil.DbHandler, referenceDh metricutil.DbHandler, dismissalId string, forceRegenerate bool) (filename string, err error) {
	filename = fmt.Sprintf("%s.pdf", dismissalId)
	fullPath := path.Join(DismissalAcceptanceLetterSubdir, filename)

	if !forceRegenerate {
		fileFound := true
		meta, err := c.DismissalStorage.GetDismissalFileMetadata(ctx, fullPath)
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

	data := &DismissalAcceptanceTemplate{}
	asnId := ""
	status := 0
	decreeDate := sql.NullString{}
	err = dh.QueryRowContext(
		ctx,
		"select asn_id, status, coalesce(alasan_pemberhentian, ''), nosurat_surat_pemberhentian, to_char(tgl_surat_pemberhentian, 'YYYY-MM-DD'), to_char(tgl_pemberhentian, 'YYYY-MM-DD'), coalesce(nomor_sk, ''), to_char(tgl_sk, 'YYYY-MM-DD') from pemberhentian where uuid_pemberhentian = $1",
		dismissalId,
	).Scan(
		&asnId,
		&status,
		&data.DismissalReason,
		&data.DocumentNumber,
		&data.DocumentDate,
		&data.DismissalDate,
		&data.DecreeNumber,
		&decreeDate,
	)
	data.DecreeDate = decreeDate.String
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrEntryNotFound
		}
		return "", ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot query pemberhentian: %w", err))
	}

	detail, err := c.getUserDetailByAsnId(ctx, profileDh, referenceDh, asnId, "")
	if err != nil {
		return "", err
	}

	if detail == nil {
		return "", ec.NewError(ErrCodeEntryNotFound, Errs[ErrCodeEntryNotFound], errors.New("ASN cannot be found in profile database"))
	}

	data.Position = detail.Position
	data.OrganizationUnit = detail.OrganizationUnit
	data.AsnGrade = detail.Rank

	err = c.generateDismissalAcceptanceLetterCtx(ctx, fullPath, data)
	if err != nil {
		if errors.Is(err, docx.ErrSiasnRendererBadTemplate) {
			return "", ec.NewError(ErrCodeDocumentGenerateBadTemplate, Errs[ErrCodeDocumentGenerateBadTemplate], err)
		}
		return "", ec.NewError(ErrCodeDocumentGenerate, Errs[ErrCodeDocumentGenerate], err)
	}

	return filename, nil
}
