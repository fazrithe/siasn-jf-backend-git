package metricutil

import (
	"context"
	"database/sql"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

// A GenericSqlMetrics defines a type that implements counter for total SQL transaction, in-flight transactions, and
// transaction durations.
type GenericSqlMetrics interface {
	IncTxTotalCommitted()
	IncTxTotalRolledBack()
	IncTxInFlight()
	DecTxInFlight()
	// Only committed transaction should be measured.
	ObserveTxDuration(duration float64)
}

type DefaultGenericSqlMetrics struct {
	TxTotal           *prometheus.CounterVec
	TxInFlight        prometheus.Gauge
	TxDurationSeconds prometheus.Observer
}

func (gsm *DefaultGenericSqlMetrics) IncTxTotalCommitted() {
	gsm.TxTotal.WithLabelValues("committed").Inc()
}

func (gsm *DefaultGenericSqlMetrics) IncTxTotalRolledBack() {
	gsm.TxTotal.WithLabelValues("rolledback").Inc()
}

func (gsm *DefaultGenericSqlMetrics) IncTxInFlight() {
	gsm.TxInFlight.Inc()
}

func (gsm *DefaultGenericSqlMetrics) DecTxInFlight() {
	gsm.TxInFlight.Dec()
}

func (gsm *DefaultGenericSqlMetrics) ObserveTxDuration(duration float64) {
	gsm.TxDurationSeconds.Observe(duration)
}

// DbHandler interface encompasses all actions related to querying and executing SQL queries.
// This include preparing, query, and exec. You can pass everywhere DbHandler instead of DB or Tx, when you do not
// really care which one.
type DbHandler interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (rows *sql.Rows, err error)
	Query(query string, args ...interface{}) (rows *sql.Rows, err error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) (row *sql.Row)
	QueryRow(query string, args ...interface{}) (row *sql.Row)
	ExecContext(ctx context.Context, query string, args ...interface{}) (result sql.Result, err error)
	Exec(query string, args ...interface{}) (result sql.Result, err error)
	PrepareContext(ctx context.Context, query string) (stmt *Stmt, err error)
	Prepare(query string) (stmt *Stmt, err error)
}

// DB is a wrapper for sql.DB that accepts GenericSqlMetrics.
// For Exec and Query, metrics are incremented and decremented automatically. However, for prepared
// statements using Prepare, a Stmt wrapper is returned instead which will increment or decrement
// metric when you call Stmt.Exec. In prepared statements, only in-flight metric is incremented during preparation,
// which means that in-flight metric measurement spans the prepare to exec code block. If the prepared statement
// itself fails during preparation, it will be decremented immediately and considered a rolled back transaction.
//
// The supplied GenericSqlMetrics can be nil, which makes DB acts like a normal DB.
type DB struct {
	*sql.DB
	genericSqlMetrics GenericSqlMetrics
}

// Stmt is a wrapper to sql.Stmt.
// See also DB.
type Stmt struct {
	*sql.Stmt
	genericSqlMetrics GenericSqlMetrics
	StartTime         time.Time
}

func NewDB(db *sql.DB, metrics GenericSqlMetrics) *DB {
	return &DB{
		DB:                db,
		genericSqlMetrics: metrics,
	}
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (rows *sql.Rows, err error) {
	if db.genericSqlMetrics == nil {
		return db.DB.QueryContext(ctx, query, args...)
	}

	startTime := time.Now()
	db.genericSqlMetrics.IncTxInFlight()
	rows, err = db.DB.QueryContext(ctx, query, args...)
	db.genericSqlMetrics.DecTxInFlight()
	if err != nil {
		db.genericSqlMetrics.IncTxTotalRolledBack()
		return
	}

	db.genericSqlMetrics.IncTxTotalCommitted()
	db.genericSqlMetrics.ObserveTxDuration(time.Since(startTime).Seconds())
	return
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (result sql.Result, err error) {
	if db.genericSqlMetrics == nil {
		return db.DB.ExecContext(ctx, query, args...)
	}

	startTime := time.Now()
	db.genericSqlMetrics.IncTxInFlight()
	result, err = db.DB.ExecContext(ctx, query, args...)
	db.genericSqlMetrics.DecTxInFlight()
	if err != nil {
		db.genericSqlMetrics.IncTxTotalRolledBack()
		return
	}

	db.genericSqlMetrics.IncTxTotalCommitted()
	db.genericSqlMetrics.ObserveTxDuration(time.Since(startTime).Seconds())
	return
}

func (db *DB) PrepareContext(ctx context.Context, query string) (stmt *Stmt, err error) {
	if db.genericSqlMetrics == nil {
		nativeStmt, err := db.DB.PrepareContext(ctx, query)
		if err != nil {
			return nil, err
		}
		return &Stmt{Stmt: nativeStmt}, nil
	}

	startTime := time.Now()
	db.genericSqlMetrics.IncTxInFlight()
	nativeStmt, err := db.DB.PrepareContext(ctx, query)
	if err != nil {
		db.genericSqlMetrics.DecTxInFlight()
		db.genericSqlMetrics.IncTxTotalRolledBack()
		return nil, err
	}

	return &Stmt{Stmt: nativeStmt, StartTime: startTime, genericSqlMetrics: db.genericSqlMetrics}, nil
}

func (db *DB) Prepare(query string) (stmt *Stmt, err error) {
	return db.PrepareContext(context.Background(), query)
}

func (stmt *Stmt) ExecContext(ctx context.Context, args ...interface{}) (result sql.Result, err error) {
	if stmt.genericSqlMetrics == nil {
		return stmt.Stmt.ExecContext(ctx, args...)
	}

	result, err = stmt.Stmt.ExecContext(ctx, args...)
	stmt.genericSqlMetrics.DecTxInFlight()
	if err != nil {
		stmt.genericSqlMetrics.IncTxTotalRolledBack()
		return
	}

	stmt.genericSqlMetrics.IncTxTotalCommitted()
	stmt.genericSqlMetrics.ObserveTxDuration(time.Since(stmt.StartTime).Seconds())
	return
}

// A variant of Tx where the metric can be nil, in which case will just act like a normal sql.Tx.
type Tx struct {
	*sql.Tx
	genericSqlMetrics GenericSqlMetrics
	startTime         time.Time
}

// Creates a new Tx.
func NewTx(tx *sql.Tx, metrics GenericSqlMetrics) *Tx {
	return &Tx{
		Tx:                tx,
		genericSqlMetrics: metrics,
	}
}

func (tx *Tx) QueryContext(ctx context.Context, query string, args ...interface{}) (rows *sql.Rows, err error) {
	if tx.genericSqlMetrics == nil {
		return tx.Tx.QueryContext(ctx, query, args...)
	}

	startTime := time.Now()
	tx.genericSqlMetrics.IncTxInFlight()
	rows, err = tx.Tx.QueryContext(ctx, query, args...)
	tx.genericSqlMetrics.DecTxInFlight()
	if err != nil {
		tx.genericSqlMetrics.IncTxTotalRolledBack()
		return
	}

	tx.genericSqlMetrics.IncTxTotalCommitted()
	tx.genericSqlMetrics.ObserveTxDuration(time.Since(startTime).Seconds())
	return
}

func (tx *Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (result sql.Result, err error) {
	if tx.genericSqlMetrics == nil {
		return tx.Tx.ExecContext(ctx, query, args...)
	}

	startTime := time.Now()
	tx.genericSqlMetrics.IncTxInFlight()
	result, err = tx.Tx.ExecContext(ctx, query, args...)
	tx.genericSqlMetrics.DecTxInFlight()
	if err != nil {
		tx.genericSqlMetrics.IncTxTotalRolledBack()
		return
	}

	tx.genericSqlMetrics.IncTxTotalCommitted()
	tx.genericSqlMetrics.ObserveTxDuration(time.Since(startTime).Seconds())
	return
}

func (tx *Tx) PrepareContext(ctx context.Context, query string) (stmt *Stmt, err error) {
	if tx.genericSqlMetrics == nil {
		nativeStmt, err := tx.Tx.PrepareContext(ctx, query)
		if err != nil {
			return nil, err
		}
		return &Stmt{Stmt: nativeStmt}, nil
	}

	startTime := time.Now()
	tx.genericSqlMetrics.IncTxInFlight()
	nativeStmt, err := tx.Tx.PrepareContext(ctx, query)
	if err != nil {
		tx.genericSqlMetrics.DecTxInFlight()
		tx.genericSqlMetrics.IncTxTotalRolledBack()
		return nil, err
	}

	return &Stmt{Stmt: nativeStmt, StartTime: startTime, genericSqlMetrics: tx.genericSqlMetrics}, nil
}

func (tx *Tx) Prepare(query string) (stmt *Stmt, err error) {
	return tx.PrepareContext(context.Background(), query)
}

// Creates a new Tx, and starts observing the duration and in-flight.
func StartTx(tx *sql.Tx, metrics GenericSqlMetrics) (mtx *Tx) {
	mtx = NewTx(tx, metrics)
	if metrics != nil {
		mtx.StartObserving()
	}
	return mtx
}

func (tx *Tx) StartObserving() {
	tx.startTime = time.Now()
	tx.genericSqlMetrics.IncTxInFlight()
}

func (tx *Tx) Commit() (err error) {
	defer func() {
		if tx.genericSqlMetrics != nil {
			tx.genericSqlMetrics.IncTxTotalCommitted()
			tx.genericSqlMetrics.ObserveTxDuration(time.Since(tx.startTime).Seconds())
			tx.genericSqlMetrics.DecTxInFlight()
		}
	}()
	return tx.Tx.Commit()
}

func (tx *Tx) Rollback() (err error) {
	defer func() {
		if tx.genericSqlMetrics != nil {
			tx.genericSqlMetrics.IncTxTotalRolledBack()
			tx.genericSqlMetrics.DecTxInFlight()
		}
	}()
	return tx.Tx.Rollback()
}
