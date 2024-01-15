package store_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fazrithe/siasn-jf-backend-git/store"
	"github.com/fazrithe/siasn-jf-backend-git/store/models"
	"github.com/google/uuid"
	"github.com/if-itb/siasn-libs-backend/auth"
	"github.com/lib/pq"
	. "github.com/onsi/gomega"
)

func TestHandleDismissalAdmissionSubmit(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	profileDb, profileMock := MustCreateMock()
	client := CreateClientNoServer(db, profileDb, nil)

	dummy := &models.DismissalAdmission{
		AsnId:           uuid.New().String(),
		DismissalReason: "2",
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
		DecreeNumber:    uuid.NewString(),
		DecreeDate:      "2022-01-01",
		ReasonDetail:    uuid.NewString(),
		AdmissionNumber: uuid.NewString(),
	}

	profileMock.ExpectQuery("select").WithArgs(dummy.AsnId).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	mock.ExpectBegin()
	mock.ExpectExec("insert").WithArgs(
		sqlmock.AnyArg(),
		dummy.AsnId,
		sqlmock.AnyArg(),
		models.DismissalAdmissionStatusCreated,
		sqlmock.AnyArg(),
		dummy.DismissalReason,
		sql.NullString{Valid: true, String: dummy.DecreeNumber},
		sql.NullString{Valid: true, String: string(dummy.DecreeDate)},
		sql.NullString{Valid: true, String: dummy.ReasonDetail},
		dummy.AdmissionNumber,
	).WillReturnResult(sqlmock.NewResult(1, 0))
	stmt := mock.ExpectPrepare("insert")
	for _, d := range dummy.TempSupportDocuments {
		stmt.ExpectExec().WithArgs(sqlmock.AnyArg(), d.Filename, d.DocumentName).WillReturnResult(sqlmock.NewResult(1, 0))
	}
	mock.ExpectCommit()

	payload, _ := json.Marshal(dummy)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/dismissal/admission/submit", bytes.NewBuffer(payload))
	client.HandleDismissalAdmissionSubmit(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)

	MustMockExpectationsMet(mock)

	result := &models.DismissalAdmission{}
	MustJsonDecode(rec.Result().Body, result)

	Expect(result.DismissalId).ToNot(BeEmpty())
}

func TestHandleDismissalAdmissionGet(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	profileDb, profileMock := MustCreateMock()
	client := CreateClientNoServer(db, profileDb, nil)

	dismissalId := uuid.New().String()
	dummy := &models.DismissalAdmission{
		DismissalId:                   dismissalId,
		AsnId:                         uuid.NewString(),
		AsnName:                       uuid.NewString(),
		AsnNip:                        uuid.NewString(),
		Status:                        1,
		StatusTs:                      models.EpochTime{},
		StatusBy:                      uuid.New().String(),
		DismissalReason:               "2",
		DismissalLetter:               nil,
		DismissalLetterSignerAsnId:    "",
		DismissalDenyReason:           "",
		DismissalDenySupportDocuments: nil,
		DismissalDate:                 models.Iso8601Date(time.Now().Format("2006-01-02")),
		TempSupportDocuments:          nil,
		SupportDocuments: []*models.Document{
			{
				Filename:     uuid.New().String(),
				DocumentName: uuid.New().String(),
			},
		},
		SubmitterAsnId:  uuid.New().String(),
		AgencyId:        uuid.New().String(),
		DecreeNumber:    uuid.NewString(),
		DecreeDate:      "2021-01-01",
		ReasonDetail:    uuid.NewString(),
		AdmissionNumber: uuid.NewString(),
	}

	mock.ExpectBegin()
	mock.ExpectQuery("select").WithArgs(dismissalId, sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{
		"asn_id",
		"status",
		"status_ts",
		"status_by",
		"coalesce(alasan_pemberhentian, '')",
		"coalesce(alasan_tidak_diberhentikan, '')",
		"nama_doc_surat_pemberhentian",
		"nosurat_surat_pemberhentian",
		"coalesce(ttd_user_id_surat_pemberhentian, '')",
		"tgl_surat_pemberhentian",
		"tgl_pemberhentian",
		"nomor_sk",
		"tgl_sk",
		"detail_alasan",
		"no_usulan",
	}).
		AddRow(
			dummy.AsnId,
			dummy.Status,
			time.Time(dummy.StatusTs),
			dummy.StatusBy,
			dummy.DismissalReason,
			dummy.DismissalDenyReason,
			sql.NullString{},
			sql.NullString{},
			dummy.DismissalLetterSignerAsnId,
			sql.NullString{},
			dummy.DismissalDate,
			dummy.DecreeNumber,
			dummy.DecreeDate,
			dummy.ReasonDetail,
			dummy.AdmissionNumber,
		))
	mock.ExpectQuery("select").WithArgs(dismissalId).WillReturnRows(sqlmock.NewRows([]string{"filename", "nama_doc"}).AddRow(dummy.SupportDocuments[0].Filename, dummy.SupportDocuments[0].DocumentName))
	profileMock.ExpectQuery("select").WithArgs(pq.Array([]string{dummy.AsnId})).WillReturnRows(sqlmock.NewRows([]string{"id", "nama", "nip"}).AddRow(dummy.AsnId, dummy.AsnName, dummy.AsnNip))
	mock.ExpectCommit()

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/dismissal/admission/get?pemberhentian_id=%s", dismissalId), nil)
	client.HandleDismissalAdmissionGet(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustMockExpectationsMet(mock)

	MustStatusCodeEqual(rec.Result(), http.StatusOK)

	result := &models.DismissalAdmission{}
	MustJsonDecode(rec.Result().Body, result)

	Expect(result.DismissalId).To(Equal(dummy.DismissalId))
	Expect(result.StatusBy).To(Equal(dummy.StatusBy))
	Expect(result.SupportDocuments).To(Equal(dummy.SupportDocuments))
}

func TestHandleDismissalAdmissionsSearch(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)
	dismissalDate := ""
	status := 0

	dummy := []*models.DismissalAdmission{
		{
			DismissalId:     uuid.New().String(),
			AsnId:           uuid.New().String(),
			Status:          1,
			StatusTs:        models.EpochTime{},
			StatusBy:        uuid.New().String(),
			DismissalReason: uuid.New().String(),
			DismissalDate:   models.Iso8601Date("2006-10-10"),
		},
	}

	rows := sqlmock.NewRows([]string{"pemberhentian_id", "asn_id", "status", "status_ts", "status_by", "alasan_pemberhentian", "alasan_tidak_diberhentikan", "tgl_pemberhentian"})
	for _, d := range dummy {
		rows.AddRow(d.DismissalId, d.AsnId, d.Status, time.Time(d.StatusTs), d.StatusBy, d.DismissalReason, d.DismissalDenyReason, string(d.DismissalDate))
	}

	mock.ExpectQuery("select").WithArgs(dismissalDate, status, sqlmock.AnyArg()).WillReturnRows(rows)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/dismissal/accept/search", nil)
	client.HandleDismissalAdmissionsSearch(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustMockExpectationsMet(mock)

	MustStatusCodeEqual(rec.Result(), http.StatusOK)

	result := make([]*models.DismissalAdmission, 0)
	MustJsonDecode(rec.Result().Body, &result)

	Expect(result).To(HaveLen(len(dummy)))
}

func TestHandleDismissalAcceptSet(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	profileDb, profileMock := MustCreateMock()
	referenceDb, referenceMock := MustCreateMock()
	client := CreateClientNoServer(db, profileDb, referenceDb)
	client.DocxRenderer = &mockDocxRenderer{}

	dummy := &models.DismissalAcceptanceRequest{
		DismissalId:                uuid.New().String(),
		DismissalLetterSignerAsnId: uuid.New().String(),
		DismissalLetter: &models.Document{
			Filename:       uuid.New().String(),
			DocumentName:   uuid.New().String(),
			DocumentNumber: uuid.New().String(),
			DocumentDate:   models.Iso8601Date(time.Now().Format("2006-01-02")),
		},
	}

	asnId := uuid.NewString()
	data := &store.DismissalAcceptanceTemplate{
		DocumentNumber:   dummy.DismissalLetter.DocumentNumber,
		DocumentDate:     string(dummy.DismissalLetter.DocumentDate),
		DecreeNumber:     uuid.NewString(),
		DecreeDate:       time.Now().Format("2006-01-02"),
		DismissalDate:    time.Now().Format("2006-01-02"),
		DismissalReason:  uuid.NewString(),
		AsnName:          uuid.NewString(),
		AsnNip:           uuid.NewString(),
		AsnGrade:         uuid.NewString(),
		Position:         uuid.NewString(),
		OrganizationUnit: uuid.NewString(),
	}

	mock.ExpectBegin()
	mock.ExpectQuery("select").WithArgs(dummy.DismissalLetterSignerAsnId, models.StaffRoleSupervisor, sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("update").WithArgs(
		models.DismissalAdmissionStatusAccepted,
		sqlmock.AnyArg(),
		dummy.DismissalLetter.DocumentName,
		dummy.DismissalLetter.DocumentNumber,
		dummy.DismissalLetterSignerAsnId,
		string(dummy.DismissalLetter.DocumentDate),
		dummy.DismissalId,
		models.DismissalAdmissionStatusCreated,
	).WillReturnRows(sqlmock.NewRows([]string{"status_ts"}).AddRow(time.Now()))
	mock.ExpectQuery("select").WithArgs(dummy.DismissalId).WillReturnRows(sqlmock.NewRows([]string{"asn_id", "status", "coalesce(alasan_pemberhentian, '')", "nosurat_surat_pemberhentian", "to_char(tgl_surat_pemberhentian, 'YYYY-MM-DD')", "to_char(tgl_pemberhentian, 'YYYY-MM-DD')", "coalesce(nomor_sk, '')", "to_char(tgl_sk, 'YYYY-MM-DD')"}).AddRow(
		asnId,
		models.DismissalAdmissionStatusAccepted,
		data.DismissalReason,
		data.DocumentNumber,
		data.DocumentDate,
		data.DismissalDate,
		data.DecreeNumber,
		data.DecreeDate,
	))
	profileMock.ExpectQuery("select").WithArgs(asnId, "").WillReturnRows(sqlmock.NewRows([]string{
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
	}).AddRow(asnId, data.AsnNip, "", data.AsnName, "", "", "", "", "", "", "", "", "", 0, uuid.NewString(), uuid.NewString()))
	referenceMock.ExpectQuery("select").WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"nama_unor", "coalesce(nama_jabatan, '')"}).AddRow(data.OrganizationUnit, data.Position))
	referenceMock.ExpectQuery("select").WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"nama", "nama_pangkat"}).AddRow(uuid.NewString(), data.AsnGrade))
	mock.ExpectCommit()

	payload, _ := json.Marshal(dummy)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/dismissal/accept/submit", bytes.NewBuffer(payload))
	client.HandleDismissalAcceptSet(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)

	MustMockExpectationsMet(mock)

	result := &models.DismissalAdmission{}
	MustJsonDecode(rec.Result().Body, result)

	Expect(result.DismissalId).ToNot(BeEmpty())
}

func TestHandleDismissalDenySet(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	dummy := &models.DismissalDenyRequest{
		DismissalId:         uuid.New().String(),
		DismissalDenyReason: uuid.New().String(),
		TempDismissalDenySupportDocuments: []*models.Document{
			{
				Filename: uuid.New().String(),
			},
		},
	}

	mock.ExpectBegin()
	mock.ExpectQuery("update").WithArgs(
		models.DismissalAdmissionStatusRejected,
		sqlmock.AnyArg(),
		dummy.DismissalDenyReason,
		dummy.DismissalId,
		models.DismissalAdmissionStatusCreated,
	).WillReturnRows(sqlmock.NewRows([]string{"status_ts"}).AddRow(time.Now()))
	docStmt := mock.ExpectPrepare("insert")
	for _, d := range dummy.TempDismissalDenySupportDocuments {
		docStmt.ExpectExec().WithArgs(dummy.DismissalId, d.Filename).WillReturnResult(sqlmock.NewResult(1, 0))
	}
	mock.ExpectCommit()

	payload, _ := json.Marshal(dummy)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/dismissal/deny/submit", bytes.NewBuffer(payload))
	client.HandleDismissalDenySet(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)

	MustMockExpectationsMet(mock)

	result := &models.DismissalAdmission{}
	MustJsonDecode(rec.Result().Body, result)

	Expect(result.DismissalId).ToNot(BeEmpty())
}

func TestHandleGetDismissalStatusStatistic(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	statisticRows := sqlmock.NewRows([]string{"status", "jumlah"}).AddRow(1, 0).AddRow(2, 1)

	mock.ExpectQuery("select").WillReturnRows(statisticRows)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/dismissal/statistic/status/get", nil)
	client.HandleGetDismissalStatusStatistic(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)
}
