package controllers

import (
	"carlosapi/pkg/config"
	"carlosapi/pkg/database"
	"carlosapi/pkg/models"
	"carlosapi/pkg/utils"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/gorilla/mux"
)

func Root(writer http.ResponseWriter, request *http.Request) {
	conf := config.GetConfig()

	res, _ := json.Marshal(conf)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
}

func GetStatus(writer http.ResponseWriter, request *http.Request) {
	recordings := models.GetRecordings()

	res, _ := json.Marshal(recordings)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
}

func GetStatusId(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	varid := vars["id"]
	id, err := strconv.ParseInt(varid, 0, 0)
	if err != nil {
		log.Printf("ID Parse Error %v\n", err.Error())
	}
	recording, _ := models.GetRecordingById(id)

	res, _ := json.Marshal(recording)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
}

func ClearDatabase(writer http.ResponseWriter, request *http.Request) {
	models.ClearDB()
	res := []byte("{'clear'='ok'}")
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
}

func CreateRecording(writer http.ResponseWriter, request *http.Request) {
	newRecording := &models.Recording{}
	utils.ParseBody(request, newRecording)
	newRecording.Id = time.Now().UnixMilli()
	newRecording.Status = models.Created
	recording := newRecording.CreateRecording()

	res, _ := json.Marshal(recording)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
}

// launched on another thread
func RunScheduling() {
	log.Println("Starting Scheduler")

	db := database.GetDB()

	for {
		var newRecordings []models.Recording
		db.Where("status=?", models.Created).Find(&newRecordings)
		for _, rec := range newRecordings {
			if rec.Time < time.Now().UnixMilli() && !config.IsRecording(){
				log.Printf("Launching %v\n", rec.Id)
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

func RunProcess(rec models.Recording) {
	time.Sleep(10 * time.Second)
	log.Printf("Finishing %v\n", rec.Id)
	rec.Status = models.Finished
	rec.Update()
	config.NoRecording()
}
