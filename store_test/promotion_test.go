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
	"github.com/if-itb/siasn-libs-backend/search"
	. "github.com/onsi/gomega"
)

func TestHandlePromotionAdmissionSubmit(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	dbProfile, mockProfile := MustCreateMock()
	dbReference, mockReference := MustCreateMock()
	client := CreateClientNoServer(db, dbProfile, dbReference)

	dummy := &models.PromotionAdmission{
		AdmissionNumber: uuid.NewString(),
		AdmissionDate:   models.Iso8601Date(time.Now().Format("2006-01-02")),
		AsnId:           uuid.New().String(),

		TestStatus: 1 + rand.Intn(2),
		TestScore:  rand.Float64() * 100,

		PromotionType:       1 + rand.Intn(3),
		PromotionPositionId: uuid.New().String(),

		PakLetter: &models.Document{
			Filename:       uuid.New().String(),
			DocumentName:   uuid.New().String(),
			DocumentNumber: uuid.New().String(),
			DocumentDate:   models.Iso8601Date(time.Now().Format("2006-01-02")),
		},
		RecommendationLetter: &models.Document{
			Filename:       uuid.New().String(),
			DocumentName:   uuid.New().String(),
			DocumentNumber: uuid.New().String(),
			DocumentDate:   models.Iso8601Date(time.Now().Format("2006-01-02")),
		},
		TestCertificate: &models.Document{
			Filename:       uuid.New().String(),
			DocumentName:   uuid.New().String(),
			DocumentNumber: uuid.New().String(),
			DocumentDate:   models.Iso8601Date(time.Now().Format("2006-01-02")),
		},
		AgencyId:       uuid.New().String(),
		SubmitterAsnId: uuid.New().String(),
	}

	mock.ExpectBegin()
	mock.ExpectExec("insert").WithArgs(
		sqlmock.AnyArg(),
		dummy.AsnId,
		dummy.AdmissionNumber,
		dummy.AdmissionDate,
		dummy.PromotionType,
		dummy.PromotionPositionId,
		models.PromotionAdmissionStatusCreated,
		dummy.SubmitterAsnId,
		dummy.TestStatus,
		sqlmock.AnyArg(),
		sql.NullString{Valid: true, String: dummy.PakLetter.Filename},
		sql.NullString{Valid: true, String: dummy.PakLetter.DocumentNumber},
		sql.NullString{Valid: true, String: string(dummy.PakLetter.DocumentDate)},
		sql.NullString{Valid: true, String: dummy.RecommendationLetter.Filename},
		sql.NullString{Valid: true, String: dummy.RecommendationLetter.DocumentNumber},
		sql.NullString{Valid: true, String: string(dummy.RecommendationLetter.DocumentDate)},
		sql.NullString{Valid: true, String: dummy.TestCertificate.Filename},
		sql.NullString{Valid: true, String: dummy.TestCertificate.DocumentNumber},
		sql.NullString{Valid: true, String: string(dummy.TestCertificate.DocumentDate)},
	).WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectCommit()

	payload, _ := json.Marshal(dummy)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/promotion/admission/submit", bytes.NewBuffer(payload))
	client.HandlePromotionAdmissionSubmit(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: dummy.SubmitterAsnId, WorkAgencyId: dummy.AgencyId}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)

	MustMockExpectationsMet(mock)
	MustMockExpectationsMet(mockProfile)
	MustMockExpectationsMet(mockReference)

	result := &models.PromotionAdmission{}
	MustJsonDecode(rec.Result().Body, result)

	Expect(result.PromotionId).ToNot(BeEmpty())
}

func TestHandlePromotionAdmissionAccept(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	dummy := &models.PromotionAdmission{
		PromotionId: uuid.New().String(),
		AgencyId:    uuid.New().String(),
		PromotionLetter: &models.Document{
			Filename:       uuid.New().String(),
			DocumentName:   uuid.New().String(),
			DocumentNumber: uuid.New().String(),
			DocumentDate:   models.Iso8601Date(time.Now().Format("2006-01-02")),
		},
		SubmitterAsnId: uuid.New().String(),
	}

	mock.ExpectBegin()
	mock.ExpectQuery("select").WithArgs(dummy.PromotionId).WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(models.PromotionAdmissionStatusCreated))
	mock.ExpectExec("update").WithArgs(
		models.PromotionAdmissionStatusAccepted,
		sqlmock.AnyArg(),
		dummy.SubmitterAsnId,
		dummy.PromotionId,
	).WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectCommit()

	payload, _ := json.Marshal(dummy)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/promotion/admission/accept", bytes.NewBuffer(payload))
	client.HandlePromotionAdmissionAccept(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: dummy.SubmitterAsnId, WorkAgencyId: dummy.AgencyId}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)

	MustMockExpectationsMet(mock)

	result := make(map[string]interface{}, 0)
	MustJsonDecode(rec.Result().Body, &result)

	Expect(result["pengangkatan_id"]).To(Equal(dummy.PromotionId))
	Expect(result).To(HaveKey("modified_at"))
}

func TestHandlePromotionAdmissionReject(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	dummy := &models.PromotionReject{
		PromotionId:    uuid.New().String(),
		RejectReason:   uuid.New().String(),
		SubmitterAsnId: uuid.New().String(),
	}

	mock.ExpectBegin()
	mock.ExpectQuery("select").WithArgs(dummy.PromotionId).WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(models.PromotionAdmissionStatusCreated))
	mock.ExpectExec("update").WithArgs(
		models.PromotionAdmissionStatusRejected,
		sqlmock.AnyArg(),
		dummy.SubmitterAsnId,
		dummy.RejectReason,
		dummy.PromotionId,
	).WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectCommit()

	payload, _ := json.Marshal(dummy)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/promotion/admission/reject", bytes.NewBuffer(payload))
	client.HandlePromotionAdmissionReject(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: dummy.SubmitterAsnId, WorkAgencyId: dummy.SubmitterAsnId}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)

	MustMockExpectationsMet(mock)

	result := make(map[string]interface{}, 0)
	MustJsonDecode(rec.Result().Body, &result)

	Expect(result["pengangkatan_id"]).To(Equal(dummy.PromotionId))
	Expect(result).To(HaveKey("modified_at"))
}

func TestHandlePromotionAdmissionSearch(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	profileDb, profileMock := MustCreateMock()
	client := CreateClientNoServer(db, profileDb, nil)

	admissionDate := models.Iso8601Date(time.Now().Format("2006-01-02"))
	admissionStatus := rand.Intn(3) + 1
	admissionType := rand.Intn(3) + 1
	countPerPage := rand.Intn(25) + 1
	pageNumber := rand.Intn(25) + 1

	dummy := &models.PromotionItem{
		PromotionId:              uuid.New().String(),
		AsnId:                    uuid.New().String(),
		Name:                     uuid.New().String(),
		Status:                   admissionStatus,
		PromotionType:            admissionType,
		RecommendationLetterDate: admissionDate,
	}

	rows := sqlmock.NewRows([]string{"uuid_pengangkatan", "asn_id", "status", "tgl_doc_surat_rekomendasi", "jenis_pengangkatan"})
	rows.AddRow(
		dummy.PromotionId,
		dummy.AsnId,
		dummy.Status,
		dummy.RecommendationLetterDate,
		dummy.PromotionType,
	)

	asnRows := sqlmock.NewRows([]string{"id", "nama"})
	asnRows.AddRow(
		dummy.AsnId,
		dummy.Name,
	)

	mock.ExpectQuery("select").WithArgs(dummy.Status, dummy.PromotionType, dummy.RecommendationLetterDate, countPerPage+1, (pageNumber-1)*countPerPage).WillReturnRows(rows)
	profileMock.ExpectQuery("select").WillReturnRows(asnRows)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/promotion/admission/search/paginated", nil)
	q := req.URL.Query()
	q.Add("tgl_doc_surat_rekomendasi", string(admissionDate))
	q.Add("status", strconv.Itoa(admissionStatus))
	q.Add("jenis_pengangkatan", strconv.Itoa(admissionType))
	q.Add("jumlah_per_halaman", strconv.Itoa(countPerPage))
	q.Add("halaman", strconv.Itoa(pageNumber))
	req.URL.RawQuery = q.Encode()
	client.HandlePromotionAdmissionSearchPaginated(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)

	var result *struct {
		Data     []*models.PromotionItem       `json:"data"`
		Metadata *search.PaginatedListMetadata `json:"metadata"`
	}
	MustJsonDecode(rec.Result().Body, &result)

	Expect(result.Data).To(HaveLen(1))
	Expect(result.Data[0]).To(Equal(dummy))
}

func TestHandleGetPromotionStatusStatistic(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	statisticRows := sqlmock.NewRows([]string{"status", "jumlah"}).AddRow(1, 0).AddRow(2, 1)

	mock.ExpectQuery("select").WillReturnRows(statisticRows)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/promotion/statistic/status/get", nil)
	client.HandleGetPromotionStatusStatistic(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)
}
