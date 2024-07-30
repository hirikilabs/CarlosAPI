package models

import(
	"carlosapi/pkg/database"
	"carlosapi/pkg/config"
	"fmt"
	"gorm.io/gorm"
	"time"
)

// strings that represent the status of the recordings
const(
	Created = "Created"
	Running = "Running"
	Finished = "Finished"
)

var db *gorm.DB

type RecordStatus string

type Recording struct {
	gorm.Model
	Id			int64	`json:"id"`
	User		string	`json:"user"`
	Time		int64	`json:"time"`
	Frequency	int		`json:"frequency"`
	SampleRate	int		`json:"sample_rate"`
	Gain		float32	`json:"gain"`
	RecTime		int		`json:"rec_time"`
	WaitTime	int		`json:"wait_time"`
	Az          float32 `json:"az"`
	El          float32 `json:"el"`
	AzRange		float32 `json:"az_range"`
	AzStep		float32 `json:"az_step"`
	ElRange		float32 `json:"el_range"`
	ElStep		float32 `json:"el_step"`
	CalcTime    int64   `json:"calc_time"`
	Status		RecordStatus `json:"status"`
}

type Notification struct {}

func init() {
	// connect to the database and create the tables if needed
	conf := config.GetConfig()
	database.ConnectDB(conf.Database)
	db = database.GetDB()
	db.AutoMigrate(&Recording{})
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
func (r* Recording) EstimateTime() {
	// (record time * wait ) * number of points * 1000 (milliseconds)
	r.CalcTime = (int64(r.RecTime) * int64(r.WaitTime)) * int64((r.AzRange / r.AzStep) * (r.ElRange / r.ElStep)) * 1000
}

// check recording fields
func (r* Recording) Check() error {
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
