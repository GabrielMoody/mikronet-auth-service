package gRPC

import (
	"context"

	"github.com/GabrielMoody/mikronet-auth-service/internal/pb"
	"github.com/GabrielMoody/mikronet-auth-service/internal/repository"
)

type GRPC struct {
	pb.UnimplementedAuthenticationServiceServer
	repo repository.AuthRepo
}

func NewgRPC(repo repository.AuthRepo) *GRPC {
	return &GRPC{
		repo: repo,
	}
}

func (a *GRPC) DeleteUser(ctx context.Context, req *pb.ReqByID) (res *pb.User, err error) {
	resRepo, err := a.repo.DeleteUser(ctx, req.Id)

	if err != nil {
		return nil, err
	}

	return &pb.User{
		Id: resRepo.ID,
	}, nil
}
