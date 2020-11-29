package hcip2

import (
	"fmt"
	"os"
)

// JSONResult is the jsonv2 type we get from the nominatim API
type JSONResult struct {
	PlaceID           int `json:"place_id"`
	Licence           string
	OSMType           string `json:"osm_type"`
	OSMID             int    `json:"osm_id"`
	Boundingbox       [4]string
	Lat               string
	Lon               string
	DisplayName       string `json:"display_name"`
	PlaceRank         int    `json:"place_rank"`
	Category          string
	Objtype           string `json:"type"`
	Importance        float64
	StateVoterIDBytes [25]byte
	StateVoterIDStr   string
}

// MakeFiles sets up files for spitting out good, no, and multi-result Nominatim searches
func MakeFiles() (goods *os.File, bads *os.File, multis *os.File) {
	goods, err := os.OpenFile("goods.csv", os.O_WRONLY+os.O_CREATE, 0664)
	// defer goods.Close()
	if err != nil {
		fmt.Printf("Error opening goods.csv: %s\n", err.Error())
		os.Exit(1)
	}

	bads, err = os.OpenFile("bads.csv", os.O_WRONLY+os.O_CREATE, 0664)
	// defer bads.Close()
	if err != nil {
		fmt.Printf("Error opening bads.csv: %s\n", err.Error())
		os.Exit(1)
	}

	multis, err = os.OpenFile("multis.csv", os.O_WRONLY+os.O_CREATE, 0664)
	// defer multis.Close()
	if err != nil {
		fmt.Printf("Error opening multis.csv: %s\n", err.Error())
		os.Exit(1)
	}

	return goods, bads, multis
}

// MakeFilesWithPrefix sets up files for spitting out good, no, and multi-result Nominatim searches
func MakeFilesWithPrefix(prefix string) (goods *os.File, bads *os.File, multis *os.File) {
	goods, err := os.OpenFile(prefix+"goods.csv", os.O_WRONLY+os.O_CREATE, 0664)
	// defer goods.Close()
	if err != nil {
		fmt.Printf("Error opening goods.csv: %s\n", err.Error())
		os.Exit(1)
	}

	bads, err = os.OpenFile(prefix+"bads.csv", os.O_WRONLY+os.O_CREATE, 0664)
	// defer bads.Close()
	if err != nil {
		fmt.Printf("Error opening bads.csv: %s\n", err.Error())
		os.Exit(1)
	}

	multis, err = os.OpenFile(prefix+"multis.csv", os.O_WRONLY+os.O_CREATE, 0664)
	// defer multis.Close()
	if err != nil {
		fmt.Printf("Error opening multis.csv: %s\n", err.Error())
		os.Exit(1)
	}

	return goods, bads, multis
}
