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

const readBatchSize = 1000
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

func makeCall(url *string) []hcip2.JSONResult {
	// Call Nominatim to geocode their address
	resp, err := http.Get(*url)
	if err != nil {
		fmt.Printf("Error calling Nominatim: %s\n", err.Error())
		os.Exit(1)
	}
	if resp.StatusCode != 200 {
		fmt.Printf("Non-OK response code from Nominatim: %d %s\n", resp.StatusCode, resp.Body)
		fmt.Printf("Came from URL %s\n", *url)
		os.Exit(1)
	}

	v := []hcip2.JSONResult{}
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		fmt.Printf("Error decoding JSON: %s\n", err.Error())
		fmt.Printf("From URL: %s\n", *url)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Full response was: %s\n", body)
		os.Exit(1)
	}
	return v
}

type myString struct {
	Data   [maxLineLength]byte
	Length int
}

func (m myString) String() string {
	return fmt.Sprintf("%s", m.Data)
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
		var lines chan myString = make(chan myString, readBatchSize)
		var goodlines chan hcip2.JSONResult = make(chan hcip2.JSONResult, readBatchSize/10)
		var badlines chan myString = make(chan myString, readBatchSize/10)
		var multilines chan myString = make(chan myString, readBatchSize/10)

		// we're going to read these in batches
		for i := 0; i < readBatchSize; i++ {
			if !scanner.Scan() {
				done = true
				break
			}
			line := scanner.Bytes()
			tosend := myString{}
			tosend.Length = copyByteArray(&tosend.Data, line)
			lines <- tosend
		}
		close(lines)

		for line := range lines {
			// we're going to cobble their street address together
			// fmt.Printf("Working on %s\n", line)
			go lookupMyStringAddress(line, config, goodlines, badlines, multilines)
		}

		innerdone := false
		for !innerdone {
			select {
			case badline := <-badlines:
				bads.Write(badline.Data[:badline.Length])
				bads.Write(newline)
				break
			case multiline := <-multilines:
				multis.Write(multiline.Data[:multiline.Length])
				multis.Write(newline)
				break
			case goodline := <-goodlines:
				goods.WriteString(fmt.Sprintf("%s,%s,%s\n", goodline.StateVoterIDBytes[:goodline.StateVoterIDLength], goodline.Lat, goodline.Lon))
				break
			case <-time.After(30 * time.Second):
				fmt.Println("timeout 1")
				innerdone = true
				break
			}
		}

		numCycles++
		end := time.Now()
		fmt.Printf("Finished %d records in %s...\n", numCycles*readBatchSize, end.Sub(start))
	}
}

func lookupMyStringAddress(line myString, config *hcip2.HciConfig, goodlines chan hcip2.JSONResult, badlines chan myString, multilines chan myString) {
	pieces := bytes.Split(line.Data[:], []byte{'|'})
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

	v := makeCall(&url)

	if len(v) == 0 {
		badlines <- line
	} else if len(v) > 1 {
		multilines <- line
	} else {
		// one record - the good case
		v[0].StateVoterIDLength = copyByteArray2(&v[0].StateVoterIDBytes, pieces[config.STATE_VOTER_ID])
		goodlines <- v[0]
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
			go func(line string) {
				// we're going to cobble their street address together
				// fmt.Printf("Working on %s\n", line)
				pieces := strings.Split(line[:], splitchar)

				if !config.FilterStr(pieces) {
					return
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

				v := makeCall(&url)

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
			}(line)
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
