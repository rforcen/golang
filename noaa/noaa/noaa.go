// noaa support

package noaa

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

const NOAA_DATA_PATH = "/media/asd/data/code/noaadata/"
const DailyTarBall = "ghcnd_all.tar.gz"
const AuxFilePrefix = "ghcnd-"

var AuxFiles = []string{"countries", "states", "elements", "stations", "inventory"}

type PosLen struct {
	pos int
	len int
}

var COUNTRY_FS = []PosLen{{0, 2}, {3, 80}}
var STATES_FS = []PosLen{{0, 2}, {3, 80}}
var ELEMENTS_FS = []PosLen{{0, 4}, {5, 100}}
var STATIONS_FS = []PosLen{{0, 11}, {12, 8}, {21, 9}, {31, 6}, {38, 2}, {41, 30}, {72, 3}, {76, 3}, {80, 5}}
var INVENTORY_FS = []PosLen{{0, 11}, {12, 8}, {21, 9}, {31, 4}, {36, 4}, {41, 4}}

var prefixMap = map[string][]PosLen{
	"countries": COUNTRY_FS,
	"states":    STATES_FS,
	"elements":  ELEMENTS_FS,
	"stations":  STATIONS_FS,
	"inventory": INVENTORY_FS,
}

type NOAA_DB struct {
	Countries map[string]string
	States    map[string]string
	Elements  map[string]string
	Stations  map[string]Station
	Inventory map[string]string
}

type Station struct {
	Id           string
	Latitude     float64
	Longitude    float64
	Elevation    float64
	State        string
	Name         string
	Gsn_flag     string
	Hcn_crn_flag string
	Wmo_id       string
}

type DailyRaw struct {
	id      [11]byte
	year    [4]byte
	month   [2]byte
	element [4]byte
	items   [31]ItemsRaw
	lf      byte
}

type ItemsRaw struct {
	value [5]byte
	mflag byte
	qflag byte
	sflag byte
}

type Daily struct {
	Id      string
	Year    int
	Month   int
	Element string
	Items   [31]Item
}

type Item struct {
	Value int
	Mflag string
	Qflag string
	Sflag string
}

////

func OpenDB() *NOAA_DB {
	return &NOAA_DB{
		Countries: ReadAuxFile("countries"),
		States:    ReadAuxFile("states"),
		Elements:  ReadAuxFile("elements"),
		Stations:  ReadStations(),
		// Inventory: ReadAuxFile("inventory"), // not yet interested
	}
}

func ReadAuxFile(file string) map[string]string {
	data, err := os.ReadFile(NOAA_DATA_PATH + AuxFilePrefix + file + ".txt")

	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	var prefix []PosLen
	if _prefix, ok := prefixMap[file]; !ok {
		log.Fatalf("file %s not found", file)
	} else {
		prefix = _prefix
	}

	m := make(map[string]string)
	for line := range strings.SplitSeq(string(data), "\n") {
		if line != "" {
			m[line[prefix[0].pos:prefix[0].len]] = strings.TrimSpace(line[prefix[1].pos:min(prefix[1].pos+prefix[1].len, len(line))])
		}
	}
	return m
}

func ReadStations() map[string]Station {
	m := make(map[string]Station)
	data, err := os.ReadFile(NOAA_DATA_PATH + AuxFilePrefix + "stations" + ".txt")
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	toFloat64 := func(s string) float64 {
		f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			log.Fatalf("failed to parse float: %v", err)
		}
		return f
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		if line != "" {
			m[line[STATIONS_FS[0].pos:STATIONS_FS[0].pos+STATIONS_FS[0].len]] = Station{
				Id:           line[STATIONS_FS[0].pos : STATIONS_FS[0].pos+STATIONS_FS[0].len],
				Latitude:     toFloat64(line[STATIONS_FS[1].pos : STATIONS_FS[1].pos+STATIONS_FS[1].len]),
				Longitude:    toFloat64(line[STATIONS_FS[2].pos : STATIONS_FS[2].pos+STATIONS_FS[2].len]),
				Elevation:    toFloat64(line[STATIONS_FS[3].pos : STATIONS_FS[3].pos+STATIONS_FS[3].len]),
				State:        line[STATIONS_FS[4].pos : STATIONS_FS[4].pos+STATIONS_FS[4].len],
				Name:         line[STATIONS_FS[5].pos : STATIONS_FS[5].pos+STATIONS_FS[5].len],
				Gsn_flag:     line[STATIONS_FS[6].pos : STATIONS_FS[6].pos+STATIONS_FS[6].len],
				Hcn_crn_flag: line[STATIONS_FS[7].pos : STATIONS_FS[7].pos+STATIONS_FS[7].len],
				Wmo_id:       line[STATIONS_FS[8].pos : STATIONS_FS[8].pos+STATIONS_FS[8].len],
			}
		}
	}
	return m
}

// daily obs

func NewDailiesRaw(dataFile []byte) []DailyRaw {
	var dailyRaw DailyRaw
	var dailyRaws []DailyRaw
	szDaily := unsafe.Sizeof(dailyRaw)

	for i := 0; i < len(dataFile); i += int(szDaily) {
		dataRec := dataFile[i : i+int(szDaily)]
		copy(dailyRaw.id[:], dataRec[:11])
		copy(dailyRaw.year[:], dataRec[11:15])
		copy(dailyRaw.month[:], dataRec[15:17])
		copy(dailyRaw.element[:], dataRec[17:21])
		for j := 0; j < 31; j++ {
			copy(dailyRaw.items[j].value[:], dataRec[21+j*8:26+j*8])
			dailyRaw.items[j].mflag = dataRec[26+j*8]
			dailyRaw.items[j].qflag = dataRec[27+j*8]
			dailyRaw.items[j].sflag = dataRec[28+j*8]
		}
		dailyRaw.lf = dataRec[156]
		dailyRaws = append(dailyRaws, dailyRaw)
	}
	return dailyRaws
}

func NewDaily(dailyRaw DailyRaw) Daily {
	var daily Daily
	daily.Id = string(dailyRaw.id[:])
	year, _ := strconv.Atoi(string(dailyRaw.year[:]))
	daily.Year = year
	month, _ := strconv.Atoi(string(dailyRaw.month[:]))
	daily.Month = month
	daily.Element = string(dailyRaw.element[:])
	for i := 0; i < 31; i++ {
		value, _ := strconv.Atoi(strings.TrimSpace(string(dailyRaw.items[i].value[:])))
		daily.Items[i].Value = value
		daily.Items[i].Mflag = string(dailyRaw.items[i].mflag)
		daily.Items[i].Qflag = string(dailyRaw.items[i].qflag)
		daily.Items[i].Sflag = string(dailyRaw.items[i].sflag)
	}
	return daily
}

// dailyraw helpers
func (dr *DailyRaw) Country() string {
	return string(dr.id[:2])
}
func (dr *DailyRaw) Id() string {
	return string(dr.id[:])
}

func (dr *DailyRaw) Year() int {
	year, _ := strconv.Atoi(string(dr.year[:]))
	return year
}
func (dr *DailyRaw) Month() int {
	month, _ := strconv.Atoi(string(dr.month[:]))
	return month
}
func (dr *DailyRaw) Element() string {
	return string(dr.element[:])
}
func (dr *DailyRaw) Value(d int) int {
	value, _ := strconv.Atoi(strings.TrimSpace(string(dr.items[d].value[:])))
	if value == -9999 && dr.Qflag(d) == ' ' {
		return 0
	}
	return value
}

func (dr *DailyRaw) Avg() float64 {
	sum := 0.0
	count := 0
	for d := 0; d < 31; d++ {
		sum += float64(dr.Value(d))
		count++
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func (dr *DailyRaw) Min() int {
	min := math.MaxInt
	for d := 0; d < 31; d++ {
		if dr.Value(d) < min {
			min = dr.Value(d)
		}
	}
	return min
}
func (dr *DailyRaw) Max() int {
	max := math.MinInt
	for d := 0; d < 31; d++ {
		if dr.Value(d) > max {
			max = dr.Value(d)
		}
	}
	return max
}
func (dr *DailyRaw) MinMaxAvg() (int, int, float64) {
	min := math.MaxInt
	max := math.MinInt
	sum := 0.0
	count := 0
	for d := 0; d < 31; d++ {
		if dr.Value(d) < min {
			min = dr.Value(d)
		}
		if dr.Value(d) > max {
			max = dr.Value(d)
		}
		sum += float64(dr.Value(d))
		count++
	}
	if count == 0 {
		return 0, 0, 0
	}
	return min, max, sum / float64(count)
}

func (dr *DailyRaw) Mflag(d int) byte {
	return dr.items[d].mflag
}
func (dr *DailyRaw) Qflag(d int) byte {
	return dr.items[d].qflag
}
func (dr *DailyRaw) Sflag(d int) byte {
	return dr.items[d].sflag
}

// /
type DailyTraverse struct {
	Filter       func(DailyRaw, Station) bool
	Progress     func(int, *DailyTraverse, *DailyRaw)
	FoundFunc    func(DailyRaw, unsafe.Pointer)
	ProgressStop int
	MaxFound     int
	CountFound   int
	MaxTraverse  int
	DataPtr      unsafe.Pointer
}

func TraverseDaily(db *NOAA_DB, dt DailyTraverse) {
	filePath := NOAA_DATA_PATH + DailyTarBall

	// Open the gzipped file
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create a new gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		log.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gzr.Close()

	// Create a new tar reader on top of the gzip reader
	tr := tar.NewReader(gzr)
	totFiles := 0

	// Iterate through the files in the archive
	for nfiles := 0; ; nfiles++ {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			log.Fatalf("failed to read next tar entry: %v", err)
		}
		if header.Size == 0 || !strings.Contains(header.Name, ".dly") { // skip empty files and non-daily files
			continue
		}
		if (dt.MaxTraverse > 0 && nfiles >= dt.MaxTraverse) || // by nfiles or #found
			(dt.MaxFound > 0 && dt.CountFound >= dt.MaxFound) {
			break
		}

		data, err := io.ReadAll(tr) // read file
		if err != nil {
			log.Fatalf("failed to read file content: %v", err)
		}

		dailysRaw := NewDailiesRaw(data) // create daily array

		station := db.Stations[string(dailysRaw[0].id[:])] // get station

		if dt.ProgressStop > 0 && dt.Progress != nil && nfiles%dt.ProgressStop == 0 { // progress
			dt.Progress(nfiles, &dt, &dailysRaw[0])
		}

		for _, drec := range dailysRaw { // traverse files = daily array
			if dt.Filter != nil && dt.Filter(drec, station) {
				if dt.FoundFunc != nil {
					dt.FoundFunc(drec, dt.DataPtr)
				}
				dt.CountFound++
			}
		}
		totFiles++
	}
}
