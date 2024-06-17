package database

import (

	//_ "github.com/lib/pq"
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

type UserData struct {
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	BirthDay    string `json:"bith_day"`
	Mail        string `json:"mail"`
	PhoneNumber string `json:"phone_number"`
}

func New() *DataBase {
	dsn := "host=user_db dbname=user_db sslmode=disable user=user password=password"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	db.AutoMigrate(&userInfo{})

	return &DataBase{db}
}

func (db *DataBase) UserExist(login string) (bool, error) {
	var info userInfo
	result := db.First(&info, "login = ?", login)
	if result.Error == gorm.ErrRecordNotFound {
		return false, nil
	} else if result.Error != nil {
		return false, result.Error
	}

	return true, nil
}

func (db *DataBase) CreateUser(login string, passwordHash []byte) error {
	info := &userInfo{Login: login, PasswordHash: passwordHash}
	result := db.Create(info)
	return result.Error
}

func (db *DataBase) GetPasswordHash(login string) ([]byte, error) {
	info := &userInfo{}
	result := db.First(&info, "login = ?", login)
	return info.PasswordHash, result.Error
}

func (db *DataBase) UpdateUserData(login string, data *UserData) error {
	result := db.Model(&userInfo{}).Where("login = ?", login).Updates(userInfo{
		Name:        data.Name,
		Surname:     data.Surname,
		BirthDay:    data.BirthDay,
		Mail:        data.Mail,
		PhoneNumber: data.PhoneNumber,
	})
	return result.Error
}

func (db *DataBase) GetUserData(login string) (*UserData, error) {
	info := &userInfo{Login: login}
	result := db.First(&info, "login = ?", login)
	return &UserData{
		Name:        info.Name,
		Surname:     info.Surname,
		BirthDay:    info.BirthDay,
		Mail:        info.Mail,
		PhoneNumber: info.PhoneNumber,
	}, result.Error
}
