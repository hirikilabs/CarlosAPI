package controllers

import (
	"carlosapi/pkg/color"
	"carlosapi/pkg/config"
	"carlosapi/pkg/database"
	"carlosapi/pkg/models"
	"carlosapi/pkg/utils"
	"carlosapi/pkg/sdrcarlos"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var updateChannel chan models.Notification

func init() {
	updateChannel = make(chan models.Notification)
}

func GetChannel() chan models.Notification {
	return updateChannel
}

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
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(`{"error": "Problem parsing ID"}`))
		return
	}
	recording, result := models.GetRecordingById(id)
	if result.Error != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte(`{"error": "No recording with that ID"}`))
		return
	}
	
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
		res := fmt.Sprintf("{'error' = '%v'}", err.Error())
		writer.Write([]byte(res))
		return
	}
	// check fields
	err = newRecording.Check()
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		res := fmt.Sprintf("{'error' = '%v'}", err.Error())
		writer.Write([]byte(res))
		return		
	}
	// ok, create recording
	newRecording.Id = time.Now().UnixMilli()
	newRecording.Status = models.Created
	recording := newRecording.CreateRecording()

	// send notification
	updateChannel <- models.Notification{}

	// ok
	log.Printf("📝" + color.Blue + " Added %v\n" + color.Reset, recording.Id)
	
	res, _ := json.Marshal(recording)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
}


// Downloads file for Id
func DownloadId(writer http.ResponseWriter, request *http.Request) {
	// get config for paths
	conf := config.GetConfig()

	// check Id
	vars := mux.Vars(request)
	varid := vars["id"]
	id, err := strconv.ParseInt(varid, 0, 0)
	if err != nil {
		log.Printf("❌ ID Parse Error %v\n", err.Error())
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(`{"error": "Error parsing ID"}`))
		return
	}

	// get from database to check
	recording, result := models.GetRecordingById(id)

	if result.Error != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte(`{"error": "No recording with requested ID"}`))
		return
	}
	if recording.Status != models.Finished {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusLocked)
		writer.Write([]byte(`{"error": "Recording not done yet"}`))
		return
	}

	// ok, send file
	http.ServeFile(writer, request, conf.RecordPath + varid + ".tar.gz")
}

// Scheduler, checks for due recordings and launches them
// launched on another thread
func RunScheduling() {
	log.Println("⏰" + color.Yellow + " Starting Scheduler" + color.Reset)

	// need to update?
	updateDatabase := true

	// get db
	db := database.GetDB()

	// recordings array
	var newRecordings []models.Recording
	
	for {
		// check channel
		select {
		case <- updateChannel:
			// get from database again
			updateDatabase = true
		default:
			// nothing
		}

		// need to update?
		if updateDatabase {
			// get requests not done yet
			db.Where("status=?", models.Created).Find(&newRecordings)
			updateDatabase = false
		}
		
		for _, rec := range newRecordings {
			if rec.Time < time.Now().UnixMilli() && !config.IsRecording() {
				log.Printf("⚡" + color.Yellow + " Launching %v\n" + color.Reset, rec.Id)
				rec.Status = models.Running
				rec.Update()
				config.Recording()
				go RunProcess(rec)
				// we need to update to see the change of a finished recording
				updateDatabase = true
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
}


// runs the recording
// launched on another thread
// TODO: do it for real
func RunProcess(rec models.Recording) {
	conf := config.GetConfig()

	// get and configure SDR
	carlosDev := &sdrcarlos.SDRCARLOS{Debug: false}

	devices := carlosDev.GetDevices()
	if devices == nil {
		log.Println("❌ Can't find any RTLSDR devices")
	}

	// use first device
	indexID := 0
	err := carlosDev.Config(indexID, rec.SampleRate, rec.Frequency, 0, rec.Gain, true)
	if err != nil {
		log.Printf("❌ SDR configure failed: %s\n", err.Error())
	}
	
	// args := fmt.Sprintf(conf.RecordCmd,
	// 	rec.SampleRate, rec.Frequency, rec.Gain, rec.RecTime,
	// 	rec.WaitTime, rec.Az, rec.El, rec.AzRange, rec.ElRange,
	// 	rec.AzStep, rec.ElStep, conf.RecordPath + strconv.FormatInt(rec.Id, 10) + ".iq")

	// log.Println("Args: ", args)
	
	// create output dir
	err = os.MkdirAll(conf.RecordPath + strconv.FormatInt(rec.Id, 10), 0755)
	if err != nil && !os.IsExist(err) {
		log.Println("❌ Error creating output directory")
		log.Println(err.Error())
	}

	// no errors and SDR detected?
	if err == nil && devices != nil {
		// move rotor to starting point
		// TODO
		
		// ranges
		for az := rec.Az - rec.AzRange/2; az <= rec.Az + rec.AzRange/2; az += rec.AzStep {
			for el := rec.El - rec.ElRange/2; el <= rec.El + rec.ElRange/2; el += rec.ElStep {
				// TODO move rotor
				
				log.Printf("🔴 Recording: (%3.1f, %3.1f)\n", az, el)

				// record
				carlosDev.ReadTime(fmt.Sprintf("%s/%d/%d-%3.1f-%3.1f.iq",
					conf.RecordPath, rec.Id, rec.Id, az, el), rec.RecTime)
			}
		}
		
		// create compressed archive
		log.Printf("🗜️ Creating compressed archive.\n")
		dirname := fmt.Sprintf("%s/%d/", conf.RecordPath, rec.Id)
		directory, err := os.Open(dirname)
		if err != nil {
			log.Println("❌ Error getting output directory")
		}
		defer directory.Close()
		files, err := directory.Readdirnames(0)
		if err != nil {
			log.Println("❌ Error listing data files")
		}
		err = utils.CreateArchive(fmt.Sprintf("%s/%d.tar.gz", conf.RecordPath, rec.Id), dirname, files)
		if err != nil {
			log.Printf("❌ Error creating compressed archive: %v", err)
		}

	}
	
	// out, err := exec.Command(conf.RecordCmd, args).Output()
    // if err != nil {
	// 	log.Println("❌ Error running record command")
    //     log.Println(err.Error() + "\n\n" + string(out))
    // }
	log.Printf("✅ Finishing %v\n", rec.Id)
	// update recording status
	rec.Status = models.Finished
	rec.Update()
	// not recording anymore
	config.NoRecording()
}
