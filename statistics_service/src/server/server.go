package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sort"
	pb "statistics/proto/statistic"
	"statistics/src/database"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	pb.UnimplementedStatisticsServiceServer
	db *database.DataBase
}

func New(db *database.DataBase) *Server {
	return &Server{
		db: db,
	}
}

func (s *Server) GetTaskStats(ctx context.Context, req *pb.GetTaskStatsRequest) (*pb.GetTaskStatsResponse, error) {
	likes, _ := s.db.CountLikes(uint(req.Id))
	views, err := s.db.CountViews(uint(req.Id))

	return &pb.GetTaskStatsResponse{
		Likes: int32(likes),
		Views: int32(views),
	}, err
}

func (s *Server) GetTopTasks(ctx context.Context, req *pb.GetTopTasksRequest) (*pb.GetTopTasksResponse, error) {
	var err error
	var topTasks []database.TaskIDCount

	if req.Sort == pb.SortBy_likes {
		topTasks, err = s.db.TopByLikes(5)
	} else {
		topTasks, err = s.db.TopByViews(5)
	}
	if err != nil {
		return nil, err
	}

	result := make([]*pb.Task, 0, len(topTasks))
	for _, task := range topTasks {
		result = append(result, &pb.Task{
			Id:           uint32(task.TaskID),
			LikesOrViews: int32(task.Count),
			Author:       req.Authors[uint32(task.TaskID)],
		})
	}

	return &pb.GetTopTasksResponse{
		Tasks: result,
	}, nil
}

func (s *Server) GetTopUsers(ctx context.Context, req *pb.GetTopUsersRequest) (*pb.GetTopUsersResponse, error) {
	likes, err := s.db.GroupedLikes()
	if err != nil {
		return nil, err
	}
	authorsLikes := make(map[string]int64)
	for _, like := range likes {
		authorsLikes[req.Authors[uint32(like.TaskID)]] += like.Count
	}
	var authors []*pb.Author
	for login, count := range authorsLikes {
		authors = append(authors, &pb.Author{
			Login: login,
			Likes: int32(count),
		})
	}
	sort.Slice(authors, func(i, j int) bool {
		return authors[i].Likes > authors[j].Likes
	})
	if len(authors) > 3 {
		authors = authors[:3]
	}

	return &pb.GetTopUsersResponse{
		Authors: authors,
	}, nil
}

func (s *Server) RegisterAndListen(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	pb.RegisterStatisticsServiceServer(grpcServer, s)

	fmt.Println("Statistics service server started.")
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
