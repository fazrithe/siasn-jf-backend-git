package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	. "github.com/fazrithe/siasn-jf-backend-git/errnum"
	"github.com/fazrithe/siasn-jf-backend-git/libs/breaker"
	"github.com/fazrithe/siasn-jf-backend-git/libs/docx"
	"github.com/fazrithe/siasn-jf-backend-git/libs/ec"
	"github.com/fazrithe/siasn-jf-backend-git/libs/httputil"
	"github.com/fazrithe/siasn-jf-backend-git/libs/logutil"
	"github.com/fazrithe/siasn-jf-backend-git/libs/metricutil"
	"github.com/fazrithe/siasn-jf-backend-git/store/object"
	"github.com/gorilla/schema"
)

// Client is a db + HTTP client to access data from SI-X and store them in local PostgreSQL database.
type Client struct {
	Db                    *sql.DB
	ProfileDb             *sql.DB
	ReferenceDb           *sql.DB
	ActivityStorage       object.ActivityStorage
	RequirementStorage    object.RequirementStorage
	DismissalStorage      object.DismissalStorage
	PromotionStorage      object.PromotionStorage
	PromotionCpnsStorage  object.PromotionCpnsStorage
	AssessmentTeamStorage object.AssessmentTeamStorage
	DocxRenderer          docx.Renderer
	SqlMetrics            metricutil.GenericSqlMetrics
	Breaker               *breaker.RateCircuitBreaker
	Logger                logutil.Logger
}

func NewClient(
	db,
	profileDb,
	referenceDb *sql.DB,
	storage object.Storage,
	docxRenderer docx.Renderer,
	sqlMetrics metricutil.GenericSqlMetrics,
	breaker *breaker.RateCircuitBreaker,
) *Client {
	return &Client{
		Db:                    db,
		ProfileDb:             profileDb,
		ReferenceDb:           referenceDb,
		ActivityStorage:       storage,
		RequirementStorage:    storage,
		DismissalStorage:      storage,
		PromotionStorage:      storage,
		PromotionCpnsStorage:  storage,
		AssessmentTeamStorage: storage,
		DocxRenderer:          docxRenderer,
		SqlMetrics:            sqlMetrics,
		Logger:                logutil.NewStdLogger(false, "store"),
		Breaker:               breaker,
	}
}

// Deprecated: use createMtxDb instead.
//
// createMtx creates a Tx (nullable metric, from metricutil) that wraps sql.Tx and prints and error to logger if error.
// It creates the TX with READ_COMITTED level of isolation.
//
// createMtx returns an error with code as ec.Error.
func (c *Client) createMtx(ctx context.Context) (mtx *metricutil.Tx, err error) {
	return c.createMtxDb(ctx, c.Db)
}

// createMtxDb creates a Tx (nullable metric, from metricutil) with a given *sql.DB that wraps sql.Tx and prints and
// error to logger if error.
// It creates the TX with READ_COMITTED level of isolation.
//
// createMtxDb returns an error with code as ec.Error.
func (c *Client) createMtxDb(ctx context.Context, db *sql.DB) (mtx *metricutil.Tx, err error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, ec.NewError(ErrCodeTxStart, Errs[ErrCodeTxStart], err)
	}
	mtx = metricutil.StartTx(tx, c.SqlMetrics)
	return
}

// completeMtx completes an mtx (an sql.Tx wrapper) with Commit if err is nil and Rollback if err is not nil.
// This function does not add error to the breaker.
func (c *Client) completeMtx(mtx *metricutil.Tx, err error) {
	if err == nil {
		_ = mtx.Commit()
		return
	}

	_ = mtx.Rollback()
}

// httpErrorVerifyListMeta writer error to response, for verifying common list metadata parameters given from
// the request like page number and count per page. If the error is returned, that means httpError has been called,
// and you don't have to write response or status code again.
//
// pageNumber must be >= 1.
// count must be >= 1 and <= 100.
func (c *Client) httpErrorVerifyListMeta(writer http.ResponseWriter, pageNumber, count int) error {
	if count < 1 || count > 100 {
		c.httpError(writer, ErrListCountPerPage)
		return ErrListCountPerPage
	}

	if pageNumber < 1 {
		c.httpError(writer, ErrListPageNumber)
		return ErrListPageNumber
	}

	return nil
}

// httpError writes error to response. Also write to log and add breaker if its a client error.
// Generic errors that are not a derivative of ec.Error will be wrapped as ec.Error.
// Errors that are derivatives of ec.Error (by errors.As test) are returned as is.
// Error code to HTTP status mapping is retrieved from ErrsToHttp.
func (c *Client) httpError(writer http.ResponseWriter, err error) {
	if err == nil {
		panic("error cannot be nil")
	}

	var status int
	var returned error
	var w *ec.Error
	if !errors.As(err, &w) {
		w = ec.Wrap(err)
		returned = w
	} else {
		returned = err
	}

	ok := false
	if status, ok = ErrsToHttp[w.Code]; !ok {
		status = http.StatusInternalServerError
	}

	if !IsClientError(w) {
		if status, ok = ErrsToHttp[w.Code]; !ok {
			status = http.StatusInternalServerError
		}
		c.Breaker.AddError(returned)
		c.Logger.Warn(returned.Error())
		// Cause is never returned by the API if this is a server error, but it will be logged.
		_ = httputil.WriteObj(writer, ec.NewErrorBasic(w.Code, w.Message), status)
	} else {
		if status, ok = ErrsToHttp[w.Code]; !ok {
			status = http.StatusBadRequest
		}
		_ = httputil.WriteObj(writer, returned, status)
	}

	return
}

// decodeRequestJson decodes HTTP request body as JSON to `obj`.
// It also utilizes metricutil.CounterBuffer to count number of bytes decoded. This function already
// takes care sending error message to requester so you don't need to do it again outside this helper function.
func (c *Client) decodeRequestJson(writer http.ResponseWriter, request *http.Request, obj interface{}) (err error) {
	if request.Body == nil {
		c.httpError(writer, ErrRequestBodyNil)
		return
	}

	err = json.NewDecoder(request.Body).Decode(obj)
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeRequestJsonDecode, Errs[ErrCodeRequestJsonDecode], err))
		return
	}

	return nil
}

// decodeRequestSchema decodes HTTP request query string to `obj` using gorilla/schema.
// This function already
// takes care sending error message to requester so you don't need to do it again outside this helper function.
func (c *Client) decodeRequestSchema(writer http.ResponseWriter, request *http.Request, obj interface{}) (err error) {
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	err = decoder.Decode(obj, request.URL.Query())
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeRequestQueryParamParse, Errs[ErrCodeRequestQueryParamParse], err))
		return err
	}

	return nil
}

// float64ToNullFloat64 is a utility function for converting string to sql.NullFloat64
func float64ToNullFloat64(f float64) *sql.NullFloat64 {
	return &sql.NullFloat64{Float64: f, Valid: f != 0}
}

// int64ToNullInt64 is a utility function for converting string to sql.NullInt64
func int64ToNullInt64(i int64) *sql.NullInt64 {
	return &sql.NullInt64{Int64: i, Valid: i != 0}
}

// stringToNullString is a utility function for converting string to sql.NullString
func stringToNullString(s string) *sql.NullString {
	return &sql.NullString{String: s, Valid: s != ""}
}
