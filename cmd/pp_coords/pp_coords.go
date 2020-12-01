package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/skemper/hcip2"
)

const (
	CountyID int = iota
	PrecinctLabel
	PrecinctDescription
	PrecinctName
	PrecinctAddr
)

var addrRegexp = regexp.MustCompile(`([^,]+),([^,]+), NC (\d{5})`)

const readBatchSize = 10000
const maxLineLength = 1000

var newline = []byte{'\n'}

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

func main() {
	goods, bads, multis := hcip2.MakeFiles()

	// we are reading just one file: 202011_VRDB_Extract.txt
	vrdb, err := os.Open("nc_polling_places.csv")
	defer vrdb.Close()
	if err != nil {
		fmt.Printf("Error opening VRDB: %s\n", err.Error())
		os.Exit(1)
	}
	reader := csv.NewReader(vrdb)

	start := time.Now()
	var lines [][]string
	var badlines [readBatchSize]string
	var numBads = 0
	var multilines [readBatchSize]string
	var numMultis = 0

	var goodlines [readBatchSize]hcip2.JSONResult
	var numGoods = 0

	lines, err = reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading CSV: %s\n", err.Error())
		os.Exit(1)
	}

	for _, line := range lines {
		oneline := strings.Join(line, ",")
		fulladdr := line[PrecinctAddr]
		addrPieces := addrRegexp.FindStringSubmatch(fulladdr)

		// now we can construct the URL for querying the database
		urlBuilder := strings.Builder{}
		urlBuilder.WriteString("http://localhost/nominatim/search?country=us&format=jsonv2&street=")
		urlBuilder.WriteString(addrPieces[0])
		urlBuilder.WriteString("&city=")
		urlBuilder.WriteString(addrPieces[1])
		urlBuilder.WriteString("&state=NC&postalcode=")
		urlBuilder.WriteString(addrPieces[2])
		url := strings.ReplaceAll(urlBuilder.String(), " ", "+")

		v := makeCall(&url)

		if len(v) == 0 {
			badlines[numBads] = oneline
			numBads++
		} else if len(v) > 1 {
			multilines[numMultis] = oneline
			numMultis++
		} else {
			// one record - the good case
			goodlines[numGoods] = v[0]
			goodlines[numGoods].StateVoterIDStr = oneline
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

	end := time.Now()
	fmt.Printf("Finished in %s...\n", end.Sub(start))

}
