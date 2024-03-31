package database

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DataBase struct {
	*gorm.DB
}

type userInfo struct {
	gorm.Model
	Login        string
	PasswordHash []byte
	Name         string
	Surname      string
	BirthDay     string
	Mail         string
	PhoneNumber  string
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

func New() *DataBase {
	dsn := "host=task_db dbname=task_db sslmode=disable user=user password=password"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("failed to connect tasks database: " + err.Error())
	}

	db.AutoMigrate(&taskInfo{})

	return &DataBase{db}
}

// func (db *DataBase) UserExist(login string) (bool, error) {
// 	var info userInfo
// 	result := db.First(&info, "login = ?", login)
// 	if result.Error == gorm.ErrRecordNotFound {
// 		return false, nil
// 	} else if result.Error != nil {
// 		return false, result.Error
// 	}

// 	return true, nil
// }

func (db *DataBase) CreateTask(data *TaskData) (uint32, error) {
	info := &taskInfo{Author: data.Author, Title: data.Title, Content: data.Content}
	result := db.Create(info)
	return uint32(info.ID), result.Error
}

func (db *DataBase) GetTaskData(id uint) (*TaskData, error) {
	info := &taskInfo{Model: gorm.Model{ID: id}}
	result := db.First(&info, "ID = ?", id)
	return &TaskData{
		Author:       info.Author,
		Title:        info.Title,
		Content:      info.Content,
		CreationTime: info.Model.CreatedAt,
	}, result.Error
}

func (db *DataBase) UpdateTaskData(data *TaskData) error {
	result := db.Model(&taskInfo{}).Where("ID = ?", data.ID).Updates(taskInfo{
		Title:   data.Title,
		Content: data.Content,
	})
	return result.Error
}

func (db *DataBase) DeleteTask(id uint) error {
	result := db.Delete(&taskInfo{}, id)
	return result.Error
}

// func (db *DataBase) GetPasswordHash(login string) ([]byte, error) {
// 	info := &userInfo{}
// 	result := db.First(&info, "login = ?", login)
// 	return info.PasswordHash, result.Error
// }

// func (db *DataBase) UpdateUserData(login string, data *UserData) error {
// 	result := db.Model(&userInfo{}).Where("login = ?", login).Updates(userInfo{
// 		Name:        data.Name,
// 		Surname:     data.Surname,
// 		BirthDay:    data.BirthDay,
// 		Mail:        data.Mail,
// 		PhoneNumber: data.PhoneNumber,
// 	})
// 	return result.Error
// }

// func (db *DataBase) GetUserData(login string) (*UserData, error) {
// 	info := &userInfo{Login: login}
// 	result := db.First(&info, "login = ?", login)
// 	return &UserData{
// 		Name:        info.Name,
// 		Surname:     info.Surname,
// 		BirthDay:    info.BirthDay,
// 		Mail:        info.Mail,
// 		PhoneNumber: info.PhoneNumber,
// 	}, result.Error
// }
