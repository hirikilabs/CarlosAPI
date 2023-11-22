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
	Coords      string  `json:"coords"`
	AzRange		int     `json:"az_range"`
	AzStep		int     `json:"az_step"`
	ElRange		int     `json:"el_range"`
	ElStep		int     `json:"el_step"`
	Path		string	`json:"path"`
	Status		RecordStatus `json:"status"`
}


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

// check recording fields
func (r* Recording) Check() error {
	// check time
	if r.Time <= time.Now().UnixMilli() {
		return fmt.Errorf("Time is in the past")
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
	db := db.Where("id=?", Id).Find(&getRecording)
	return &getRecording, db
}
