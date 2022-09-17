package gnss

import (
	"fmt"
	"time"

	deamon "github.com/stratoberry/go-gpsd"
)

var fix []string = []string{"unkown", "no fix", "2D", "3D"}

type GPSD struct {
	Alt         float64   `json:"alt"`
	SatsVisible []GSVInfo `json:"satsGpsVisible"`
	Fix         string    `json:"fix"`
	Hdop        float64   `json:"hdop"`
	Pdop        float64   `json:"pdop"`
	Vdop        float64   `json:"vdop"`
	Longitude   float64   `json:"longitude"`
	Latitude    float64   `json:"latitude"`
	Timestamp   time.Time `json:"timestamp"`
}

type GSVInfo struct {
	SVPRNNumber float64 `json:"prn"`       // SV PRN number, pseudo-random noise or gold code
	Elevation   float64 `json:"elevation"` // Elevation in degrees, 90 maximum
	Azimuth     float64 `json:"azimuth"`   // Azimuth, degrees from true north, 000 to 359
	SNR         float64 `json:"snr"`       // SNR, 00-99 dB (null when not tracking)
	Used        bool    `json:"used"`
	Type        string  `json:"type"`
}

func NewGPSD(server string, startLongitude, startLatitude float64) *GPSD {
	gpsd := GPSD{}
	return &gpsd
}

func (gpsd *GPSD) Start() {
	var gps *deamon.Session
	var err error

	if gps, err = deamon.Dial(deamon.DefaultAddress); err != nil {
		panic(fmt.Sprintf("Failed to connect to GPSD: %s", err))
	}

	gps.AddFilter("TPV", gpsd.tpvFilter)
	gps.AddFilter("SKY", gpsd.skyfilter)
	done := gps.Watch()
	<-done
}

func (gpsd *GPSD) tpvFilter(r interface{}) {
	tpv := r.(*deamon.TPVReport)

	mode := tpv.Mode
	gpsd.Fix = fix[mode]
	gpsd.Alt = tpv.Alt
	gpsd.Timestamp = tpv.Time
	gpsd.Latitude = tpv.Lat
	gpsd.Longitude = tpv.Lon
}

func (gpsd *GPSD) skyfilter(r interface{}) {
	sky := r.(*deamon.SKYReport)
	gpsd.Hdop = sky.Hdop
	gpsd.Pdop = sky.Pdop
	gpsd.Vdop = sky.Vdop
	sats := sky.Satellites
	gpsd.SatsVisible = make([]GSVInfo, 0)

	for _, sat := range sats {
		newSat := GSVInfo{
			SVPRNNumber: sat.PRN,
			Elevation:   sat.El,
			Azimuth:     sat.Az,
			SNR:         sat.Ss,
			Used:        sat.Used,
		}
		if (newSat.SVPRNNumber >= 1) && (newSat.SVPRNNumber <= 63) {
			newSat.Type = "GPS"

		} else if (newSat.SVPRNNumber >= 64) && (newSat.SVPRNNumber <= 96) {
			newSat.Type = "Glonass"
		} else {
			newSat.Type = "GALILEO"
		}
		gpsd.SatsVisible = append(gpsd.SatsVisible, newSat)
	}
}
