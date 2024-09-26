package models

import (
	"carlosapi/pkg/config"
	"carlosapi/pkg/database"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// strings that represent the status of the recordings
const (
	Created  = "Created"
	Running  = "Running"
	Recorded = "Recorded"
	Finished = "Finished"
)

var db *gorm.DB

type RecordStatus string

type User struct {
	Username string
	Password string
}

type Recording struct {
	gorm.Model
	Id         int64        `json:"id"`
	User       string       `json:"user"`
	Password   string       `json:"password"`
	Time       int64        `json:"time"`      // unix timestamp
	Frequency  int          `json:"frequency"` // Hz
	SampleRate int          `json:"sample_rate"`
	Gain       int          `json:"gain"`      // integer, real gain/10
	RecTime    int64        `json:"rec_time"`  // ms
	WaitTime   int64        `json:"wait_time"` // ms
	Az         float64      `json:"az"`
	El         float64      `json:"el"`
	AzRange    float64      `json:"az_range"`
	AzStep     float64      `json:"az_step"`
	ElRange    float64      `json:"el_range"`
	ElStep     float64      `json:"el_step"`
	CalcTime   int64        `json:"calc_time"` // ms
	Status     RecordStatus `json:"status"`
}

type Notification struct{}

func init() {
	// connect to the database and create the tables if needed
	conf := config.GetConfig()
	database.ConnectDB(conf.Database)
	db = database.GetDB()
	db.AutoMigrate(&Recording{})
	db.AutoMigrate(&User{})
}

// add a recording to the database
func (r *Recording) CreateRecording() *Recording {
	db.Create(&r)
	return r
}

// update a recording
func (r *Recording) Update() *Recording {
	db.Save(&r)
	return r
}

// calculate estimated time for the recording
func (r *Recording) EstimateTime() {
	// (record time + wait time) * number of points
	r.CalcTime = (int64(r.RecTime) + int64(r.WaitTime)) * int64((r.AzRange/r.AzStep)*(r.ElRange/r.ElStep))
}

// check recording fields
func (r *Recording) Check() error {
	// check time
	if r.Time <= time.Now().UnixMilli() {
		return fmt.Errorf("Time is in the past")
	}
	if r.RecTime < 1 {
		return fmt.Errorf("Record time too short")
	}
	if r.WaitTime < 0 {
		return fmt.Errorf("Wait time can't be negative")
	}
	if r.AzRange < 0 || r.AzStep < 0 || r.ElStep < 0 || r.ElRange < 0 {
		return fmt.Errorf("Movement ranges and steps can't be negative")
	}

	return nil
}

// clear all the data in the database
// TODO: don't expose this API or remove
func ClearDB() {
	db.Where("1 = 1").Delete(&Recording{})
}

// Get all recordings
func GetRecordings() []Recording {
	var Recordings []Recording
	db.Find(&Recordings)
	return Recordings
}

// Get a recording by it's ID
func GetRecordingById(Id int64) (*Recording, *gorm.DB) {
	var getRecording Recording
	result := db.Where("id=?", Id).First(&getRecording)
	return &getRecording, result
}

// Users
func GetUsers() []User {
	var Users []User
	db.Find(&Users)
	return Users
}

func GetUserByName(name string) (*User, *gorm.DB) {
	var getUser User
	result := db.Where("user=?", name).First(&getUser)
	return &getUser, result
}
