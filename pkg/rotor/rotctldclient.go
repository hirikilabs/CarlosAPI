package rotor

import (
	"fmt"
	"net"
	"errors"
	"bytes"
	"strings"
	"strconv"
)

type RotCtl struct {
	Host string
	Port string
	Az float64
	El float64
	Conn *net.TCPConn
}

func NewRotCtl(host string, port string) *RotCtl {
	return &RotCtl {
		Host: host,
		Port: port,
		Az: 0.0,
		El: 0.0,
		Conn: nil,
	}
}

func (r *RotCtl) Connect() error {
	// get server
	tcpServer, err := net.ResolveTCPAddr("tcp", r.Host + ":" + r.Port)

	if err != nil {
		return err
	}

	// connect
	r.Conn, err = net.DialTCP("tcp", nil, tcpServer)
	if err != nil {
		return err
	}

	return nil
}

func (r *RotCtl) Disconnect() error {
	if r.Conn != nil {
		r.Conn.Close()
	} else {
		return errors.New("Not connected")
	}

	return nil
}

func (r *RotCtl) GetPos() (float64, float64, error) {
	
	_, err := r.Conn.Write([]byte("p\n"))
	if err != nil {
		return 0.0, 0.0, err
	}

	// create buffer for data
	rData := make([]byte, 1024)
	n, err := r.Conn.Read(rData)
	if err != nil {
		return 0.0, 0.0, err
	}

	// parse data
	az, el, err := ParsePos(rData[:n])
	if err != nil {
		return 0.0, 0.0, err
	}

	r.Az = az
	r.El = el
	return az, el, nil
}

func (r *RotCtl) SetPos(az float64, el float64) error {

	pos_str := fmt.Sprintf("P %.1f %.1f\n", az, el)
	
	_, err := r.Conn.Write([]byte(pos_str))
	if err != nil {
		return err
	}

	// create buffer for data
	rData := make([]byte, 1024)
	_, err = r.Conn.Read(rData)
	if err != nil {
		return err
	}

	// parse data
	if !strings.Contains(string(rData), "RPRT 0") {
		return errors.New("Problem setting rotor position")
	}
	
	return nil
}


func (r *RotCtl) InPos(az float64, el float64) bool {
	posaz, posel, err := r.GetPos()
	if err != nil {
		return false
	}
	
	if (posaz == az || posaz - 360 == az || posaz + 360 == az) && posel == el {
		return true
	} else {
		return false
	}
}


func ParsePos(data []byte) (float64, float64, error) {
	lines := bytes.Split(data, []byte("\n"))
	if len(lines) < 2 {
		return 0.0, 0.0, errors.New("Not enough lines")
	}

	az, err := strconv.ParseFloat(string(lines[0]), 64)
	el, err := strconv.ParseFloat(string(lines[1]), 64)
	if err != nil {
		return 0.0, 0.0, errors.New("Problem parsing numbers")
	}

	return az, el, nil
	
}

