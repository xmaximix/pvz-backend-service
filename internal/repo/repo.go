package repo

import (
	"context"
	"errors"
	"pvz-backend-service/internal/model"
	"pvz-backend-service/lib/e"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

var _ Repository = (*repo)(nil)

type Repository interface {
	CreateUser(ctx context.Context, email, hash, role string) (model.User, error)
	GetUserByEmail(ctx context.Context, email string) (model.User, error)
	CreatePVZ(ctx context.Context, city string) (model.PVZ, error)
	ListPVZ(ctx context.Context, start, end string, limit, offset int) ([]model.PVZ, error)
	OpenReception(ctx context.Context, pvzID string) (model.Reception, error)
	GetOpenReception(ctx context.Context, pvzID string) (model.Reception, error)
	AddProduct(ctx context.Context, receptionID, typ string) (model.Product, error)
	DeleteLastProduct(ctx context.Context, receptionID string) error
	CloseReception(ctx context.Context, receptionID string) error
}

type repo struct {
	db DB
	sb sq.StatementBuilderType
}

func New(db DB) Repository {
	return &repo{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *repo) CreateUser(ctx context.Context, email, hash, role string) (model.User, error) {
	id := uuid.NewString()
	sql, args, _ := r.sb.
		Insert("users").
		Columns("id", "email", "password_hash", "role").
		Values(id, email, hash, role).
		ToSql()
	_, err := r.db.Exec(ctx, sql, args...)
	return model.User{ID: id, Email: email, PasswordHash: hash, Role: role}, e.WrapIfErr("create user", err)
}

func (r *repo) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	row := r.db.QueryRow(ctx, "SELECT id,email,password_hash,role FROM users WHERE email=$1", email)
	var u model.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role); err != nil {
		return u, err
	}
	return u, nil
}

func (r *repo) CreatePVZ(ctx context.Context, city string) (model.PVZ, error) {
	if city != "Москва" && city != "Санкт-Петербург" && city != "Казань" {
		return model.PVZ{}, e.Wrap("invalid city", nil)
	}
	id := uuid.NewString()
	sql, args, _ := r.sb.
		Insert("pvz").
		Columns("id", "city").
		Values(id, city).
		Suffix("RETURNING registration_date").
		ToSql()
	var reg time.Time
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&reg); err != nil {
		return model.PVZ{}, e.WrapIfErr("create pvz", err)
	}
	return model.PVZ{ID: id, City: city, RegistrationDate: reg}, nil
}

func (r *repo) ListPVZ(ctx context.Context, start, end string, limit, offset int) ([]model.PVZ, error) {
	b := r.sb.
		Select("id", "city", "registration_date").
		From("pvz").
		OrderBy("registration_date DESC").
		Limit(uint64(limit)).Offset(uint64(offset))
	if start != "" && end != "" {
		b = b.Where(sq.And{sq.GtOrEq{"registration_date": start}, sq.LtOrEq{"registration_date": end}})
	}
	sql, args, _ := b.ToSql()
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, e.WrapIfErr("list pvz", err)
	}
	defer rows.Close()
	var res []model.PVZ
	for rows.Next() {
		var p model.PVZ
		rows.Scan(&p.ID, &p.City, &p.RegistrationDate)
		res = append(res, p)
	}
	return res, nil
}

func (r *repo) OpenReception(ctx context.Context, pvzID string) (model.Reception, error) {
	var cnt int
	if err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM reception WHERE pvz_id=$1 AND status='in_progress'", pvzID,
	).Scan(&cnt); err != nil {
		return model.Reception{}, err
	}
	if cnt > 0 {
		return model.Reception{}, errors.New("open reception exists")
	}

	id := uuid.NewString()
	var dt time.Time
	if err := r.db.QueryRow(ctx,
		"INSERT INTO reception (id,pvz_id,status) VALUES ($1,$2,'in_progress') RETURNING date_time",
		id, pvzID,
	).Scan(&dt); err != nil {
		return model.Reception{}, err
	}
	return model.Reception{ID: id, PVZID: pvzID, DateTime: dt, Status: "in_progress"}, nil
}

func (r *repo) GetOpenReception(ctx context.Context, pvzID string) (model.Reception, error) {
	row := r.db.QueryRow(ctx, "SELECT id,pvz_id,date_time,status FROM reception WHERE pvz_id=$1 AND status='in_progress'", pvzID)
	var rec model.Reception
	if err := row.Scan(&rec.ID, &rec.PVZID, &rec.DateTime, &rec.Status); err != nil {
		return rec, err
	}
	return rec, nil
}

func (r *repo) AddProduct(ctx context.Context, receptionID, typ string) (model.Product, error) {
	id := uuid.NewString()
	sql, args, _ := r.sb.
		Insert("product").
		Columns("id", "reception_id", "type").
		Values(id, receptionID, typ).
		Suffix("RETURNING date_time").
		ToSql()
	var dt time.Time
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&dt); err != nil {
		return model.Product{}, e.WrapIfErr("add product", err)
	}
	return model.Product{ID: id, ReceptionID: receptionID, DateTime: dt, Type: typ}, nil
}

func (r *repo) DeleteLastProduct(ctx context.Context, receptionID string) error {
	_, err := r.db.Exec(ctx, `
        DELETE FROM product
        WHERE id IN (
          SELECT id FROM product
          WHERE reception_id=$1
          ORDER BY date_time DESC, id DESC
          LIMIT 1
        )`, receptionID,
	)
	return err
}

func (r *repo) CloseReception(ctx context.Context, receptionID string) error {
	tag, err := r.db.Exec(ctx, "UPDATE reception SET status='close' WHERE id=$1 AND status='in_progress'", receptionID)
	if err != nil {
		return e.Wrap("close reception", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("reception already closed")
	}
	return nil
}
