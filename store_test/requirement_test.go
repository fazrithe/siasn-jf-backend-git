package store_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fazrithe/siasn-jf-backend-git/store/models"
	"github.com/google/uuid"
	"github.com/if-itb/siasn-libs-backend/auth"
	"github.com/lib/pq"
	. "github.com/onsi/gomega"
)

func TestHandleRequirementAdmissionSubmit(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	profileDb, profileMock := MustCreateMock()
	referenceDb, referenceMock := MustCreateMock()
	client := CreateClientNoServer(db, profileDb, referenceDb)

	dummy := &models.RequirementAdmission{
		AdmissionTimestamp: models.EpochTime(time.Now()),
		PositionGrade:      uuid.New().String(),
		AgencyId:           uuid.New().String(),
		FiscalYear:         uuid.NewString(),
		AdmissionNumber:    uuid.NewString(),
		RequirementCounts: []*models.RequirementCount{
			{
				OrganizationUnitId: uuid.NewString(),
				Count:              10,
			},
			{
				OrganizationUnitId: uuid.NewString(),
				Count:              12,
			},
		},
		TempCoverLetter: &models.Document{
			Filename:     uuid.New().String(),
			DocumentName: uuid.New().String(),
		},
		TempEstimationDocuments: []*models.Document{
			{
				Filename:     uuid.New().String(),
				DocumentName: uuid.New().String(),
			},
			{
				Filename:     uuid.New().String(),
				DocumentName: uuid.New().String(),
			},
		},
	}

	unorRows := sqlmock.NewRows([]string{"id"})
	unorIds := make([]string, 0)
	for _, rc := range dummy.RequirementCounts {
		unorIds = append(unorIds, rc.OrganizationUnitId)
		unorRows.AddRow(rc.OrganizationUnitId)
	}
	referenceMock.ExpectQuery("select").WithArgs(pq.Array(unorIds), sqlmock.AnyArg()).WillReturnRows(unorRows)
	mock.ExpectBegin()
	mock.ExpectExec("insert").WithArgs(
		sqlmock.AnyArg(),
		sqlmock.AnyArg(),
		models.RequirementAdmissionStatusCreated,
		dummy.PositionGrade,
		sqlmock.AnyArg(),
		dummy.TempCoverLetter.DocumentName,
		sqlmock.AnyArg(),
		dummy.FiscalYear,
		dummy.AdmissionNumber,
	).WillReturnResult(sqlmock.NewResult(1, 0))
	profileMock.ExpectQuery("select").WithArgs(dummy.PositionGrade, pq.Array(unorIds)).WillReturnRows(sqlmock.NewRows([]string{"jabatan_fungsional_id", "unor_id", "count(*)"})) // Assume that no rows are returned, this could work too
	reqCountStmt := mock.ExpectPrepare("insert")
	for _, rc := range dummy.RequirementCounts {
		reqCountStmt.ExpectExec().WithArgs(sqlmock.AnyArg(), rc.OrganizationUnitId, rc.Count, 0).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	mock.ExpectExec("insert").WithArgs(sqlmock.AnyArg(), models.RequirementAdmissionStatusCreated, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))
	stmt := mock.ExpectPrepare("insert")
	for _, _ = range dummy.TempEstimationDocuments {
		stmt.ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))
	}
	mock.ExpectCommit()

	payload, _ := json.Marshal(dummy)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/requirement/admission/submit", bytes.NewBuffer(payload))
	client.HandleRequirementAdmissionSubmit(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)

	MustMockExpectationsMet(mock)

	result := &models.RequirementAdmission{}
	MustJsonDecode(rec.Result().Body, result)

	Expect(result.RequirementId).ToNot(BeEmpty())
}

func TestHandleRequirementAdmissionEdit(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	profileDb, profileMock := MustCreateMock()
	referenceDb, referenceMock := MustCreateMock()
	client := CreateClientNoServer(db, profileDb, referenceDb)

	dummy := &models.RequirementAdmission{
		RequirementId:      uuid.New().String(),
		AdmissionTimestamp: models.EpochTime(time.Now()),
		PositionGrade:      uuid.New().String(),
		AgencyId:           uuid.New().String(),
		FiscalYear:         uuid.NewString(),
		AdmissionNumber:    uuid.NewString(),
		RequirementCounts: []*models.RequirementCount{
			{
				OrganizationUnitId: uuid.NewString(),
				Count:              10,
			},
			{
				OrganizationUnitId: uuid.NewString(),
				Count:              12,
			},
		},
		TempCoverLetter: &models.Document{
			Filename:     uuid.New().String(),
			DocumentName: uuid.New().String(),
		},
		TempEstimationDocuments: []*models.Document{
			{
				Filename:     uuid.New().String(),
				DocumentName: uuid.New().String(),
			},
			{
				Filename:     uuid.New().String(),
				DocumentName: uuid.New().String(),
			},
		},
	}

	admissionStatus := 1

	unorRows := sqlmock.NewRows([]string{"id"})
	unorIds := make([]string, 0)
	for _, rc := range dummy.RequirementCounts {
		unorIds = append(unorIds, rc.OrganizationUnitId)
		unorRows.AddRow(rc.OrganizationUnitId)
	}
	referenceMock.ExpectQuery("select").WithArgs(pq.Array(unorIds), sqlmock.AnyArg()).WillReturnRows(unorRows)
	mock.ExpectBegin()
	mock.ExpectQuery("update").WithArgs(
		dummy.PositionGrade,
		dummy.TempCoverLetter.DocumentName,
		dummy.FiscalYear,
		dummy.AdmissionNumber,
		dummy.RequirementId,
		dummy.AgencyId,
	).WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(admissionStatus))
	mock.ExpectExec("delete").WithArgs(dummy.RequirementId).WillReturnResult(sqlmock.NewResult(1, 0))
	profileMock.ExpectQuery("select").WithArgs(dummy.PositionGrade, pq.Array(unorIds)).WillReturnRows(sqlmock.NewRows([]string{"jabatan_fungsional_id", "unor_id", "count(*)"})) // Assume that no rows are returned, this could work too
	reqCountStmt := mock.ExpectPrepare("insert")
	for _, rc := range dummy.RequirementCounts {
		reqCountStmt.ExpectExec().WithArgs(sqlmock.AnyArg(), rc.OrganizationUnitId, rc.Count, 0).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	mock.ExpectExec("insert").WithArgs(sqlmock.AnyArg(), models.RequirementAdmissionStatusCreated, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))
	stmt := mock.ExpectPrepare("insert")
	for range dummy.TempEstimationDocuments {
		stmt.ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))
	}
	mock.ExpectCommit()

	payload, _ := json.Marshal(dummy)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/requirement/admission/edit", bytes.NewBuffer(payload))
	client.HandleRequirementAdmissionEdit(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: dummy.AgencyId}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)

	MustMockExpectationsMet(mock)

	result := &models.RequirementAdmission{}
	MustJsonDecode(rec.Result().Body, result)

	Expect(result.RequirementId).Should(Equal(dummy.RequirementId))
}

func TestHandleRequirementAdmissionSearch(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	admissionDate := "2006-01-02"
	admissionStatus := rand.Intn(4) + 1

	dummy := &models.RequirementAdmissionResult{
		RequirementId:      uuid.New().String(),
		Status:             admissionStatus,
		AdmissionTimestamp: models.EpochTime(time.Now()),
		PositionGrade:      uuid.New().String(),
	}

	rows := sqlmock.NewRows([]string{"kebutuhan_id", "tgl_usulan", "status", "jabatan_fungsional"})
	rows.AddRow(dummy.RequirementId, time.Time(dummy.AdmissionTimestamp), dummy.Status, dummy.PositionGrade)

	mock.ExpectQuery("select").WillReturnRows(rows)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/requirement/admission/search", nil)
	q := req.URL.Query()
	q.Add("tgl_usulan", admissionDate)
	q.Add("status", strconv.Itoa(admissionStatus))
	req.URL.RawQuery = q.Encode()
	client.HandleRequirementAdmissionSearch(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)

	var result []*models.RequirementAdmissionResult
	MustJsonDecode(rec.Result().Body, &result)

	Expect(result).To(HaveLen(1))
	Expect(result[0].RequirementId).To(Equal(dummy.RequirementId))
}

func TestHandleRequirementAdmissionDetailGet(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	referenceDb, referenceMock := MustCreateMock()
	client := CreateClientNoServer(db, nil, referenceDb)

	dummy := &models.RequirementAdmissionDetail{
		RequirementId:      uuid.New().String(),
		AdmissionTimestamp: models.EpochTime(time.Now()),
		Status:             models.RequirementAdmissionStatusRevision,
		PositionGrade:      uuid.New().String(),
		FiscalYear:         uuid.NewString(),
		AdmissionNumber:    uuid.NewString(),
		RequirementCounts: []*models.RequirementCount{
			{
				OrganizationUnitId: uuid.NewString(),
				OrganizationUnit:   uuid.NewString(),
				Count:              1,
			},
		},
		CoverLetter: &models.Document{
			Filename:     uuid.New().String(),
			DocumentName: uuid.New().String(),
		},
		EstimationDocuments: []*models.Document{
			{
				Filename:     uuid.New().String(),
				DocumentName: uuid.New().String(),
			},
			{
				Filename:     uuid.New().String(),
				DocumentName: uuid.New().String(),
			},
		},
		RevisionRequirementCounts: []*models.RequirementCount{},
	}

	requirementRows := sqlmock.NewRows([]string{"kebutuhan_id", "tgl_usulan", "status", "jabatan_fungsional", "filename_sp", "nama_doc_sp", "catatan_sp", "tahun_anggaran", "no_usulan", "alasan_perbaikan"})
	requirementRows.AddRow(dummy.RequirementId, time.Time(dummy.AdmissionTimestamp), dummy.Status, dummy.PositionGrade, dummy.CoverLetter.Filename, dummy.CoverLetter.DocumentName, dummy.CoverLetter.Note, dummy.FiscalYear, dummy.AdmissionNumber, dummy.RevisionReason)
	reqCountRows := sqlmock.NewRows([]string{"unor_id", "jlh_kebutuhan", "rekomendasi_jlh_kebutuhan", "bezetting_jlh_kebutuhan"})
	unorRows := sqlmock.NewRows([]string{"id", "nama_organisasi"})
	for _, rc := range dummy.RequirementCounts {
		reqCountRows.AddRow(rc.OrganizationUnitId, rc.Count, rc.CountRecommendation, rc.CountBezetting)
		unorRows.AddRow(rc.OrganizationUnitId, rc.OrganizationUnit)
	}

	estDocRows := sqlmock.NewRows([]string{"filename", "nama_doc_perhitungan", "catatan"})
	for _, document := range dummy.EstimationDocuments {
		estDocRows.AddRow(document.Filename, document.DocumentName, document.Note)
	}

	mock.ExpectBegin()
	mock.ExpectQuery("select").WillReturnRows(requirementRows)
	mock.ExpectQuery("select").WillReturnRows(reqCountRows)
	referenceMock.ExpectQuery("select").WillReturnRows(unorRows)
	mock.ExpectQuery("select").WillReturnRows(estDocRows)
	mock.ExpectQuery("select").WillReturnError(sql.ErrNoRows)
	mock.ExpectCommit()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/requirement/admission/get", nil)
	q := req.URL.Query()
	q.Add("kebutuhan_id", dummy.RequirementId)
	req.URL.RawQuery = q.Encode()
	client.HandleRequirementAdmissionDetailGet(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)

	var result *models.RequirementAdmissionDetail
	MustJsonDecode(rec.Result().Body, &result)

	Expect(result.RequirementId).To(Equal(dummy.RequirementId))
	Expect(time.Time(result.AdmissionTimestamp).Unix()).To(Equal(time.Time(dummy.AdmissionTimestamp).Unix()))
	Expect(result.Status).To(Equal(dummy.Status))
	Expect(result.PositionGrade).To(Equal(dummy.PositionGrade))
}

func TestHandleRequirementVerificationSet(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	dummy := &models.RequirementVerificationRequest{
		RequirementId:   uuid.New().String(),
		AgencyId:        uuid.New().String(),
		CoverLetterNote: uuid.NewString(),
		EstimationDocumentNotes: []*struct {
			Filename string `json:"nama_file"`
			Note     string `json:"catatan"`
		}{
			{
				Filename: uuid.NewString(),
				Note:     uuid.NewString(),
			},
		},
		RequirementCounts: []*models.RequirementCountRecommendation{
			{
				OrganizationUnitId:  uuid.NewString(),
				CountRecommendation: rand.Intn(100),
			},
			{
				OrganizationUnitId:  uuid.NewString(),
				CountRecommendation: rand.Intn(100),
			},
		},
	}

	mock.ExpectBegin()
	mock.ExpectQuery("select").WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(models.RequirementAdmissionStatusCreated))
	mock.ExpectExec("update").WithArgs(models.RequirementAdmissionStatusAccepted, dummy.CoverLetterNote, dummy.RequirementId, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))
	for _, doc := range dummy.EstimationDocumentNotes {
		mock.ExpectQuery("update").WithArgs(doc.Note, doc.Filename, dummy.RequirementId).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	}
	countStmt := mock.ExpectPrepare("update")
	for _, count := range dummy.RequirementCounts {
		countStmt.ExpectExec().WithArgs(count.CountRecommendation, dummy.RequirementId, count.OrganizationUnitId).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectExec("insert").WithArgs(dummy.RequirementId, models.RequirementAdmissionStatusAccepted, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectCommit()

	payload, _ := json.Marshal(dummy)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/requirement/verify/submit", bytes.NewBuffer(payload))
	client.HandleRequirementVerificationSet(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)

	result := map[string]interface{}{}
	MustJsonDecode(rec.Result().Body, &result)

	Expect(result).To(HaveKey("kebutuhan_id"))
	Expect(result).To(HaveKey("modified_at"))
	Expect(result["kebutuhan_id"]).To(Equal(dummy.RequirementId))
}

func TestHandleRequirementDenySet(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	dummy := &models.RequirementRevisionRequest{
		RequirementId:   uuid.New().String(),
		AgencyId:        uuid.New().String(),
		DenyTimestamp:   models.EpochTime(time.Now()),
		RevisionReason:  uuid.New().String(),
		SubmitterAsnId:  uuid.New().String(),
		CoverLetterNote: uuid.NewString(),
		EstimationDocumentNotes: []*struct {
			Filename string `json:"nama_file"`
			Note     string `json:"catatan"`
		}{
			{
				Filename: uuid.NewString(),
				Note:     uuid.NewString(),
			},
		},
	}

	mock.ExpectBegin()
	mock.ExpectQuery("select").WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(models.RequirementAdmissionStatusCreated))
	mock.ExpectExec("update").WithArgs(models.RequirementAdmissionStatusRevision, dummy.RevisionReason, dummy.CoverLetterNote, dummy.RequirementId, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))
	for _, doc := range dummy.EstimationDocumentNotes {
		mock.ExpectQuery("update").WithArgs(doc.Note, doc.Filename, dummy.RequirementId).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	}
	countStmt := mock.ExpectPrepare("update")
	for _, count := range dummy.RequirementCounts {
		countStmt.ExpectExec().WithArgs(count.CountRecommendation, dummy.RequirementId, count.OrganizationUnitId).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectExec("insert").WithArgs(dummy.RequirementId, models.RequirementAdmissionStatusRevision, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectCommit()

	payload, _ := json.Marshal(dummy)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/requirement/verify/deny", bytes.NewBuffer(payload))
	client.HandleRequirementDenySet(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)

	result := map[string]interface{}{}
	MustJsonDecode(rec.Result().Body, &result)

	Expect(result).To(HaveKey("kebutuhan_id"))
	Expect(result).To(HaveKey("modified_at"))
	Expect(result["kebutuhan_id"]).To(Equal(dummy.RequirementId))
}

func TestHandleRequirementVerifiersGet(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	pdb, profileMock := MustCreateMock()
	client := CreateClientNoServer(db, pdb, nil)

	user := &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}

	dummy := &models.RequirementVerifier{
		AsnId:                   time.Now().String(),
		RequirementVerifierName: time.Now().Format("2006 01 02"),
	}

	mock.ExpectQuery("select").WithArgs(models.StaffRoleSupervisor, user.WorkAgencyId).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(dummy.AsnId))
	profileMock.ExpectQuery("select").WithArgs(pq.Array([]string{dummy.AsnId})).WillReturnRows(sqlmock.NewRows([]string{"id", "nama"}).AddRow(dummy.AsnId, dummy.RequirementVerifierName))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/requirement/verifier/get", nil)
	client.HandleRequirementVerifiersGet(rec, auth.InjectUserDetail(req, user))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)

	var result []*models.RequirementVerifier
	MustJsonDecode(rec.Result().Body, &result)

	Expect(result).To(HaveLen(1))
	Expect(result[0].AsnId).To(Equal(dummy.AsnId))
	Expect(result[0].RequirementVerifierName).To(Equal(dummy.RequirementVerifierName))
}

func TestHandleBulkSubmitRecommendationLetter(t *testing.T) {
	RegisterTestingT(t)

	t.Skip("the client request the feature immediately and as it has been too complex, the unit test is postponed")

	//db, mock := MustCreateMock()
	//referenceDb, referenceMock := MustCreateMock()
	//client := CreateClientNoServer(db, nil, referenceDb)
	//client.DocxRenderer = &mockDocxRenderer{}
	//
	//type submitReq struct {
	//	RequirementIds           []string         `json:"kebutuhan_id"`
	//	IsPresigned              bool             `json:"sudah_ttd"`
	//	TempRecommendationLetter *models.Document `json:"temp_surat_rekomendasi"`
	//}
	//dummy := &submitReq{
	//	RequirementIds: []string{uuid.NewString(), uuid.NewString()},
	//	IsPresigned:    false,
	//	TempRecommendationLetter: &models.Document{
	//		Filename:       uuid.NewString(),
	//		DocumentName:   uuid.NewString(),
	//		DocumentNumber: uuid.NewString(),
	//		DocumentDate:   "2020-01-01",
	//		SignerId:       uuid.NewString(),
	//	},
	//}
	//
	//user := &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}
	//reqIdRows := sqlmock.NewRows([]string{"kebutuhan_id", "instansi_id", "jabatan_fungsional", "status"})
	//for _, rid := range dummy.RequirementIds {
	//	reqIdRows.AddRow(rid, user.WorkAgencyId, uuid.NewString(), models.RequirementAdmissionStatusAccepted)
	//}
	//reqCountRows := sqlmock.NewRows([]string{"kebutuhan_id", "unor_id", "jlh_kebutuhan", "coalesce(rekomendasi_jlh_kebutuhan, 0)", "bezetting_jlh_kebutuhan"})
	//
	//mock.ExpectBegin()
	//referenceMock.ExpectBegin()
	//mock.ExpectQuery("select").WillReturnRows(reqIdRows).WithArgs(pq.Array(dummy.RequirementIds))
	//mock.ExpectQuery("select").WillReturnRows(reqCountRows)
	//referenceMock.ExpectQuery("select").WillReturnRows()
	//referenceMock.ExpectQuery("select").WillReturnRows()
	//referenceMock.ExpectQuery("select").WillReturnRows()
	//mock.ExpectQuery("select").WillReturnRows(sqlmock.NewRows([]string{"role_peg"}).AddRow(models.StaffRoleSupervisor)).WithArgs(dummy.TempRecommendationLetter.SignerId)
	//mock.ExpectExec("update").WithArgs(models.RequirementAdmissionStatusAcceptedWithRecommendation, pq.Array(dummy.RequirementIds)).WillReturnResult(sqlmock.NewResult(1, 0))
	//stmt := mock.ExpectPrepare("insert")
	//statusStmt := mock.ExpectPrepare("insert")
	//for _, rid := range dummy.RequirementIds {
	//	stmt.ExpectExec().WithArgs(
	//		dummy.TempRecommendationLetter.Filename,
	//		rid,
	//		dummy.TempRecommendationLetter.DocumentDate,
	//		dummy.TempRecommendationLetter.DocumentName,
	//		dummy.TempRecommendationLetter.DocumentNumber,
	//		sql.NullString{Valid: dummy.TempRecommendationLetter.Note != "", String: dummy.TempRecommendationLetter.Note},
	//		dummy.TempRecommendationLetter.SignerId,
	//		false,
	//		sql.NullTime{},
	//	).WillReturnResult(sqlmock.NewResult(1, 0))
	//	statusStmt.ExpectExec().WithArgs(models.RequirementAdmissionStatusAcceptedWithRecommendation, rid, user.AsnId).WillReturnResult(sqlmock.NewResult(1, 0))
	//}
	//referenceMock.ExpectCommit()
	//mock.ExpectCommit()
	//
	//payload, _ := json.Marshal(dummy)
	//rec := httptest.NewRecorder()
	//req, _ := http.NewRequest("POST", "/api/v1/requirement/verify/bulk-submit/recommendation-letter", bytes.NewBuffer(payload))
	//client.HandleRequirementVerificationBulkSubmitRecommendationLetter(rec, auth.InjectUserDetail(req, user))
	//
	//MustStatusCodeEqual(rec.Result(), http.StatusOK)
	//MustMockExpectationsMet(mock)
	//
	//var result map[string][]string
	//MustJsonDecode(rec.Result().Body, &result)
	//
	//Expect(result).To(HaveKey("kebutuhan_id"))
}

func TestHandleGetRequirementStatusStatistic(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	statisticRows := sqlmock.NewRows([]string{"status", "jumlah"}).AddRow(1, 0).AddRow(2, 1)

	mock.ExpectQuery("select").WillReturnRows(statisticRows)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/requirement/statistic/status/get", nil)
	client.HandleGetRequirementStatusStatistic(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)
}
