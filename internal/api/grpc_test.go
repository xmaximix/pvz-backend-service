package api

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"pvz-backend-service/internal/model"
)

type stubRepo struct {
	listFn func(ctx context.Context, start, end string, limit, offset int) ([]model.PVZ, error)
}

func (s *stubRepo) CreateUser(ctx context.Context, email, hash, role string) (model.User, error) {
	panic("not used")
}
func (s *stubRepo) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	panic("not used")
}
func (s *stubRepo) CreatePVZ(ctx context.Context, city string) (model.PVZ, error) {
	panic("not used")
}
func (s *stubRepo) ListPVZ(ctx context.Context, start, end string, limit, offset int) ([]model.PVZ, error) {
	return s.listFn(ctx, start, end, limit, offset)
}
func (s *stubRepo) OpenReception(ctx context.Context, pvzID string) (model.Reception, error) {
	panic("not used")
}
func (s *stubRepo) GetOpenReception(ctx context.Context, pvzID string) (model.Reception, error) {
	panic("not used")
}
func (s *stubRepo) AddProduct(ctx context.Context, receptionID, typ string) (model.Product, error) {
	panic("not used")
}
func (s *stubRepo) DeleteLastProduct(ctx context.Context, receptionID string) error {
	panic("not used")
}
func (s *stubRepo) CloseReception(ctx context.Context, receptionID string) error {
	panic("not used")
}

func TestGetPVZList_Success(t *testing.T) {
	entries := []model.PVZ{
		{ID: "p1", City: "Москва"},
		{ID: "p2", City: "Казань"},
	}
	stub := &stubRepo{
		listFn: func(ctx context.Context, start, end string, limit, offset int) ([]model.PVZ, error) {
			assert.Equal(t, "", start)
			assert.Equal(t, "", end)
			assert.Equal(t, 100, limit)
			assert.Equal(t, 0, offset)
			return entries, nil
		},
	}

	server := &grpcServer{store: stub}
	resp, err := server.GetPVZList(context.Background(), &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Len(t, resp.Pvz, 2)
	assert.Equal(t, "p1", resp.Pvz[0].Id)
	assert.Equal(t, "Москва", resp.Pvz[0].City)
	assert.Equal(t, "p2", resp.Pvz[1].Id)
	assert.Equal(t, "Казань", resp.Pvz[1].City)
}

func TestGetPVZList_Failure(t *testing.T) {
	stub := &stubRepo{
		listFn: func(ctx context.Context, start, end string, limit, offset int) ([]model.PVZ, error) {
			return nil, errors.New("db error")
		},
	}

	server := &grpcServer{store: stub}
	resp, err := server.GetPVZList(context.Background(), &emptypb.Empty{})
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "failed to list PVZs")
}
