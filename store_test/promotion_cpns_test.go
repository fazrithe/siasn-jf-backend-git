package store_test

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fazrithe/siasn-jf-backend-git/libs/auth"
	"github.com/fazrithe/siasn-jf-backend-git/store/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
	. "github.com/onsi/gomega"
)

func TestHandlePromotionCpnsAdmissionGet(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	profileDb, profileMock := MustCreateMock()
	referenceDb, referenceMock := MustCreateMock()
	client := CreateClientNoServer(db, profileDb, referenceDb)

	dummy := &models.PromotionCpnsAdmission{
		PromotionCpnsId:     uuid.NewString(),
		PromotionPositionId: uuid.NewString(),
		PromotionPosition:   uuid.NewString(),
		AdmissionNumber:     uuid.NewString(),
		AdmissionDate:       models.Iso8601Date(time.Now().Format("2006-01-02")),
		AsnId:               uuid.NewString(),
		AsnName:             uuid.NewString(),
		AsnNip:              uuid.NewString(),
		OrganizationUnitId:  uuid.NewString(),
		OrganizationUnit:    uuid.NewString(),
		FirstCreditNumber:   rand.Intn(100) + 1,
		PakLetter: &models.Document{
			DocumentName:   uuid.NewString(),
			DocumentNumber: uuid.NewString(),
			DocumentDate:   models.Iso8601Date(time.Now().Format("2006-01-02")),
		},
		PromotionLetter: &models.Document{
			DocumentName:   uuid.NewString(),
			DocumentNumber: uuid.NewString(),
			DocumentDate:   models.Iso8601Date(time.Now().Format("2006-01-02")),
		},
	}

	mock.ExpectQuery("select").WithArgs(dummy.PromotionCpnsId).WillReturnRows(sqlmock.NewRows([]string{
		"asn_id",
		"jabatan_fungsional_tujuan_id",
		"angka_kredit_pertama",
		"unor_id",
		"tgl_usulan",
		"no_usulan",
		"nama_doc_pak",
		"no_doc_pak",
		"tgl_doc_pak",
		"nama_doc_surat_pengangkatan",
		"no_doc_surat_pengangkatan",
		"tgl_doc_surat_pengangkatan",
		"status",
		"status_by",
	}).AddRow(
		&dummy.AsnId,
		&dummy.PromotionPositionId,
		&dummy.FirstCreditNumber,
		&dummy.OrganizationUnitId,
		&dummy.AdmissionDate,
		&dummy.AdmissionNumber,
		&dummy.PakLetter.DocumentName,
		&dummy.PakLetter.DocumentNumber,
		&dummy.PakLetter.DocumentDate,
		&dummy.PromotionLetter.DocumentName,
		&dummy.PromotionLetter.DocumentNumber,
		&dummy.PromotionLetter.DocumentDate,
		&dummy.Status,
		uuid.NewString(),
	))
	profileMock.ExpectQuery("select").WithArgs(pq.Array([]string{dummy.AsnId})).WillReturnRows(sqlmock.NewRows([]string{"id", "nip", "nama"}).AddRow(dummy.AsnId, dummy.AsnNip, dummy.AsnName))
	referenceMock.ExpectQuery("select").WithArgs(pq.Array([]string{dummy.OrganizationUnitId})).WillReturnRows(sqlmock.NewRows([]string{"id", "nama"}).AddRow(dummy.OrganizationUnitId, dummy.OrganizationUnit))
	referenceMock.ExpectQuery("select").WithArgs(pq.Array([]string{dummy.PromotionPositionId})).WillReturnRows(sqlmock.NewRows([]string{"id", "nama"}).AddRow(dummy.PromotionPositionId, dummy.PromotionPosition))

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/promotion-cpns/admission/get?pengangkatan_cpns_id=%s", dummy.PromotionCpnsId), nil)
	client.HandlePromotionCpnsAdmissionGet(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: dummy.SubmitterAsnId}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)

	MustMockExpectationsMet(mock)

	result := &models.PromotionCpnsAdmission{}
	MustJsonDecode(rec.Result().Body, result)

	Expect(result).To(Equal(dummy))
}

func TestHandlePromotionCpnsAdmissionSearchPaginated(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	profileDb, profileMock := MustCreateMock()
	referenceDb, referenceMock := MustCreateMock()
	client := CreateClientNoServer(db, profileDb, referenceDb)

	admissionDate := models.Iso8601Date(time.Now().Format("2006-01-02"))
	admissionStatus := rand.Intn(3) + 1
	countPerPage := rand.Intn(25) + 1
	pageNumber := rand.Intn(25) + 1

	dummy := []*models.PromotionCpnsItem{
		{
			PromotionCpnsId:     uuid.NewString(),
			AsnId:               uuid.NewString(),
			AsnNip:              uuid.NewString(),
			AsnName:             uuid.NewString(),
			Status:              rand.Intn(3) + 1,
			PromotionPositionId: uuid.NewString(),
			PromotionPosition:   uuid.NewString(),
			FirstCreditNumber:   rand.Intn(3) + 1,
			OrganizationUnitId:  uuid.NewString(),
			OrganizationUnit:    uuid.NewString(),
			AdmissionNumber:     uuid.NewString(),
			AdmissionDate:       models.Iso8601Date(time.Now().Format("2006-01-02")),
		},
		{
			PromotionCpnsId:     uuid.NewString(),
			AsnId:               uuid.NewString(),
			AsnNip:              uuid.NewString(),
			AsnName:             uuid.NewString(),
			Status:              rand.Intn(3) + 1,
			PromotionPositionId: uuid.NewString(),
			PromotionPosition:   uuid.NewString(),
			FirstCreditNumber:   rand.Intn(3) + 1,
			OrganizationUnitId:  uuid.NewString(),
			OrganizationUnit:    uuid.NewString(),
			AdmissionNumber:     uuid.NewString(),
			AdmissionDate:       models.Iso8601Date(time.Now().Format("2006-01-02")),
		},
	}

	rows := sqlmock.NewRows([]string{
		"pengangkatan_cpns_id",
		"asn_id",
		"jabatan_fungsional_tujuan_id",
		"angka_kredit_pertama",
		"unor_id",
		"tgl_usulan",
		"no_usulan",
		"status",
	})
	unorRows := sqlmock.NewRows([]string{"id", "nama"})
	positionRows := sqlmock.NewRows([]string{"id", "nama"})
	asnRows := sqlmock.NewRows([]string{"id", "nip", "nama"})
	promotionCpnsIds := make([]string, 0)
	unorIds := make([]string, 0)
	positionIds := make([]string, 0)
	asnIds := make([]string, 0)
	for _, d := range dummy {
		promotionCpnsIds = append(promotionCpnsIds, d.PromotionCpnsId)
		unorIds = append(unorIds, d.OrganizationUnitId)
		positionIds = append(positionIds, d.PromotionPositionId)
		asnIds = append(asnIds, d.AsnId)
		rows.AddRow(
			d.PromotionCpnsId,
			d.AsnId,
			d.PromotionPositionId,
			d.FirstCreditNumber,
			d.OrganizationUnitId,
			d.AdmissionDate,
			d.AdmissionNumber,
			d.Status,
		)
		unorRows.AddRow(d.OrganizationUnitId, d.OrganizationUnit)
		positionRows.AddRow(d.PromotionPositionId, d.PromotionPosition)
		asnRows.AddRow(d.AsnId, d.AsnNip, d.AsnName)
	}

	referenceMock.ExpectBegin()
	mock.ExpectQuery("select").WithArgs(admissionStatus, admissionDate, countPerPage+1, (pageNumber-1)*countPerPage).WillReturnRows(rows)
	profileMock.ExpectQuery("select").WithArgs(pq.Array(asnIds)).WillReturnRows(asnRows)
	referenceMock.ExpectQuery("select").WithArgs(pq.Array(unorIds)).WillReturnRows(unorRows)
	referenceMock.ExpectQuery("select").WithArgs(pq.Array(positionIds)).WillReturnRows(positionRows)
	referenceMock.ExpectCommit()

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/promotion-cpns/admission/search", nil)
	q := req.URL.Query()
	q.Set("tgl_usulan", string(admissionDate))
	q.Set("status", strconv.Itoa(admissionStatus))
	q.Set("jumlah_per_halaman", strconv.Itoa(countPerPage))
	q.Set("halaman", strconv.Itoa(pageNumber))
	req.URL.RawQuery = q.Encode()
	client.HandlePromotionCpnsAdmissionSearchPaginated(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.NewString()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)

	MustMockExpectationsMet(mock)

	result := &models.IdPaginatedList{}
	MustJsonDecode(rec.Result().Body, result)

	Expect(result.Metadata.Subtotal).To(Equal(len(dummy)))
}

func TestHandleGetPromotionCpnsStatusStatistic(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(db, nil, nil)

	statisticRows := sqlmock.NewRows([]string{"status", "jumlah"}).AddRow(1, 0).AddRow(2, 1)

	mock.ExpectQuery("select").WillReturnRows(statisticRows)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/promotion-cpns/statistic/status/get", nil)
	client.HandleGetPromotionCpnsStatusStatistic(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)
}
