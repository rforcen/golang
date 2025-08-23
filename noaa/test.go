package main

import (
	"fmt"
	"noaa/noaa"
	"os"
	"sort"
	"time"
	"unsafe"

	"github.com/go-gota/gota/dataframe"
	"github.com/wcharczuk/go-chart/v2"
)

func TestOpenDB() {
	t0 := time.Now()

	db := noaa.OpenDB()
	fmt.Println("NOAA_DB\nlap to open all aux files into maps: ", time.Since(t0), "\n")

	fmt.Printf("Countries: %+v\n", db.Countries["US"])
	fmt.Printf("States   : %+v\n", db.States["AL"])
	fmt.Printf("Elements : %+v\n", db.Elements["TMAX"])
	fmt.Printf("Stations : %+v\n", db.Stations["USW00094728"])

	fmt.Println("\n")

}

func TestGetDailyObs() { // create an array of DailyObsStats and a DataFrame with it
	type DailyObsStat struct {
		Id      string
		Country string
		Element string
		Year    int
		Month   int
		Max     int
		Min     int
		Avg     float64
	}

	doStat := make([]DailyObsStat, 0)

	db := noaa.OpenDB()

	dt := noaa.DailyTraverse{
		Filter: func(dr noaa.DailyRaw, station noaa.Station) bool {
			return dr.Year() >= 2020 && dr.Element() == "TMAX"
		},
		Progress: func(nfile int, dt *noaa.DailyTraverse, dr *noaa.DailyRaw) {
			fmt.Printf("file #: %d, found: %d %d %v %v\r", nfile, dt.CountFound, dr.Year(), dr.Country(), dr.Id())
		},
		FoundFunc: func(dr noaa.DailyRaw, dataPtr unsafe.Pointer) {
			dos := (*[]DailyObsStat)(dataPtr)
			*dos = append(*dos, DailyObsStat{
				Id:      dr.Id(),
				Country: dr.Country(),
				Element: dr.Element(),
				Year:    dr.Year(),
				Month:   dr.Month(),

				Max:     dr.Max(),
				Min:     dr.Min(),
				Avg:     dr.Avg(),
			})
		},
		DataPtr:      unsafe.Pointer(&doStat),
		ProgressStop: 1000,
		MaxTraverse:  100000,
		MaxFound:     100000,
	}

	t0 := time.Now()

	noaa.TraverseDaily(db, dt)

	fmt.Printf("\n\nlap to read daily: %v, items: %d\n", time.Since(t0), len(doStat))

	// create the data frame
	df := dataframe.LoadStructs(doStat)
	arrayRange:=func(start, end int) []int {
		a:=make([]int, end-start)
		for i:=start; i<end; i++ {
			a[i-start]=i
		}
		return a
	}
	df10 := df.Subset(arrayRange(0, 10))

	fmt.Println(df10)	
}

func TestTraverseDailyObs() {

	db := noaa.OpenDB()

	type YearStat struct { // [year]max of tmax
		yearTMaxMap map[int]int
	}
	ym := YearStat{yearTMaxMap: make(map[int]int)}
	t0 := time.Now()

	dt := noaa.DailyTraverse{
		Filter: func(dr noaa.DailyRaw, station noaa.Station) bool {
			return dr.Year() >= 2000 && dr.Element() == "TMAX"
		},
		Progress: func(nfile int, dt *noaa.DailyTraverse, dr *noaa.DailyRaw) {
			fmt.Printf("file #: %d, found: %d %d %v %v\r", nfile, dt.CountFound, dr.Year(), dr.Country(), dr.Id())
		},
		FoundFunc: func(dr noaa.DailyRaw, dataPtr unsafe.Pointer) {
			if max_temp, ym := dr.Max(), (*YearStat)(dataPtr); max_temp < 750 { // tenths of degrees
				if tmax, ok := ym.yearTMaxMap[dr.Year()]; ok {
					ym.yearTMaxMap[dr.Year()] = noaa.MaxInt(tmax, max_temp)
				} else {
					ym.yearTMaxMap[dr.Year()] = max_temp
				}
			}
		},
		DataPtr:      unsafe.Pointer(&ym),
		ProgressStop: 1000,
		// MaxTraverse:  10000,
		// MaxFound:     1000000,
	}
	noaa.TraverseDaily(db, dt)
	fmt.Printf("\n\nlap to read daily: %v\n", time.Since(t0))

	// create a sorted array of years and max temps
	years := make([]float64, 0, len(ym.yearTMaxMap))
	for year := range ym.yearTMaxMap {
		years = append(years, float64(year))
	}
	sort.Float64s(years)
	max_temps := make([]float64, len(years))
	for i, temp := range years {
		max_temps[i] = float64(ym.yearTMaxMap[int(temp)]) / 10.0 // tmax in tenths of degrees
	}

	// plot chart
	linearRegression := func(x, y []float64) (lineX, lineY []float64) {
		n := float64(len(x))
		var sumX, sumY, sumXY, sumX2 float64
		for i := 0; i < int(n); i++ {
			sumX += x[i]
			sumY += y[i]
			sumXY += x[i] * y[i]
			sumX2 += x[i] * x[i]
		}

		m := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
		b := (sumY / n) - (m * sumX / n)

		lineX = make([]float64, len(x))
		lineY = make([]float64, len(y))
		for i := 0; i < len(x); i++ {
			lineX[i] = x[i]
			lineY[i] = m*x[i] + b
		}
		return lineX, lineY
	}
	lineX, lineY := linearRegression(years, max_temps)

	// plot them to chart.png
	graph := chart.Chart{
		Title: "global TMAX for year > 2000",
		Series: []chart.Series{
			chart.ContinuousSeries{
				Name:    "years/TMAX",
				XValues: years,
				YValues: max_temps,
			},
			chart.ContinuousSeries{
				Name:    "Linear Regression",
				XValues: lineX,
				YValues: lineY,
			},
		},
	}

	f, _ := os.Create("chart.png")
	defer f.Close()
	graph.Render(chart.PNG, f)
}

func main() {
	TestGetDailyObs()
}
