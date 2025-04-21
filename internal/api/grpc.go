package api

import (
	"context"

	pvzpb "pvz-backend-service/api/pvz/v1"
	"pvz-backend-service/internal/repo"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type grpcServer struct {
	pvzpb.UnimplementedPVZServiceServer
	store repo.Repository
}

func RegisterGRPC(s grpc.ServiceRegistrar, store repo.Repository) {
	srv := &grpcServer{store: store}
	pvzpb.RegisterPVZServiceServer(s, srv)
}

func (g *grpcServer) GetPVZList(ctx context.Context, _ *emptypb.Empty) (*pvzpb.GetPVZListResponse, error) {
	pvzs, err := g.store.ListPVZ(ctx, "", "", 100, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list PVZs: %v", err)
	}
	resp := &pvzpb.GetPVZListResponse{}
	for _, p := range pvzs {
		resp.Pvz = append(resp.Pvz, &pvzpb.PVZ{
			Id:   p.ID,
			City: p.City,
		})
	}
	return resp, nil
}
