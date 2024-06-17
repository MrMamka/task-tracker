package database

import (
	"errors"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var ErrPermissionDenied = errors.New("permission denied")

type DataBase struct {
	*gorm.DB
}

type taskInfo struct {
	gorm.Model
	Author  string
	Title   string
	Content string
}

type UserData struct {
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	BirthDay    string `json:"bith_day"`
	Mail        string `json:"mail"`
	PhoneNumber string `json:"phone_number"`
}

type TaskData struct {
	ID           uint
	Author       string
	Title        string
	Content      string
	CreationTime time.Time
}

func (ti taskInfo) toTaskData() TaskData {
	return TaskData{
		ID:           ti.ID,
		Author:       ti.Author,
		Title:        ti.Title,
		Content:      ti.Content,
		CreationTime: ti.Model.CreatedAt,
	}
}

func New() *DataBase {
	dsn := "host=task_db dbname=task_db sslmode=disable user=user password=password"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect tasks database: " + err.Error())
	}

	db.AutoMigrate(&taskInfo{})

	return &DataBase{db}
}

func (db *DataBase) CreateTask(data *TaskData) (uint32, error) {
	info := &taskInfo{Author: data.Author, Title: data.Title, Content: data.Content}
	result := db.Create(info)
	return uint32(info.ID), result.Error
}

func (db *DataBase) CheckTaskPermission(id uint, author string) error {
	info := &taskInfo{Model: gorm.Model{ID: id}}
	result := db.First(&info, "ID = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if info.Author != author {
		return ErrPermissionDenied
	}
	return nil
}

func (db *DataBase) GetTaskData(id uint, author string) (*TaskData, error) {
	info := &taskInfo{Model: gorm.Model{ID: id}}
	result := db.First(&info, "ID = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	data := info.toTaskData()
	return &data, nil
}

func (db *DataBase) UpdateTaskData(data *TaskData) error {
	if err := db.CheckTaskPermission(data.ID, data.Author); err != nil {
		return err
	}
	result := db.Model(&taskInfo{}).Where("ID = ?", data.ID).Updates(taskInfo{
		Title:   data.Title,
		Content: data.Content,
	})
	if result.Error != nil {
		return result.Error
	}
	return db.Model(&taskInfo{}).Where("ID = ?", data.ID).Updates(taskInfo{
		Title:   data.Title,
		Content: data.Content,
	}).Error
}

func (db *DataBase) DeleteTask(id uint, author string) error {
	if err := db.CheckTaskPermission(id, author); err != nil {
		return err
	}
	return db.Delete(&taskInfo{}, id).Error
}

func (db *DataBase) GetTasks(offset, batchSize int) ([]TaskData, error) {
	var tasks []taskInfo
	result := db.Limit(batchSize).Offset(offset).Find(&tasks)
	data := make([]TaskData, 0, len(tasks))
	for _, info := range tasks {
		data = append(data, info.toTaskData())
	}
	return data, result.Error
}
