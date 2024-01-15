package store_test

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/if-itb/siasn-jf-backend/store/models"
	"github.com/if-itb/siasn-libs-backend/auth"
	. "github.com/onsi/gomega"
)

func TestHandlePositionGradesGet(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	client := CreateClientNoServer(nil, nil, db)

	dummy := &models.PositionGrade{
		PositionGradeId:   time.Now().String(),
		PositionGradeName: time.Now().Format("2006 01 02"),
	}

	mock.ExpectQuery("select").WillReturnRows(sqlmock.NewRows([]string{"id", "nama"}).AddRow(dummy.PositionGradeId, dummy.PositionGradeName))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/generic/position/get", nil)
	client.HandlePositionGradesGet(rec, auth.InjectUserDetail(req, &auth.Asn{AsnId: uuid.New().String(), WorkAgencyId: uuid.New().String()}))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)

	var result []*models.PositionGrade
	MustJsonDecode(rec.Result().Body, &result)

	Expect(result).To(HaveLen(1))
	Expect(result[0].PositionGradeId).To(Equal(dummy.PositionGradeId))
}

func TestHandlePositionGradeBezettingGet(t *testing.T) {
	RegisterTestingT(t)

	db, mock := MustCreateMock()
	profileDb, profileMock := MustCreateMock()
	client := CreateClientNoServer(db, profileDb, nil)

	dummyUser := &auth.Asn{
		AsnId:        uuid.New().String(),
		WorkAgencyId: uuid.New().String(),
	}

	dummyBezetting := &models.Bezetting{
		PositionGradeId:    uuid.NewString(),
		OrganizationUnitId: uuid.NewString(),
		Count:              rand.Intn(100),
	}

	profileMock.ExpectQuery("select").WithArgs(dummyBezetting.PositionGradeId, dummyBezetting.OrganizationUnitId).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(dummyBezetting.Count))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/generic/bezetting?jabatan_id=%s&unit_organisasi_id=%s", dummyBezetting.PositionGradeId, dummyBezetting.OrganizationUnitId), nil)
	client.HandlePositionGradeBezettingGet(rec, auth.InjectUserDetail(req, dummyUser))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(mock)
	MustMockExpectationsMet(profileMock)

	result := &models.Bezetting{}
	MustJsonDecode(rec.Result().Body, result)

	Expect(result).To(Equal(dummyBezetting))
}

func TestHandleListOrganizationUnits(t *testing.T) {
	RegisterTestingT(t)

	referenceDb, referenceMock := MustCreateMock()
	client := CreateClientNoServer(nil, nil, referenceDb)

	asn := &auth.Asn{AsnId: uuid.NewString(), WorkAgencyId: uuid.NewString()}

	dummy := []*models.OrganizationUnit{
		{
			OrganizationUnitId:   uuid.NewString(),
			OrganizationUnitName: uuid.NewString(),
		},
		{
			OrganizationUnitId:   uuid.NewString(),
			OrganizationUnitName: uuid.NewString(),
		},
	}

	rows := sqlmock.NewRows([]string{"id", "nama"})
	for _, d := range dummy {
		rows.AddRow(d.OrganizationUnitId, d.OrganizationUnitName)
	}

	referenceMock.ExpectQuery("select").WillReturnRows(rows).WithArgs(asn.WorkAgencyId)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/generic/unit/list", nil)
	client.HandleListOrganizationUnits(rec, auth.InjectUserDetail(req, asn))

	MustStatusCodeEqual(rec.Result(), http.StatusOK)
	MustMockExpectationsMet(referenceMock)

	result := make([]*models.OrganizationUnit, 0)
	MustJsonDecode(rec.Result().Body, &result)

	Expect(result).To(Equal(dummy))
}
