package store_test

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/if-itb/siasn-jf-backend/store"
	"github.com/if-itb/siasn-jf-backend/store/models"
	"github.com/if-itb/siasn-libs-backend/auth"
	"github.com/lib/pq"
	. "github.com/onsi/gomega"
)

func TestHandleActivityAdmissionSubmit(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	profileDb, profileMock := MustCreateMock()
	client := CreateClientNoServer(db, profileDb, nil)

	rand.Seed(time.Now().UnixNano())
	asns := []*auth.Asn{
		{AsnId: uuid.New().String(), NewNip: strconv.Itoa(rand.Intn(10000)), OldNip: strconv.Itoa(rand.Intn(10000))},
		{AsnId: uuid.New().String(), NewNip: strconv.Itoa(rand.Intn(10000)), OldNip: strconv.Itoa(rand.Intn(10000))},
	}

	dummy := &models.ActivityAdmission{
		ActivityId:    uuid.New().String(),
		Name:          time.Now().Format("2006"),
		Status:        models.ActivityAdmissionStatusCreated,
		Type:          models.ActivityTypeWorkshop,
		Description:   uuid.New().String(),
		StartDate:     "2020-01-01",
		EndDate:       "2020-01-01",
		PositionGrade: uuid.New().String(),
		AgencyId:      uuid.New().String(),
		Extra:         uuid.New().String(),
		TempSupportDocuments: []*models.Document{
			{
				Filename:     uuid.New().String(),
				DocumentName: uuid.New().String(),
			},
			{
				Filename:     uuid.New().String(),
				DocumentName: uuid.New().String(),
			},
		},
		Attendees:       []string{asns[0].AsnId, asns[1].AsnId},
		TrainingYear:    time.Now().Year(),
		Duration:        time.Now().Second(),
		OrganizerAgency: "",
		AdmissionNumber: uuid.NewString(),
	}

	asnRows := sqlmock.NewRows([]string{"id", "nip_baru", "nip_lama"})
	asnRows.AddRow(asns[0].AsnId, asns[0].NewNip, asns[0].OldNip)
	asnRows.AddRow(asns[1].AsnId, asns[1].NewNip, asns[1].OldNip)

	profileMock.ExpectQuery("select").WithArgs(pq.Array(dummy.Attendees), sqlmock.AnyArg()).WillReturnRows(asnRows)

	mock.ExpectBegin()
	mock.ExpectExec("insert").WithArgs(
		sqlmock.AnyArg(),
		dummy.Name,
		models.ActivityAdmissionStatusCreated,
		dummy.Type,
		dummy.Description,
		sqlmock.AnyArg(),
		string(dummy.StartDate),
		string(dummy.EndDate),
		dummy.PositionGrade,
		sqlmock.AnyArg(),
		dummy.Extra,
		dummy.TrainingYear,
		dummy.Duration,
		sql.NullString{Valid: dummy.OrganizerAgency != "", String: dummy.OrganizerAgency},
		dummy.AdmissionNumber,
	).WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectExec("insert").WithArgs(sqlmock.AnyArg(), models.ActivityAdmissionStatusCreated, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))
	pegawaiStmt := mock.ExpectPrepare("insert")
	pesertaKegiatanStmt := mock.ExpectPrepare("insert")
	for _, asn := range asns {
		pegawaiStmt.ExpectExec().WithArgs(asn.AsnId, asn.NewNip, asn.OldNip).WillReturnResult(sqlmock.NewResult(1, 0))
		pesertaKegiatanStmt.ExpectExec().WithArgs(sqlmock.AnyArg(), asn.AsnId).WillReturnResult(sqlmock.NewResult(1, 0))
	}
	docStmt := mock.ExpectPrepare("insert")
	for _, _ = range dummy.TempSupportDocuments {
		docStmt.ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))
	}
	mock.ExpectCommit()

	payload, _ := json.Marshal(dummy)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/activity/admission/submit", bytes.NewBuffer(payload))
	client.HandleActivityAdmissionSubmit(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)

	MustMockExpectationsMet(profileMock)
	MustMockExpectationsMet(mock)

	result := &models.ActivityAdmission{}
	MustJsonDecode(rec.Result().Body, result)

	Expect(result.ActivityId).ToNot(BeEmpty())
}

func TestHandleActivityStatusCsrSet(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	dummy := &models.ActivityCsrRequest{
		ActivityId:     uuid.New().String(),
		SubmitterAsnId: uuid.New().String(),
		AttendeesPassing: []*models.ActivityCsrAttendee{
			{
				AsnId:          uuid.New().String(),
				IsPassing:      false,
				ReasonRejected: uuid.New().String(),
			},
		},
	}

	mock.ExpectBegin()
	mock.ExpectQuery("select").WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(models.ActivityAdmissionStatusAccepted))
	for _, asn := range dummy.AttendeesPassing {
		mock.ExpectQuery("update").WithArgs(asn.IsPassing, sqlmock.AnyArg(), sqlmock.AnyArg(), sql.NullString{String: asn.ReasonRejected, Valid: !asn.IsPassing}, dummy.ActivityId, asn.AsnId).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	}
	mock.ExpectExec("update").WithArgs(models.ActivityAdmissionStatusCertRequest, dummy.ActivityId, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectExec("insert").WithArgs(dummy.ActivityId, models.ActivityAdmissionStatusCertRequest, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectCommit()

	payload, _ := json.Marshal(dummy)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/activity/csr/set", bytes.NewBuffer(payload))
	client.HandleActivityStatusCsrSubmit(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)

	result := map[string]interface{}{}
	MustJsonDecode(rec.Result().Body, &result)

	Expect(result).To(HaveKey("kegiatan_id"))
	Expect(result).To(HaveKey("modified_at"))
	Expect(result["kegiatan_id"]).To(Equal(dummy.ActivityId))
}

func TestHandleActivityCertGenSubmit(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	acg := &models.ActivityCertGenRequest{
		ActivityId: uuid.New().String(),
		AttendeesPassing: []*models.ActivityAttendeeCsr{
			{
				DocumentNumber: uuid.New().String(),
				DocumentDate:   models.Iso8601Date(time.Now().Format("2006-01-02")),
				SignerAsnId:    uuid.New().String(),
				AttendeeAsnId:  uuid.New().String(),
				Type:           models.ActivityCertTypeCert,
			},
		},
	}

	mock.ExpectBegin()
	mock.ExpectQuery("select").WillReturnRows(sqlmock.NewRows([]string{"status", "jenis"}).AddRow(models.ActivityAdmissionStatusAccepted, models.ActivityTypeAccreditation))
	stmt := mock.ExpectPrepare("insert")
	attendeeStmt := mock.ExpectPrepare("update")
	for _, doc := range acg.AttendeesPassing {
		attendeeStmt.ExpectQuery().WithArgs(doc.IsPassing, sql.NullString{Valid: !doc.IsPassing, String: doc.ReasonRejected}, sqlmock.AnyArg(), doc.AttendeeAsnId, acg.ActivityId).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
		if doc.IsPassing {
			stmt.ExpectExec().WithArgs(acg.ActivityId, doc.AttendeeAsnId, doc.DocumentNumber, string(doc.DocumentDate), doc.Type, doc.SignerAsnId, sql.NullFloat64{Valid: doc.Score <= 0, Float64: float64(doc.Score)}).WillReturnResult(sqlmock.NewResult(1, 0))
		}
	}
	mock.ExpectExec("update").WithArgs(models.ActivityAdmissionStatusCertPublished, acg.ActivityId, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectExec("insert").WithArgs(acg.ActivityId, models.ActivityAdmissionStatusCertPublished, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))

	mock.ExpectCommit()

	payload, _ := json.Marshal(acg)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/activity/certgen/submit", bytes.NewBuffer(payload))
	client.HandleActivityCertGenSubmit(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)

	result := map[string]interface{}{}
	MustJsonDecode(rec.Result().Body, &result)

	Expect(result).To(HaveKey("kegiatan_id"))
	Expect(result).To(HaveKey("modified_at"))
	Expect(result["kegiatan_id"]).To(Equal(acg.ActivityId))
}

func TestHandleActivityAdmissionSearch(t *testing.T) {
	t.Skip("not tested against the paginated version")

	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	admissionDate := "2006-01-02"
	admissionStatus := rand.Intn(4) + 1
	admissionType := rand.Intn(6) + 1

	dummy := &models.ActivityAdmission{
		ActivityId:         uuid.New().String(),
		Name:               time.Now().Format("2006"),
		Status:             admissionStatus,
		Type:               admissionType,
		Description:        uuid.New().String(),
		AdmissionTimestamp: models.EpochTime(time.Unix(time.Now().Unix(), 0)),
		StartDate:          "2020-01-01",
		EndDate:            "2020-01-01",
		PositionGrade:      uuid.New().String(),
		AgencyId:           uuid.New().String(),
		Extra:              uuid.New().String(),
	}

	rows := sqlmock.NewRows([]string{"kegiatan_id", "nama", "status", "jenis", "deskripsi", "tgl_usulan", "tgl_mulai", "tgl_selesai", "jabatan_jenjang", "instansi_id", "data_tambahan", "tahun_diklat", "durasi", "instansi_penyelenggara", "no_usulan"})
	rows.AddRow(
		dummy.ActivityId,
		dummy.Name,
		dummy.Status,
		dummy.Type,
		dummy.Description,
		time.Time(dummy.AdmissionTimestamp),
		dummy.StartDate,
		dummy.EndDate,
		dummy.PositionGrade,
		dummy.AgencyId,
		dummy.Extra,
		dummy.TrainingYear,
		dummy.Duration,
		sql.NullString{Valid: dummy.OrganizerAgency != "", String: dummy.OrganizerAgency},
		dummy.AdmissionNumber,
	)

	mock.ExpectQuery("select").WillReturnRows(rows)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/activity/admission/search", nil)
	q := req.URL.Query()
	q.Add("tgl_usulan", admissionDate)
	q.Add("status", strconv.Itoa(admissionStatus))
	q.Add("jenis", strconv.Itoa(admissionType))
	req.URL.RawQuery = q.Encode()
	client.HandleActivityAdmissionSearch(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)

	var result []*models.ActivityAdmission
	MustJsonDecode(rec.Result().Body, &result)

	Expect(result).To(HaveLen(1))
	Expect(result[0].ActivityId).To(Equal(dummy.ActivityId))
}

func TestHandleActivityAdmissionDetail(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	profileDb, profileMock := MustCreateMock()
	client := CreateClientNoServer(db, profileDb, nil)

	activityId := uuid.New().String()

	dummy := &models.ActivityAdmission{
		ActivityId:         activityId,
		Name:               time.Now().Format("2006"),
		Status:             rand.Intn(4) + 1,
		Type:               rand.Intn(6) + 1,
		Description:        uuid.New().String(),
		AdmissionTimestamp: models.EpochTime(time.Unix(time.Now().Unix(), 0)), // Prevents conversion difference to Unix epoch
		StartDate:          "2020-01-01",
		EndDate:            "2020-01-01",
		PositionGrade:      uuid.New().String(),
		AgencyId:           "",
		Extra:              uuid.New().String(),
		AttendeesDetail: []*models.ActivityAttendee{
			{
				AsnId:          uuid.New().String(),
				Nip:            uuid.NewString(),
				IsAccepted:     rand.Int31()%2 == 2,
				ReasonRejected: uuid.New().String(),
				AsnName:        uuid.New().String(),
			},
		},
		SupportDocuments: []*models.Document{
			{
				Filename:     uuid.New().String(),
				DocumentName: uuid.New().String(),
			},
		},
	}

	activityRows := sqlmock.NewRows([]string{
		"kegiatan_id",
		"nama",
		"status",
		"jenis",
		"deskripsi",
		"tgl_usulan",
		"tgl_mulai",
		"tgl_selesai",
		"jabatan_jenjang",
		"instansi_id",
		"data_tambahan",
		"tahun_diklat",
		"durasi",
		"instansi_penyelenggara",
		"no_usulan",
	})
	activityRows.AddRow(
		dummy.ActivityId,
		dummy.Name,
		dummy.Status,
		dummy.Type,
		dummy.Description,
		time.Time(dummy.AdmissionTimestamp),
		dummy.StartDate,
		dummy.EndDate,
		dummy.PositionGrade,
		dummy.AgencyId,
		dummy.Extra,
		dummy.TrainingYear,
		dummy.Duration,
		dummy.OrganizerAgency,
		dummy.AdmissionNumber,
	)

	// Set all certificate fields to null.
	attendeeRows := sqlmock.NewRows([]string{"pegawai_user_id", "isaccepted", "accepted_rejected_reason", "ispass", "pass_rejected_reason", "s.nosurat", "s.tgl_surat", "s.jenis", "s.ttd_user_id", "s.createdat", "nilai"})
	attendeeRows.AddRow(dummy.AttendeesDetail[0].AsnId, dummy.AttendeesDetail[0].IsAccepted, dummy.AttendeesDetail[0].ReasonRejected, dummy.AttendeesDetail[0].IsPassing, dummy.AttendeesDetail[0].ReasonNotPassing, sql.NullString{}, sql.NullString{}, sql.NullInt32{}, sql.NullString{}, sql.NullTime{}, sql.NullFloat64{})

	documentRows := sqlmock.NewRows([]string{"filename", "nama_doc"})
	documentRows.AddRow(dummy.SupportDocuments[0].Filename, dummy.SupportDocuments[0].DocumentName)

	attendeeNameRows := sqlmock.NewRows([]string{"id", "nip", "nama"})
	attendeeNameRows.AddRow(dummy.AttendeesDetail[0].AsnId, dummy.AttendeesDetail[0].Nip, dummy.AttendeesDetail[0].AsnName)

	mock.ExpectBegin()
	mock.ExpectQuery("select").WithArgs(driver.Value(activityId)).WillReturnRows(activityRows)
	mock.ExpectQuery("select").WithArgs(driver.Value(activityId)).WillReturnRows(attendeeRows)
	mock.ExpectQuery("select").WithArgs(driver.Value(activityId)).WillReturnRows(documentRows)
	mock.ExpectCommit()

	profileMock.ExpectQuery("select").WithArgs(driver.Value(pq.Array([]string{dummy.AttendeesDetail[0].AsnId}))).WillReturnRows(attendeeNameRows)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/activity/admission/detail", nil)
	q := req.URL.Query()
	q.Add("kegiatan_id", activityId)
	req.URL.RawQuery = q.Encode()
	client.HandleActivityAdmissionDetail(rec, req)

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)

	var result *models.ActivityAdmission
	MustJsonDecode(rec.Result().Body, &result)

	//Expect(result).To(Equal(dummy))
}

func TestHandleActivityVerificationSet(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	activityId := uuid.New().String()

	mock.ExpectBegin()
	mock.ExpectQuery("select").WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(models.ActivityAdmissionStatusCreated))
	mock.ExpectQuery("update").WithArgs(true, sqlmock.AnyArg(), sqlmock.AnyArg(), sql.NullString{String: "REASON", Valid: false}, activityId, "TESTID").WillReturnRows(sqlmock.NewRows([]string{"isaccepted"}).AddRow(true))
	mock.ExpectExec("update").WithArgs(models.ActivityAdmissionStatusAccepted, activityId, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectExec("insert").WithArgs(activityId, models.ActivityAdmissionStatusAccepted, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectCommit()

	payload, _ := json.Marshal(&models.ActivityVerificationRequest{
		ActivityId: activityId,
		AttendeesAcceptance: []*models.ActivityAttendee{{
			AsnId:          "TESTID",
			IsAccepted:     true,
			ReasonRejected: "REASON",
		}},
		SubmitterAsnId: "TEST",
	})

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/activity/admission/verify", bytes.NewBuffer(payload))
	client.HandleActivityVerificationSet(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)

	result := map[string]interface{}{}
	MustJsonDecode(rec.Result().Body, &result)

	Expect(result).To(HaveKey("kegiatan_id"))
	Expect(result).To(HaveKey("modified_at"))
	Expect(result["kegiatan_id"]).To(Equal(activityId))
}

func TestHandleActivityAdmissionSearchPembina(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	agencyId := uuid.NewString()
	admissionDate := "2006-01-02"
	admissionStatus := rand.Intn(4) + 1
	admissionType := rand.Intn(6) + 1

	dummy := &models.ActivityAdmission{
		ActivityId:         uuid.New().String(),
		Name:               time.Now().Format("2006"),
		Status:             admissionStatus,
		Type:               admissionType,
		Description:        uuid.New().String(),
		AdmissionTimestamp: models.EpochTime(time.Unix(time.Now().Unix(), 0)),
		StartDate:          "2020-01-01",
		EndDate:            "2020-01-01",
		PositionGrade:      uuid.New().String(),
		AgencyId:           agencyId,
		Extra:              uuid.New().String(),
	}

	rows := sqlmock.NewRows([]string{"kegiatan_id", "nama", "status", "jenis", "deskripsi", "tgl_usulan", "tgl_mulai", "tgl_selesai", "jabatan_jenjang", "instansi_id", "data_tambahan"})
	rows.AddRow(dummy.ActivityId, dummy.Name, dummy.Status, dummy.Type, dummy.Description, time.Time(dummy.AdmissionTimestamp), dummy.StartDate, dummy.EndDate, dummy.PositionGrade, dummy.AgencyId, dummy.Extra)

	mock.ExpectQuery("select").WillReturnRows(rows)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/activity/admission/search-pembina", nil)
	q := req.URL.Query()
	q.Add("instansi_id", agencyId)
	q.Add("tgl_usulan", admissionDate)
	q.Add("status", strconv.Itoa(admissionStatus))
	q.Add("jenis", strconv.Itoa(admissionType))
	req.URL.RawQuery = q.Encode()
	client.HandleActivityAdmissionSearchPembina(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)

	var result []*models.ActivityAdmission
	MustJsonDecode(rec.Result().Body, &result)

	Expect(result).To(HaveLen(1))
	Expect(result[0].ActivityId).To(Equal(dummy.ActivityId))
}

func TestHandleActivityCertGenDocDownload(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	profileDb, profileMock := MustCreateMock()
	referenceDb, referenceMock := MustCreateMock()
	client := CreateClientNoServer(db, profileDb, referenceDb)
	client.DocxRenderer = &mockDocxRenderer{}

	activityId := uuid.NewString()
	attendeeAsnId := uuid.NewString()
	dummy := &store.ActivityCertificateTemplate{
		AttendeeName:       uuid.NewString(),
		AttendeeNip:        uuid.NewString(),
		AttendeeBirthday:   uuid.NewString(),
		AttendeeBirthplace: uuid.NewString(),
		FunctionalPosition: uuid.NewString(),
		Agency:             uuid.NewString(),
		OrganizerAgency:    uuid.NewString(),
		Qualification:      "Terima",
		ActivityName:       uuid.NewString(),
		Duration:           "10",
		AdmissionNumber:    uuid.NewString(),
		StartDate:          uuid.NewString(),
		EndDate:            uuid.NewString(),
		Description:        uuid.NewString(),
		DocumentNumber:     uuid.NewString(),
		DocumentDate:       uuid.NewString(),
	}

	functionalPositionId := uuid.NewString()
	agencyId := uuid.NewString()
	referenceMock.ExpectBegin()
	profileMock.ExpectBegin()
	mock.ExpectQuery("select").WithArgs(activityId, attendeeAsnId).WillReturnRows(sqlmock.NewRows([]string{
		"k.nama",
		"status",
		"deskripsi",
		"tgl_mulai",
		"tgl_selesai",
		"jabatan_jenjang",
		"instansi_id",
		"durasi",
		"coalesce(instansi_penyelenggara, '')",
		"no_usulan",
		"nosurat",
		"tgl_surat",
		"isaccepted",
	}).AddRow(
		dummy.AttendeeName,
		0,
		dummy.Description,
		dummy.StartDate,
		dummy.EndDate,
		functionalPositionId,
		agencyId,
		dummy.Duration,
		dummy.OrganizerAgency,
		dummy.AdmissionNumber,
		dummy.DocumentNumber,
		dummy.DocumentDate,
		true,
	))
	referenceMock.ExpectQuery("select").WithArgs(pq.Array([]string{agencyId})).WillReturnRows(sqlmock.NewRows([]string{"id", "nama"}).AddRow(agencyId, dummy.Agency))
	referenceMock.ExpectQuery("select").WithArgs(pq.Array([]string{functionalPositionId})).WillReturnRows(sqlmock.NewRows([]string{"id", "nama"}).AddRow(functionalPositionId, dummy.FunctionalPosition))
	profileMock.ExpectQuery("select").WithArgs(attendeeAsnId, "").WillReturnRows(sqlmock.NewRows([]string{
		"pns.id",
		"nip_baru",
		"coalesce(nip_lama, '')",
		"coalesce(nama, '')",
		"coalesce(nomor_id_document, '')",
		"coalesce(nomor_hp, '')",
		"tgl_lhr",
		"instansi_induk_id",
		"coalesce(instansi_induk_nama, '')",
		"instansi_kerja_id",
		"coalesce(instansi_kerja_nama, '')",
		"coalesce(jabatan_fungsional_id, '')",
		"coalesce(jabatan_fungsional_umum_id, '')",
		"jenis_jabatan_id",
		"coalesce(unor_id, '')",
		"golongan_id",
	}).AddRow(attendeeAsnId, dummy.AttendeeNip, "", dummy.AttendeeName, "", "", dummy.AttendeeBirthday, "", "", "", "", "", "", 0, uuid.NewString(), uuid.NewString()))
	referenceMock.ExpectQuery("select").WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"nama_unor", "coalesce(nama_jabatan, '')"}).AddRow(uuid.NewString(), uuid.NewString()))
	referenceMock.ExpectQuery("select").WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"nama", "nama_pangkat"}).AddRow(uuid.NewString(), uuid.NewString()))
	referenceMock.ExpectCommit()
	profileMock.ExpectCommit()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/activity/certgen/download", nil)
	q := req.URL.Query()
	q.Add("kegiatan_id", activityId)
	q.Add("peserta_user_id", attendeeAsnId)
	q.Add("force", strconv.FormatBool(true))
	req.URL.RawQuery = q.Encode()
	client.HandleActivityCertGenDocDownload(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusFound)
	MustMockExpectationsMet(mock)
	MustMockExpectationsMet(referenceMock)
	MustMockExpectationsMet(profileMock)
}

func TestHandleGetActivityStatusStatistic(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	statisticRows := sqlmock.NewRows([]string{"status", "jumlah"}).AddRow(1, 0).AddRow(2, 1)

	mock.ExpectQuery("select").WillReturnRows(statisticRows)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/statistic/activity/status/get", nil)
	client.HandleGetActivityStatusStatistic(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)
}
