package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/TomOnTime/utfutil"
	"github.com/skemper/hcip2"
)

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

var buckets map[string]int = make(map[string]int)

var voters map[string]Voter = make(map[string]Voter)

// var electRegex = regexp.MustCompile("^\\d+/\\d+/\\d+\\s+(CONGRESSIONAL\\s+)?(GENERAL|PRIMARY),")

func getBucketName(pieces []string) string {
	if voter, ok := voters[strings.Trim(pieces[ncid], "\"")]; ok {
		age := voter.Age / 10
		return fmt.Sprintf("%s_%s_%s_%d_%c_%s_%s_%s",
			strings.Trim(pieces[electionDesc], "\""),
			voter.Race_desc,
			voter.Ethnic_desc,
			age*10,
			voter.Sex_code,
			voter.Party_cd,
			strings.Trim(pieces[votedPartyCode], "\""),
			strings.Trim(pieces[votingMethod], "\""))
	}
	return "novoter"
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

func main() {
	loadVoterDatabase()

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

		// check if this is an election we care about, first
		// if !electRegex.MatchString(pieces[electionLabel]) {
		// 	continue
		// }

		bucket := getBucketName(pieces)
		if _, ok := buckets[bucket]; !ok {
			buckets[bucket] = 0
		}
		buckets[bucket]++
		count++
		if count%1000000 == 0 {
			fmt.Printf("Finished %d rows of history...\n", count)
		}
	}

	// dump it to CSV
	outFile, err := os.OpenFile("graph1.csv", os.O_CREATE+os.O_WRONLY, 0644)
	defer outFile.Close()
	writer := csv.NewWriter(outFile)
	for k, v := range buckets {
		// fmt.Printf("Working on bucket %s\n", k)
		pieces := strings.Split(k, "_")
		writer.Write(append(pieces, strconv.Itoa(v)))
	}
}
