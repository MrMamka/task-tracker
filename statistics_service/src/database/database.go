package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DataBase struct {
	*gorm.DB
}

type likeStat struct {
	gorm.Model
	login string
	id    uint
}

type viewStat struct {
	gorm.Model
	login string
	id    uint
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
	info := &likeStat{login: stat.Login, id: stat.TaskID}
	result := db.Create(info)
	return result.Error
}

func (db *DataBase) AddView(stat Statistic) error {
	info := &viewStat{login: stat.Login, id: stat.TaskID}
	result := db.Create(info)
	return result.Error
}
