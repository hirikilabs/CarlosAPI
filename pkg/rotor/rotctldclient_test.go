package rotor

import (
	"testing"
)


func TestConnect(t *testing.T) {
	rot := NewRotCtl("172.16.30.11", "4533")
	err := rot.Connect()
	if err != nil {
		t.Errorf("Error connecting %v", err.Error())
	}

	err = rot.Disconnect()
	if err != nil {
		t.Errorf("Error disconnecting %v", err.Error())
	}
}

func TestGetPos(t *testing.T) {
	rot := NewRotCtl("172.16.30.11", "4533")
	err := rot.Connect()
	if err != nil {
		t.Errorf("Error connecting to rotor %v", err.Error())
	}

	az, el, err := rot.GetPos()
	if err != nil {
		t.Errorf("Error getting rotor position %v", err.Error())
	}

	// limits
	if az < -180.0 || az > 540.0 || el < -20.0 || el > 210.0 {
		t.Errorf("Rotor position out of limits")
	}
	t.Logf("Position: %f, %f", az, el)

	err = rot.Disconnect()
	if err != nil {
		t.Errorf("Error disconnecting from rotor %v", err.Error())
	}
}

func TestSetPos(t *testing.T) {
	rot := NewRotCtl("172.16.30.11", "4533")
	err := rot.Connect()
	if err != nil {
		t.Errorf("Error connecting to rotor %v", err.Error())
	}


	err = rot.SetPos(10.0, 10.0)
	if err != nil {
		t.Errorf("Error setting rotor position %v", err.Error())
	}
	
	az, el, err := rot.GetPos()
	if err != nil {
		t.Errorf("Error getting rotor position %v", err.Error())
	}

	// limits
	if az < -180.0 || az > 540.0 || el < -20.0 || el > 210.0 {
		t.Errorf("Rotor position out of limits")
	}
	t.Logf("Position: %f, %f", az, el)

	err = rot.Disconnect()
	if err != nil {
		t.Errorf("Error disconnecting from rotor %v", err.Error())
	}
}

