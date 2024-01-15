package store_test

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/if-itb/siasn-jf-backend/store/models"
	"github.com/if-itb/siasn-libs-backend/auth"
	"github.com/lib/pq"
	. "github.com/onsi/gomega"
)

func TestHandleAssessmentTeamAdmissionSubmit(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	rand.Seed(time.Now().UnixNano())

	asn := &auth.Asn{AsnId: uuid.NewString(), WorkAgencyId: uuid.NewString()}

	dummy := &models.AssessmentTeamAdmission{
		AdmissionNumber:      uuid.NewString(),
		FunctionalPositionId: uuid.NewString(),
		AdmissionDate:        models.Iso8601Date(time.Now().Format("2006-01-02")),
		TempSupportDocuments: []*models.Document{
			{
				Filename:     uuid.NewString(),
				DocumentName: uuid.NewString(),
			},
			{
				Filename:     uuid.NewString(),
				DocumentName: uuid.NewString(),
			},
		},
		Assessors: []*models.Assessor{
			{
				AsnId: uuid.NewString(),
				Role:  rand.Intn(3) + 1,
			}, {
				AsnId: uuid.NewString(),
				Role:  rand.Intn(3) + 1,
			}, {
				AsnId: uuid.NewString(),
				Role:  rand.Intn(3) + 1,
			},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("insert").WithArgs(
		sqlmock.AnyArg(),
		asn.AsnId,
		asn.WorkAgencyId,
		dummy.FunctionalPositionId,
		dummy.AdmissionDate,
		dummy.AdmissionNumber,
		models.AssessmentTeamStatusCreated,
		asn.AsnId,
	).WillReturnResult(sqlmock.NewResult(1, 0))

	assessorStmt := mock.ExpectPrepare("insert")
	for _, assessor := range dummy.Assessors {
		assessorStmt.ExpectExec().WithArgs(sqlmock.AnyArg(), assessor.AsnId, assessor.Role).WillReturnResult(sqlmock.NewResult(1, 0))
	}
	docStmt := mock.ExpectPrepare("insert")
	for _, _ = range dummy.TempSupportDocuments {
		docStmt.ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))
	}
	mock.ExpectCommit()

	payload, _ := json.Marshal(dummy)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/assessment-team/admission/submit", bytes.NewBuffer(payload))
	client.HandleAssessmentTeamAdmissionSubmit(rec, auth.InjectUserDetail(req, asn))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)

	result := &struct {
		AssessmentTeamId string `json:"tim_penilaian_id"`
	}{}
	MustJsonDecode(rec.Result().Body, result)

	Expect(result.AssessmentTeamId).ToNot(BeEmpty())
}

func TestHandleAssessmentTeamGet(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	profileDb, profileMock := MustCreateMock()
	referenceDb, referenceMock := MustCreateMock()
	client := CreateClientNoServer(db, profileDb, referenceDb)

	dummy := &models.AssessmentTeam{
		AssessmentTeamId:     uuid.NewString(),
		AdmissionNumber:      uuid.NewString(),
		FunctionalPositionId: uuid.NewString(),
		FunctionalPosition:   uuid.NewString(),
		Status:               rand.Intn(2) + 1,
		AgencyId:             uuid.NewString(),
		Agency:               uuid.NewString(),
		AdmissionDate:        models.Iso8601Date(time.Now().Format("2006-01-02")),
		SupportDocuments: []*models.Document{
			{
				Filename:     uuid.NewString(),
				DocumentName: uuid.NewString(),
				CreatedAt:    models.EpochTime(time.Unix(time.Now().Unix(), 0)),
			},
			{
				Filename:     uuid.NewString(),
				DocumentName: uuid.NewString(),
				CreatedAt:    models.EpochTime(time.Unix(time.Now().Unix(), 0)),
			},
		},
		Assessors: []*models.Assessor{
			{
				AsnId:          uuid.NewString(),
				Nip:            uuid.NewString(),
				Name:           uuid.NewString(),
				Role:           rand.Intn(3) + 1,
				Status:         rand.Intn(2) + 1,
				ReasonRejected: uuid.NewString(),
			},
			{
				AsnId:          uuid.NewString(),
				Nip:            uuid.NewString(),
				Name:           uuid.NewString(),
				Role:           rand.Intn(3) + 1,
				Status:         rand.Intn(2) + 1,
				ReasonRejected: uuid.NewString(),
			},
			{
				AsnId:          uuid.NewString(),
				Nip:            uuid.NewString(),
				Name:           uuid.NewString(),
				Role:           rand.Intn(3) + 1,
				Status:         rand.Intn(2) + 1,
				ReasonRejected: uuid.NewString(),
			},
		},
		RecommendationLetter: &models.Document{
			Filename:       uuid.NewString(),
			DocumentName:   uuid.NewString(),
			DocumentNumber: uuid.NewString(),
			DocumentDate:   models.Iso8601Date(time.Now().Format("2006-01-02")),
			CreatedAt:      models.EpochTime(time.Unix(time.Now().Unix(), 0)),
		},
	}

	rows := sqlmock.NewRows([]string{
		"instansi_id",
		"jabatan_fungsional_id",
		"tgl_usulan",
		"no_usulan",
		"status",
	})
	rows.AddRow(
		dummy.AgencyId,
		dummy.FunctionalPositionId,
		dummy.AdmissionDate,
		dummy.AdmissionNumber,
		dummy.Status,
	)

	referenceMock.ExpectQuery("select").WithArgs(pq.Array([]string{dummy.FunctionalPositionId})).WillReturnRows(sqlmock.NewRows([]string{"id", "nama"}).AddRow(dummy.FunctionalPositionId, dummy.FunctionalPosition))
	referenceMock.ExpectQuery("select").WithArgs(pq.Array([]string{dummy.AgencyId})).WillReturnRows(sqlmock.NewRows([]string{"id", "nama"}).AddRow(dummy.AgencyId, dummy.Agency))

	assessorDetailRows := sqlmock.NewRows([]string{"id", "nip", "nama"})

	for _, assessor := range dummy.Assessors {
		assessorDetailRows.AddRow(assessor.AsnId, assessor.Nip, assessor.Name)
	}

	profileMock.ExpectQuery("select").WithArgs(pq.Array([]string{dummy.Assessors[0].AsnId, dummy.Assessors[1].AsnId, dummy.Assessors[2].AsnId})).WillReturnRows(assessorDetailRows)

	assessorRows := sqlmock.NewRows([]string{"asn_id", "peran", "status", "alasan_ditolak"})
	for _, assessor := range dummy.Assessors {
		assessorRows.AddRow(assessor.AsnId, assessor.Role, assessor.Status, assessor.ReasonRejected)
	}
	documentRows := sqlmock.NewRows([]string{"filename", "nama_doc", "createdat"})
	documentRows.AddRow(dummy.SupportDocuments[0].Filename, dummy.SupportDocuments[0].DocumentName, time.Time(dummy.SupportDocuments[0].CreatedAt))
	documentRows.AddRow(dummy.SupportDocuments[1].Filename, dummy.SupportDocuments[1].DocumentName, time.Time(dummy.SupportDocuments[1].CreatedAt))

	letterRows := sqlmock.NewRows([]string{"filename", "nama_doc", "no_surat", "tgl_surat", "createdat"})
	letterRows.AddRow(dummy.RecommendationLetter.Filename, dummy.RecommendationLetter.DocumentName, dummy.RecommendationLetter.DocumentNumber, dummy.RecommendationLetter.DocumentDate, time.Time(dummy.RecommendationLetter.CreatedAt))

	mock.ExpectQuery("select").WithArgs(dummy.AssessmentTeamId).WillReturnRows(rows)
	mock.ExpectQuery("select").WithArgs(dummy.AssessmentTeamId).WillReturnRows(assessorRows)
	mock.ExpectQuery("select").WithArgs(dummy.AssessmentTeamId).WillReturnRows(documentRows)

	if dummy.Status == models.AssessmentTeamStatusVerified {
		mock.ExpectQuery("select").WithArgs(dummy.AssessmentTeamId).WillReturnRows(letterRows)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/assessment-team/get", nil)
	q := req.URL.Query()
	q.Add("tim_penilaian_id", dummy.AssessmentTeamId)
	req.URL.RawQuery = q.Encode()
	client.HandleAssessmentTeamGet(rec, req)

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)
	MustMockExpectationsMet(profileMock)
	MustMockExpectationsMet(referenceMock)

	var result *models.AssessmentTeam
	MustJsonDecode(rec.Result().Body, &result)

	Expect(result).To(Equal(dummy))
}

func TestHandleAssessmentTeamSearch(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	referenceDb, referenceMock := MustCreateMock()
	client := CreateClientNoServer(db, nil, referenceDb)

	asn := &auth.Asn{AsnId: uuid.NewString(), WorkAgencyId: uuid.NewString()}

	admissionDate := models.Iso8601Date(time.Now().Format("2006-01-02"))
	admissionStatus := rand.Intn(2) + 1
	countPerPage := rand.Intn(25) + 2
	pageNumber := rand.Intn(25) + 1

	dummy := []*models.AssessmentTeamItem{
		{
			AssessmentTeamId: uuid.NewString(),
			AdmissionNumber:  uuid.NewString(),
			Status:           rand.Intn(2) + 1,
			Agency:           uuid.NewString(),
			AgencyId:         uuid.NewString(),
			AdmissionDate:    models.Iso8601Date(time.Now().Format("2006-01-02")),
		},
		{
			AssessmentTeamId: uuid.NewString(),
			AdmissionNumber:  uuid.NewString(),
			Status:           rand.Intn(2) + 1,
			Agency:           uuid.NewString(),
			AgencyId:         uuid.NewString(),
			AdmissionDate:    models.Iso8601Date(time.Now().Format("2006-01-02")),
		},
	}

	rows := sqlmock.NewRows([]string{
		"tim_penilaian_id",
		"no_usulan",
		"instansi_id",
		"tgl_usulan",
		"status",
	})
	agencyRows := sqlmock.NewRows([]string{"id", "nama"})
	agencyIds := make([]string, 0)
	for _, d := range dummy {
		agencyIds = append(agencyIds, d.AgencyId)
		rows.AddRow(
			d.AssessmentTeamId,
			d.AdmissionNumber,
			d.AgencyId,
			d.AdmissionDate,
			d.Status,
		)
		agencyRows.AddRow(d.AgencyId, d.Agency)
	}

	mock.ExpectQuery("select").WithArgs(sqlmock.AnyArg(), admissionStatus, sqlmock.AnyArg(), admissionDate, countPerPage+1, (pageNumber-1)*countPerPage).WillReturnRows(rows)
	referenceMock.ExpectQuery("select").WithArgs(pq.Array(agencyIds)).WillReturnRows(agencyRows)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/assessment-team/search", nil)
	q := req.URL.Query()
	q.Set("tgl_usulan", string(admissionDate))
	q.Set("status", strconv.Itoa(admissionStatus))
	q.Set("jumlah_per_halaman", strconv.Itoa(countPerPage))
	q.Set("halaman", strconv.Itoa(pageNumber))
	req.URL.RawQuery = q.Encode()
	client.HandleAssessmentTeamSearch(rec, auth.InjectUserDetail(req, asn))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)

	MustMockExpectationsMet(mock)

	result := &models.IdPaginatedList{}
	MustJsonDecode(rec.Result().Body, result)

	Expect(result.Metadata.Subtotal).To(Equal(len(dummy)))
}

func TestHandleAssessmentTeamVerificationSubmit(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	rand.Seed(time.Now().UnixNano())

	asn := &auth.Asn{AsnId: uuid.NewString(), WorkAgencyId: uuid.NewString()}

	dummy := &models.AssessmentTeamVerification{
		AssessmentTeamId: uuid.NewString(),
		TempRecommendationLetter: &models.Document{
			DocumentName:   uuid.NewString(),
			DocumentNumber: uuid.NewString(),
			DocumentDate:   models.Iso8601Date(time.Now().Format("2006-01-02")),
		},
		Assessors: []*models.Assessor{
			{
				AsnId:          uuid.NewString(),
				Status:         rand.Intn(2) + 1,
				ReasonRejected: uuid.NewString(),
			},
			{
				AsnId:          uuid.NewString(),
				Status:         rand.Intn(2) + 1,
				ReasonRejected: uuid.NewString(),
			},
			{
				AsnId:          uuid.NewString(),
				Status:         rand.Intn(2) + 1,
				ReasonRejected: uuid.NewString(),
			},
		},
	}

	mock.ExpectBegin()
	mock.ExpectQuery("select").WithArgs(
		dummy.AssessmentTeamId,
	).WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(models.AssessmentTeamStatusCreated))

	now := time.Now()
	mock.ExpectQuery("update").WithArgs(
		models.AssessmentTeamStatusVerified,
		asn.AsnId,
		dummy.AssessmentTeamId,
	).WillReturnRows(sqlmock.NewRows([]string{"status_ts"}).AddRow(now))

	assessorStmt := mock.ExpectPrepare("update")
	for _, assessor := range dummy.Assessors {
		assessorStmt.ExpectExec().WithArgs(assessor.Status, assessor.ReasonRejected, dummy.AssessmentTeamId, assessor.AsnId).WillReturnResult(sqlmock.NewResult(1, 0))
	}

	mock.ExpectExec("insert").WithArgs(dummy.AssessmentTeamId, dummy.AssessmentTeamId, dummy.TempRecommendationLetter.DocumentName, dummy.TempRecommendationLetter.DocumentNumber, dummy.TempRecommendationLetter.DocumentDate).WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectCommit()

	payload, _ := json.Marshal(dummy)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/assessment-team/verification/submit", bytes.NewBuffer(payload))
	client.HandleAssessmentTeamVerificationSubmit(rec, auth.InjectUserDetail(req, asn))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)

	result := &struct {
		AssessmentTeamId string `json:"tim_penilaian_id"`
		UpdatedAt        int64  `json:"updated_at"`
	}{}
	MustJsonDecode(rec.Result().Body, result)

	Expect(result.AssessmentTeamId).To(Equal(dummy.AssessmentTeamId))
	Expect(result.UpdatedAt).To(Equal(now.Unix()))
}
