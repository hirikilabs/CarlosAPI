package sdrcarlos

import (
	//"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	
	rtl "github.com/jpoirier/gortlsdr"
)

type RTLDevice struct {
	Vendor string
	Product string
	Serial string
}

// CARLOS holds a device context.
type SDRCARLOS struct {
	Dev *rtl.Context
	Wg  *sync.WaitGroup
	Debug bool
}

// gets connected devices
func (u *SDRCARLOS) GetDevices() []RTLDevice {
	var devices []RTLDevice

	// if no devices return a nil slice
	if c := rtl.GetDeviceCount(); c == 0 {
		return devices
	} else {
		for i := 0; i < c; i++ {
			m, p, s, _ := rtl.GetDeviceUsbStrings(i)
			dev := RTLDevice {
				Vendor: m,
				Product: p,
				Serial: s,
			}
			devices = append(devices, dev)		
		}
	}
	
	return devices
}

// read does synchronous specific reads.
func (u *SDRCARLOS) Read(filename string) {
	defer u.Wg.Done()

	if u.Debug {
		log.Println("Entered SDRCARLOS read() ...")
	}
	
	// create file
    f, err := os.Create(filename)
    if err != nil {
		log.Fatal(err)
    }
	defer f.Close()
	
	var readCnt uint64
	var buffer = make([]uint8, rtl.DefaultBufLength)
	for {
		nRead, err := u.Dev.ReadSync(buffer, rtl.DefaultBufLength)
		if err != nil {
			break
		}
		if nRead > 0 {
			if u.Debug {
				fmt.Printf("\rnRead %d: readCnt: %d", nRead, readCnt)
			}
			readCnt++
			_, err = f.Write(buffer)
			if err != nil {
				log.Fatal(err)
			}

		}
	}
}

// ReadTime does syncronous read for a period of time
func (u *SDRCARLOS) ReadTime(filename string, milliseconds int64) {
	if u.Debug {
		log.Println("Entered SDRCARLOS ReadTime() ...")
	}

	// create file
    f, err := os.Create(filename)
    if err != nil {
		log.Fatal(err)
    }
	defer f.Close()
	
	var readCnt uint64
	//var buffer = make([]uint8, rtl.DefaultBufLength)
	var buffer = make([]uint8, 1024)
	
	
	// get current time
	start := time.Now()
	
	for {
		nRead, err := u.Dev.ReadSync(buffer, 1024)
		if err != nil {
			// log.Printf("\tReadSync Failed - error: %s\n", err)
			break
		}
		// log.Printf("\tReadSync %d\n", nRead)
		if nRead > 0 {
			// buf := buffer[:nRead]
			if u.Debug {
				fmt.Printf("\rnRead %d: readCnt: %d", nRead, readCnt)
			}
			readCnt++
			_, err = f.Write(buffer)
			if err != nil {
				log.Fatal(err)
			}


		}
		// check time
		t := time.Now()
		elapsed := t.Sub(start)
		if elapsed.Milliseconds() >= milliseconds {
			break
		}
	}
	if u.Debug {
		log.Println("End ReadTime() ...")
	}
	//u.Wg.Done()
}

// shutdown
func (u *SDRCARLOS) Shutdown() {
	if u.Debug {
		fmt.Println()
		log.Println("\nEntered SDRCARLOS shutdown() ...")
		log.Println("SDRCARLOS shutdown(): closing Device ...")
	}
	u.Dev.Close() // preempt the blocking ReadSync call
	if u.Debug { 
		log.Println("SDRCARLOS shutdown(): calling .Wait() ...")
	}
	u.Wg.Wait() // Wait for the goroutine to shutdown
	if u.Debug {
		log.Println("SDRCARLOS shutdown(): .Wait() returned...")
	}
}

// sdrConfig configures the Device.
func (u *SDRCARLOS) Config(indexID int, samplerate int, freq int, bw int, gain int, bias bool) (err error) {
	if u.Dev, err = rtl.Open(indexID); err != nil {
		if u.Debug {
			log.Printf("\tSDRCARLOS Open Failed...\n")
		}
		return
	}
	if u.Debug {
		log.Printf("\tGetTunerType: %s\n", u.Dev.GetTunerType())
	}

	//---------- Set Tuner Gain ----------
	err = u.Dev.SetTunerGainMode(true)
	if err != nil {
		u.Dev.Close()
		if u.Debug {
			log.Printf("\tSetTunerGainMode Failed - error: %s\n", err)
		}
		return
	}
	if u.Debug {
		log.Printf("\tSetTunerGainMode Successful\n")
	}

	gains, err := u.Dev.GetTunerGains()
	if err != nil {
		if u.Debug {
			log.Printf("\tGetTunerGains Failed - error: %s\n", err)
		}
	} else if len(gains) > 0 {
		for _, g := range gains {
			if u.Debug {
				log.Printf("\tPossible gain value: %d\n", g)
			}
		}
	}

	err = u.Dev.SetTunerGain(gain)
	if err != nil {
		u.Dev.Close()
		if u.Debug {
			log.Printf("\tSetTunerGain Failed - error: %s\n", err)
		}
		return
	}
	if u.Debug {
		log.Printf("\tSetTunerGain Successful\n")
	}

	//---------- Get/Set Sample Rate ----------
	//samplerate := 2083334
	err = u.Dev.SetSampleRate(samplerate)
	if err != nil {
		u.Dev.Close()
		if u.Debug {
			log.Printf("\tSetSampleRate Failed - error: %s\n", err)
		}
		return
	}
	if u.Debug {
		log.Printf("\tSetSampleRate - rate: %d\n", samplerate)
		log.Printf("\tGetSampleRate: %d\n", u.Dev.GetSampleRate())
	}

	//---------- Get/Set Center Freq ----------
	err = u.Dev.SetCenterFreq(freq)
	if err != nil {
		u.Dev.Close()
		if u.Debug {
			log.Printf("\tSetCenterFreq Failed, error: %s\n", err)
		}
		return
	}
	if u.Debug {
		log.Printf("\tSetCenterFreq Successful\n")
		log.Printf("\tGetCenterFreq: %d\n", u.Dev.GetCenterFreq())
	}

	//---------- Set Bandwidth ----------
	if u.Debug {
		log.Printf("\tSetting Bandwidth: %d\n", bw)
	}
	if err = u.Dev.SetTunerBw(bw); err != nil {
		u.Dev.Close()
		if u.Debug {
			log.Printf("\tSetTunerBw %d Failed, error: %s\n", bw, err)
		}
		return
	}
	if u.Debug {
		log.Printf("\tSetTunerBw %d Successful\n", bw)
	}

	if err = u.Dev.ResetBuffer(); err != nil {
		u.Dev.Close()
		if u.Debug {
			log.Printf("\tResetBuffer Failed - error: %s\n", err)
		}
		return
	}
	if u.Debug {
		log.Printf("\tResetBuffer Successful\n")
	}

	//---------- Get/Set Freq Correction ----------
	freqCorr := u.Dev.GetFreqCorrection()
	if u.Debug {
		log.Printf("\tGetFreqCorrection: %d\n", freqCorr)
	}
	err = u.Dev.SetFreqCorrection(freqCorr)
	if err != nil {
		u.Dev.Close()
		if u.Debug {
			log.Printf("\tSetFreqCorrection %d Failed, error: %s\n", freqCorr, err)
		}
		return
	}
	if u.Debug {
		log.Printf("\tSetFreqCorrection %d Successful\n", freqCorr)
	}

	// Bias-T
	err = u.Dev.SetBiasTee(true)
	if err != nil {
		u.Dev.Close()
		if u.Debug {
			log.Printf("SetBiasTee %v Failed, error %s\n", bias, err)
		}
	}

	return
}

// sigAbort
func (u *SDRCARLOS) SigAbort() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)
	<-ch
	u.Shutdown()
	os.Exit(0)
}
