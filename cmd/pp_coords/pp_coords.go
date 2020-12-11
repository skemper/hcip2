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
	PollingPlaceName
	PollingPlaceAddr
)

var addrRegexp = regexp.MustCompile(`([^,]+),([^,]+), NC (\d{5})`)

const readBatchSize = 10000
const maxLineLength = 1000

var newline = []byte{'\n'}

func makeCall(url *string) *[]hcip2.JSONResult {
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

	v := new([]hcip2.JSONResult)
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

// query1 decomposes the entire address and feeds the structed data to the API
func query1(addrPieces []string) *[]hcip2.JSONResult {
	urlBuilder := strings.Builder{}
	urlBuilder.WriteString("http://localhost/nominatim/search?country=us&format=jsonv2&street=")
	urlBuilder.WriteString(addrPieces[1])
	urlBuilder.WriteString("&city=")
	urlBuilder.WriteString(addrPieces[2])
	urlBuilder.WriteString("&state=NC&postalcode=")
	urlBuilder.WriteString(addrPieces[3])
	url := strings.ReplaceAll(urlBuilder.String(), " ", "+")
	fmt.Println(url)
	return makeCall(&url)
}

// query2 asks just the location name and the ZIP code
func query2(name string, addrPieces []string) *[]hcip2.JSONResult {
	urlBuilder := strings.Builder{}
	urlBuilder.WriteString("http://localhost/nominatim/search?country=us&format=jsonv2&q=")
	urlBuilder.WriteString(name)
	urlBuilder.WriteString(", ")
	urlBuilder.WriteString(addrPieces[3])
	url := strings.ReplaceAll(urlBuilder.String(), " ", "+")
	fmt.Println(url)
	return makeCall(&url)
}

// query3 is like query1, but without the city
func query3(addrPieces []string) *[]hcip2.JSONResult {
	urlBuilder := strings.Builder{}
	urlBuilder.WriteString("http://localhost/nominatim/search?country=us&format=jsonv2&street=")
	urlBuilder.WriteString(addrPieces[1])
	urlBuilder.WriteString("&state=NC&postalcode=")
	urlBuilder.WriteString(addrPieces[3])
	url := strings.ReplaceAll(urlBuilder.String(), " ", "+")
	fmt.Println(url)
	return makeCall(&url)
}

// query4 looks for the name of the polling place, in North Carolina.  it's a Hail Mary, but it works in at least one case
func query4(name string) *[]hcip2.JSONResult {
	urlBuilder := strings.Builder{}
	urlBuilder.WriteString("http://localhost/nominatim/search?country=us&format=jsonv2&q=")
	urlBuilder.WriteString(name)
	urlBuilder.WriteString(", NC, USA")
	url := strings.ReplaceAll(urlBuilder.String(), " ", "+")
	fmt.Println(url)
	return makeCall(&url)
}

func main() {
	_goods, _bads, _multis := hcip2.MakeFiles()

	// we are reading just one file: 202011_VRDB_Extract.txt
	vrdb, err := os.Open("nc_polling_places.csv")
	defer vrdb.Close()
	if err != nil {
		fmt.Printf("Error opening VRDB: %s\n", err.Error())
		os.Exit(1)
	}
	reader := csv.NewReader(vrdb)
	goods := csv.NewWriter(_goods)
	bads := csv.NewWriter(_bads)
	multis := csv.NewWriter(_multis)

	start := time.Now()
	var lines [][]string
	var badlines [readBatchSize][]string
	var numBads = 0
	var multilines [readBatchSize][]string
	var numMultis = 0

	var goodlines [readBatchSize][]string
	var numGoods = 0

	lines, err = reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading CSV: %s\n", err.Error())
		os.Exit(1)
	}

	count := 0
	for _, line := range lines {
		count++
		oneline := strings.Join(line, ",")
		fmt.Printf("\n")
		fmt.Println(oneline)
		fulladdr := line[PollingPlaceAddr]
		addrPieces := addrRegexp.FindStringSubmatch(fulladdr)
		if addrPieces == nil {
			fmt.Printf("Line %s doesn't match regex\n", fulladdr)
		}

		v := query1(addrPieces)
		if len(*v) == 1 {
			goodlines[numGoods] = append(line, (*v)[0].Lat, (*v)[0].Lon)
			numGoods++
			continue
		}

		v = query2(line[PollingPlaceName], addrPieces)
		if len(*v) == 1 {
			goodlines[numGoods] = append(line, (*v)[0].Lat, (*v)[0].Lon)
			numGoods++
			continue
		}

		v = query3(addrPieces)
		if len(*v) == 1 {
			goodlines[numGoods] = append(line, (*v)[0].Lat, (*v)[0].Lon)
			numGoods++
			continue
		}

		v = query4(line[PollingPlaceName])
		if len(*v) == 1 {
			goodlines[numGoods] = append(line, (*v)[0].Lat, (*v)[0].Lon)
			numGoods++
			continue
		}

		badlines[numBads] = line
		numBads++
	}

	for i := 0; i < numBads; i++ {
		bads.Write(badlines[i])
		goods.Write(append(badlines[i], "", ""))
	}

	for i := 0; i < numMultis; i++ {
		multis.Write(multilines[i])
		goods.Write(append(multilines[i], "", ""))
	}

	for i := 0; i < numGoods; i++ {
		goods.Write(goodlines[i])
	}
	goods.Flush()
	bads.Flush()
	multis.Flush()

	end := time.Now()
	fmt.Printf("Finished (read: %d, wrote: %d) in %s...\n", count, numGoods+numBads+numMultis, end.Sub(start))

}
