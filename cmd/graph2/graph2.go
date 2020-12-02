package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/TomOnTime/utfutil"
	"github.com/skemper/hcip2"
)

// county ID -> name mappings
var counties = []string{"", // nothing at ID 0
	"ALAMANCE",
	"ALEXANDER",
	"ALLEGHANY",
	"ANSON",
	"ASHE",
	"AVERY",
	"BEAUFORT",
	"BERTIE",
	"BLADEN",
	"BRUNSWICK",
	"BUNCOMBE",
	"BURKE",
	"CABARRUS",
	"CALDWELL",
	"CAMDEN",
	"CARTERET",
	"CASWELL",
	"CATAWBA",
	"CHATHAM",
	"CHEROKEE",
	"CHOWAN",
	"CLAY",
	"CLEVELAND",
	"COLUMBUS",
	"CRAVEN",
	"CUMBERLAND",
	"CURRITUCK",
	"DARE",
	"DAVIDSON",
	"DAVIE",
	"DUPLIN",
	"DURHAM",
	"EDGECOMBE",
	"FORSYTH",
	"FRANKLIN",
	"GASTON",
	"GATES",
	"GRAHAM",
	"GRANVILLE",
	"GREENE",
	"GUILFORD",
	"HALIFAX",
	"HARNETT",
	"HAYWOOD",
	"HENDERSON",
	"HERTFORD",
	"HOKE",
	"HYDE",
	"IREDELL",
	"JACKSON",
	"JOHNSTON",
	"JONES",
	"LEE",
	"LENOIR",
	"LINCOLN",
	"MACON",
	"MADISON",
	"MARTIN",
	"MCDOWELL",
	"MECKLENBURG",
	"MITCHELL",
	"MONTGOMERY",
	"MOORE",
	"NASH",
	"NEWHANOVER",
	"NORTHAMPTON",
	"ONSLOW",
	"ORANGE",
	"PAMLICO",
	"PASQUOTANK",
	"PENDER",
	"PERQUIMANS",
	"PERSON",
	"PITT",
	"POLK",
	"RANDOLPH",
	"RICHMOND",
	"ROBESON",
	"ROCKINGHAM",
	"ROWAN",
	"RUTHERFORD",
	"SAMPSON",
	"SCOTLAND",
	"STANLY",
	"STOKES",
	"SURRY",
	"SWAIN",
	"TRANSYLVANIA",
	"TYRRELL",
	"UNION",
	"VANCE",
	"WAKE",
	"WARREN",
	"WASHINGTON",
	"WATAUGA",
	"WAYNE",
	"WILKES",
	"WILSON",
	"YADKIN",
	"YANCEY",
}

const (
	countyID        int = iota //                County identification number
	countyDesc                 //            varchar(20)        County name
	voterRegNum                //          char(12)           Voter registration number (unique to county)
	electionLabel              //           char(10)           Election date label
	electionDesc               //          varchar(230)       Election description
	votingMethod               //          varchar(10)        Voting method
	votedPartyCode             //         varchar(3)         Voted party code
	votedPartyDesc             //       varchar(60)        Voted party name
	pctLabel                   //              varchar(6)         Precinct code label
	pctDescription             //        varchar(60)        Precinct name
	ncid                       //                   varchar(12)        NCID (North Carolina Identification) number
	votedCountyID              //        varchar(3)         The county id number in which the voter voted in the election (see election_lbl)
	votedCountyDesc            //      varchar(60)        The county name in which the voter voted in the election (see election_lbl)
	vtdLabel                   //              varchar(6)         Voter tabulation district label
	vtdDescription             //        varchar(60)        Voter tabulation district name
)

const (
	PopFips int = iota
	PopCounty
	PopRegion
	PopCOG
	PopMSA
	PopYear
	PopRace
	PopSex
	Popage0to2
	Popage3to4
	Popage5
	Popage6to9
	Popage10to13
	Popage14
	Popage15
	Popage16to17
	Popage18to19
	Popage20to24
	Popage25to34
	Popage35to44
	Popage45to54
	Popage55to59
	Popage60to64
	Popage65to74
	Popage75to84
	Popage85to99
	Popage100
	Poptotal
	Popmedage
	Popage0to17
	Popage18to24
	Popage25to44
	Popage45to64
	Popage65plus
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
	// Cong_dist_abbrv                     //			char  4         Congressional district abbreviation
	// Super_court_abbrv                   //		char  4         Supreme Court abbreviation
	// Judic_dist_abbrv                    //		char  4         Judicial district abbreviation
	// NC_senate_abbrv                     //			char  4         NC Senate district abbreviation
	// NC_house_abbrv                      //			char  4         NC House district abbreviation
	Cancellation_dt time.Time //			char 10         Cancellation date
	Vtd_abbrv       string    //			char  6         Voter tabuluation district abbreviation
	Vtd_desc        string    //			char 30         Voter tabuluation district name
	Age_group       string    //			char 35         Age group range
}

// output: election, ethnicity, age_group, gender, party_affil, method, count

type Stat struct {
	residents  int
	registered int
	voted      int
}

var buckets map[string]*Stat = make(map[string]*Stat)

var voters map[string]Voter = make(map[string]Voter)

var elections map[string]bool = make(map[string]bool)

func getBucketNameHistory(pieces []string) string {
	if voter, ok := voters[strings.Trim(pieces[ncid], "\"")]; ok {
		ageBucket := "BADAGE"
		for k, v := range agebuckets {
			if voter.Age >= v[1] && voter.Age <= v[2] {
				ageBucket = k
			}
		}
		return fmt.Sprintf("%s_%s_%s_%c%s",
			strings.Trim(pieces[electionDesc], "\""),
			strings.Trim(pieces[countyDesc], "\""),
			voter.Race_desc,
			voter.Sex_code,
			ageBucket)
	}
	return "novoter"
}

func getBucketNamePopulation(pieces []string) string {
	var race string
	switch pieces[PopRace] {
	case "aian":
		race = "INDIAN AMERICAN or ALASKA NATIVE"
		break
	case "asian":
		race = "ASIAN"
		break
	case "black":
		race = "BLACK or AFRICAN AMERICAN"
		break
	case "other":
		race = "OTHER"
		break
	case "white":
		race = "WHITE"
		break
	default:
		race = strings.ToLower(pieces[PopRace])
		break
	}
	var sex byte
	switch pieces[PopSex] {
	case "female":
		sex = 'F'
		break
	case "male":
		sex = 'M'
		break
	default:
		sex = 'U'
		break
	}
	return fmt.Sprintf("%s_%s_%c",
		strings.ToUpper(pieces[PopCounty]),
		race,
		sex)
}

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
		voter := Voter{
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
		// fmt.Printf("Added voter %s\n", voter.Ncid)
	}
	fmt.Printf("Loaded VRDB in %s...\n", time.Now().Sub(start))
}

var electRegex = regexp.MustCompile(`^\d+/\d+/\d+\s+(CONGRESSIONAL\s+)?(GENERAL|PRIMARY)$`)

func processVoterHistory() {
	theFile, err := os.Open("ncvhis_Statewide.txt")
	if err != nil {
		fmt.Printf("Error reading voting history database: %s\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(theFile)

	scanner.Scan() // skip the header row

	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		pieces := strings.Split(line, "\t")

		// see if we know about this election already
		elec := strings.Trim(pieces[electionDesc], "\"")
		if !electRegex.MatchString(elec) {
			continue
		}
		if _, ok := elections[elec]; !ok {
			elections[elec] = true
		}

		bucket := getBucketNameHistory(pieces)
		if _, ok := buckets[bucket]; !ok {
			// fmt.Printf("Created bucket %s\n", bucket)
			buckets[bucket] = new(Stat)
			buckets[bucket].residents = 0
			buckets[bucket].registered = 0
			buckets[bucket].voted = 0
		}
		buckets[bucket].voted++

		count++
		if count%1000000 == 0 {
			fmt.Printf("Finished %d rows of history...\n", count)
		}
	}
}

var agebuckets = map[string][]int{
	"_1819": {Popage18to19, 18, 19},
	"_2024": {Popage20to24, 20, 24},
	"_2534": {Popage25to34, 25, 34},
	"_3544": {Popage35to44, 35, 44},
	"_4554": {Popage45to54, 45, 54},
	"_5559": {Popage55to59, 55, 59},
	"_6064": {Popage60to64, 60, 64},
	"_6574": {Popage65to74, 65, 74},
	"_7584": {Popage75to84, 75, 84},
	"_8599": {Popage85to99, 85, 99},
	"_100":  {Popage100, 100, 1000},
}

func processPopulationData() {
	theFile, err := os.Open("NCprojectionsbyagegrp2019.csv")
	if err != nil {
		fmt.Printf("Error reading voting history database: %s\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(theFile)

	scanner.Scan() // skip the header row

	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		pieces := strings.Split(line, ",")

		if pieces[PopSex] == "Total" {
			continue // this is redundant data
		}

		if pieces[PopRace] == "Total" {
			continue // this is redundant data
		}

		bucketFragment := getBucketNamePopulation(pieces)
		rowYear := pieces[PopYear]

		// loop through the elections and add the data
		for election := range elections {
			elecYear := election[6:10]
			if rowYear != elecYear {
				continue
			}

			prefix := election + "_" + bucketFragment

			// loop through all the possible age buckets
			for k, v := range agebuckets {
				// only mess with the bucket if we have actual data
				if pieces[v[0]] != "" {
					bucket := prefix + k
					if _, ok := buckets[bucket]; !ok {
						// fmt.Printf("WARNING: no bucket found for popdata %s\n", bucket)
						buckets[bucket] = new(Stat)
						buckets[bucket].residents = 0
						buckets[bucket].registered = 0
						buckets[bucket].voted = 0
					}
					data := buckets[bucket]

					count, err := strconv.Atoi(pieces[v[0]])
					if err != nil {
						fmt.Printf("WARNING: invalid count %s for '%s'\n", pieces[Popage18to19], line)
					}
					data.residents += count

				}
			}
		}

		count++
		if count%1000000 == 0 {
			fmt.Printf("Finished %d rows of popdata...\n", count)
		}
	}
}

func processRegisteredVoters() {
	count := 0
	for _, voter := range voters {
		ageBucket := "BADAGE"
		for k, v := range agebuckets {
			if voter.Age >= v[1] && voter.Age <= v[2] {
				ageBucket = k
			}
		}

		for election := range elections {
			bucket := fmt.Sprintf("%s_%s_%s_%c%s",
				election,
				counties[voter.County_id],
				voter.Race_desc,
				voter.Sex_code,
				ageBucket)

			if _, ok := buckets[bucket]; !ok {
				// fmt.Printf("WARNING: no bucket found for regdata %s\n", bucket)
				buckets[bucket] = new(Stat)
				buckets[bucket].residents = 0
				buckets[bucket].registered = 0
				buckets[bucket].voted = 0
			}
			buckets[bucket].registered++
		}
		count++
		if count%1000000 == 0 {
			fmt.Printf("Finished %d rows of regdata...\n", count)
		}
	}
}

func main() {
	outFile, err := os.OpenFile("graph2.csv", os.O_CREATE+os.O_WRONLY, 0644)
	defer outFile.Close()
	if err != nil {
		fmt.Printf("Error opening CSV for writing results: %s\n", err)
		os.Exit(1)
	}

	loadVoterDatabase()

	processVoterHistory()

	processPopulationData()

	processRegisteredVoters()

	// dump it to CSV

	writer := csv.NewWriter(outFile)
	for k, v := range buckets {
		// fmt.Printf("Working on bucket %s\n", k)
		pieces := strings.Split(k, "_")
		writer.Write(append(pieces, strconv.Itoa(v.residents), strconv.Itoa(v.registered), strconv.Itoa(v.voted)))
	}
}
