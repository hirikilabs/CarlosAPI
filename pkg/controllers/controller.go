package controllers

import (
	"carlosapi/pkg/color"
	"carlosapi/pkg/config"
	"carlosapi/pkg/database"
	"carlosapi/pkg/models"
	"carlosapi/pkg/rotor"
	"carlosapi/pkg/sdrcarlos"
	"carlosapi/pkg/utils"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html/template"
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

// "/api" return configuration parameters
func ApiRoot(writer http.ResponseWriter, request *http.Request) {
	conf := config.GetConfig()

	res, _ := json.Marshal(conf)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
}

// "/api/status" returns all the recording stored
func ApiGetStatus(writer http.ResponseWriter, request *http.Request) {
	recordings := models.GetRecordings()

	res, _ := json.Marshal(recordings)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
}

// "/api/status/id" returns the status of a recording identified by it's ID
func ApiGetStatusId(writer http.ResponseWriter, request *http.Request) {
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

// "/api/clear" clears all the data
// TODO: don't expose this API on prodution
func ApiClearDatabase(writer http.ResponseWriter, request *http.Request) {
	models.ClearDB()
	res := []byte(`{"clear": "ok"}`)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
}

// inserts a new recording, returns the ID
func InsertRecording(rec models.Recording) *models.Recording {
	// ok, create recording
	rec.Id = time.Now().UnixMilli()
	rec.EstimateTime()
	rec.Status = models.Created
	recording := rec.CreateRecording()

	// send notification
	updateChannel <- models.Notification{}

	return recording
}

// creates a new recording
// TODO: validate fields
func ApiCreateRecording(writer http.ResponseWriter, request *http.Request) {
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
	// newRecording.Id = time.Now().UnixMilli()
	// newRecording.EstimateTime()
	// newRecording.Status = models.Created
	// recording := newRecording.CreateRecording()

	// // send notification
	// updateChannel <- models.Notification{}
	recording := InsertRecording(*newRecording)

	// ok
	log.Printf("📝"+color.Blue+" Added %v\n"+color.Reset, recording.Id)

	res, _ := json.Marshal(recording)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)
}

// Downloads file for Id
func ApiDownloadId(writer http.ResponseWriter, request *http.Request) {
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

	// ok, send file (add header for filename)
	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", varid+".tar.gz"))
	http.ServeFile(writer, request, conf.RecordPath+varid+".tar.gz")
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
		case <-updateChannel:
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
				log.Printf("⚡"+color.Yellow+" Launching %v\n"+color.Reset, rec.Id)
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

	// connect to rotor
	rot := rotor.NewRotCtl(conf.RotorHost, conf.RotorPort)
	err = rot.Connect()
	if err != nil {
		log.Printf("❌ Can't connect to rotor: %s\n", err.Error())
	}

	// args := fmt.Sprintf(conf.RecordCmd,
	// 	rec.SampleRate, rec.Frequency, rec.Gain, rec.RecTime,
	// 	rec.WaitTime, rec.Az, rec.El, rec.AzRange, rec.ElRange,
	// 	rec.AzStep, rec.ElStep, conf.RecordPath + strconv.FormatInt(rec.Id, 10) + ".iq")

	// log.Println("Args: ", args)

	// create output dir
	err = os.MkdirAll(conf.RecordPath+strconv.FormatInt(rec.Id, 10), 0755)
	if err != nil && !os.IsExist(err) {
		log.Println("❌ Error creating output directory")
		log.Println(err.Error())
	}

	// no errors and SDR detected?
	if err == nil && devices != nil {
		// move rotor to starting point
		log.Printf("📡 Moving rotor to start point...\n")

		err := rot.SetPos(rec.Az-rec.AzRange/2, rec.El-rec.ElRange/2)
		if err != nil {
			log.Printf("❌ Error moving rotor, %s\n", err.Error())
		}

		// wait
		for !rot.InPos(rec.Az-rec.AzRange/2, rec.El-rec.ElRange/2) {
			time.Sleep(1 * time.Second)
		}
		log.Println("📍 Rotor in place.")

		// ranges
		for az := rec.Az - rec.AzRange/2; az <= rec.Az+rec.AzRange/2; az += rec.AzStep {
			for el := rec.El - rec.ElRange/2; el <= rec.El+rec.ElRange/2; el += rec.ElStep {
				// move rotor
				err := rot.SetPos(az, el)
				if err != nil {
					log.Printf("❌ Error moving rotor, %s\n", err.Error())
				}

				// wait
				log.Printf("⏳ Waiting...")
				time.Sleep(time.Duration(rec.WaitTime) * time.Millisecond)

				log.Printf("🔴 Recording: (%3.1f, %3.1f)\n", az, el)

				// record
				carlosDev.ReadTime(fmt.Sprintf("%s/%d/%d-%3.1f-%3.1f.iq",
					conf.RecordPath, rec.Id, rec.Id, az, el), rec.RecTime)
			}
		}
		// close SDR
		carlosDev.Shutdown()

		// finished recording
		rec.Status = models.Recorded
		rec.Update()
		config.NoRecording()

		// create compressed archive
		log.Printf("🗜️  Creating compressed archive.\n")
		dirname := fmt.Sprintf("%s%d/", conf.RecordPath, rec.Id)
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

		// remove uncompressed data
		log.Printf("🗑️  Removing uncompressed data.\n")
		err = os.RemoveAll(dirname)
		if err != nil {
			log.Printf("❌ Error deleting uncompressed data: %v", err)
		}
	}

	// close rotor connection
	rot.Disconnect()

	// out, err := exec.Command(conf.RecordCmd, args).Output()
	// if err != nil {
	// 	log.Println("❌ Error running record command")
	//     log.Println(err.Error() + "\n\n" + string(out))
	// }
	log.Printf("✅ Finishing %v\n", rec.Id)
	// update status
	rec.Status = models.Finished
	rec.Update()
}

// Webs

func WebRoot(writer http.ResponseWriter, request *http.Request) {
	tmpl, err := template.ParseFiles("html/index.html")
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Problem loading web page"))
		return
	}

	recordings := models.GetRecordings()
	err = tmpl.Execute(writer, recordings)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Problem rendering web page"))
		return
	}

	return
}

// shows request web
func WebMakeRequest(writer http.ResponseWriter, request *http.Request) {
	tmpl, err := template.ParseFiles("html/request.html")
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Problem loading web page"))
		return
	}

	err = tmpl.Execute(writer, nil)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Problem rendering web page"))
		return
	}

	return
}

func WebCreateRecording(writer http.ResponseWriter, request *http.Request) {
	// prepare template
	tmpl, err := template.ParseFiles("html/request_answer.html")
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Problem loading web page"))
		writer.Write([]byte(err.Error()))
		return
	}

	errorParsing := false
	recAnswer := &models.RecordingAnswer{}

	// get and validate form values

	// check user
	formUser := request.PostFormValue("user")
	dbUser, _ := models.GetUserByName(formUser)
	if dbUser == nil {
		recAnswer.ErrorMessage = "Wrong Username"
		errorParsing = true
	}

	// check password
	formPassword := request.PostFormValue("password")
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(formPassword)))
	if (hash != dbUser.Password) && !errorParsing {
		recAnswer.ErrorMessage = "Wrong Password"
		errorParsing = true
	}

	// check date and time
	formDate := request.PostFormValue("date")
	formTime := request.PostFormValue("time")

	timeString := formDate + " " + formTime
	theTime, err := time.ParseInLocation("2006-01-02 15:04", timeString, time.Local)
	if err != nil && !errorParsing {
		recAnswer.ErrorMessage = "Can't parse time or date: " + err.Error()
		errorParsing = true
	}

	// check SDR config
	formFreq := request.PostFormValue("frequency")
	formSrate := request.PostFormValue("sample_rate")
	formGain := request.PostFormValue("gain")

	freq, err := strconv.Atoi(formFreq)
	if (err != nil || freq <= 0) && !errorParsing {
		recAnswer.ErrorMessage = "Wrong frequency"
		errorParsing = true
	}
	srate, err := strconv.Atoi(formSrate)
	if (err != nil || srate <= 0) && !errorParsing {
		recAnswer.ErrorMessage = "Wrong sample rate"
		errorParsing = true
	}

	gain, err := strconv.ParseFloat(formGain, 64)
	if (err != nil || gain <= 0) && !errorParsing {
		recAnswer.ErrorMessage = "Wrong gain"
		errorParsing = true
	}

	// check recording times
	formRecTime := request.PostFormValue("record_time")
	formWaitTime := request.PostFormValue("wait_time")

	rectime, err := strconv.Atoi(formRecTime)
	if (err != nil || rectime <= 0) && !errorParsing {
		recAnswer.ErrorMessage = "Wrong record time"
		errorParsing = true
	}
	waittime, err := strconv.Atoi(formWaitTime)
	if (err != nil || waittime <= 0) && !errorParsing {
		recAnswer.ErrorMessage = "Wrong wait time"
		errorParsing = true
	}

	// check coordinates
	formAz := request.PostFormValue("az")
	formAzRange := request.PostFormValue("az-range")
	formAzStep := request.PostFormValue("az-step")

	az, err := strconv.ParseFloat(formAz, 64)
	if (err != nil || az <= 0) && !errorParsing {
		recAnswer.ErrorMessage = "Wrong azimuth"
		errorParsing = true
	}
	azrange, err := strconv.ParseFloat(formAzRange, 64)
	if (err != nil || azrange <= 0) && !errorParsing {
		recAnswer.ErrorMessage = "Wrong azimuth range"
		errorParsing = true
	}

	azstep, err := strconv.ParseFloat(formAzStep, 64)
	if (err != nil || azstep <= 0) && !errorParsing {
		recAnswer.ErrorMessage = "Wrong azimuth step"
		errorParsing = true
	}

	formEl := request.PostFormValue("el")
	formElRange := request.PostFormValue("el-range")
	formElStep := request.PostFormValue("el-step")

	el, err := strconv.ParseFloat(formEl, 64)
	if (err != nil || el <= 0) && !errorParsing {
		recAnswer.ErrorMessage = "Wrong elevation"
		errorParsing = true
	}
	elrange, err := strconv.ParseFloat(formElRange, 64)
	if (err != nil || elrange <= 0) && !errorParsing {
		recAnswer.ErrorMessage = "Wrong elevation range"
		errorParsing = true
	}

	elstep, err := strconv.ParseFloat(formElStep, 64)
	if (err != nil || elstep <= 0) && !errorParsing {
		recAnswer.ErrorMessage = "Wrong elevation step"
		errorParsing = true
	}

	recAnswer.Rec = models.Recording{
		User:       formUser,
		Password:   formPassword,
		Time:       theTime.UnixMilli(),
		Frequency:  freq,
		SampleRate: srate,
		Gain:       int(gain * 10),
		RecTime:    int64(rectime),
		WaitTime:   int64(waittime),
		Az:         az,
		AzRange:    azrange,
		AzStep:     azstep,
		El:         el,
		ElRange:    elrange,
		ElStep:     elstep,
	}

	if !errorParsing {
		// ok, insert it
		recording := InsertRecording(recAnswer.Rec)
		// ok
		log.Printf("📝"+color.Blue+" Added %v\n"+color.Reset, recording.Id)
		recAnswer.Rec.Id = recording.Id
		recAnswer.StringTime = time.UnixMilli(recAnswer.Rec.Time).Local().String()
	}

	err = tmpl.Execute(writer, recAnswer)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Problem rendering web page"))
		return
	}

	return
}
