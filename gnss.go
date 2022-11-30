package gnss

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

var fix []string = []string{"unkown", "no fix", "2D", "3D"}
var satSystem []string = []string{"GPS", "unknown", "Galileo", "Beidou", "unknown", "QZSS", "GLONASS"}

func NewGPSD(server string, startLongitude, startLatitude float64) *GPSD {
	gpsd := GPSD{
		server: server,
	}
	return &gpsd
}

func (gpsd *GPSD) Start() error {
	var err error

	// connect to server
	gpsd.gpsd, err = net.Dial("tcp", gpsd.server)
	if err != nil {
		return err
	}
	fmt.Fprintf(gpsd.gpsd, "?WATCH={\"enable\":true,\"json\":true}")
	go gpsd.loop()

	return nil
}

func (gpsd *GPSD) loop() {
	reader := bufio.NewReader(gpsd.gpsd)
	for {
		buffer, _ := reader.ReadBytes('\n')
		msg := &basemsg{}
		json.Unmarshal(buffer, msg)

		switch msg.Class {
		case "TPV":
			gpsd.tpvFilter(buffer)
		case "SKY":
			gpsd.skyfilter(buffer)
		case "PPS":
			gpsd.ppsFilter(buffer)
		default:
			if gpsd.debug {
				fmt.Println(string(buffer))
			}
		}
	}
}

func (gpsd *GPSD) SetDebug(debug bool) {
	gpsd.debug = debug
}

func (gpsd *GPSD) tpvFilter(data []byte) {
	tpv := TPVReport{}
	err := json.Unmarshal(data, &tpv)
	if err != nil {
		return
	}
	gpsd.Position.Fix = fix[tpv.Mode]
	gpsd.Position.Altitude = tpv.Alt
	gpsd.TimeData.Timestamp = tpv.Time
	gpsd.Position.Latitude = tpv.Lat
	gpsd.Position.Longitude = tpv.Lon
	gpsd.Position.MagneticVariance = tpv.Magvar
	gpsd.TimeData.Leapseconds = tpv.Leapseconds
	//gpsd.TPVReport = tpv
}

func (gpsd *GPSD) ppsFilter(data []byte) {
	pps := PPSReport{}
	err := json.Unmarshal(data, &pps)
	if err != nil {
		return
	}
	gpsd.TimeData.Clock_Nsec = pps.Clock_Nsec
	gpsd.TimeData.Clock_Sec = pps.Clock_Sec
	gpsd.TimeData.Precision = pps.Precision
	gpsd.TimeData.Real_Nsec = pps.Real_Nsec
	gpsd.TimeData.Real_Sec = pps.Real_Sec
	gpsd.TimeData.Qerr = pps.Qerr
	gpsd.TimeData.Shm = pps.Shm
}

func (gpsd *GPSD) skyfilter(data []byte) {
	sky := SKYReport{}
	err := json.Unmarshal(data, &sky)
	if err != nil {
		return
	}

	gpsd.DilutionOfPrecision.Xdop = sky.Xdop
	gpsd.DilutionOfPrecision.Ydop = sky.Ydop
	gpsd.DilutionOfPrecision.Vdop = sky.Vdop
	gpsd.DilutionOfPrecision.Tdop = sky.Tdop
	gpsd.DilutionOfPrecision.Hdop = sky.Hdop
	gpsd.DilutionOfPrecision.Pdop = sky.Pdop
	gpsd.DilutionOfPrecision.Gdop = sky.Gdop

	sats := sky.Satellites
	gpsd.SatsVisible = make([]SatInfo, 0)

	for _, sat := range sats {
		newSat := SatInfo{
			PRNNumber:        sat.PRN,
			Elevation:        sat.El,
			Azimuth:          sat.Az,
			SignalNoiseRatio: sat.Ss,
			Used:             sat.Used,
			Type:             satSystem[sat.Gnssid],
			Svid:             sat.Svid,
			Sigid:            sat.Sigid,
			Freqid:           sat.Freqid,
			Health:           sat.Health,
		}
		gpsd.SatsVisible = append(gpsd.SatsVisible, newSat)
	}
}
