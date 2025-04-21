package repo

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func setupMockRepo(t *testing.T) (Repository, pgxmock.PgxPoolIface) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	return New(mock), mock
}

func TestCreatePVZ_Success(t *testing.T) {
	r, mock := setupMockRepo(t)
	cols := []string{"registration_date"}
	mock.ExpectQuery(regexp.QuoteMeta(
		"INSERT INTO pvz (id,city) VALUES ($1,$2) RETURNING registration_date",
	)).
		WithArgs(pgxmock.AnyArg(), "Москва").
		WillReturnRows(pgxmock.NewRows(cols).AddRow(time.Date(2025, 4, 19, 0, 0, 0, 0, time.UTC)))

	pvz, err := r.CreatePVZ(context.Background(), "Москва")
	assert.NoError(t, err)
	assert.Equal(t, "Москва", pvz.City)
}

func TestCreatePVZ_InvalidCity(t *testing.T) {
	r, _ := setupMockRepo(t)
	_, err := r.CreatePVZ(context.Background(), "London")
	assert.Error(t, err)
}

func TestListPVZ_NoFilter(t *testing.T) {
	r, mock := setupMockRepo(t)
	rows := pgxmock.NewRows([]string{"id", "city", "registration_date"}).
		AddRow("p1", "Москва", time.Date(2025, 4, 19, 0, 0, 0, 0, time.UTC)).
		AddRow("p2", "Казань", time.Date(2025, 4, 20, 0, 0, 0, 0, time.UTC))
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, city, registration_date FROM pvz ORDER BY registration_date DESC LIMIT 10 OFFSET 0",
	)).
		WillReturnRows(rows)

	pvzs, err := r.ListPVZ(context.Background(), "", "", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, pvzs, 2)
	assert.Equal(t, "p1", pvzs[0].ID)
}

func TestListPVZ_WithFilter(t *testing.T) {
	r, mock := setupMockRepo(t)
	rows := pgxmock.NewRows([]string{"id", "city", "registration_date"}).
		AddRow("p3", "СПб", time.Date(2025, 4, 21, 0, 0, 0, 0, time.UTC))
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, city, registration_date FROM pvz WHERE (registration_date >= $1 AND registration_date <= $2) ORDER BY registration_date DESC LIMIT 5 OFFSET 1",
	)).
		WithArgs("2025-04-19T00:00:00Z", "2025-04-21T23:59:59Z").
		WillReturnRows(rows)

	pvzs, err := r.ListPVZ(
		context.Background(),
		"2025-04-19T00:00:00Z",
		"2025-04-21T23:59:59Z",
		5, 1,
	)
	assert.NoError(t, err)
	assert.Len(t, pvzs, 1)
	assert.Equal(t, "p3", pvzs[0].ID)
}

func TestOpenReception_Conflict(t *testing.T) {
	r, mock := setupMockRepo(t)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT COUNT(*) FROM reception WHERE pvz_id=$1 AND status='in_progress'",
	)).
		WithArgs("p1").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

	_, err := r.OpenReception(context.Background(), "p1")
	assert.EqualError(t, err, "open reception exists")
}

func TestOpenReception_Success(t *testing.T) {
	r, mock := setupMockRepo(t)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT COUNT(*) FROM reception WHERE pvz_id=$1 AND status='in_progress'",
	)).
		WithArgs("p1").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(
		"INSERT INTO reception (id,pvz_id,status) VALUES ($1,$2,'in_progress') RETURNING date_time",
	)).
		WithArgs(pgxmock.AnyArg(), "p1").
		WillReturnRows(pgxmock.NewRows([]string{"date_time"}).AddRow(time.Now().UTC()))

	rec, err := r.OpenReception(context.Background(), "p1")
	assert.NoError(t, err)
	assert.Equal(t, "p1", rec.PVZID)
}

func TestAddProduct_Success(t *testing.T) {
	r, mock := setupMockRepo(t)
	mock.ExpectQuery(regexp.QuoteMeta(
		"INSERT INTO product (id,reception_id,type) VALUES ($1,$2,$3) RETURNING date_time",
	)).
		WithArgs(pgxmock.AnyArg(), "r1", "электроника").
		WillReturnRows(pgxmock.NewRows([]string{"date_time"}).AddRow(time.Now().UTC()))

	prod, err := r.AddProduct(context.Background(), "r1", "электроника")
	assert.NoError(t, err)
	assert.Equal(t, "r1", prod.ReceptionID)
}

func TestDeleteLastProduct_Success(t *testing.T) {
	r, mock := setupMockRepo(t)
	mock.ExpectExec(regexp.QuoteMeta(
		"DELETE FROM product WHERE id IN ( SELECT id FROM product WHERE reception_id=$1 ORDER BY date_time DESC, id DESC LIMIT 1 )",
	)).
		WithArgs("r1").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := r.DeleteLastProduct(context.Background(), "r1")
	assert.NoError(t, err)
}

func TestCloseReception_AlreadyClosed(t *testing.T) {
	r, mock := setupMockRepo(t)
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE reception SET status='close' WHERE id=$1 AND status='in_progress'",
	)).
		WithArgs("r1").
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err := r.CloseReception(context.Background(), "r1")
	assert.EqualError(t, err, "reception already closed")
}

func TestCreateUser_Success(t *testing.T) {
	r, mock := setupMockRepo(t)
	mock.ExpectExec(regexp.QuoteMeta(
		"INSERT INTO users (id,email,password_hash,role) VALUES ($1,$2,$3,$4)",
	)).
		WithArgs(pgxmock.AnyArg(), "a@b", "hash", "employee").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	user, err := r.CreateUser(context.Background(), "a@b", "hash", "employee")
	assert.NoError(t, err)
	assert.Equal(t, "a@b", user.Email)
}

func TestGetUserByEmail_Success(t *testing.T) {
	r, mock := setupMockRepo(t)
	cols := []string{"id", "email", "password_hash", "role"}
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id,email,password_hash,role FROM users WHERE email=$1",
	)).
		WithArgs("a@b").
		WillReturnRows(pgxmock.NewRows(cols).
			AddRow("u1", "a@b", "hsh", "employee"),
		)

	u, err := r.GetUserByEmail(context.Background(), "a@b")
	assert.NoError(t, err)
	assert.Equal(t, "u1", u.ID)
	assert.Equal(t, "employee", u.Role)
}

func TestGetOpenReception_NotFound(t *testing.T) {
	r, mock := setupMockRepo(t)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id,pvz_id,date_time,status FROM reception WHERE pvz_id=$1 AND status='in_progress'",
	)).
		WithArgs("p1").
		WillReturnRows(pgxmock.NewRows([]string{"id", "pvz_id", "date_time", "status"}))
	_, err := r.GetOpenReception(context.Background(), "p1")
	assert.Error(t, err)
}

func TestGetOpenReception_Success(t *testing.T) {
	r, mock := setupMockRepo(t)
	now := time.Now().UTC()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id,pvz_id,date_time,status FROM reception WHERE pvz_id=$1 AND status='in_progress'",
	)).
		WithArgs("p1").
		WillReturnRows(pgxmock.NewRows([]string{"id", "pvz_id", "date_time", "status"}).
			AddRow("r1", "p1", now, "in_progress"),
		)
	rec, err := r.GetOpenReception(context.Background(), "p1")
	assert.NoError(t, err)
	assert.Equal(t, "r1", rec.ID)
}

func TestCloseReception_Success(t *testing.T) {
	r, mock := setupMockRepo(t)
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE reception SET status='close' WHERE id=$1 AND status='in_progress'",
	)).
		WithArgs("r1").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := r.CloseReception(context.Background(), "r1")
	assert.NoError(t, err)
}

func TestDeleteLastProduct_None(t *testing.T) {
	r, mock := setupMockRepo(t)
	mock.ExpectExec(regexp.QuoteMeta(
		"DELETE FROM product WHERE id IN ( SELECT id FROM product WHERE reception_id=$1 ORDER BY date_time DESC, id DESC LIMIT 1 )",
	)).
		WithArgs("r1").
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := r.DeleteLastProduct(context.Background(), "r1")
	assert.NoError(t, err)
}
