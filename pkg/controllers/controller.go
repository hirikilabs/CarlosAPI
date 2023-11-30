package controllers

import (
	"carlosapi/pkg/color"
	"carlosapi/pkg/config"
	"carlosapi/pkg/database"
	"carlosapi/pkg/models"
	"carlosapi/pkg/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// "/" return configuration parameters
func Root(writer http.ResponseWriter, request *http.Request) {
	conf := config.GetConfig()

	res, _ := json.Marshal(conf)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
}

// "/status" returns all the recording stored
func GetStatus(writer http.ResponseWriter, request *http.Request) {
	recordings := models.GetRecordings()

	res, _ := json.Marshal(recordings)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
}

// "/status/id" returns the status of a recording identified by it's ID
func GetStatusId(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	varid := vars["id"]
	id, err := strconv.ParseInt(varid, 0, 0)
	if err != nil {
		log.Printf("❌ ID Parse Error %v\n", err.Error())
	}
	recording, _ := models.GetRecordingById(id)

	res, _ := json.Marshal(recording)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
}

// "/clear" clears all the data
// TODO: don't expose this API on prodution
func ClearDatabase(writer http.ResponseWriter, request *http.Request) {
	models.ClearDB()
	res := []byte("{'clear'='ok'}")
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
}

// creates a new recording
// TODO: validate fields
func CreateRecording(writer http.ResponseWriter, request *http.Request) {
	// parse JSON
	newRecording := &models.Recording{}
	err := utils.ParseBody(request, newRecording)
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		res := fmt.Sprintf("'error' = '%v'}", err.Error())
		writer.Write([]byte(res))
		return
	}
	// check fields
	err = newRecording.Check()
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		res := fmt.Sprintf("'error' = '%v'}", err.Error())
		writer.Write([]byte(res))
		return		
	}
	// ok, create recording
	newRecording.Id = time.Now().UnixMilli()
	newRecording.Status = models.Created
	recording := newRecording.CreateRecording()
	log.Printf("📝" + color.Blue + " Added %v\n" + color.Reset, recording.Id)
	
	res, _ := json.Marshal(recording)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
}

// Scheduler, checks for due recordings and launches them
// launched on another thread
func RunScheduling() {
	log.Println("🗓️" + color.Red + " Starting Scheduler" + color.Reset)

	db := database.GetDB()

	for {
		var newRecordings []models.Recording
		db.Where("status=?", models.Created).Find(&newRecordings)
		for _, rec := range newRecordings {
			if rec.Time < time.Now().UnixMilli() && !config.IsRecording(){
				log.Printf("⚡" + color.Yellow + " Launching %v\n" + color.Reset, rec.Id)
				rec.Status = models.Running
				rec.Update()
				config.Recording()
				go RunProcess(rec)
				break
			}
		}
		time.Sleep(1 * time.Minute)
	}
}


// runs the record software
// launched on another thread
// TODO: do it for real
func RunProcess(rec models.Recording) {
	//time.Sleep(10 * time.Second)
	conf := config.GetConfig()
	
	cmd := fmt.Sprintf(conf.RecordCmd,
		rec.SampleRate, rec.Frequency, rec.Gain, rec.RecTime,
		rec.WaitTime, rec.Coords, rec.AzRange, rec.ElRange,
		rec.AzStep, rec.AzRange, conf.RecordPath + strconv.FormatInt(rec.Id, 10))
	_, err := exec.Command(cmd).Output()
    if err != nil {
        log.Fatal(err)
    }
	log.Printf("✅ Finishing %v\n", rec.Id)
	// update recording status
	rec.Status = models.Finished
	rec.Update()
	// not recording anymore
	config.NoRecording()
}
