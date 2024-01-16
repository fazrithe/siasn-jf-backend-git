package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
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

// CreateCoverLetterFilename creates cover letter filename from given requirement ID.
// The filename includes the directory path. The extension is always .pdf.
func (c *Client) CreateCoverLetterFilename(requirementId string) (filename string) {
	return path.Join(RequirementCoverLetterSubdir, fmt.Sprintf("%s.pdf", requirementId))
}

// CheckRequirementAdmissionInsert checks requirement admission request object for validity, before inserting the entry
// to the database.
func (c *Client) CheckRequirementAdmissionInsert(ctx context.Context, request *models.RequirementAdmission) (err error) {
	if request.PositionGrade == "" {
		return ec.NewErrorBasic(ErrCodeRequirementAdmissionPositionGradeEmpty, Errs[ErrCodeRequirementAdmissionPositionGradeEmpty])
	}

	if request.RequirementCounts == nil || len(request.RequirementCounts) <= 0 {
		return ec.NewError(ErrCodeRequirementAdmissionRequirementCountInvalid, Errs[ErrCodeRequirementAdmissionRequirementCountInvalid], errors.New("requirement counts cannot be empty or nil"))
	}

	for _, rc := range request.RequirementCounts {
		if rc.OrganizationUnitId == "" {
			return ec.NewError(ErrCodeRequirementAdmissionRequirementCountInvalid, Errs[ErrCodeRequirementAdmissionRequirementCountInvalid], errors.New("unit_organisasi_id cannot be empty"))
		}

		if rc.Count <= 0 {
			return ec.NewError(ErrCodeRequirementAdmissionRequirementCountInvalid, Errs[ErrCodeRequirementAdmissionRequirementCountInvalid], errors.New("jumlah kebutuhan must be > 0"))
		}
	}

	if request.TempCoverLetter == nil || request.TempCoverLetter.DocumentName == "" || request.TempCoverLetter.Filename == "" {
		return ec.NewErrorBasic(ErrCodeRequirementAdmissionCoverLetterInvalid, Errs[ErrCodeRequirementAdmissionCoverLetterInvalid])
	}

	if len(request.TempEstimationDocuments) <= 0 {
		return ec.NewErrorBasic(ErrCodeRequirementAdmissionNoEstimationDocs, Errs[ErrCodeRequirementAdmissionNoEstimationDocs])
	}

	if request.FiscalYear == "" {
		return ec.NewErrorBasic(ErrCodeRequirementAdmissionNoFiscalYear, Errs[ErrCodeRequirementAdmissionNoFiscalYear])
	}

	if request.AdmissionNumber == "" {
		return ec.NewErrorBasic(ErrCodeRequirementAdmissionNoAdmissionNumber, Errs[ErrCodeRequirementAdmissionNoAdmissionNumber])
	}

	referenceMdb := metricutil.NewDB(c.ReferenceDb, c.SqlMetrics)

	unorIdMap := make(map[string]struct{})
	unorIds := make([]string, 0)
	for _, rc := range request.RequirementCounts {
		unorIds = append(unorIds, rc.OrganizationUnitId)
		unorIdMap[rc.OrganizationUnitId] = struct{}{}
	}

	unorRows, err := referenceMdb.QueryContext(ctx, "select id from unor where id = ANY($1) and instansi_id = $2", pq.Array(unorIds), request.AgencyId)
	if err != nil {
		return ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve unor: %w", err))
	}
	defer unorRows.Close()

	retrievedUnorIdMap := make(map[string]struct{})
	notFoundUnorIds := make([]string, 0)
	for unorRows.Next() {
		unorId := ""
		err = unorRows.Scan(&unorId)
		if err != nil {
			return ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve unor: %w", err))
		}

		retrievedUnorIdMap[unorId] = struct{}{}
	}
	if err = unorRows.Err(); err != nil {
		return ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve unor: %w", err))
	}

	for uid, _ := range unorIdMap {
		if _, ok := retrievedUnorIdMap[uid]; !ok {
			notFoundUnorIds = append(notFoundUnorIds, uid)
		}
	}

	if len(notFoundUnorIds) > 0 {
		e := ec.NewError(ErrCodeRequirementAdmissionRequirementCountInvalid, Errs[ErrCodeRequirementAdmissionRequirementCountInvalid], errors.New("some unor ID cannot be found"))
		e.Data = map[string]interface{}{
			"invalid_unit_organisasi_id": notFoundUnorIds,
		}
		return e
	}

	return nil
}

// CheckRequirementAdmissionEdit checks requirement admission request object for validity, before replacing the entry
// to the database.
func (c *Client) CheckRequirementAdmissionEdit(ctx context.Context, request *models.RequirementAdmission) (err error) {
	if request.PositionGrade == "" {
		return ec.NewErrorBasic(ErrCodeRequirementAdmissionPositionGradeEmpty, Errs[ErrCodeRequirementAdmissionPositionGradeEmpty])
	}

	if request.RequirementCounts == nil || len(request.RequirementCounts) <= 0 {
		return ec.NewError(ErrCodeRequirementAdmissionRequirementCountInvalid, Errs[ErrCodeRequirementAdmissionRequirementCountInvalid], errors.New("requirement counts cannot be empty or nil"))
	}

	for _, rc := range request.RequirementCounts {
		if rc.OrganizationUnitId == "" {
			return ec.NewError(ErrCodeRequirementAdmissionRequirementCountInvalid, Errs[ErrCodeRequirementAdmissionRequirementCountInvalid], errors.New("unit_organisasi_id cannot be empty"))
		}

		if rc.Count <= 0 {
			return ec.NewError(ErrCodeRequirementAdmissionRequirementCountInvalid, Errs[ErrCodeRequirementAdmissionRequirementCountInvalid], errors.New("jumlah kebutuhan must be > 0"))
		}
	}

	if request.FiscalYear == "" {
		return ec.NewErrorBasic(ErrCodeRequirementAdmissionNoFiscalYear, Errs[ErrCodeRequirementAdmissionNoFiscalYear])
	}

	if request.AdmissionNumber == "" {
		return ec.NewErrorBasic(ErrCodeRequirementAdmissionNoAdmissionNumber, Errs[ErrCodeRequirementAdmissionNoAdmissionNumber])
	}

	referenceMdb := metricutil.NewDB(c.ReferenceDb, c.SqlMetrics)

	unorIdMap := make(map[string]struct{})
	unorIds := make([]string, 0)
	for _, rc := range request.RequirementCounts {
		unorIds = append(unorIds, rc.OrganizationUnitId)
		unorIdMap[rc.OrganizationUnitId] = struct{}{}
	}

	unorRows, err := referenceMdb.QueryContext(ctx, "select id from unor where id = ANY($1) and instansi_id = $2", pq.Array(unorIds), request.AgencyId)
	if err != nil {
		return ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve unor: %w", err))
	}
	defer unorRows.Close()

	retrievedUnorIdMap := make(map[string]struct{})
	notFoundUnorIds := make([]string, 0)
	for unorRows.Next() {
		unorId := ""
		err = unorRows.Scan(&unorId)
		if err != nil {
			return ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve unor: %w", err))
		}

		retrievedUnorIdMap[unorId] = struct{}{}
	}
	if err = unorRows.Err(); err != nil {
		return ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve unor: %w", err))
	}

	for uid, _ := range unorIdMap {
		if _, ok := retrievedUnorIdMap[uid]; !ok {
			notFoundUnorIds = append(notFoundUnorIds, uid)
		}
	}

	if len(notFoundUnorIds) > 0 {
		e := ec.NewError(ErrCodeRequirementAdmissionRequirementCountInvalid, Errs[ErrCodeRequirementAdmissionRequirementCountInvalid], errors.New("some unor ID cannot be found"))
		e.Data = map[string]interface{}{
			"invalid_unit_organisasi_id": notFoundUnorIds,
		}
		return e
	}

	return nil
}

// bulkRetrieveBezetting bulks calculate bezetting for all organization IDs.
func (c *Client) bulkRetrieveBezetting(ctx context.Context, positionId string, organizationUnitIds []string) (results []*bezettingResult, err error) {
	mdbProfile := metricutil.NewDB(c.ProfileDb, c.SqlMetrics)

	rows, err := mdbProfile.QueryContext(ctx, "select jabatan_fungsional_id, unor_id, count(*) from pns where jabatan_fungsional_id = $1 and unor_id = ANY($2) group by jabatan_fungsional_id, unor_id", positionId, pq.Array(organizationUnitIds))
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve bezetting count: %w", err))
	}
	defer rows.Close()

	results = make([]*bezettingResult, 0)
	for rows.Next() {
		result := &bezettingResult{}
		err = rows.Scan(&result.PositionId, &result.OrganizationUnitId, &result.BezettingCount)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve bezetting count: %w", err))
		}
		results = append(results, result)
	}
	if err = rows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve bezetting count: %w", err))
	}

	return results, nil
}

// InsertRequirementAdmissionCtx inserts a new requirement admission entry.
// This will also move estimation documents and cover letter to permanent location in object storage.
func (c *Client) InsertRequirementAdmissionCtx(ctx context.Context, request *models.RequirementAdmission) (requirementId string, err error) {
	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return "", err
	}
	defer func() {
		c.completeMtx(mtx, err)
	}()

	err = c.CheckRequirementAdmissionInsert(ctx, request)
	if err != nil {
		return "", err
	}

	requirementId = uuid.NewString()

	_, err = mtx.ExecContext(
		ctx,
		"insert into kebutuhan(kebutuhan_id, tgl_usulan, status, jabatan_fungsional, filename_sp, nama_doc_sp, instansi_id, tahun_anggaran, no_usulan) values($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		requirementId,
		time.Time(request.AdmissionTimestamp),
		models.RequirementAdmissionStatusCreated,
		request.PositionGrade,
		path.Base(c.CreateCoverLetterFilename(requirementId)),
		request.TempCoverLetter.DocumentName,
		request.AgencyId,
		request.FiscalYear,
		request.AdmissionNumber,
	)
	if err != nil {
		return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to kebutuhan: %w", err))
	}

	// Collect all unor_id to retrieve bezetting.
	organizationUnitIds := make([]string, 0)
	for _, rc := range request.RequirementCounts {
		organizationUnitIds = append(organizationUnitIds, rc.OrganizationUnitId)
	}

	bezettingCounts, err := c.bulkRetrieveBezetting(ctx, request.PositionGrade, organizationUnitIds)
	if err != nil {
		return "", err
	}

	// Index bezetting result
	bezettingCountMap := make(map[string]*bezettingResult)
	for _, br := range bezettingCounts {
		bezettingCountMap[br.OrganizationUnitId] = br
	}

	reqCountStmt, err := mtx.PrepareContext(ctx, "insert into jumlah_kebutuhan(kebutuhan_id, unor_id, jlh_kebutuhan, bezetting_jlh_kebutuhan) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return "", ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], fmt.Errorf("cannot prepare insert entry to jumlah_kebutuhan: %w", err))
	}

	for _, rc := range request.RequirementCounts {
		bzCount := 0
		bez, ok := bezettingCountMap[rc.OrganizationUnitId]
		if ok {
			bzCount = bez.BezettingCount
		}
		_, err = reqCountStmt.ExecContext(ctx, requirementId, rc.OrganizationUnitId, rc.Count, bzCount)
		if err != nil {
			return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to jumlah_kebutuhan: %w", err))
		}
	}

	_, err = mtx.ExecContext(
		ctx,
		"insert into kebutuhan_statushist(kebutuhan_kebutuhan_id, status, modified_at_ts, user_id) values($1, $2, $3, $4)",
		requirementId,
		models.RequirementAdmissionStatusCreated,
		time.Now(),
		request.SubmitterAsnId,
	)
	if err != nil {
		return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to kebutuhan_statushist: %w", err))
	}

	filenameToDocName := make(map[string]string)
	var filenames []string
	for _, doc := range request.TempEstimationDocuments {
		filenames = append(filenames, path.Join(RequirementEstimationDocSubdir, doc.Filename))
		filenameToDocName[doc.Filename] = doc.DocumentName
	}
	estimationResults, err := c.RequirementStorage.SaveRequirementFiles(ctx, filenames)
	if err != nil {
		if err == object.ErrTempFileNotFound {
			return "", ErrStorageFileNotFound
		}
		return "", ec.NewError(ErrCodeStorageCopyFail, Errs[ErrCodeStorageCopyFail], fmt.Errorf("cannot save requirement estimation docs: %w", err))
	}

	stmt, err := mtx.PrepareContext(ctx, "insert into doc_perhitungan(kebutuhan_kebutuhan_id, filename, nama_doc_perhitungan, createdat) values($1, $2, $3, $4)")
	if err != nil {
		return "", ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], fmt.Errorf("cannot prepare statement for doc_perhitungan table: %w", err))
	}

	for _, result := range estimationResults {
		// Strip the directory from the filename and store only the basename
		basename := path.Base(result.Filename)
		docName := filenameToDocName[basename]
		_, err = stmt.ExecContext(ctx, requirementId, basename, docName, result.CreatedAt)
		if err != nil {
			return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to doc_perhitungan: %w", err))
		}
	}

	_, err = c.RequirementStorage.SaveRequirementFile(
		ctx,
		path.Join(RequirementCoverLetterSubdir, request.TempCoverLetter.Filename),
		c.CreateCoverLetterFilename(requirementId),
	)
	if err != nil {
		return "", ec.NewError(ErrCodeStorageCopyFail, Errs[ErrCodeStorageCopyFail], fmt.Errorf("cannot save cover letter: %w", err))
	}

	return requirementId, err
}

// EditRequirementAdmissionCtx updates an existing admission.
//
// Its behaviour is similar to InsertRequirementAdmissionCtx. The difference is if no cover letter provided, it will
// not insert a new cover letter and the estimation documents provided will be added to the list of estimation documents
// the admission has.
func (c *Client) EditRequirementAdmissionCtx(ctx context.Context, newRequirement *models.RequirementAdmission) (err error) {
	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return err
	}
	defer func() {
		c.completeMtx(mtx, err)
	}()

	err = c.CheckRequirementAdmissionEdit(ctx, newRequirement)
	if err != nil {
		return err
	}

	coverLetterDocName := ""
	if newRequirement.TempCoverLetter != nil {
		coverLetterDocName = newRequirement.TempCoverLetter.DocumentName
	}

	// Update requirement admission and get the current admission status
	var currentAdmissionStatus int
	err = mtx.QueryRowContext(ctx,
		"update kebutuhan set jabatan_fungsional = $1, nama_doc_sp = case when $2 = '' then nama_doc_sp else $2 end, tahun_anggaran = $3, no_usulan = $4 where kebutuhan_id = $5 and instansi_id = $6 returning status",
		newRequirement.PositionGrade,
		coverLetterDocName,
		newRequirement.FiscalYear,
		newRequirement.AdmissionNumber,
		newRequirement.RequirementId,
		newRequirement.AgencyId,
	).Scan(&currentAdmissionStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			// This means no requirement admission is found with the provided requirement id and agency id
			return ec.NewErrorBasic(ErrCodeEntryNotFound, Errs[ErrCodeEntryNotFound])
		}
		return ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot update entry in kebutuhan: %w", err))
	}

	_, err = mtx.ExecContext(ctx, "delete from jumlah_kebutuhan where kebutuhan_id = $1", newRequirement.RequirementId)
	if err != nil {
		return ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot delete entries in jumlah_kebutuhan: %w", err))
	}

	// Collect all unor_id to retrieve bezetting.
	organizationUnitIds := make([]string, 0)
	for _, rc := range newRequirement.RequirementCounts {
		organizationUnitIds = append(organizationUnitIds, rc.OrganizationUnitId)
	}

	bezettingCounts, err := c.bulkRetrieveBezetting(ctx, newRequirement.PositionGrade, organizationUnitIds)
	if err != nil {
		return err
	}

	// Index bezetting result
	bezettingCountMap := make(map[string]*bezettingResult)
	for _, br := range bezettingCounts {
		bezettingCountMap[br.OrganizationUnitId] = br
	}

	reqCountStmt, err := mtx.PrepareContext(ctx, "insert into jumlah_kebutuhan(kebutuhan_id, unor_id, jlh_kebutuhan, bezetting_jlh_kebutuhan) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], fmt.Errorf("cannot prepare insert entry to jumlah_kebutuhan: %w", err))
	}

	for _, rc := range newRequirement.RequirementCounts {
		bzCount := 0
		bez, ok := bezettingCountMap[rc.OrganizationUnitId]
		if ok {
			bzCount = bez.BezettingCount
		}
		_, err = reqCountStmt.ExecContext(ctx, newRequirement.RequirementId, rc.OrganizationUnitId, rc.Count, bzCount)
		if err != nil {
			return ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to jumlah_kebutuhan: %w", err))
		}
	}

	// Add new entry to kebutuhan_statushist (audit history)
	_, err = mtx.ExecContext(
		ctx,
		"insert into kebutuhan_statushist(kebutuhan_kebutuhan_id, status, modified_at_ts, user_id) values($1, $2, $3, $4)",
		newRequirement.RequirementId,
		currentAdmissionStatus,
		time.Now(),
		newRequirement.SubmitterAsnId,
	)
	if err != nil {
		return ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to kebutuhan_statushist: %w", err))
	}

	// Save the temp cover letter if provided
	if newRequirement.TempCoverLetter != nil {
		_, err := c.RequirementStorage.SaveRequirementFile(
			ctx,
			path.Join(RequirementCoverLetterSubdir, newRequirement.TempCoverLetter.Filename),
			c.CreateCoverLetterFilename(newRequirement.RequirementId),
		)
		if err != nil {
			return ec.NewError(ErrCodeStorageCopyFail, Errs[ErrCodeStorageCopyFail], fmt.Errorf("cannot save cover letter: %w", err))
		}
	}

	// Return early if no new estimation documents to save
	if len(newRequirement.TempEstimationDocuments) == 0 {
		return
	}

	// Save new estimation documents
	filenameToDocName := make(map[string]string)
	var filenames []string
	for _, doc := range newRequirement.TempEstimationDocuments {
		filenames = append(filenames, path.Join(RequirementEstimationDocSubdir, doc.Filename))
		filenameToDocName[doc.Filename] = doc.DocumentName
	}
	estimationResults, err := c.RequirementStorage.SaveRequirementFiles(ctx, filenames)
	if err != nil {
		return ec.NewError(ErrCodeStorageCopyFail, Errs[ErrCodeStorageCopyFail], fmt.Errorf("cannot save requirement estimation docs: %w", err))
	}

	stmt, err := mtx.PrepareContext(ctx, "insert into doc_perhitungan(kebutuhan_kebutuhan_id, filename, nama_doc_perhitungan, createdat) values($1, $2, $3, $4)")
	if err != nil {
		return ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], fmt.Errorf("cannot prepare statement for doc_perhitungan table: %w", err))
	}

	for _, result := range estimationResults {
		// Strip the directory from the filename and store only the basename
		basename := path.Base(result.Filename)
		docName := filenameToDocName[basename]
		_, err = stmt.ExecContext(ctx, newRequirement.RequirementId, basename, docName, result.CreatedAt)
		if err != nil {
			return ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to doc_perhitungan: %w", err))
		}
	}

	return
}

// Deprecated: use the paginated version.
// SearchRequirementAdmissionsCtx searches for the list of requirement admissions of a particular work agency ID (instansi kerja).
// It will return empty slice if no requirements are found.
func (c *Client) SearchRequirementAdmissionsCtx(ctx context.Context, searchFilter *RequirementAdmissionSearchFilter) (admissions []*models.RequirementAdmissionResult, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)

	queryBuilder := strings.Builder{}
	var queryArgs []interface{}

	queryBuilder.WriteString("select kebutuhan_id, tgl_usulan, status, jabatan_fungsional from kebutuhan where instansi_id = $")
	queryArgs = append(queryArgs, searchFilter.AgencyId)
	queryBuilder.WriteString(strconv.Itoa(len(queryArgs)))

	if searchFilter.AdmissionStatus != 0 {
		queryBuilder.WriteString(" and status = $")
		queryArgs = append(queryArgs, searchFilter.AdmissionStatus)
		queryBuilder.WriteString(strconv.Itoa(len(queryArgs)))
	}

	if searchFilter.AdmissionDate != "" {
		queryBuilder.WriteString(" and tgl_usulan::date = $")
		queryArgs = append(queryArgs, string(searchFilter.AdmissionDate))
		queryBuilder.WriteString(strconv.Itoa(len(queryArgs)))
	}

	query := queryBuilder.String()
	admissionRows, err := mdb.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer admissionRows.Close()

	admissions = []*models.RequirementAdmissionResult{}
	for admissionRows.Next() {
		admission := &models.RequirementAdmissionResult{}
		err = admissionRows.Scan(
			&admission.RequirementId,
			&admission.AdmissionTimestamp,
			&admission.Status,
			&admission.PositionGrade,
		)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
		admissions = append(admissions, admission)
	}
	if err = admissionRows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	return
}

// SearchRequirementAdmissionsPaginatedCtx searches for the list of requirement admissions of a particular work agency ID (instansi kerja).
// It will return empty slice if no requirements are found.
func (c *Client) SearchRequirementAdmissionsPaginatedCtx(ctx context.Context, filter *RequirementAdmissionSearchFilter) (result *search.PaginatedList[*models.RequirementAdmissionResult], err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)

	admissionRows, err := mdb.QueryContext(
		ctx,
		"select kebutuhan_id, tgl_usulan, status, jabatan_fungsional, no_usulan, tahun_anggaran from kebutuhan where instansi_id = $1 and ($2 = 0 or status = $2) and ($3::date is null or tgl_usulan::date = $3) order by tgl_usulan desc limit $4 offset $5",
		filter.AgencyId,
		filter.AdmissionStatus,
		sql.NullString{Valid: string(filter.AdmissionDate) != "", String: string(filter.AdmissionDate)},
		filter.CountPerPage+1,
		(filter.PageNumber-1)*filter.CountPerPage,
	)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer admissionRows.Close()

	admissions := make([]*models.RequirementAdmissionResult, 0)
	for admissionRows.Next() {
		admission := &models.RequirementAdmissionResult{}
		err = admissionRows.Scan(
			&admission.RequirementId,
			&admission.AdmissionTimestamp,
			&admission.Status,
			&admission.PositionGrade,
			&admission.AdmissionNumber,
			&admission.FiscalYear,
		)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
		admissions = append(admissions, admission)
	}
	if err = admissionRows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	hasNext := false
	if len(admissions) > filter.CountPerPage {
		hasNext = true
		admissions = admissions[:filter.CountPerPage]
	}

	return &search.PaginatedList[*models.RequirementAdmissionResult]{
		Data:     admissions,
		Metadata: search.CreatePaginatedListMetadataNoTotalNext(filter.PageNumber, len(admissions), hasNext),
	}, nil
}

// GetRequirementAdmissionDetailCtx retrieve the detail of a requirement admission with matched id.
func (c *Client) GetRequirementAdmissionDetailCtx(ctx context.Context, requirementId string, agencyId string) (admission *models.RequirementAdmissionDetail, err error) {
	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return nil, err
	}
	defer func() {
		c.completeMtx(mtx, err)
	}()

	referenceMdb := metricutil.NewDB(c.ReferenceDb, c.SqlMetrics)

	admission = &models.RequirementAdmissionDetail{
		CoverLetter:         &models.Document{},
		EstimationDocuments: []*models.Document{},
	}
	err = mtx.QueryRowContext(ctx, "select kebutuhan_id, tgl_usulan, status, jabatan_fungsional, filename_sp, nama_doc_sp, coalesce(catatan_sp, ''), tahun_anggaran, no_usulan, coalesce(alasan_perbaikan, '') from kebutuhan where kebutuhan_id = $1 and instansi_id = $2 for share", requirementId, agencyId).Scan(
		&admission.RequirementId,
		&admission.AdmissionTimestamp,
		&admission.Status,
		&admission.PositionGrade,
		&admission.CoverLetter.Filename,
		&admission.CoverLetter.DocumentName,
		&admission.CoverLetter.Note,
		&admission.FiscalYear,
		&admission.AdmissionNumber,
		&admission.RevisionReason,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrEntryNotFound
		}
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve requirement: %w", err))
	}

	admission.RequirementCounts = make([]*models.RequirementCount, 0)
	unorIds := make([]string, 0)

	reqCountRows, err := mtx.QueryContext(ctx, "select unor_id, jlh_kebutuhan, coalesce(rekomendasi_jlh_kebutuhan, 0), bezetting_jlh_kebutuhan from jumlah_kebutuhan where kebutuhan_id = $1 for share", requirementId)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve requirement count: %w", err))
	}
	defer reqCountRows.Close()

	for reqCountRows.Next() {
		reqCount := &models.RequirementCount{}
		err = reqCountRows.Scan(&reqCount.OrganizationUnitId, &reqCount.Count, &reqCount.CountRecommendation, &reqCount.CountBezetting)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve requirement count: %w", err))
		}
		admission.RequirementCounts = append(admission.RequirementCounts, reqCount)
		unorIds = append(unorIds, reqCount.OrganizationUnitId) // Duplicates are fine
	}
	if err = reqCountRows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve requirement count: %w", err))
	}

	unorMap := make(map[string]string)
	unorRows, err := referenceMdb.QueryContext(ctx, "select id, nama_unor from unor where id = ANY($1)", pq.Array(unorIds))
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve unor: %w", err))
	}
	defer unorRows.Close()

	for unorRows.Next() {
		unorId := ""
		unorName := ""
		err = unorRows.Scan(&unorId, &unorName)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve unor: %w", err))
		}

		unorMap[unorId] = unorName
	}
	if err = unorRows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve unor: %w", err))
	}

	for _, rc := range admission.RequirementCounts {
		rc.OrganizationUnit = unorMap[rc.OrganizationUnitId]
	}
	if admission.RevisionRequirementCounts != nil {
		for _, rc := range admission.RevisionRequirementCounts {
			rc.OrganizationUnit = unorMap[rc.OrganizationUnitId]
		}
	}

	documentRows, err := mtx.QueryContext(ctx, "select filename, nama_doc_perhitungan, coalesce(catatan, '') from doc_perhitungan where kebutuhan_kebutuhan_id = $1 for share", requirementId)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve estimation document: %w", err))
	}
	defer documentRows.Close()

	for documentRows.Next() {
		document := &models.Document{}
		err = documentRows.Scan(&document.Filename, &document.DocumentName, &document.Note)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve estimation document: %w", err))
		}
		admission.EstimationDocuments = append(admission.EstimationDocuments, document)
	}
	if err = documentRows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve estimation document: %w", err))
	}

	signedAt := sql.NullTime{}
	recommendationLetter := &models.Document{}
	err = mtx.QueryRowContext(ctx, "select filename, nosurat, tgl_surat, coalesce(ttd_user_id, ''), nama_doc, coalesce(catatan, ''), createdat, is_signed, signedat from surat_rekomendasi_kebutuhan where kebutuhan_id = $1", requirementId).
		Scan(
			&recommendationLetter.Filename,
			&recommendationLetter.DocumentNumber,
			&recommendationLetter.DocumentDate,
			&recommendationLetter.SignerId,
			&recommendationLetter.DocumentName,
			&recommendationLetter.Note,
			(*time.Time)(&recommendationLetter.CreatedAt),
			&recommendationLetter.IsSigned,
			&signedAt,
		)
	if err != nil && err != sql.ErrNoRows {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve recommendation letter: %w", err))
	}

	if err != sql.ErrNoRows {
		if signedAt.Valid {
			recommendationLetter.SignedAt = models.EpochTime(signedAt.Time)
		}
		admission.RecommendationLetter = recommendationLetter
	}

	return admission, nil
}

// SetRequirementStatusAcceptedCtx sets a requirement admission status as models.RequirementAdmissionStatusAccepted.
// If the status of the requirement is not models.RequirementAdmissionStatusCreated or models.RequirementAdmissionStatusRevision, this will return error code errnum.ErrCodeRequirementVerificationStatusProcessedFurther.
// If the status is already models.RequirementAdmissionStatusAccepted, this will return error code errnum.ErrCodeRequirementVerificationStatusAlreadyAccepted.
func (c *Client) SetRequirementStatusAcceptedCtx(ctx context.Context, request *models.RequirementVerificationRequest) (modifiedAt time.Time, err error) {
	requirementId, err := uuid.Parse(request.RequirementId)
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
	err = mtx.QueryRowContext(ctx, "select status from kebutuhan where kebutuhan_id = $1 and instansi_id = $2 for update", requirementId, request.AgencyId).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, ErrEntryNotFound
		}

		return time.Time{}, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	if currentStatus == models.RequirementAdmissionStatusAccepted {
		return time.Time{}, ec.NewErrorBasic(ErrCodeRequirementVerificationStatusAlreadyAccepted, Errs[ErrCodeRequirementVerificationStatusAlreadyAccepted])
	}

	if currentStatus != models.RequirementAdmissionStatusCreated && currentStatus != models.RequirementAdmissionStatusRevision {
		return time.Time{}, ec.NewErrorBasic(ErrCodeRequirementVerificationStatusProcessedFurther, Errs[ErrCodeRequirementVerificationStatusProcessedFurther])
	}

	_, err = mtx.ExecContext(
		ctx,
		"update kebutuhan set status = $1, alasan_perbaikan = NULL, catatan_sp = $2 where kebutuhan_id = $3 and instansi_id = $4",
		models.RequirementAdmissionStatusAccepted,
		sql.NullString{Valid: request.CoverLetterNote != "", String: request.CoverLetterNote},
		requirementId.String(),
		request.AgencyId,
	)
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], err)
	}

	for _, doc := range request.EstimationDocumentNotes {
		d := 0
		err = mtx.QueryRowContext(ctx, "update doc_perhitungan set catatan = $1 where filename = $2 and kebutuhan_kebutuhan_id = $3 returning 1",
			sql.NullString{Valid: doc.Note != "", String: doc.Note},
			doc.Filename,
			requirementId,
		).Scan(&d)
		if err != nil {
			e := ErrEntryNotFound
			e.Data = map[string]string{
				"invalid_dokumen_perhitungan": doc.Filename,
			}
			return time.Time{}, e
		}
	}

	countStmt, err := mtx.PrepareContext(ctx, "update jumlah_kebutuhan set rekomendasi_jlh_kebutuhan = $1 where kebutuhan_id = $2 and unor_id = $3")
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], err)
	}
	defer countStmt.Close()

	for _, count := range request.RequirementCounts {
		_, err = countStmt.ExecContext(ctx, count.CountRecommendation, requirementId, count.OrganizationUnitId)
		if err != nil {
			return time.Time{}, ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("unable to update count recommendation: %w", err))
		}
	}

	modifiedAt = time.Now()
	_, err = mtx.ExecContext(
		ctx,
		"insert into kebutuhan_statushist(kebutuhan_kebutuhan_id, status, modified_at_ts, user_id) values($1, $2, $3, $4)",
		requirementId.String(),
		models.RequirementAdmissionStatusAccepted,
		modifiedAt,
		request.SubmitterAsnId,
	)
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], err)
	}
	return modifiedAt, nil
}

// SetRequirementStatusRevisionCtx sets a requirement admission status as models.RequirementAdmissionStatusDenied.
// If the status of the requirement is not models.RequirementAdmissionStatusCreated, models.RequirementAdmissionStatusDenied
// or models.RequirementAdmissionStatusRevision, this will return error code errnum.ErrCodeRequirementVerificationStatusProcessedFurther.
// If the status is already models.RequirementAdmissionStatusAccepted, this will return error code errnum.ErrCodeRequirementVerificationStatusAlreadyAccepted.
func (c *Client) SetRequirementStatusRevisionCtx(ctx context.Context, request *models.RequirementRevisionRequest) (modifiedAt time.Time, err error) {
	requirementId, err := uuid.Parse(request.RequirementId)
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
	err = mtx.QueryRowContext(ctx, "select status from kebutuhan where kebutuhan_id = $1 and instansi_id = $2 for update", requirementId, request.AgencyId).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, ErrEntryNotFound
		}

		return time.Time{}, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	if currentStatus == models.RequirementAdmissionStatusAccepted {
		return time.Time{}, ec.NewErrorBasic(ErrCodeRequirementVerificationStatusAlreadyAccepted, Errs[ErrCodeRequirementVerificationStatusAlreadyAccepted])
	}

	if currentStatus != models.RequirementAdmissionStatusCreated && currentStatus != models.RequirementAdmissionStatusRevision && currentStatus != models.RequirementAdmissionStatusDenied {
		return time.Time{}, ec.NewErrorBasic(ErrCodeRequirementVerificationStatusProcessedFurther, Errs[ErrCodeRequirementVerificationStatusProcessedFurther])
	}

	_, err = mtx.ExecContext(
		ctx,
		"update kebutuhan set status = $1, alasan_perbaikan = $2, catatan_sp = $3 where kebutuhan_id = $4 and instansi_id = $5",
		models.RequirementAdmissionStatusRevision,
		request.RevisionReason,
		sql.NullString{Valid: request.CoverLetterNote != "", String: request.CoverLetterNote},
		requirementId.String(),
		request.AgencyId,
	)
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], err)
	}

	for _, doc := range request.EstimationDocumentNotes {
		d := 0
		err = mtx.QueryRowContext(ctx, "update doc_perhitungan set catatan = $1 where filename = $2 and kebutuhan_kebutuhan_id = $3 returning 1",
			sql.NullString{Valid: doc.Note != "", String: doc.Note},
			doc.Filename,
			requirementId,
		).Scan(&d)
		if err != nil {
			e := ErrEntryNotFound
			e.Data = map[string]string{
				"invalid_dokumen_perhitungan": doc.Filename,
			}
			return time.Time{}, e
		}
	}

	countStmt, err := mtx.PrepareContext(ctx, "update jumlah_kebutuhan set rekomendasi_jlh_kebutuhan = $1 where kebutuhan_id = $2 and unor_id = $3")
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], err)
	}
	defer countStmt.Close()

	for _, count := range request.RequirementCounts {
		_, err = countStmt.ExecContext(ctx, count.CountRecommendation, requirementId, count.OrganizationUnitId)
		if err != nil {
			return time.Time{}, ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("unable to update count recommendation: %w", err))
		}
	}

	modifiedAt = time.Now()
	_, err = mtx.ExecContext(
		ctx,
		"insert into kebutuhan_statushist(kebutuhan_kebutuhan_id, status, modified_at_ts, user_id) values($1, $2, $3, $4)",
		requirementId.String(),
		models.RequirementAdmissionStatusRevision,
		modifiedAt,
		request.SubmitterAsnId,
	)
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], err)
	}

	return modifiedAt, nil
}

// GetRequirementVerifiersCtx searches for the list of verifiers of a particular work agency ID (instansi kerja).
// It will return empty slice if no verifier are found.
func (c *Client) GetRequirementVerifiersCtx(ctx context.Context, workAgencyId string) (verifiers []*models.RequirementVerifier, err error) {
	mpdb := metricutil.NewDB(c.ProfileDb, c.SqlMetrics)
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)

	idRows, err := mdb.QueryContext(ctx, "select user_id from pegawai where role_peg = $1 and instansi = $2", models.StaffRoleSupervisor, workAgencyId)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer idRows.Close()
	userIds := make([]string, 0)
	for idRows.Next() {
		userId := ""
		err = idRows.Scan(&userId)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
		userIds = append(userIds, userId)
	}
	if err = idRows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	verifierRows, err := mpdb.QueryContext(ctx, "select id, nama from orang where id = any ($1)", pq.Array(userIds))
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer verifierRows.Close()
	verifiers = make([]*models.RequirementVerifier, 0)
	for verifierRows.Next() {
		verifier := &models.RequirementVerifier{}
		err = verifierRows.Scan(
			&verifier.AsnId,
			&verifier.RequirementVerifierName,
		)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
		verifiers = append(verifiers, verifier)
	}
	if err = verifierRows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	return
}

// BulkSubmitRequirementRecommendationLetterCtx submits a recommendation letter for multiple requirements at once.
// Set isPresigned to true if recommendation letter has been signed before uploading.
func (c *Client) BulkSubmitRequirementRecommendationLetterCtx(ctx context.Context, requirementIds []string, recommendationLetter *models.Document, submitterAsnId string) (filename string, err error) {
	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return "", err
	}
	referenceMtx, err := c.createMtxDb(ctx, c.ReferenceDb)
	if err != nil {
		return "", err
	}
	defer func() {
		c.completeMtx(referenceMtx, err)
	}()
	defer func() {
		c.completeMtx(mtx, err)
	}()

	rows, err := mtx.QueryContext(ctx, "select kebutuhan_id, instansi_id, jabatan_fungsional, status from kebutuhan where kebutuhan_id = ANY($1) for share", pq.Array(requirementIds))
	if err != nil {
		return "", ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot select kebutuhan: %w", err))
	}
	defer rows.Close()

	retrievedRequirementIds := make(map[string]*models.RequirementAdmissionDetail)
	functionalPositionIdMap := map[string]string{}
	agencyId := ""
	for rows.Next() {
		rd := &models.RequirementAdmissionDetail{}
		err = rows.Scan(&rd.RequirementId, &rd.AgencyId, &rd.PositionGrade, &rd.Status)
		if err != nil {
			return "", ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
		if agencyId == "" {
			agencyId = rd.AgencyId
		} else if agencyId != "" && agencyId != rd.AgencyId {
			return "", ec.NewErrorBasic(ErrCodeRequirementVerificationGenerateMustBeFromSameAgency, Errs[ErrCodeRequirementVerificationGenerateMustBeFromSameAgency])
		}

		retrievedRequirementIds[rd.RequirementId] = rd
		functionalPositionIdMap[rd.PositionGrade] = ""
	}
	if err = rows.Err(); err != nil {
		return "", ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	invalidStatusRequirementIds := make([]string, 0)
	notFoundRequirementIds := make([]string, 0)
	for _, r := range requirementIds {
		if req, ok := retrievedRequirementIds[r]; !ok {
			notFoundRequirementIds = append(notFoundRequirementIds, r)
		} else {
			if req.Status != models.RequirementAdmissionStatusAccepted {
				invalidStatusRequirementIds = append(invalidStatusRequirementIds, r)
			}
		}
	}

	if len(notFoundRequirementIds) > 0 {
		e := ec.NewErrorBasic(ErrCodeEntryNotFound, Errs[ErrCodeEntryNotFound])
		e.Data = map[string]interface{}{
			"invalid_kebutuhan_id": notFoundRequirementIds,
		}
		return "", e
	}

	if len(invalidStatusRequirementIds) > 0 {
		e := ec.NewErrorBasic(ErrCodeRequirementVerificationStatusNotAccepted, Errs[ErrCodeRequirementVerificationStatusNotAccepted])
		e.Data = map[string]interface{}{
			"invalid_kebutuhan_id": invalidStatusRequirementIds,
		}
		return "", e
	}

	requirementRows, err := mtx.QueryContext(ctx, "select kebutuhan_id, unor_id, jlh_kebutuhan, coalesce(rekomendasi_jlh_kebutuhan, 0), bezetting_jlh_kebutuhan from jumlah_kebutuhan where kebutuhan_id = ANY($1)", pq.Array(requirementIds))
	if err != nil {
		return "", ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve jumlah kebutuhan: %w", err))
	}
	defer requirementRows.Close()

	unorIdMap := map[string]string{}
	for requirementRows.Next() {
		ctReqId := ""
		ct := &models.RequirementCount{}
		err = requirementRows.Scan(&ctReqId, &ct.OrganizationUnitId, &ct.Count, &ct.CountRecommendation, &ct.CountBezetting)
		if err != nil {
			return "", ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve jumlah kebutuhan: %w", err))
		}

		r, ok := retrievedRequirementIds[ctReqId]
		if ok {
			unorIdMap[ct.OrganizationUnitId] = ""
			r.RequirementCounts = append(r.RequirementCounts, ct)
		}
	}

	// Retrieve unor names.
	unorIds := make([]string, 0)
	for unorId, _ := range unorIdMap {
		unorIds = append(unorIds, unorId)
	}

	unor, err := c.getOrganizationUnitNamesByAgencyId(ctx, referenceMtx, agencyId, unorIds)
	if err != nil {
		return "", err
	}

	// Retrieve JF names.
	functionalPositionIds := make([]string, 0)
	for functionalPositionId, _ := range functionalPositionIdMap {
		functionalPositionIds = append(functionalPositionIds, functionalPositionId)
	}

	// Retrieve agency name.
	agencies, err := c.getAgencyNames(ctx, referenceMtx, []string{agencyId})
	if err != nil {
		return "", err
	}

	positions, err := c.getFunctionalPositionNames(ctx, referenceMtx, functionalPositionIds)
	if err != nil {
		return "", err
	}

	for _, r := range retrievedRequirementIds {
		r.FunctionalPosition = positions[r.PositionGrade]
		if r.RequirementCounts != nil {
			for _, rc := range r.RequirementCounts {
				rc.OrganizationUnit = unor[rc.OrganizationUnitId]
			}
		}
	}

	staffRole := ""
	err = mtx.QueryRowContext(ctx, "select role_peg from pegawai where user_id = $1", recommendationLetter.SignerId).Scan(&staffRole)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ec.NewError(ErrCodeRoleUnauthorized, Errs[ErrCodeRoleUnauthorized], errors.New("signer user not found"))
		}
		return "", ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot select pegawai: %w", err))
	}

	if staffRole != models.StaffRoleSupervisor {
		return "", ec.NewError(ErrCodeRoleUnauthorized, Errs[ErrCodeRoleUnauthorized], fmt.Errorf("user %s cannot sign the document", recommendationLetter.SignerId))
	}

	_, err = mtx.ExecContext(ctx, "update kebutuhan set status = $1 where kebutuhan_id = ANY($2)", models.RequirementAdmissionStatusAcceptedWithRecommendation, pq.Array(requirementIds))
	if err != nil {
		return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot update kebutuhan status: %w", err))
	}

	stmt, err := mtx.PrepareContext(ctx, "insert into surat_rekomendasi_kebutuhan(filename, kebutuhan_id, tgl_surat, nama_doc, nosurat, catatan, ttd_user_id, is_signed, signedat) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) on conflict(kebutuhan_id) do update set filename = excluded.filename, tgl_surat = excluded.tgl_surat, nama_doc = excluded.nama_doc, nosurat = excluded.nosurat, catatan = excluded.catatan, ttd_user_id = excluded.ttd_user_id, is_signed = excluded.is_signed, signedat = excluded.signedat")
	if err != nil {
		return "", ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], fmt.Errorf("cannot prepare to insert to surat_rekomendasi_kebutuhan: %w", err))
	}

	statusStmt, err := mtx.PrepareContext(ctx, "insert into kebutuhan_statushist(status, modified_at_ts, kebutuhan_kebutuhan_id, user_id) values ($1, current_timestamp, $2, $3)")
	if err != nil {
		return "", ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], fmt.Errorf("cannot prepare to insert to kebutuhan_statushist: %w", err))
	}

	filename, _ = c.RequirementStorage.GenerateRequirementFilename("application/pdf")
	filename = path.Join(RequirementRecommendationLetterSubdir, filename)
	baseFilename := path.Base(filename)
	signedAt := sql.NullTime{}
	for _, rid := range requirementIds {
		_, err = stmt.ExecContext(
			ctx,
			baseFilename,
			rid,
			string(recommendationLetter.DocumentDate),
			recommendationLetter.DocumentName,
			recommendationLetter.DocumentNumber,
			sql.NullString{Valid: recommendationLetter.Note != "", String: recommendationLetter.Note},
			recommendationLetter.SignerId,
			false,
			signedAt,
		)
		if err != nil {
			return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to surat_rekomendasi_kebutuhan: %w", err))
		}

		_, err = statusStmt.ExecContext(ctx, models.RequirementAdmissionStatusAcceptedWithRecommendation, rid, submitterAsnId)
		if err != nil {
			return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to kebutuhan_statushist: %w", err))
		}
	}

	firstFunctionalPosition := ""
	if len(functionalPositionIds) > 0 {
		firstFunctionalPosition = positions[functionalPositionIds[0]]
	}
	fistAgencyName := agencies[agencyId] // Assume that all requirements have the same agency ID, so only 1 ID is in the map
	err = c.generateRequirementRecommendationLetterCtx(ctx, filename, c.createRequirementRecommendationLetterTemplate(recommendationLetter.DocumentNumber, string(recommendationLetter.DocumentDate), firstFunctionalPosition, fistAgencyName, retrievedRequirementIds))
	if err != nil {
		if errors.Is(err, docx.ErrSiasnRendererBadTemplate) {
			return "", ec.NewError(ErrCodeDocumentGenerateBadTemplate, Errs[ErrCodeDocumentGenerateBadTemplate], err)
		}
		return "", ec.NewError(ErrCodeDocumentGenerate, Errs[ErrCodeDocumentGenerate], err)
	}

	return baseFilename, nil
}

// createRequirementRecommendationLetterTemplate creates template data from a map of requirements (key must be requirement ID).
// firstFunctionalPosition and firstAgency both refers to the first requirement JF and agency.
func (c *Client) createRequirementRecommendationLetterTemplate(documentNumber, documentDate string, firstFunctionalPosition string, firstAgency string, requirements map[string]*models.RequirementAdmissionDetail) (data *RequirementRecommendationLetterTemplate) {
	data = &RequirementRecommendationLetterTemplate{
		DocumentNumber:   documentNumber,
		DocumentDate:     documentDate,
		Position:         firstFunctionalPosition,
		Agency:           firstAgency,
		CalculationCount: 0,
		RequirementCount: 0,
		Requirements:     nil,
	}

	entries := make([]*RequirementRecommendationLetterTemplateEntry, 0)
	totalCalculation := 0
	totalRequirement := 0
	for _, req := range requirements {
		ous := make([]*RequirementRecommendationLetterTemplateEntryUnor, 0)
		subtotalBezetting := 0
		subtotalCalculation := 0
		subtotalRequirement := 0
		subtotalRecommendation := 0
		for _, reqOu := range req.RequirementCounts {
			subtotalBezetting += reqOu.CountBezetting
			subtotalCalculation += reqOu.Count
			subtotalRecommendation += reqOu.CountRecommendation
			subtotalRequirement += reqOu.CountBezetting - reqOu.Count
			ous = append(ous, &RequirementRecommendationLetterTemplateEntryUnor{
				OrganizationUnit: reqOu.OrganizationUnit,
				Bezetting:        reqOu.CountBezetting,
				Calculation:      reqOu.Count,
				Requirement:      reqOu.CountBezetting - reqOu.Count,
				Recommendation:   reqOu.CountRecommendation,
			})
		}

		totalRequirement += subtotalRequirement
		totalCalculation += subtotalCalculation

		entries = append(entries, &RequirementRecommendationLetterTemplateEntry{
			FunctionalPosition:     req.FunctionalPosition,
			SubtotalBezetting:      subtotalBezetting,
			SubtotalCalculation:    subtotalCalculation,
			SubtotalRequirement:    subtotalRequirement,
			SubtotalRecommendation: subtotalRecommendation,
			OrganizationUnits:      ous,
		})
	}

	data.Requirements = entries
	data.CalculationCount = totalCalculation
	data.RequirementCount = totalRequirement

	return data
}

// SignRequirementRecommendationLetterCtx signs a requirement recommendation letter, async.
func (c *Client) SignRequirementRecommendationLetterCtx(ctx context.Context, filename string) (err error) {
	// TODO implement real signing
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)
	d := 0
	err = mdb.QueryRowContext(ctx, "update surat_rekomendasi_kebutuhan set is_signed = true, signedat = current_timestamp where filename = $1 returning 1", filename).Scan(&d)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrEntryNotFound
		}

		return ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	return nil
}

// GetRequirementStatusStatisticCtx returns the number of requirement items for each status.
func (c *Client) GetRequirementStatusStatisticCtx(ctx context.Context) (statistics []*models.StatisticStatus, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)
	rows, err := mdb.QueryContext(ctx, "select status, jumlah from kebutuhan_status_statistik")
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer rows.Close()

	statistics = make([]*models.StatisticStatus, 0)
	statisticMap := make(map[int]*models.StatisticStatus)
	for status := range models.RequirementAdmissionStatuses {
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
