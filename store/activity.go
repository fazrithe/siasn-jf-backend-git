package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	. "github.com/if-itb/siasn-jf-backend/errnum"
	"github.com/if-itb/siasn-jf-backend/store/models"
	"github.com/if-itb/siasn-jf-backend/store/object"
	"github.com/if-itb/siasn-libs-backend/auth"
	"github.com/if-itb/siasn-libs-backend/docx"
	"github.com/if-itb/siasn-libs-backend/ec"
	"github.com/if-itb/siasn-libs-backend/metricutil"
	"github.com/if-itb/siasn-libs-backend/search"
	"github.com/lib/pq"
)

// CheckActivityAdmissionInsert checks activity admission request object for validity, before inserting the entry
// to the database.
func (c *Client) CheckActivityAdmissionInsert(request *models.ActivityAdmission) (err error) {
	if len(request.Attendees) <= 0 {
		return ec.NewErrorBasic(ErrCodeActivityAdmissionInsertNoAttendees, Errs[ErrCodeActivityAdmissionInsertNoAttendees])
	}

	if request.Name == "" {
		return ec.NewErrorBasic(ErrCodeActivityAdmissionInsertNameEmpty, Errs[ErrCodeActivityAdmissionInsertNameEmpty])
	}

	if request.PositionGrade == "" {
		return ec.NewErrorBasic(ErrCodeActivityAdmissionPositionGradeEmpty, Errs[ErrCodeActivityAdmissionPositionGradeEmpty])
	}

	if _, ok := models.ActivityTypes[request.Type]; !ok {
		return ec.NewError(ErrCodeActivityAdmissionInsertTypeInvalid, Errs[ErrCodeActivityAdmissionInsertTypeInvalid], models.ErrActivityTypeInvalid)
	}

	startDate, err := request.StartDate.Time()
	if err != nil {
		return ec.NewError(ErrCodeActivityAdmissionDatePeriodInvalid, Errs[ErrCodeActivityAdmissionDatePeriodInvalid], err)
	}
	endDate, err := request.EndDate.Time()
	if err != nil {
		return ec.NewError(ErrCodeActivityAdmissionDatePeriodInvalid, Errs[ErrCodeActivityAdmissionDatePeriodInvalid], err)
	}

	if endDate.Before(startDate) {
		return ec.NewError(ErrCodeActivityAdmissionDatePeriodInvalid, Errs[ErrCodeActivityAdmissionDatePeriodInvalid], errors.New("end date is before start date"))
	}

	if request.TrainingYear < 1970 {
		return ec.NewErrorBasic(ErrCodeActivityAdmissionTrainingYearInvalid, Errs[ErrCodeActivityAdmissionTrainingYearInvalid])
	}

	if request.Duration < 0 {
		return ec.NewErrorBasic(ErrCodeActivityAdmissionDurationInvalid, Errs[ErrCodeActivityAdmissionDurationInvalid])
	}

	if request.AdmissionNumber == "" {
		return ec.NewErrorBasic(ErrCodeActivityAdmissionNumberInvalid, Errs[ErrCodeActivityAdmissionNumberInvalid])
	}

	return nil
}

// InsertActivityAdmissionCtx inserts a new admission request (pengajuan kegiatan).
// It also moves filenames defined request.TempSupportDocuments from temporary storage to permanent storage in
// object storage.
func (c *Client) InsertActivityAdmissionCtx(ctx context.Context, request *models.ActivityAdmission) (activityId string, err error) {
	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return "", err
	}

	defer func() {
		c.completeMtx(mtx, err)
	}()

	err = c.CheckActivityAdmissionInsert(request)
	if err != nil {
		return "", err
	}

	// TODO: verify the given agency ID and jabatan jenjang

	asns, err := c.getAsnBulkByIdForActivityInsert(ctx, request.Attendees, request.AgencyId)
	if err != nil {
		return "", err
	}

	activityIdBytes := uuid.New()
	query := "INSERT INTO kegiatan (kegiatan_id, nama, status, jenis, deskripsi, tgl_usulan, tgl_mulai, tgl_selesai, jabatan_jenjang, instansi_id, data_tambahan, tahun_diklat, durasi, instansi_penyelenggara, no_usulan) VALUES ($1, $2,  $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)"

	// Execute the SQL INSERT statement with the data
	_, err = mtx.Exec(query, activityIdBytes, request.Name, models.ActivityAdmissionStatusCreated, request.Type, request.Description, time.Time(request.AdmissionTimestamp), string(request.StartDate), string(request.EndDate), request.PositionGrade, request.AgencyId, request.Extra, request.TrainingYear, request.Duration, sql.NullString{Valid: request.OrganizerAgency != "", String: request.OrganizerAgency}, request.AdmissionNumber)
	if err != nil {
		log.Fatal(err)
	}

	queryHist := "INSERT INTO kegiatan_status_hist (kegiatan_kegiatan_id, status, modified_at_ts, user_id) VALUES ($1, $2, $3, $4)"

	// Execute the SQL INSERT statement with the data
	_, err = mtx.Exec(queryHist, activityIdBytes, models.ActivityAdmissionStatusCreated, time.Now(), request.SubmitterAsnId)
	if err != nil {
		log.Fatal(err)
	}

	// _, err = mtx.ExecContext(
	// 	ctx,
	// 	"insert into kegiatan_status_hist(kegiatan_kegiatan_id, status, modified_at_ts, user_id) VALUES($1, $2, $3, $4)",
	// 	activityIdBytes,
	// 	models.ActivityAdmissionStatusCreated,
	// 	time.Now(),
	// 	request.SubmitterAsnId,
	// )
	// if err != nil {
	// 	return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to kegiatan_status_hist: %w", err))
	// }

	pegawaiStmt, err := mtx.PrepareContext(ctx, "insert into pegawai(user_id, nip_baru, nip_lama) VALUES($1, $2, $3) on conflict(user_id) do nothing")
	if err != nil {
		return "", ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], fmt.Errorf("cannot prepare statement for pegawai table: %w", err))
	}

	pesertaKegiatanStmt, err := mtx.PrepareContext(ctx, "insert into perserta_kegiatan(kegiatan_kegiatan_id, pegawai_user_id) VALUES($1, $2)")
	if err != nil {
		return "", ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], fmt.Errorf("cannot prepare statement for peserta_kegiatan table: %w", err))
	}

	for _, asn := range asns {
		_, err = pegawaiStmt.ExecContext(ctx, asn.AsnId, sql.NullString{Valid: asn.NewNip != "", String: asn.NewNip}, sql.NullString{Valid: asn.OldNip != "", String: asn.OldNip})
		if err != nil {
			return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to pegawai: %w", err))
		}

		_, err = pesertaKegiatanStmt.ExecContext(ctx, activityIdBytes, asn.AsnId)
		if err != nil {
			return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to peserta_kegiatan: %w", err))
		}
	}

	docStmt, err := mtx.PrepareContext(ctx, "insert into dokumen_pendukung(kegiatan_kegiatan_id, filename, nama_doc, createdat) values($1, $2, $3, $4)")
	if err != nil {
		return "", ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], fmt.Errorf("cannot prepare statement for dokumen_pendukung table: %w", err))
	}

	filenameToDocName := make(map[string]string)
	var filenames []string
	for _, doc := range request.TempSupportDocuments {
		filenames = append(filenames, path.Join(ActivitySupportDocSubdir, doc.Filename))
		filenameToDocName[doc.Filename] = doc.DocumentName
	}
	results, err := c.ActivityStorage.SaveActivityFiles(ctx, filenames)
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
		_, err = docStmt.ExecContext(ctx, activityIdBytes, basename, docName, result.CreatedAt)
		if err != nil {
			return "", ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to dokumen_pendukung: %w", err))
		}
	}

	return activityIdBytes.String(), nil
}

// getAsnBulkByIdForActivityInsert works just like getAsnBulkForActivityInsert, but with ASN ID.
func (c *Client) getAsnBulkByIdForActivityInsert(ctx context.Context, asnIds []string, agencyId string) (asns []*auth.Asn, err error) {
	mdb := metricutil.NewDB(c.ProfileDb, c.SqlMetrics)

	oldNipRows, err := mdb.QueryContext(ctx, "select id, coalesce(nip_baru, ''), coalesce(nip_lama, '') from pns where id = ANY($1) and instansi_kerja_id = $2", pq.Array(asnIds), agencyId)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	for oldNipRows.Next() {
		asn := &auth.Asn{}
		err = oldNipRows.Scan(&asn.AsnId, &asn.NewNip, &asn.OldNip)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
		asns = append(asns, asn)
	}
	defer oldNipRows.Close()

	if err = oldNipRows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	if len(asns) != len(asnIds) {
		return nil, ec.NewError(ErrCodeActivityAdmissionInsertAsnNotFound, Errs[ErrCodeActivityAdmissionInsertAsnNotFound], err)
	}

	return asns, nil
}

// Deprecated: use GetUserDetailByNipWorkAgencyId directly.
// GetAsnNewCtx searches for an ASN based on new NIP and work agency ID (instansi kerja).
// It will return nil in case the profile cannot be found.
func (c *Client) GetAsnNewCtx(ctx context.Context, newNip, workAgencyId string) (asn *auth.Asn, err error) {
	return c.GetUserDetailByNipWorkAgencyId(ctx, newNip, workAgencyId)
}

// Deprecated: use GetUserDetailByNipWorkAgencyId directly.
// GetAsnOldCtx searches for an ASN based on old NIP and work agency ID (instansi kerja).
// It will return nil in case the profile cannot be found.
func (c *Client) GetAsnOldCtx(ctx context.Context, oldNip, workAgencyId string) (asn *auth.Asn, err error) {
	return c.GetUserDetailByNipWorkAgencyId(ctx, oldNip, workAgencyId)
}

// updateActivityStatusCtxDh updates an activity status.
// This will also insert a new status update history entry. This method already returns an error in form of error code.
func (c *Client) updateActivityStatusCtxDh(ctx context.Context, dh metricutil.DbHandler, activityId uuid.UUID, submitterAsnId string, agencyId string, status int) (modifiedAt time.Time, err error) {
	_, err = dh.ExecContext(
		ctx,
		"update kegiatan set status = $1 where kegiatan_id = $2 and instansi_id = $3",
		status,
		activityId.String(),
		agencyId,
	)
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], err)
	}

	modifiedAt = time.Now()
	_, err = dh.ExecContext(
		ctx,
		"insert into kegiatan_status_hist(kegiatan_kegiatan_id, status, modified_at_ts, user_id) values($1, $2, $3, $4)",
		activityId.String(),
		status,
		modifiedAt,
		submitterAsnId,
	)
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], err)
	}

	return modifiedAt, nil
}

// Deprecated: CSR status is removed.
// SetActivityStatusCsrCtx sets an activity admission status to certificate request.
// If the status of the activity is not models.ActivityAdmissionStatusAccepted, this will return error code errnum.ErrCodeActivityCsrStatusSetNotAccepted.
// If the status is already models.ActivityAdmissionStatusCertRequest, this will return error code errnum.ErrCodeActivityCsrStatusSetOngoing.
func (c *Client) SetActivityStatusCsrCtx(ctx context.Context, request *models.ActivityCsrRequest) (modifiedAt time.Time, err error) {
	mtx, err := c.createMtx(ctx)
	if err != nil {
		return time.Time{}, err
	}

	defer func() {
		c.completeMtx(mtx, err)
	}()

	activityIdBytes, err := uuid.Parse(request.ActivityId)
	if err != nil {
		return time.Time{}, ErrUuidInvalid
	}

	currentStatus := 0
	err = mtx.QueryRowContext(ctx, "select status from kegiatan where kegiatan_id = $1 and instansi_id = $2 for update", request.ActivityId, request.AgencyId).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, ErrEntryNotFound
		}

		return time.Time{}, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	if currentStatus == models.ActivityAdmissionStatusCertRequest {
		return time.Time{}, ec.NewErrorBasic(ErrCodeActivityCsrStatusSetOngoing, Errs[ErrCodeActivityCsrStatusSetOngoing])
	}

	if currentStatus != models.ActivityAdmissionStatusAccepted {
		return time.Time{}, ec.NewErrorBasic(ErrCodeActivityCsrStatusSetNotAccepted, Errs[ErrCodeActivityCsrStatusSetNotAccepted])
	}

	if len(request.AttendeesPassing) <= 0 {
		return time.Time{}, ec.NewErrorBasic(ErrCodeActivityVerificationStatusNoAttendees, Errs[ErrCodeActivityVerificationStatusNoAttendees])
	}

	ignored := 0
	for _, asn := range request.AttendeesPassing {
		reasonRejected := sql.NullString{String: asn.ReasonRejected, Valid: !asn.IsPassing}
		err = mtx.QueryRowContext(ctx,
			"update perserta_kegiatan set ispass = $1, passts = $2, passby = $3, pass_rejected_reason = $4 where kegiatan_kegiatan_id = $5 and pegawai_user_id = $6 returning 1",
			asn.IsPassing, time.Now(), request.SubmitterAsnId, reasonRejected, request.ActivityId, asn.AsnId).Scan(&ignored)
		if err != nil {
			if err == sql.ErrNoRows {
				return time.Time{}, ErrEntryNotFound
			}

			return time.Time{}, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
	}

	modifiedAt, err = c.updateActivityStatusCtxDh(ctx, mtx, activityIdBytes, request.SubmitterAsnId, request.AgencyId, models.ActivityAdmissionStatusCertRequest)
	if err != nil {
		return time.Time{}, err
	}

	return modifiedAt, nil
}

// SetActivityStatusAcceptedCtx sets an activity admission status as models.ActivityAdmissionStatusAccepted.
// If the status of the activity is not models.ActivityAdmissionStatusCreated, this will return error code errnum.ErrCodeActivityVerificationStatusProcessedFurther.
// If the status is already models.ActivityAdmissionStatusAccepted, this will return error code errnum.ErrCodeActivityVerificationStatusAlreadyAccepted.
func (c *Client) SetActivityStatusAcceptedCtx(ctx context.Context, request *models.ActivityVerificationRequest) (modifiedAt time.Time, err error) {
	activityId, err := uuid.Parse(request.ActivityId)
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodeUuidInvalid, Errs[ErrCodeUuidInvalid], err)
	}

	mtx, err := c.createMtx(ctx)
	if err != nil {
		return time.Time{}, err
	}

	defer func() {
		c.completeMtx(mtx, err)
	}()

	currentStatus := 0
	err = mtx.QueryRowContext(ctx, "select status from kegiatan where kegiatan_id = $1 and instansi_id = $2 for update", request.ActivityId, request.AgencyId).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, ErrEntryNotFound
		}

		return time.Time{}, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	if currentStatus == models.ActivityAdmissionStatusAccepted {
		return time.Time{}, ec.NewErrorBasic(ErrCodeActivityVerificationStatusAlreadyAccepted, Errs[ErrCodeActivityVerificationStatusAlreadyAccepted])
	}

	if currentStatus != models.ActivityAdmissionStatusCreated {
		return time.Time{}, ec.NewErrorBasic(ErrCodeActivityVerificationStatusProcessedFurther, Errs[ErrCodeActivityVerificationStatusProcessedFurther])
	}

	if len(request.AttendeesAcceptance) <= 0 {
		return time.Time{}, ec.NewErrorBasic(ErrCodeActivityVerificationStatusNoAttendees, Errs[ErrCodeActivityVerificationStatusNoAttendees])
	}

	isAccepted := false
	for _, asn := range request.AttendeesAcceptance {
		reasonRejected := sql.NullString{String: asn.ReasonRejected, Valid: !asn.IsAccepted}
		err = mtx.QueryRowContext(ctx,
			"update perserta_kegiatan set isaccepted = $1, acceptedts = $2, acceptedby = $3, accepted_rejected_reason = $4 where kegiatan_kegiatan_id = $5 and pegawai_user_id = $6 returning 1",
			asn.IsAccepted, time.Now(), request.SubmitterAsnId, reasonRejected, request.ActivityId, asn.AsnId).Scan(&isAccepted)
		if err != nil {
			if err == sql.ErrNoRows {
				return time.Time{}, ErrEntryNotFound
			}

			return time.Time{}, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
	}

	modifiedAt, err = c.updateActivityStatusCtxDh(ctx, mtx, activityId, request.SubmitterAsnId, request.AgencyId, models.ActivityAdmissionStatusAccepted)
	if err != nil {
		return time.Time{}, err
	}

	return modifiedAt, nil
}

// CreateCertFilename creates a full object name from activity ID, attendee ASN ID and certificate type.
// Will return error if certificate type is unknown.
func (c *Client) CreateCertFilename(activityId, attendeeAsnId string, certType int) (filename string, err error) {
	subdir, ok := ActivityCertTypeToSubdir[certType]
	if !ok {
		return "", ec.NewErrorBasic(ErrCodeActivityCertTypeUnsupported, Errs[ErrCodeActivityCertTypeUnsupported])
	}

	return path.Join(subdir, fmt.Sprintf("%s-%s.pdf", activityId, attendeeAsnId)), nil
}

// VerifyActivityCertCtx verifies that an upload request is valid and returns a full object filename to be uploaded
// to object storage. Upload request is not valid if activity status is not models.ActivityAdmissionStatusAccepted.
func (c *Client) VerifyActivityCertCtx(ctx context.Context, request *models.ActivityCertGenUploadRequest) (filename string, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)

	currentStatus := 0
	currentType := 0
	err = mdb.QueryRowContext(ctx, "select status, jenis from kegiatan where kegiatan_id = $1 and instansi_id = $2 for share", request.ActivityId, request.AgencyId).Scan(&currentStatus, &currentType)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrEntryNotFound
		}

		return "", ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	if currentStatus != models.ActivityAdmissionStatusAccepted {
		return "", ec.NewErrorBasic(ErrCodeActivityCertSubmitNotRequested, Errs[ErrCodeActivityCertSubmitNotRequested])
	}

	if request.Type == models.ActivityCertTypePak && currentType != models.ActivityTypeMutationExam {
		return "", ec.NewErrorBasic(ErrCodeActivityCertSubmitPakUnsupported, Errs[ErrCodeActivityCertSubmitPakUnsupported])
	}

	return c.CreateCertFilename(request.ActivityId, request.AttendeeAsnId, request.Type)
}

// InsertActivityCertCtx inserts all the certificates/PAKs to object storage and database.
// This will mark the activity status to models.ActivityAdmissionStatusCertPublished. Certificates must already exist
// in temporary storage. You can handle the error code ErrCodeStorageCopyFail in case the temporary files do not exist,
// for example due to temporary storage being purged at the exact time certificates are submitted.
//
// May return ErrCodeEntryNotFound if activity cannot be found. May also return ErrCodeActivityCertSubmitNotRequested
// if activity status is not models.ActivityAdmissionStatusAccepted.
//
// PAK document type is only supported for activity with type uji kompetensi perpindahan jabatan (models.ActivityTypeMutationExam).
func (c *Client) InsertActivityCertCtx(ctx context.Context, cert *models.ActivityCertGenRequest) (modifiedAt time.Time, err error) {
	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return time.Time{}, err
	}

	defer func() {
		c.completeMtx(mtx, err)
	}()

	activityIdUuid, err := uuid.Parse(cert.ActivityId)
	if err != nil {
		return time.Time{}, ErrUuidInvalid
	}

	if len(cert.AttendeesPassing) <= 0 {
		return time.Time{}, ec.NewErrorBasic(ErrCodeActivityCertSubmitNoDocs, Errs[ErrCodeActivityCertSubmitNoDocs])
	}

	currentStatus := 0
	currentType := 0
	err = mtx.QueryRowContext(ctx, "select status, jenis from kegiatan where kegiatan_id = $1 and instansi_id = $2 for update", cert.ActivityId, cert.AgencyId).Scan(&currentStatus, &currentType)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, ErrEntryNotFound
		}

		return time.Time{}, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	if currentStatus != models.ActivityAdmissionStatusAccepted {
		return time.Time{}, ec.NewErrorBasic(ErrCodeActivityCertSubmitNotRequested, Errs[ErrCodeActivityCertSubmitNotRequested])
	}

	stmt, err := mtx.PrepareContext(ctx, "insert into sertifikat(persertakegiatan_kegiatan_id, persertakegiatan_user_id, nosurat, tgl_surat, jenis, ttd_user_id, createdat, nilai) values($1, $2, $3, $4, $5, $6, current_timestamp, $7)")
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], fmt.Errorf("cannot prepare sertifikat statement: %w", err))
	}
	defer stmt.Close()

	passingStmt, err := mtx.PrepareContext(ctx, "update perserta_kegiatan set ispass = $1, passts = current_timestamp, pass_rejected_reason = $2, passby = $3 where pegawai_user_id = $4 and kegiatan_kegiatan_id = $5 and isaccepted returning 1")
	if err != nil {
		return time.Time{}, ec.NewError(ErrCodePrepareFail, Errs[ErrCodePrepareFail], fmt.Errorf("cannot prepare perserta_kegiatan statement: %w", err))
	}
	defer passingStmt.Close()

	for _, attendee := range cert.AttendeesPassing {
		if attendee.Type == models.ActivityCertTypePak && currentType != models.ActivityTypeMutationExam {
			return time.Time{}, ec.NewErrorBasic(ErrCodeActivityCertSubmitPakUnsupported, Errs[ErrCodeActivityCertSubmitPakUnsupported])
		}

		d := 0
		err = passingStmt.QueryRowContext(ctx, attendee.IsPassing, sql.NullString{Valid: !attendee.IsPassing, String: attendee.ReasonRejected}, cert.SubmitterAsnId, attendee.AttendeeAsnId, cert.ActivityId).Scan(&d)
		if err != nil {
			if err == sql.ErrNoRows { // Don't create a certificate for attendees that cannot be found
				err = nil
				continue
			}
			return time.Time{}, ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to perserta_kegiatan: %w", err))
		}

		if attendee.IsPassing {
			_, err = stmt.ExecContext(ctx, cert.ActivityId, attendee.AttendeeAsnId, attendee.DocumentNumber, string(attendee.DocumentDate), attendee.Type, attendee.SignerAsnId, sql.NullFloat64{Valid: attendee.Score <= 0, Float64: float64(attendee.Score)})
			if err != nil {
				return time.Time{}, ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to sertifikat: %w", err))
			}
		}

		nameFile := "doc1-1699499155.docx"

		c.copyFileSertificate(nameFile, attendee.AttendeeAsnId)

	}

	modifiedAt, err = c.updateActivityStatusCtxDh(ctx, mtx, activityIdUuid, cert.SubmitterAsnId, cert.AgencyId, models.ActivityAdmissionStatusCertPublished)
	if err != nil {
		return time.Time{}, err
	}

	jsonData := `{
				"name": "` + cert.Signing.JabtanInstansiPengusul + `",
				"jabatan_instansi_pengusul": "` + cert.Signing.JabtanInstansiPengusul + `",
				"nama_pejabat_instansi_pengusul": "` + cert.Signing.NamaPejabatInstansiPengusul + `",
				"nip_pejabat_instansi_pengusul": "` + cert.Signing.NipPejabatInstansiPnegusul + `",
				"jabatan_instansi_penyelenggara": "` + cert.Signing.JabatanInstansiPenyelenggara + `",
				"nama_pejabat_instansi_penyelenggara": "` + cert.Signing.NamaPejabatInstansiPenyelenggara + `",
				"nip_pejabat_instansi_penyelenggara": "` + cert.Signing.NipPejabatInstansiPenyelanggara + `",
				"jabatan_pemateri": "` + cert.Signing.JabatanPemateri + `",
				"nama_pemateri": "` + cert.Signing.NamaPemateri + `",
				"nip_pemateri": "` + cert.Signing.NipPemateri + `"
				}`

	recordID := cert.ActivityId
	templateID := cert.TemplateId
	query := "UPDATE kegiatan SET penandatangan = $1, template_id = $2 WHERE kegiatan_id = $3"
	result, err := mtx.Exec(query, jsonData, templateID, recordID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(result)

	// currentDir, err := os.Getwd()
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }

	// nameFile := "doc1-1699499155.docx"
	// sourceFile, err := os.Open(currentDir + "\\uploads\\" + nameFile)
	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }
	// defer sourceFile.Close()

	// destinationFile, err := os.Create(currentDir + "\\uploads\\sertificate\\" + nameFile)
	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }
	// defer destinationFile.Close()

	// _, err = io.Copy(destinationFile, sourceFile)
	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }

	return modifiedAt, nil
}

func (c *Client) copyFileSertificate(src, dst string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}
	sourceFile, err := os.Open(currentDir + "\\uploads\\" + src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(currentDir + "\\uploads\\sertificate\\" + dst + ".pdf")
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

// Deprecated: use the paginated version.
// SearchActivityAdmissionsCtx searches for the list of activity admissions of a particular
// work agency ID (instansi kerja).
// It will return empty slice if no admissions are found.
func (c *Client) SearchActivityAdmissionsCtx(ctx context.Context, searchFilter *ActivityAdmissionSearchFilter) (admissions []*models.ActivityAdmission, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)

	queryBuilder := strings.Builder{}
	var queryArgs []interface{}

	queryBuilder.WriteString("select kegiatan_id, nama, status, jenis, deskripsi, tgl_usulan, tgl_mulai, tgl_selesai, jabatan_jenjang, instansi_id, data_tambahan from kegiatan where instansi_id = $")
	queryArgs = append(queryArgs, searchFilter.WorkAgencyId)
	queryBuilder.WriteString(strconv.Itoa(len(queryArgs)))

	if searchFilter.AdmissionStatus != 0 {
		queryBuilder.WriteString(" and status = $")
		queryArgs = append(queryArgs, searchFilter.AdmissionStatus)
		queryBuilder.WriteString(strconv.Itoa(len(queryArgs)))
	}

	if searchFilter.AdmissionType != 0 {
		queryBuilder.WriteString(" and jenis = $")
		queryArgs = append(queryArgs, searchFilter.AdmissionType)
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

	admissions = []*models.ActivityAdmission{}
	for admissionRows.Next() {
		admission := &models.ActivityAdmission{}
		err = admissionRows.Scan(
			&admission.ActivityId,
			&admission.Name,
			&admission.Status,
			&admission.Type,
			&admission.Description,
			&admission.AdmissionTimestamp,
			&admission.StartDate,
			&admission.EndDate,
			&admission.PositionGrade,
			&admission.AgencyId,
			&admission.Extra,
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

// SearchActivityAdmissionsPaginatedCtx searches for the list of activity admissions of a particular
// work agency ID (instansi kerja).
// It will return empty slice if no admissions are found.
func (c *Client) SearchActivityAdmissionsPaginatedCtx(ctx context.Context, filter *ActivityAdmissionSearchFilter) (result *search.PaginatedList, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)

	admissionRows, err := mdb.QueryContext(
		ctx,
		"select kegiatan_id, nama, status, jenis, deskripsi, tgl_usulan, tgl_mulai, tgl_selesai, jabatan_jenjang, instansi_id, data_tambahan, tahun_diklat, durasi, coalesce(instansi_penyelenggara, ''), no_usulan from kegiatan where instansi_id = $1 and ($2 <= 0 or status = $2) and ($3 <= 0 or jenis = $3) and ($4::date is null or tgl_usulan::date = $4) order by tgl_usulan desc limit $5 offset $6",
		filter.WorkAgencyId,
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

	admissions := make([]*models.ActivityAdmission, 0)
	for admissionRows.Next() {
		admission := &models.ActivityAdmission{}
		err = admissionRows.Scan(
			&admission.ActivityId,
			&admission.Name,
			&admission.Status,
			&admission.Type,
			&admission.Description,
			&admission.AdmissionTimestamp,
			&admission.StartDate,
			&admission.EndDate,
			&admission.PositionGrade,
			&admission.AgencyId,
			&admission.Extra,
			&admission.TrainingYear,
			&admission.Duration,
			&admission.OrganizerAgency,
			&admission.AdmissionNumber,
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

	return &search.PaginatedList{
		Data:     admissions,
		Metadata: search.CreatePaginatedListMetadataNoTotalNext(filter.PageNumber, len(admissions), hasNext),
	}, nil
}

// GetActivityAdmissionDetailCtx composes the details af an activity admission which includes the information of the
// activity itself, the list of attendees, and the supporting documents.
// If supplied, agencyId will be compared to the activity admission's agencyId.
// If the agencyId is an empty string, it's assumed that the user is a 'pembina' and can see any activity admission
// detail.
func (c *Client) GetActivityAdmissionDetailCtx(ctx context.Context, activityAdmissionId string, agencyId string) (admission *models.ActivityAdmission, err error) {
	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return nil, ec.NewError(ErrCodeTxStart, Errs[ErrCodeTxStart], err)
	}

	defer func() {
		c.completeMtx(mtx, err)
	}()

	admission = &models.ActivityAdmission{
		AttendeesDetail:  []*models.ActivityAttendee{},
		SupportDocuments: []*models.Document{},
	}

	err = mtx.QueryRowContext(ctx, "select kegiatan_id, nama, status, jenis, deskripsi, tgl_usulan, tgl_mulai, tgl_selesai, jabatan_jenjang, instansi_id, data_tambahan, tahun_diklat, durasi, coalesce(instansi_penyelenggara, ''), no_usulan from kegiatan where kegiatan_id = $1", activityAdmissionId).
		Scan(
			&admission.ActivityId,
			&admission.Name,
			&admission.Status,
			&admission.Type,
			&admission.Description,
			&admission.AdmissionTimestamp,
			&admission.StartDate,
			&admission.EndDate,
			&admission.PositionGrade,
			&admission.AgencyId,
			&admission.Extra,
			&admission.TrainingYear,
			&admission.Duration,
			&admission.OrganizerAgency,
			&admission.AdmissionNumber,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrEntryNotFound
		}
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	if agencyId != "" && agencyId != admission.AgencyId {
		// The user is not a 'pembina' and the agencyId doesn't match
		return nil, ec.NewErrorBasic(ErrCodeActivityAdmissionDetailForbidden, Errs[ErrCodeActivityAdmissionDetailForbidden])
	}

	// We'll query about the attendees acceptance
	var attendeesId []string

	attendeeRows, err := mtx.QueryContext(ctx, "select pegawai_user_id, isaccepted, coalesce(accepted_rejected_reason, ''), ispass, coalesce(pass_rejected_reason, ''), s.nosurat, s.tgl_surat, s.jenis, s.ttd_user_id, s.createdat, nilai from perserta_kegiatan left join sertifikat s on perserta_kegiatan.pegawai_user_id = s.persertakegiatan_user_id and perserta_kegiatan.kegiatan_kegiatan_id = s.persertakegiatan_kegiatan_id where kegiatan_kegiatan_id = $1", activityAdmissionId)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer attendeeRows.Close()

	for attendeeRows.Next() {
		activityAttendee := &models.ActivityAttendee{}
		// Nullable certificate columns.
		var (
			noSurat   = sql.NullString{}
			tglSurat  = sql.NullString{}
			jenis     = sql.NullInt32{}
			ttdUserId = sql.NullString{}
			createdAt = sql.NullTime{}
			nilai     = sql.NullFloat64{}
		)

		err = attendeeRows.Scan(
			&activityAttendee.AsnId,
			&activityAttendee.IsAccepted,
			&activityAttendee.ReasonRejected,
			&activityAttendee.IsPassing,
			&activityAttendee.ReasonNotPassing,
			&noSurat,
			&tglSurat,
			&jenis,
			&ttdUserId,
			&createdAt,
			&nilai,
		)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
		if noSurat.Valid && tglSurat.Valid && jenis.Valid {
			activityAttendee.Certificate = &models.ActivityCertificate{
				DocumentNumber: noSurat.String,
				DocumentDate:   models.Iso8601Date(tglSurat.String),
				SignerAsnId:    ttdUserId.String,
				AttendeeAsnId:  activityAttendee.AsnId,
				Type:           int(jenis.Int32),
				Score:          float32(nilai.Float64),
			}
		}

		attendeesId = append(attendeesId, activityAttendee.AsnId)
		admission.AttendeesDetail = append(admission.AttendeesDetail, activityAttendee)
	}
	if err = attendeeRows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	// Get support documents
	documentRows, err := mtx.QueryContext(ctx, "select filename, nama_doc from dokumen_pendukung where kegiatan_kegiatan_id = $1", activityAdmissionId)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer documentRows.Close()

	for documentRows.Next() {
		supportDocument := &models.Document{}
		err = documentRows.Scan(
			&supportDocument.Filename,
			&supportDocument.DocumentName,
		)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
		admission.SupportDocuments = append(admission.SupportDocuments, supportDocument)
	}
	if err = documentRows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	// Need to get the name of the attendees from the profile db
	// No need to query the names if no attendees found.
	if len(attendeesId) > 0 {
		profileMdb := metricutil.NewDB(c.ProfileDb, c.SqlMetrics)
		asns, err := c.getAsnNipNames(ctx, profileMdb, attendeesId)
		if err != nil {
			return nil, err
		}

		for _, a := range admission.AttendeesDetail {
			a.Nip = asns[a.AsnId].Nip
			a.AsnName = asns[a.AsnId].AsnName
		}
	}

	return admission, nil
}

// VerifyActivityRecommendationLetterCtx verifies that an upload request is valid.
// Upload request is not valid if requirement status is >= models.ActivityAdmissionStatusAccepted.
func (c *Client) VerifyActivityRecommendationLetterCtx(ctx context.Context, activityId, agencyId string) (err error) {
	parsedActivityId, err := uuid.Parse(activityId)
	if err != nil {
		return ec.NewError(ErrCodeUuidInvalid, Errs[ErrCodeUuidInvalid], err)
	}

	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return err
	}

	defer func() {
		c.completeMtx(mtx, err)
	}()

	err = mtx.QueryRowContext(ctx, "select kegiatan_id from kegiatan where kegiatan_id = $1 and instansi_id = $2", parsedActivityId, agencyId).Scan(&parsedActivityId)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrEntryNotFound
		}

		return ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	currentStatus := 0
	err = mtx.QueryRowContext(ctx, "select status from kegiatan where kegiatan_id = $1 and instansi_id = $2 for update", parsedActivityId, agencyId).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrEntryNotFound
		}

		return ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	if currentStatus >= models.ActivityAdmissionStatusAccepted {
		return ec.NewErrorBasic(ErrCodeActivityVerificationStatusAlreadyAccepted, Errs[ErrCodeActivityVerificationStatusAlreadyAccepted])
	}

	return
}

// Deprecated: no longer needed.
// SubmitActivityRecommendationLetterCtx inserts a new admission request for activity recommendation letter.
// It also moves filename defined admission.RecommendationLetter from temporary storage to permanent storage in object
// storage.
func (c *Client) SubmitActivityRecommendationLetterCtx(ctx context.Context, admission *models.ActivityRecommendationLetterAdmission) (err error) {
	activityIdBytes, err := uuid.Parse(admission.ActivityId)
	if err != nil {
		return ErrUuidInvalid
	}

	filepath := path.Join(ActivityRecommendationLetterSubdir, fmt.Sprintf("%s.pdf", admission.ActivityId))
	_, err = c.ActivityStorage.SaveActivityFiles(ctx, []string{filepath})
	if err != nil {
		if err == object.ErrTempFileNotFound {
			return ErrStorageFileNotFound
		}
		return ec.NewError(ErrCodeStorageCopyFail, Errs[ErrCodeStorageCopyFail], fmt.Errorf("cannot save recommendation letter: %w", err))
	}

	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)
	_, err = mdb.ExecContext(
		ctx,
		"insert into surat_rekomendasi_kegiatan(kegiatan_id, nosurat, tgl_surat, ttd_user_id, nama_doc) values ($1,$2,$3,$4,$5)",
		activityIdBytes,
		admission.RecommendationLetter.DocumentNumber,
		string(admission.RecommendationLetter.DocumentDate),
		admission.SignerAsnId,
		admission.RecommendationLetter.DocumentName,
	)
	if err != nil {
		return ec.NewError(ErrCodeExecFail, Errs[ErrCodeExecFail], fmt.Errorf("cannot insert entry to surat_rekomendasi_kegiatan: %w", err))
	}

	return nil
}

// GenerateActivityCertificateCtx generates a certificate and store it in the storage.
// Filename is generated from activityId and attendeeAsnId concatenated together. The base filename is returned,
// it can be accessed in the certificate subdir in permanent bucket.
//
// To help reduce performance load, it is only generated once, unless forceRegenerate is set to true.
func (c *Client) GenerateActivityCertificateCtx(ctx context.Context, activityId, attendeeAsnId string, forceRegenerate bool) (filename string, err error) {
	filename = fmt.Sprintf("%s-%s.pdf", activityId, attendeeAsnId)
	fullPath := path.Join(ActivityCertSubdir, filename)

	if !forceRegenerate {
		fileFound := true
		meta, err := c.ActivityStorage.GetActivityFileMetadata(ctx, fullPath)
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

	status := 0 // TODO: status should probably be checked, maybe to reject downloading certificates for certain status
	functionalPositionId := ""
	agencyId := ""
	isAccepted := false
	data := &ActivityCertificateTemplate{}
	err = mdb.QueryRowContext(
		ctx,
		"select k.nama, status, deskripsi, to_char(tgl_mulai, 'YYYY-MM-DD'), to_char(tgl_selesai, 'YYYY-MM-DD'), jabatan_jenjang, instansi_id, durasi, coalesce(instansi_penyelenggara, ''), no_usulan, nosurat, to_char(tgl_surat, 'YYYY-MM-DD'), isaccepted from sertifikat s join kegiatan k on s.persertakegiatan_kegiatan_id = k.kegiatan_id join perserta_kegiatan p on s.persertakegiatan_user_id = p.pegawai_user_id and s.persertakegiatan_kegiatan_id = p.kegiatan_kegiatan_id where k.kegiatan_id = $1 and s.persertakegiatan_user_id = $2",
		activityId,
		attendeeAsnId,
	).Scan(
		&data.ActivityName,
		&status,
		&data.Description,
		&data.StartDate,
		&data.EndDate,
		&functionalPositionId,
		&agencyId,
		&data.Duration,
		&data.OrganizerAgency,
		&data.AdmissionNumber,
		&data.DocumentNumber,
		&data.DocumentDate,
		&isAccepted,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrEntryNotFound
		}
		return "", ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	data.Qualification = "Ditolak"
	if isAccepted {
		data.Qualification = "Diterima"
	}

	agencies, err := c.getAgencyNames(ctx, referenceMtx, []string{agencyId})
	if err != nil {
		return "", err
	}

	positions, err := c.getFunctionalPositionNames(ctx, referenceMtx, []string{functionalPositionId})
	if err != nil {
		return "", err
	}

	detail, err := c.getUserDetailByAsnId(ctx, profileMtx, referenceMtx, attendeeAsnId, "")
	if err != nil {
		return "", err
	}

	if detail == nil {
		return "", ec.NewError(ErrCodeEntryNotFound, Errs[ErrCodeEntryNotFound], errors.New("ASN cannot be found in profile database"))
	}

	data.Agency = agencies[agencyId]
	data.FunctionalPosition = positions[functionalPositionId]
	data.AttendeeName = detail.Name
	data.AttendeeBirthday = detail.Birthday
	data.AttendeeNip = detail.NewNip
	// TODO: replace with real attendee image
	//data.AttendeePicture = &docx.InlineImage{ImageDescriptor: "2286BCE4-E5F7-11E5-A342-EFC6098C20DC_1.jpeg", Width: 10, Height: 20}

	err = c.generateActivityCertificateCtx(ctx, fullPath, data)
	if err != nil {
		if errors.Is(err, docx.ErrSiasnRendererBadTemplate) {
			return "", ec.NewError(ErrCodeDocumentGenerateBadTemplate, Errs[ErrCodeDocumentGenerateBadTemplate], err)
		}
		return "", ec.NewError(ErrCodeDocumentGenerate, Errs[ErrCodeDocumentGenerate], err)
	}

	return filename, nil
}

// GetActivityStatusStatisticCtx returns the number of activity items for each status.
func (c *Client) GetActivityStatusStatisticCtx(ctx context.Context) (statistics []*models.StatisticStatus, err error) {
	mdb := metricutil.NewDB(c.Db, c.SqlMetrics)
	rows, err := mdb.QueryContext(ctx, "select status, jumlah from kegiatan_status_statistik")
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer rows.Close()

	statistics = make([]*models.StatisticStatus, 0)
	statisticMap := make(map[int]*models.StatisticStatus)
	for status := range models.ActivityAdmissionStatuses {
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
