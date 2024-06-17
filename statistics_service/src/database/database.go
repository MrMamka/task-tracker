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

func (db *DataBase) EnsureLike(stat Statistic) error {
	var info likeStat
	result := db.First(&info, "user_login = ? AND task_id = ?", stat.Login, stat.TaskID)
	if result.Error == nil {
		return nil
	} else if result.Error != gorm.ErrRecordNotFound {
		return result.Error
	}

	info = likeStat{UserLogin: stat.Login, TaskID: stat.TaskID}
	result = db.Create(&info)
	return result.Error
}

func (db *DataBase) EnsureView(stat Statistic) error {
	var info viewStat
	result := db.First(&info, "user_login = ? AND task_id = ?", stat.Login, stat.TaskID)
	if result.Error == nil {
		return nil
	} else if result.Error != gorm.ErrRecordNotFound {
		return result.Error
	}

	info = viewStat{UserLogin: stat.Login, TaskID: stat.TaskID}
	result = db.Create(&info)
	return result.Error
}

func (db *DataBase) CountLikes(taskID uint) (count int64, err error) {
	result := db.Model(&likeStat{}).Where("task_id = ?", taskID).Count(&count)
	err = result.Error
	return
}

func (db *DataBase) CountViews(taskID uint) (count int64, err error) {
	result := db.Model(&viewStat{}).Where("task_id = ?", taskID).Count(&count)
	err = result.Error
	return
}

type TaskIDCount struct {
	TaskID uint
	Count  int64
}

func (db *DataBase) TopByLikes(n int) ([]TaskIDCount, error) {
	var tasks []TaskIDCount
	result := db.Model(&likeStat{}).
		Group("task_id").
		Select("task_id, COUNT(*) AS count").
		Order("count DESC").
		Limit(n).
		Scan(&tasks)
	return tasks, result.Error
}

func (db *DataBase) TopByViews(n int) ([]TaskIDCount, error) {
	var tasks []TaskIDCount
	result := db.Model(&viewStat{}).
		Group("task_id").
		Select("task_id, COUNT(*) AS count").
		Order("count DESC").
		Limit(n).
		Scan(&tasks)
	return tasks, result.Error
}

func (db *DataBase) GroupedLikes() ([]TaskIDCount, error) {
	var tasks []TaskIDCount
	result := db.Model(&likeStat{}).
		Group("task_id").
		Select("task_id, COUNT(*) AS count").
		Scan(&tasks)
	return tasks, result.Error
}
