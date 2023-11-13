package models

import(
	"carlosapi/pkg/database"
	"carlosapi/pkg/config"
	"gorm.io/gorm"
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
	AzStart		float32 `json:"az_start"`
	AzEnd		float32 `json:"az_end"`
	ElStart		float32 `json:"el_start"`
	ElEnd		float32 `json:"el_end"`
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
