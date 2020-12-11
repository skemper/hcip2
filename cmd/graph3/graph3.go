package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/TomOnTime/utfutil"
	"github.com/skemper/hcip2"
)

// voter_coords.csv
const (
	vcNcid int = iota
	vcLat
	vcLon
)

type Voter struct {
	County_id                int       //			char  3         County identification number
	Ncid                     string    //				char 12         North Carolina identification number (NCID) of voter
	Status_cd                byte      //			char  1	        Status code for voter registration
	Voter_status_desc        string    //		char 10         Satus code descriptions.
	Reason_cd                string    //			char  2         Reason code for voter registration status
	Voter_status_reason_desc string    //	char 60         Reason code description
	Race_code                string    // char  3         Race code
	Race_desc                string    //			char 35         Race description
	Ethnic_code              string    //			char  2         Ethnicity code
	Ethnic_desc              string    //			char 30         Ethnicity description
	Party_cd                 string    //			char  3         Party affiliation code
	Sex_code                 byte      //			char  1         Gender code
	Age                      int       //				char  3         Age
	Registr_dt               time.Time //			char 10         Voter registration date
	Precinct_abbrv           string    //			char  6         Precinct abbreviation
	Precinct_desc            string    //			char 30         Precinct name
	Cancellation_dt          time.Time //			char 10         Cancellation date
	Vtd_abbrv                string    //			char  6         Voter tabuluation district abbreviation
	Vtd_desc                 string    //			char 30         Voter tabuluation district name
	Age_group                string    //			char 35         Age group range
	Lat                      float64
	Lon                      float64
}

var voters map[string]*Voter = make(map[string]*Voter)

// precinct_coords.csv
const (
	ppCountyIDIdx int = iota
	ppPrecinctCodeIdx
	ppPrecinctNameIdx
	ppNameIdx
	ppAddressIdx
	ppLatIdx
	ppLonIdx
)

type Precinct struct {
	countyID     int
	precinctCode string
	precinctName string
	ppName       string
	ppAddress    string
	ppLat        float64
	ppLon        float64
	distances    []float64
	avgDistance  *big.Float
}

var precincts map[string]*Precinct = make(map[string]*Precinct)

func loadVoterDatabase() {
	start := time.Now()
	vrdb, err := utfutil.OpenFile("VR_Snapshot_20201103.txt", utfutil.WINDOWS)
	defer vrdb.Close()
	if err != nil {
		fmt.Printf("Error opening VRDB: %s\n", err)
		os.Exit(1)
	}
	scanner := bufio.NewScanner(vrdb)
	scanner.Scan() // skip the header line
	for scanner.Scan() {
		line := scanner.Text()
		pieces := strings.Split(line, "\t")
		countyID, err := strconv.Atoi(strings.TrimSpace(pieces[hcip2.County_id]))
		if err != nil {
			fmt.Printf("Error converting county %s to integer for %s: %s\n", pieces[hcip2.County_id], pieces[hcip2.Ncid], err)
			os.Exit(1)
		}
		age, err := strconv.Atoi(strings.TrimSpace(pieces[hcip2.Age]))
		if err != nil {
			fmt.Printf("Error converting age %s to integer for %s: %s\n", pieces[hcip2.County_id], pieces[hcip2.Ncid], err)
			os.Exit(1)
		}
		regDate, err := time.Parse("2006-01-02", pieces[hcip2.Registr_dt])
		if err != nil {
			fmt.Printf("Error converting regdate %s to date for %s: %s\n", pieces[hcip2.Registr_dt], pieces[hcip2.Ncid], err)
			os.Exit(1)
		}
		canDate, err := time.Parse("2006-01-02", pieces[hcip2.Cancellation_dt])
		if err != nil {
			fmt.Printf("Error converting candate %s to date for %s: %s\n", pieces[hcip2.Cancellation_dt], pieces[hcip2.Ncid], err)
			os.Exit(1)
		}
		voter := &Voter{
			County_id:                countyID,
			Ncid:                     strings.Trim(pieces[hcip2.Ncid], "\""),
			Status_cd:                pieces[hcip2.Status_cd][0],
			Voter_status_desc:        strings.Trim(pieces[hcip2.Voter_status_desc], "\""),
			Reason_cd:                strings.Trim(pieces[hcip2.Reason_cd], "\""),
			Voter_status_reason_desc: strings.Trim(pieces[hcip2.Voter_status_reason_desc], "\""),
			Race_code:                strings.Trim(pieces[hcip2.Race_code], "\""),
			Race_desc:                strings.Trim(pieces[hcip2.Race_desc], "\""),
			Ethnic_code:              strings.Trim(pieces[hcip2.Ethnic_code], "\""),
			Ethnic_desc:              strings.Trim(pieces[hcip2.Ethnic_desc], "\""),
			Party_cd:                 strings.Trim(pieces[hcip2.Party_cd], "\""),
			Sex_code:                 pieces[hcip2.Sex_code][0],
			Age:                      age,
			Registr_dt:               regDate,
			Precinct_abbrv:           strings.Trim(pieces[hcip2.Precinct_abbrv], "\""),
			Precinct_desc:            strings.Trim(pieces[hcip2.Precinct_desc], "\""),
			Cancellation_dt:          canDate,
			Vtd_abbrv:                strings.Trim(pieces[hcip2.Vtd_abbrv], "\""),
			Vtd_desc:                 strings.Trim(pieces[hcip2.Vtd_desc], "\""),
			Age_group:                strings.Trim(pieces[hcip2.Age_group], "\""),
		}
		voters[voter.Ncid] = voter
	}
	fmt.Printf("Loaded VRDB in %s...\n", time.Now().Sub(start))
}

func checkIsBadLatLong(lat float64, lon float64) bool {
	return lat < 33.842316 ||
		lat > 36.588117 ||
		lon < -84.321869 ||
		lon > -75.460621
}

func loadVoterCoords() {
	theFile, err := os.Open("voter_coords.csv")
	defer theFile.Close()
	if err != nil {
		fmt.Printf("Error opening voter_coords.csv: %s\n", err)
		os.Exit(1)
	}

	reader := csv.NewReader(theFile)
	lines, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading CSV data: %s\n", err)
	}

	count := 0
	countBadLatLong := 0
	for _, line := range lines {
		if count%1000000 == 0 {
			fmt.Printf("Processed %d rows of voter coordinates...\n", count)
		}

		count++

		if line[vcLat] == "" {
			continue
		}

		if voter, ok := voters[line[vcNcid]]; ok {
			// fmt.Printf("Adding coordinates for voter %s\n", line[vcNcid])
			lat, err := strconv.ParseFloat(line[vcLat], 64)
			if err != nil {
				fmt.Printf("Couldn't convert lat %s to float\n", line[vcLat])
				continue
			}
			lon, err := strconv.ParseFloat(line[vcLon], 64)
			if err != nil {
				fmt.Printf("Couldn't convert lon %s to float\n", line[vcLon])
				continue
			}

			if checkIsBadLatLong(lat, lon) {
				countBadLatLong++
				continue
			}

			voter.Lat = lat
			voter.Lon = lon
			// fmt.Printf("Coordinates are %f, %f\n", voters[line[vcNcid]].Lat, voters[line[vcNcid]].Lon)
		} else {
			fmt.Printf("Missing voter %s!?\n", line[vcNcid])
		}
	}
	fmt.Printf("** Found a total of %d bad voter lat/longs\n", countBadLatLong)
}

func getPrecinctLabel(countyID string, precinctCode string) string {
	return fmt.Sprintf("%s_%s", countyID, precinctCode)
}

func loadPrecinctData() {
	theFile, err := os.Open("precinct_coords.csv")
	defer theFile.Close()
	if err != nil {
		fmt.Printf("Error opening precinct_coords.csv: %s\n", err)
		os.Exit(1)
	}

	reader := csv.NewReader(theFile)
	lines, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading CSV data: %s\n", err)
	}

	count := 0
	countBadLatLong := 0
	for _, line := range lines {
		if count%1000 == 0 {
			fmt.Printf("Processed %d rows of precinct coordinates...\n", count)
		}

		count++

		// we couldn't find a location for this polling place
		if line[ppLatIdx] == "" {
			continue
		}

		countyID, err := strconv.Atoi(line[ppCountyIDIdx])
		if err != nil {
			fmt.Printf("Couldn't convert %s to valid county ID: %s", line[ppCountyIDIdx], err)
			continue
		}
		lat, err := strconv.ParseFloat(line[ppLatIdx], 64)
		if err != nil {
			fmt.Printf("Couldn't convert lat %s to float\n", line[ppLatIdx])
			continue
		}
		lon, err := strconv.ParseFloat(line[ppLonIdx], 64)
		if err != nil {
			fmt.Printf("Couldn't convert lon %s to float\n", line[ppLonIdx])
			continue
		}

		if checkIsBadLatLong(lat, lon) {
			countBadLatLong++
			continue
		}

		p := new(Precinct)

		p.countyID = countyID
		p.precinctCode = line[ppPrecinctCodeIdx]
		p.precinctName = line[ppPrecinctNameIdx]
		p.ppName = line[ppNameIdx]
		p.ppAddress = line[ppAddressIdx]
		p.ppLat = lat
		p.ppLon = lon

		label := getPrecinctLabel(line[ppCountyIDIdx], line[ppPrecinctCodeIdx])
		precincts[label] = p
		fmt.Printf("Added precinct %s\n", label)
		fmt.Printf("Coordinates are %f, %f\n", precincts[label].ppLat, precincts[label].ppLon)
	}

	fmt.Printf("** Found a total of %d bad precinct lat/longs\n", countBadLatLong)
}

// earthRadius is the radius of the Earth, in meters
const earthRadius = 6378.137

func haversineDistance(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	latA := lat1 * (math.Pi / 180.0)
	latB := lat2 * (math.Pi / 180.0)

	a1 := math.Sin(dLat/2) * math.Sin(dLat/2)
	a2 := math.Sin(dLon/2) * math.Sin(dLon/2) * math.Cos(latA) * math.Cos(latB)

	a := a1 + a2

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c * 1000
}

func calculateDistances() {
	count := 0
	noVoterLoc := 0
	noPrecinctLoc := 0
	for _, voter := range voters {
		count++
		if voter.Lat == +0.0 {
			// no data, skip
			noVoterLoc++
			continue
		}

		pLabel := getPrecinctLabel(strconv.Itoa(voter.County_id), voter.Precinct_abbrv)
		// fmt.Printf("Looking for precinct %s\n", pLabel)
		if precinct, ok := precincts[pLabel]; ok {
			distance := haversineDistance(voter.Lat, voter.Lon, precinct.ppLat, precinct.ppLon)
			precinct.distances = append(precinct.distances, distance)
		} else {
			// no precinct location, skip
			noPrecinctLoc++
		}

		if count%1000000 == 0 {
			fmt.Printf("Finished measuring %d distances (%d bad p / %d bad v)...\n", count, noPrecinctLoc, noVoterLoc)
		}
	}
}

func getAverageDistances() {
	countPrecincts := 0
	for _, precinct := range precincts {
		// fmt.Printf("Working precinct %s\n", k)
		countPrecincts++
		count := 0
		sum := big.NewFloat(0.0)
		// var sum float64 = 0.0
		for _, distance := range precinct.distances {
			// fmt.Printf("Distance: %f\n", distance)
			count++
			sum = sum.Add(sum, big.NewFloat(distance))
			// sum += distance
		}

		if count == 0 {
			continue // nothing to see here
		}

		precinct.avgDistance = sum.Quo(sum, big.NewFloat(float64(count)))
		// fmt.Printf("** AVERAGE: %f\n", precinct.avgDistance)

		if countPrecincts%1000 == 0 {
			fmt.Printf("Finished averaging %d precincts...\n", countPrecincts)
		}
	}
}

func main() {
	outFile, err := os.OpenFile("graph3.csv", os.O_CREATE+os.O_WRONLY, 0644)
	defer outFile.Close()
	if err != nil {
		fmt.Printf("Error opening CSV for writing results: %s\n", err)
		os.Exit(1)
	}

	loadVoterDatabase()

	loadVoterCoords()

	loadPrecinctData()

	calculateDistances()

	getAverageDistances()

	// dump it to CSV
	writer := csv.NewWriter(outFile)
	writer.Write([]string{"COUNTY", "PRECINCT", "AVERAGE_DISTANCE"})
	for k, v := range precincts {
		pieces := strings.Split(k, "_")
		countyID, _ := strconv.Atoi(pieces[0])
		var avgDist float64
		if v.avgDistance == nil {
			avgDist = 0
		} else {
			avgDist, _ = v.avgDistance.Float64()
		}
		writer.Write([]string{hcip2.Counties[countyID], pieces[1], strconv.FormatFloat(avgDist, 'f', 4, 64)})
	}
}
