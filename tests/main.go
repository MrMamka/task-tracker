package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func main() {
	addr := flag.String("addr", "http://localhost:8080", "Address of user service")
	flag.Parse()

	generator := NewGenerator()
	login := generator.GenerateLogin()
	password := generator.GeneratePassword()

	manager := NewTaskManager(*addr)
	manager.RegisterAndLogin(login, password)

	manager.TestAddAndCheckTask()
	manager.TestLikeAndCheckTask()

	fmt.Println("All tests passed")
}

type Generator struct {
	*rand.Rand
}

func NewGenerator() Generator {
	seed := time.Now().UnixNano()
	source := rand.NewSource(seed)
	return Generator{rand.New(source)}
}

func (g Generator) GeneratePassword() string {
	password := []byte("LongPassword12345")
	g.Shuffle(len(password), func(i, j int) {
		password[i], password[j] = password[j], password[i]
	})

	return string(password)
}

func (g Generator) GenerateLogin() string {
	suffix := g.Intn(1000000)
	return fmt.Sprintf("TestLogin%d", suffix)
}

type TaskManager struct {
	addr   string
	jwt    string
	login  string
	taskID int
	client *http.Client
}

func NewTaskManager(addr string) *TaskManager {
	return &TaskManager{addr: addr, client: &http.Client{}}
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (m *TaskManager) RegisterAndLogin(login, password string) {
	m.login = login

	request := RegisterRequest{
		Login:    login,
		Password: password,
	}
	requestBody, err := json.Marshal(request)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, m.addr+"/register", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		log.Fatal("non created status code")
	}

	req, err = http.NewRequest(http.MethodPost, m.addr+"/login", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err = m.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatal("non ok status code")
	}

	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "jwt" {
			m.jwt = cookie.Value
			return
		}
	}
}

type AddTaskRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type AddTaskResponse struct {
	ID int `json:"id"`
}

type GetTaskResponse struct {
	ID      int    `json:"id"`
	Author  string `json:"author"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (m *TaskManager) TestAddAndCheckTask() {
	title := "TestTitle"
	content := "TestContent"
	request := AddTaskRequest{
		Title:   title,
		Content: content,
	}
	requestBody, err := json.Marshal(request)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, m.addr+"/create-task", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "jwt="+m.jwt)
	resp, err := m.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var responseValue AddTaskResponse
	if err = json.Unmarshal(data, &responseValue); err != nil {
		log.Fatal(err)
	}

	m.taskID = responseValue.ID
	req, err = http.NewRequest(http.MethodGet, m.addr+"/task", nil)
	if err != nil {
		log.Fatal(err)
	}
	q := req.URL.Query()
	q.Set("id", strconv.Itoa(m.taskID))
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Cookie", "jwt="+m.jwt)
	resp, err = m.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.StatusCode)
	}
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var taskValue GetTaskResponse
	if err = json.Unmarshal(data, &taskValue); err != nil {
		log.Fatal(err)
	}

	if taskValue.Author != m.login ||
		taskValue.Title != title ||
		taskValue.Content != content ||
		taskValue.ID != m.taskID {
		log.Fatalf("wrong task info; got: %#v", taskValue)
	}
}

type TaskStats struct {
	ID    int `json:"id"`
	Likes int `json:"likes"`
	Views int `json:"views"`
}

func (m *TaskManager) TestLikeAndCheckTask() {
	req, err := http.NewRequest(http.MethodPost, m.addr+"/like", nil)
	if err != nil {
		log.Fatal(err)
	}
	q := req.URL.Query()
	q.Set("id", strconv.Itoa(m.taskID))
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Cookie", "jwt="+m.jwt)
	resp, err := m.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.StatusCode)
	}

	time.Sleep(time.Second)

	req, err = http.NewRequest(http.MethodGet, m.addr+"/task-stats", nil)
	if err != nil {
		log.Fatal(err)
	}
	q = req.URL.Query()
	q.Set("id", strconv.Itoa(m.taskID))
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Cookie", "jwt="+m.jwt)
	resp, err = m.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var taskValue TaskStats
	if err = json.Unmarshal(data, &taskValue); err != nil {
		log.Fatal(err)
	}

	if taskValue.ID != m.taskID ||
		taskValue.Likes != 1 ||
		taskValue.Views != 0 {
		log.Fatalf("wrong task stats; got: %#v", taskValue)
	}
}
