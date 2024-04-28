package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	pb "userservice/proto"
	"userservice/src/auth"
	"userservice/src/broker"
	"userservice/src/database"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	mux     *chi.Mux
	db      *database.DataBase
	auth    *auth.AuthService
	taskMan pb.TaskServiceClient
	broker  *broker.Broker
}

func New(db *database.DataBase) (*Server, func()) {
	broker, close := broker.New()
	return &Server{
		mux:    chi.NewRouter(),
		db:     db,
		auth:   auth.New(db),
		broker: broker,
	}, close
}

func (s *Server) Register() {
	s.mux.Post("/register", s.register)
	s.mux.Post("/login", s.login)
	s.mux.Put("/update-info", s.updateInfo)
	s.mux.Get("/info", s.getInfo)

	s.mux.Post("/create-task", s.createTask)
	s.mux.Get("/task", s.getTask)
	s.mux.Put("/update-task-info", s.updateTaskInfo)
	s.mux.Delete("/task", s.deleteTask)
	s.mux.Get("/tasks", s.getTasks)

	s.mux.Post("/like", s.addLike)
	s.mux.Post("/view", s.addView)
}

func (s *Server) Listen(addr string) {
	tasksManAddr := os.Getenv("TASKS_MANAGER_ADDR")
	if tasksManAddr == "" {
		tasksManAddr = "tasks_manager:8081"
	}
	conn, err := grpc.Dial(tasksManAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	s.taskMan = pb.NewTaskServiceClient(conn)

	fmt.Println("Server started.")
	err = http.ListenAndServe(addr, s.mux)
	log.Fatalf("Server stopped: %v", err)
}

type LoginPassword struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var user LoginPassword
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Can not parse body: %v", err)
		return
	}
	defer r.Body.Close()

	login := user.Login
	password := user.Password

	if err := s.auth.ValidateLogin(login); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad login: %v", err)
		return
	}

	if err := s.auth.ValidatePassword(password); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad password: %v", err)
		return
	}

	if err := s.auth.CreateUser(login, password); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Can not create user: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var user LoginPassword
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Can not parse body: %v", err)
		return
	}
	defer r.Body.Close()

	login := user.Login
	password := user.Password

	if err := s.auth.CheckPassword(login, password); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Wrong login or password: %v", err)
		return
	}

	token, err := s.auth.CreateToken(login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Can not generate token: %v", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "jwt",
		Value: token,
	})
	w.WriteHeader(http.StatusOK)
}

func (s *Server) updateInfo(w http.ResponseWriter, r *http.Request) {
	login, ok := s.auth.CheckAuth(w, r)
	if !ok {
		return
	}

	var userData database.UserData
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Can not parse body: %v", err)
		return
	}
	defer r.Body.Close()

	if err := s.auth.ValidateUserData(&userData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad user data: %v", err)
		return
	}

	err := s.db.UpdateUserData(login, &userData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Can not update user data: %v", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) getInfo(w http.ResponseWriter, r *http.Request) {
	login, ok := s.auth.CheckAuth(w, r)
	if !ok {
		return
	}

	data, err := s.db.GetUserData(login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Can not get user data: %v", err)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(login)
	encoder.Encode(data)
}

type TaskData struct {
	ID           uint       `json:"id,omitempty"`
	Author       string     `json:"author,omitempty"`
	Title        string     `json:"title,omitempty"`
	Content      string     `json:"content,omitempty"`
	CreationTime *time.Time `json:"creation_time,omitempty"`
}

func protoToTaskData(task *pb.Task) TaskData {
	time := task.CreationTime.AsTime()
	return TaskData{
		ID:           uint(task.Id),
		Author:       task.Author,
		Title:        task.Title,
		Content:      task.Content,
		CreationTime: &time,
	}
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	login, ok := s.auth.CheckAuth(w, r)
	if !ok {
		return
	}

	var taskData TaskData
	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Can not parse body: %v", err)
		return
	}
	defer r.Body.Close()

	resp, err := s.taskMan.CreateTask(context.Background(), &pb.CreateTaskRequest{
		Author:  login,
		Title:   taskData.Title,
		Content: taskData.Content,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Can not create new task: %v", err)
		return
	}

	json.NewEncoder(w).Encode(TaskData{ID: uint(resp.Id)})
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	login, ok := s.auth.CheckAuth(w, r)
	if !ok {
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Can not parse query: %v", err)
		return
	}

	resp, err := s.taskMan.GetTask(context.Background(), &pb.GetTaskRequest{
		Id:     uint32(id),
		Author: login,
	})
	if err != nil {
		if strings.Contains(err.Error(), "permission denied") {
			w.WriteHeader(http.StatusForbidden)
		}
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Can not get task: %v", err)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(protoToTaskData(resp))
	w.WriteHeader(http.StatusOK)
}

func (s *Server) updateTaskInfo(w http.ResponseWriter, r *http.Request) {
	login, ok := s.auth.CheckAuth(w, r) // TODO: check login
	if !ok {
		return
	}

	var taskData TaskData
	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Can not parse body: %v", err)
		return
	}
	defer r.Body.Close()

	if taskData.ID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "ID is required")
		return
	}

	_, err := s.taskMan.UpdateTask(context.Background(), &pb.UpdateTaskRequest{
		Author:  login,
		Id:      uint32(taskData.ID),
		Title:   taskData.Title,
		Content: taskData.Content,
	})
	if err != nil {
		if strings.Contains(err.Error(), "permission denied") {
			w.WriteHeader(http.StatusForbidden)
		}
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Can not update task: %v", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) deleteTask(w http.ResponseWriter, r *http.Request) {
	login, ok := s.auth.CheckAuth(w, r)
	if !ok {
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Can not parse query: %v", err)
		return
	}

	_, err = s.taskMan.DeleteTask(context.Background(), &pb.DeleteTaskRequest{
		Id:     uint32(id),
		Author: login,
	})
	if err != nil {
		if strings.Contains(err.Error(), "permission denied") {
			w.WriteHeader(http.StatusForbidden)
		}
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Can not delete task: %v", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) getTasks(w http.ResponseWriter, r *http.Request) {
	login, ok := s.auth.CheckAuth(w, r)
	if !ok {
		return
	}

	batchSizeStr := r.URL.Query().Get("batch_size")
	if batchSizeStr == "" {
		batchSizeStr = "1"
	}
	batchSize, err := strconv.Atoi(batchSizeStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Can not parse query: %v", err)
		return
	}

	offsetStr := r.URL.Query().Get("offset")
	if batchSizeStr == "" {
		batchSizeStr = "0"
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Can not parse query: %v", err)
		return
	}

	tasksResp, err := s.taskMan.GetTasks(context.Background(), &pb.GetTasksRequest{
		BatchSize: uint32(batchSize),
		Offset:    uint32(offset),
		Author:    login,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Can not delete task: %v", err)
		return
	}

	type TasksOffset struct {
		Tasks  []TaskData `json:"tasks"`
		Offset uint32     `json:"offset"`
	}

	tasksOffset := TasksOffset{
		Offset: tasksResp.Offset,
		Tasks:  make([]TaskData, 0, len(tasksResp.Tasks)),
	}

	for _, task := range tasksResp.Tasks {
		tasksOffset.Tasks = append(tasksOffset.Tasks, protoToTaskData(task))
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(tasksOffset)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) addLike(w http.ResponseWriter, r *http.Request) {
	login, ok := s.auth.CheckAuth(w, r)
	if !ok {
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Can not parse query: %v", err)
		return
	}

	s.broker.SendLike(broker.Statistic{
		Login: login,
		TaskID: uint(id),
	})
}

func (s *Server) addView(w http.ResponseWriter, r *http.Request) {
	login, ok := s.auth.CheckAuth(w, r)
	if !ok {
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Can not parse query: %v", err)
		return
	}

	s.broker.SendView(broker.Statistic{
		Login: login,
		TaskID: uint(id),
	})
}
