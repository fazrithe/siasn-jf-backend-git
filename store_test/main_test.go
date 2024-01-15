package store_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fazrithe/siasn-jf-backend-git/store"
	"github.com/fazrithe/siasn-jf-backend-git/store/object"
	"github.com/if-itb/siasn-libs-backend/breaker"
	"github.com/if-itb/siasn-libs-backend/logutil"
	. "github.com/onsi/gomega"
)

func CreateClient(db, profileDb, referenceDb *sql.DB, h http.Handler) (*httptest.Server, *store.Client) {
	s := httptest.NewServer(h)
	rcb := &breaker.RateCircuitBreaker{
		Limit:    1000,
		Cooldown: 1,
		Logger:   logutil.NewStdLogger(false, "test"),
		Tripped:  make(chan struct{}),
	}
	c := &store.Client{
		Db:                    db,
		ProfileDb:             profileDb,
		ReferenceDb:           referenceDb,
		SqlMetrics:            nil,
		ActivityStorage:       &object.MockStorage{},
		RequirementStorage:    &object.MockStorage{},
		DismissalStorage:      &object.MockStorage{},
		PromotionStorage:      &object.MockStorage{},
		AssessmentTeamStorage: &object.MockStorage{},
		Breaker:               rcb,
		Logger:                logutil.NewStdLogger(false, "test"),
	}

	return s, c
}

func CreateClientNoServer(db, profileDb, referenceDb *sql.DB) *store.Client {
	_, client := CreateClient(db, profileDb, referenceDb, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
	return client
}

func MustCreateMock() (db *sql.DB, mock sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	Expect(err).ToNot(HaveOccurred(), "an error '%s' was not expected when opening a stub database connection", err)
	return
}

func MustMockExpectationsMet(mock sqlmock.Sqlmock) {
	err := mock.ExpectationsWereMet()
	Expect(err).ShouldNot(HaveOccurred(), "some SQL expectations were not met: %v", err)
}

func MustStatusCodeEqual(resp *http.Response, statusCode int) {
	Expect(resp.StatusCode).To(Equal(statusCode), "status code not expected: %d, want %d", resp.StatusCode, statusCode)
}

func MustJsonDecode(reader io.Reader, object interface{}) {
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(object)
	Expect(err).ShouldNot(HaveOccurred(), "result should be able to be decoded: %v", err)
}

type mockDocxRenderer struct{}

func (m *mockDocxRenderer) Render(data interface{}, templatePath string, outputPath string) (err error) {
	return m.RenderCtx(context.Background(), data, templatePath, outputPath)
}

func (m *mockDocxRenderer) RenderCtx(ctx context.Context, data interface{}, templatePath string, outputPath string) (err error) {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

func (m *mockDocxRenderer) RenderAsPdf(data interface{}, templatePath string, outputPath string) (err error) {
	return m.RenderAsPdfCtx(context.Background(), data, templatePath, outputPath)
}

func (m *mockDocxRenderer) RenderAsPdfCtx(ctx context.Context, data interface{}, templatePath string, outputPath string) (err error) {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}
