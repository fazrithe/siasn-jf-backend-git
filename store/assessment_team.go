package store

import (
	"context"
	"database/sql"
	"fmt"
	"path"
	"time"

	. "github.com/fazrithe/siasn-jf-backend-git/errnum"
	"github.com/fazrithe/siasn-jf-backend-git/libs/ec"
	"github.com/fazrithe/siasn-jf-backend-git/libs/metricutil"
	"github.com/fazrithe/siasn-jf-backend-git/libs/search"
	"github.com/fazrithe/siasn-jf-backend-git/store/models"
	"github.com/fazrithe/siasn-jf-backend-git/store/object"
	"github.com/google/uuid"
)

// CheckAssessmentTeamAdmissionInsert checks assessment team admission request object for validity, before inserting the entry
// to the database.
func (c *Client) CheckAssessmentTeamAdmissionInsert(request *models.AssessmentTeamAdmission) (err error) {
	if len(request.Assessors) < 3 {
		return ErrAssessmentTeamAdmissionAssessorCountInvalid
	}

	if len(request.Assessors)%2 == 0 {
		return ErrAssessmentTeamAdmissionAssessorCountEven
	}

	if request.FunctionalPositionId == "" {
		return ErrAssessmentTeamAdmissionFunctionalPositionIdInvalid
	}

	if request.AdmissionNumber == "" {
		return ErrAssessmentTeamAdmissionNumberInvalid
	}

	for _, assessor := range request.Assessors {
		if _, ok := models.AssessmentTeamAssessorRoles[assessor.Role]; !ok {
			return ErrAssessmentTeamAssessorRoleInvalid
		}
	}

	return nil
}

// InsertAssessmentTeamAdmissionCtx inserts a new admission request. It also moves filenames defined
// request.TempSupportDocuments from temporary storage to permanent storage in object storage.
func (c *Client) InsertAssessmentTeamAdmissionCtx(ctx context.Context, request *models.AssessmentTeamAdmission) (admissionId string, err error) {
	if err = c.CheckAssessmentTeamAdmissionInsert(request); err != nil {
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
		`insert into tim_penilaian (
                               tim_penilaian_id,
                               asn_id,
			                   instansi_id,
                               jabatan_fungsional_id,
                               tgl_usulan,
                               no_usulan,
                               status,
                               status_ts,
                               status_by)
                               values ($1, $2, $3, $4, $5, $6, $7, current_timestamp, $8)`,
		admissionId,
		request.SubmitterAsnId,
		request.AgencyId,
		request.FunctionalPositionId,
		request.AdmissionDate,
		request.AdmissionNumber,
		models.AssessmentTeamStatusCreated,
		request.SubmitterAsnId,
	)
	if err != nil {
		return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], err)
	}

	assessorStmt, err := mtx.PrepareContext(ctx, "insert into anggota_tim_penilaian(tim_penilaian_id, asn_id, peran) VALUES($1, $2, $3)")
	if err != nil {
		return "", ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], fmt.Errorf("cannot prepare statement for anggota_tim_penilaian table: %w", err))
	}

	for _, asn := range request.Assessors {
		_, err = assessorStmt.ExecContext(ctx, admissionId, asn.AsnId, asn.Role)
		if err != nil {
			return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to anggota_tim_penilaian: %w", err))
		}
	}

	docStmt, err := mtx.PrepareContext(ctx, "insert into dokumen_pendukung_tim_penilaian(tim_penilaian_id, filename, nama_doc, createdat) values($1, $2, $3, $4)")
	if err != nil {
		return "", ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], fmt.Errorf("cannot prepare statement for dokumen_pendukung_tim_penilaian table: %w", err))
	}

	filenameToDocName := make(map[string]string)
	var filenames []string
	for _, doc := range request.TempSupportDocuments {
		filenames = append(filenames, path.Join(AssessmentTeamSupportDocSubdir, doc.Filename))
		filenameToDocName[doc.Filename] = doc.DocumentName
	}
	results, err := c.AssessmentTeamStorage.SaveAssessmentTeamFiles(ctx, filenames)
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
		_, err = docStmt.ExecContext(ctx, admissionId, basename, docName, result.CreatedAt)
		if err != nil {
			return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to dokumen_pendukung_tim_penilaian: %w", err))
		}
	}

	return admissionId, nil
}

// GetAssessmentTeamCtx retrieves detail about assessment team detail.
func (c *Client) GetAssessmentTeamCtx(ctx context.Context, assessmentTeamId string) (assessmentTeam *models.AssessmentTeam, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)
	profileMdb := metricutil.NewDB(c.ProfileDb, c.SqlMetrics)
	referenceMdb := metricutil.NewDB(c.ReferenceDb, c.SqlMetrics)

	assessmentTeam = &models.AssessmentTeam{
		AssessmentTeamId: assessmentTeamId,
		SupportDocuments: []*models.Document{},
	}

	err = mdb.QueryRowContext(ctx, `
select
	instansi_id,
	jabatan_fungsional_id,
	tgl_usulan,
	no_usulan,
	status
from tim_penilaian where tim_penilaian_id = $1
`, assessmentTeamId).Scan(
		&assessmentTeam.AgencyId,
		&assessmentTeam.FunctionalPositionId,
		&assessmentTeam.AdmissionDate,
		&assessmentTeam.AdmissionNumber,
		&assessmentTeam.Status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrEntryNotFound
		}
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	positions, err := c.getFunctionalPositionNames(ctx, referenceMdb, []string{assessmentTeam.FunctionalPositionId})
	if err != nil {
		return nil, err
	}

	agencies, err := c.getAgencyNames(ctx, referenceMdb, []string{assessmentTeam.AgencyId})
	if err != nil {
		return nil, err
	}

	assessmentTeam.FunctionalPosition = positions[assessmentTeam.FunctionalPositionId]
	assessmentTeam.Agency = agencies[assessmentTeam.AgencyId]

	var assessorIds []string
	assessorRows, err := mdb.QueryContext(ctx, "select asn_id, peran, status, coalesce(alasan_ditolak, '') from anggota_tim_penilaian where tim_penilaian_id = $1", assessmentTeamId)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer assessorRows.Close()

	for assessorRows.Next() {
		assessor := &models.Assessor{}

		err = assessorRows.Scan(
			&assessor.AsnId,
			&assessor.Role,
			&assessor.Status,
			&assessor.ReasonRejected,
		)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}

		assessorIds = append(assessorIds, assessor.AsnId)
		assessmentTeam.Assessors = append(assessmentTeam.Assessors, assessor)
	}
	if err = assessorRows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	asns, err := c.getAsnNipNames(ctx, profileMdb, assessorIds)
	if err != nil {
		return nil, err
	}

	for _, a := range assessmentTeam.Assessors {
		a.Nip = asns[a.AsnId].Nip
		a.Name = asns[a.AsnId].AsnName
	}

	// Get support documents
	documentRows, err := mdb.QueryContext(ctx, "select filename, nama_doc, createdat from dokumen_pendukung_tim_penilaian where tim_penilaian_id = $1", assessmentTeamId)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer documentRows.Close()

	for documentRows.Next() {
		supportDocument := &models.Document{}
		err = documentRows.Scan(
			&supportDocument.Filename,
			&supportDocument.DocumentName,
			&supportDocument.CreatedAt,
		)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
		assessmentTeam.SupportDocuments = append(assessmentTeam.SupportDocuments, supportDocument)
	}
	if err = documentRows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	// Get recommendation letters
	if assessmentTeam.Status == models.AssessmentTeamStatusVerified {
		assessmentTeam.RecommendationLetter = &models.Document{}
		err = mdb.QueryRowContext(ctx, "select filename, nama_doc, no_surat, tgl_surat, createdat from surat_rekomendasi_tim_penilaian where tim_penilaian_id = $1", assessmentTeamId).Scan(
			&assessmentTeam.RecommendationLetter.Filename,
			&assessmentTeam.RecommendationLetter.DocumentName,
			&assessmentTeam.RecommendationLetter.DocumentNumber,
			&assessmentTeam.RecommendationLetter.DocumentDate,
			&assessmentTeam.RecommendationLetter.CreatedAt,
		)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
	}

	return assessmentTeam, nil
}

// SearchAssessmentTeamsCtx searches for the list of assessment team admissions with a particular filter.
// It will return empty slice if no admissions are found.
func (c *Client) SearchAssessmentTeamsCtx(ctx context.Context, filter *AssessmentTeamAdmissionSearchFilter) (result *search.PaginatedList[*models.AssessmentTeamItem], err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)
	referenceMdb := metricutil.NewDB(c.ReferenceDb, c.SqlMetrics)

	admissionRows, err := mdb.QueryContext(
		ctx,
		"select tim_penilaian_id, no_usulan, instansi_id, tgl_usulan, status from tim_penilaian where ($1 = '' or instansi_id = $1) and ($2 <= 0 or status = $2) and ($3 = '' or asn_id = $3) and ($4 = '' or tgl_usulan = $4::date) order by tgl_usulan desc limit $5 offset $6",
		filter.AgencyId,
		filter.AdmissionStatus,
		filter.SubmitterAsnId,
		filter.AdmissionDate,
		filter.CountPerPage+1,
		(filter.PageNumber-1)*filter.CountPerPage,
	)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer admissionRows.Close()

	var agencyIds []string
	admissions := make([]*models.AssessmentTeamItem, 0)
	for admissionRows.Next() {
		admission := &models.AssessmentTeamItem{}
		err = admissionRows.Scan(
			&admission.AssessmentTeamId,
			&admission.AdmissionNumber,
			&admission.AgencyId,
			&admission.AdmissionDate,
			&admission.Status,
		)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
		admissions = append(admissions, admission)
		agencyIds = append(agencyIds, admission.AgencyId)
	}
	if err = admissionRows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	agencies, err := c.getAgencyNames(ctx, referenceMdb, agencyIds)
	if err != nil {
		return nil, err
	}

	for _, admission := range admissions {
		admission.Agency = agencies[admission.AgencyId]
	}

	hasNext := false
	if len(admissions) > filter.CountPerPage {
		hasNext = true
		admissions = admissions[:filter.CountPerPage]
	}

	return &search.PaginatedList[*models.AssessmentTeamItem]{
		Data:     admissions,
		Metadata: search.CreatePaginatedListMetadataNoTotalNext(filter.PageNumber, len(admissions), hasNext),
	}, nil
}

// SetAssessmentTeamVerificationCtx set assessment team status to verified. It also moves filenames defined
// request.TempRecommendationLetter from temporary storage to permanent storage in object storage.
func (c *Client) SetAssessmentTeamVerificationCtx(ctx context.Context, request *models.AssessmentTeamVerification) (updatedAt time.Time, err error) {
	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return time.Time{}, err
	}

	defer func() {
		c.completeMtx(mtx, err)
	}()

	currentStatus := 0
	err = mtx.QueryRowContext(ctx, "select status from tim_penilaian where tim_penilaian_id = $1 for update", request.AssessmentTeamId).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, ErrEntryNotFound
		}

		return time.Time{}, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	if currentStatus != models.AssessmentTeamStatusCreated {
		return time.Time{}, ec.NewError(ErrCodeAssessmentTeamStatusNotCreated, Errs[ErrCodeAssessmentTeamStatusNotCreated], err)
	}

	err = mtx.QueryRowContext(
		ctx,
		"update tim_penilaian set status = $1, status_ts = current_timestamp, status_by = $2 where tim_penilaian_id = $3 returning status_ts",
		models.AssessmentTeamStatusVerified,
		request.SubmitterAsnId,
		request.AssessmentTeamId,
	).Scan(&updatedAt)
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], err)
	}

	assessorStmt, err := mtx.PrepareContext(ctx, "update anggota_tim_penilaian set status = $1, alasan_ditolak = $2 where tim_penilaian_id = $3 and asn_id = $4")
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], fmt.Errorf("cannot prepare statement for anggota_tim_penilaian table: %w", err))
	}

	for _, asn := range request.Assessors {
		if _, ok := models.AssessmentTeamAssessorStatuses[asn.Status]; !ok {
			return time.Time{}, ErrAssessmentTeamAssessorStatusInvalid
		}

		_, err = assessorStmt.ExecContext(ctx, asn.Status, sql.NullString{Valid: asn.ReasonRejected != "", String: asn.ReasonRejected}, request.AssessmentTeamId, asn.AsnId)
		if err != nil {
			return time.Time{}, ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot update entry of anggota_tim_penilaian: %w", err))
		}
	}

	filepath := path.Join(AssessmentTeamRecommendationLetterSubdir, fmt.Sprintf("%s.pdf", request.AssessmentTeamId))
	_, err = c.AssessmentTeamStorage.SaveAssessmentTeamFiles(ctx, []string{filepath})
	if err != nil {
		if err == object.ErrTempFileNotFound {
			return time.Time{}, ErrStorageFileNotFound
		}
		return time.Time{}, ec.NewError(ErrCodeStorageCopyFail, Errs[ErrCodeStorageCopyFail], fmt.Errorf("cannot save recommendation letter: %w", err))
	}

	_, err = mtx.ExecContext(
		ctx,
		"insert into surat_rekomendasi_tim_penilaian(tim_penilaian_id, filename, nama_doc, no_surat, tgl_surat, createdat) values($1, $2, $3, $4, $5, current_timestamp)",
		request.AssessmentTeamId, request.AssessmentTeamId, request.TempRecommendationLetter.DocumentName, request.TempRecommendationLetter.DocumentNumber, request.TempRecommendationLetter.DocumentDate,
	)
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to surat_rekomendasi_tim_penilaian: %w", err))
	}

	return updatedAt, nil
}
