package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/TomOnTime/utfutil"
	"github.com/skemper/hcip2"
)

const readBatchSize = 10000
const maxLineLength = 1000

var newline = []byte{'\n'}

func copyByteArray(dst *[maxLineLength]byte, src []byte) int {
	slen := len(src)
	for i := 0; i < slen; i++ {
		(*dst)[i] = src[i]
	}
	return slen
}

func copyByteArray2(dst *[25]byte, src []byte) int {
	slen := len(src)
	for i := 0; i < slen; i++ {
		(*dst)[i] = src[i]
	}
	return slen
}

func main() {
	goods, bads, multis := hcip2.MakeFiles()

	var config hcip2.HciConfig = hcip2.Configs[os.Args[1]]

	switch os.Args[3] {
	case "b":
		doBytes(&config, goods, bads, multis)
		break
	case "s":
		doStrings(&config, goods, bads, multis)
	}
}

func makeCall(url *string) ([]hcip2.JSONResult, error) {
	// Call Nominatim to geocode their address
	resp, err := http.Get(*url)
	if err != nil {
		return nil, fmt.Errorf("Error calling Nominatim: %s", err.Error())
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-OK response code from %s: %d %s\n", *url, resp.StatusCode, resp.Body)
	}

	v := []hcip2.JSONResult{}
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		fmt.Printf("Error decoding JSON: %s\n", err.Error())
		fmt.Printf("From URL: %s\n", *url)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Full response was: %s\n", body)
		return nil, err
	}
	return v, nil
}

func doBytes(config *hcip2.HciConfig, goods *os.File, bads *os.File, multis *os.File) {
	vrdbFilename := os.Args[2]
	vrdb, err := os.Open(vrdbFilename)
	defer vrdb.Close()
	if err != nil {
		fmt.Printf("Error opening VRDB file %s: %s\n", vrdbFilename, err.Error())
		os.Exit(1)
	}
	scanner := bufio.NewScanner(vrdb)
	scanner.Scan() // skip the first line

	numCycles := 0
	done := false

	for !done {
		start := time.Now()
		var lines [readBatchSize][maxLineLength]byte
		var lineLengths [readBatchSize]int
		var badlines [readBatchSize][maxLineLength]byte
		var badlineLengths [readBatchSize]int
		var numBads = 0
		var multilines [readBatchSize][maxLineLength]byte
		var multilineLengths [readBatchSize]int
		var numMultis = 0

		var goodlines [readBatchSize]hcip2.JSONResult
		var goodlineVoterIDLengths [readBatchSize]int
		var numGoods = 0

		// we're going to read these in batches
		for i := 0; i < readBatchSize; i++ {
			if !scanner.Scan() {
				done = true
				break
			}
			line := scanner.Bytes()
			lineLengths[i] = copyByteArray(&lines[i], line)
		}

		for i, line := range lines {
			// we're going to cobble their street address together
			// fmt.Printf("Working on %s\n", line)
			pieces := bytes.Split(line[:], []byte{'|'})
			// fmt.Printf("Split to %s\n", pieces)
			roadBuilder := strings.Builder{}
			for _, piece := range config.RoadNoUnit {
				for i := range pieces[piece] {
					if pieces[piece][i] == ' ' {
						pieces[piece][i] = '+'
					}
				}
				roadBuilder.Write(pieces[piece])
				roadBuilder.WriteRune('+')
			}
			streetAddr := roadBuilder.String()

			// now we can construct the URL for querying the database
			urlBuilder := strings.Builder{}
			urlBuilder.WriteString("http://localhost/nominatim/search?country=us&format=jsonv2&street=")
			urlBuilder.WriteString(streetAddr)
			urlBuilder.WriteString("&city=")
			for i := range pieces[config.CITY] {
				if pieces[config.CITY][i] == ' ' {
					pieces[config.CITY][i] = '+'
				}
			}
			urlBuilder.Write(pieces[config.CITY])
			urlBuilder.WriteString("&state=")
			urlBuilder.Write(pieces[config.STATE])
			urlBuilder.WriteString("&postalcode=")
			urlBuilder.Write(pieces[config.ZIP])
			url := urlBuilder.String()

			v, err := makeCall(&url)
			if err != nil {
				fmt.Println("Error geocoding: %s\n", err)
				fmt.Println("Line was %s\n", line)
			}

			if len(v) == 0 {
				badlines[numBads] = line
				badlineLengths[numBads] = lineLengths[i]
				numBads++
			} else if len(v) > 1 {
				multilines[numMultis] = line
				multilineLengths[numMultis] = lineLengths[i]
				numMultis++
			} else {
				// one record - the good case
				goodlines[numGoods] = v[0]
				goodlineVoterIDLengths[numGoods] = copyByteArray2(&goodlines[numGoods].StateVoterIDBytes, pieces[config.STATE_VOTER_ID])
				numGoods++
			}
		}

		for i := 0; i < numBads; i++ {
			bads.Write(badlines[i][:badlineLengths[i]])
			bads.Write(newline)
		}

		for i := 0; i < numMultis; i++ {
			multis.Write(multilines[i][:multilineLengths[i]])
			multis.Write(newline)
		}

		for i := 0; i < numGoods; i++ {
			goods.WriteString(fmt.Sprintf("%s,%s,%s\n", goodlines[i].StateVoterIDBytes[:goodlineVoterIDLengths[i]], goodlines[i].Lat, goodlines[i].Lon))
		}

		numCycles++
		end := time.Now()
		fmt.Printf("Finished %d records in %s...\n", numCycles*readBatchSize, end.Sub(start))
	}
}

func doStrings(config *hcip2.HciConfig, goods *os.File, bads *os.File, multis *os.File) {
	// we are reading just one file: 202011_VRDB_Extract.txt
	vrdb, err := utfutil.OpenFile("VR_Snapshot_20201103.txt", utfutil.WINDOWS)
	defer vrdb.Close()
	if err != nil {
		fmt.Printf("Error opening VRDB: %s\n", err.Error())
		os.Exit(1)
	}
	scanner := bufio.NewScanner(vrdb)
	scanner.Scan() // skip the first line

	numCycles := 0
	done := false
	splitchar := "\t"

	for !done {
		start := time.Now()
		var lines [readBatchSize]string
		var badlines [readBatchSize]string
		var numBads = 0
		var multilines [readBatchSize]string
		var numMultis = 0

		var goodlines [readBatchSize]hcip2.JSONResult
		var numGoods = 0

		// we're going to read these in batches
		for i := 0; i < readBatchSize; i++ {
			if !scanner.Scan() {
				done = true
				break
			}
			lines[i] = scanner.Text()
		}

		for _, line := range lines {
			// we're going to cobble their street address together
			// fmt.Printf("Working on %s\n", line)
			pieces := strings.Split(line[:], splitchar)

			if !config.FilterStr(pieces) {
				continue
			}

			// fmt.Printf("Split to %s\n", pieces)
			roadBuilder := strings.Builder{}
			for _, piece := range config.RoadNoUnit {
				val := strings.ReplaceAll(pieces[piece], " ", "+")
				roadBuilder.WriteString(val)
				roadBuilder.WriteRune('+')
			}
			streetAddr := roadBuilder.String()

			// now we can construct the URL for querying the database
			urlBuilder := strings.Builder{}
			urlBuilder.WriteString("http://localhost/nominatim/search?country=us&format=jsonv2&street=")
			urlBuilder.WriteString(streetAddr)
			urlBuilder.WriteString("&city=")
			urlBuilder.WriteString(strings.ReplaceAll(pieces[config.CITY], " ", "+"))
			urlBuilder.WriteString("&state=")
			urlBuilder.WriteString(pieces[config.STATE])
			urlBuilder.WriteString("&postalcode=")
			urlBuilder.WriteString(pieces[config.ZIP])
			url := urlBuilder.String()

			v, err := makeCall(&url)
			if err != nil {
				fmt.Println("Error geocoding: %s\n", err)
				fmt.Println("Line was %s\n", line)
			}

			if len(v) == 0 {
				badlines[numBads] = line
				numBads++
			} else if len(v) > 1 {
				multilines[numMultis] = line
				numMultis++
			} else {
				// one record - the good case
				goodlines[numGoods] = v[0]
				goodlines[numGoods].StateVoterIDStr = pieces[config.STATE_VOTER_ID]
				numGoods++
			}
		}

		for i := 0; i < numBads; i++ {
			bads.WriteString(badlines[i])
			bads.Write(newline)
		}

		for i := 0; i < numMultis; i++ {
			multis.WriteString(multilines[i])
			multis.Write(newline)
		}

		for i := 0; i < numGoods; i++ {
			goods.WriteString(fmt.Sprintf("%s,%s,%s\n", goodlines[i].StateVoterIDStr, goodlines[i].Lat, goodlines[i].Lon))
		}

		numCycles++
		end := time.Now()
		fmt.Printf("Finished %d records in %s...\n", numCycles*readBatchSize, end.Sub(start))
	}
}
