package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DataBase struct {
	*gorm.DB
}

type likeStat struct {
	gorm.Model
	UserLogin string
	TaskID    uint
}

type viewStat struct {
	gorm.Model
	UserLogin string
	TaskID    uint
}

type Statistic struct {
	Login  string
	TaskID uint
}

func New() *DataBase {
	dsn := "host=statistics_db dbname=statistics_db sslmode=disable user=user password=password"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect tasks database: " + err.Error())
	}

	db.AutoMigrate(&likeStat{})
	db.AutoMigrate(&viewStat{})

	return &DataBase{db}
}

func (db *DataBase) AddLike(stat Statistic) error {
	info := &likeStat{UserLogin: stat.Login, TaskID: stat.TaskID}

	result := db.Create(info)

	var statInfo likeStat
	db.First(&stat)
	fmt.Printf("Like added: %#v\n", statInfo)

	return result.Error
}

func (db *DataBase) AddView(stat Statistic) error {
	info := &viewStat{UserLogin: stat.Login, TaskID: stat.TaskID}
	result := db.Create(info)
	return result.Error
}
