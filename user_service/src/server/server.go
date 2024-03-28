package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"userservice/src/auth"
	"userservice/src/database"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	mux  *chi.Mux
	db   *database.DataBase
	auth *auth.AuthService
}

func New(db *database.DataBase) *Server {
	return &Server{
		mux:  chi.NewRouter(),
		db:   db,
		auth: auth.New(db),
	}
}

func (s *Server) Register() {
	s.mux.Post("/register", s.register)
	s.mux.Post("/login", s.login)
	s.mux.Put("/update-info", s.updateInfo)
	s.mux.Get("/info", s.getInfo)
}

func (s *Server) Listen(addr string) {
	fmt.Println("Server started.")
	err := http.ListenAndServe(addr, s.mux)
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
