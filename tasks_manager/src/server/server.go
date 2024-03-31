package server

import (
	"context"
	"fmt"
	"log"
	"net"
	pb "tasksmanager/proto"
	"tasksmanager/src/database"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedTaskServiceServer
	db *database.DataBase
}

func (s *Server) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	data := &database.TaskData{
		Author:  req.Author,
		Title:   req.Title,
		Content: req.Content,
	}
	id, err := s.db.CreateTask(data)
	return &pb.CreateTaskResponse{Id: id}, err
}

func (s *Server) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.Task, error) {
	data, err := s.db.GetTaskData(uint(req.Id))
	return &pb.Task{
		Id:           req.Id,
		Author:       data.Author,
		Title:        data.Title,
		Content:      data.Content,
		CreationTime: timestamppb.New(data.CreationTime),
	}, err
}

func (s *Server) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*emptypb.Empty, error) {
	err := s.db.UpdateTaskData(&database.TaskData{
		ID:      uint(req.Id),
		Content: req.Content,
		Title:   req.Title,
	})
	return nil, err
}

func (s *Server) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*emptypb.Empty, error) {
	return nil, s.db.DeleteTask(uint(req.Id))
}

func (s *Server) GetTasks(*emptypb.Empty, pb.TaskService_GetTasksServer) error {
	return nil
}

func New(db *database.DataBase) *Server {
	return &Server{
		db: db,
	}
}

func (s *Server) RegisterAndListen(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	pb.RegisterTaskServiceServer(grpcServer, s)

	fmt.Println("Tasks manager server started.")
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
