package hcip2

// import "fmt"

// type Column int

const (
	Snapshot_dt              int = iota //			char 10         Date of snapshot
	County_id                           //			char  3         County identification number
	County_desc                         //			char 15         County description
	Voter_reg_num                       //			char 12         Voter registration number (unique by county)
	Ncid                                //				char 12         North Carolina identification number (NCID) of voter
	Status_cd                           //			char  1	        Status code for voter registration
	Voter_status_desc                   //		char 10         Satus code descriptions.
	Reason_cd                           //			char  2         Reason code for voter registration status
	Voter_status_reason_desc            //	char 60         Reason code description
	Absent_ind                          //			char  1         <not used>
	Name_prefx_cd                       //			char  4         <not used>
	Last_name                           //			char 25         Voter last name
	First_name                          //			char 20         Voter first name
	Midl_name                           //			char 20         Voter middle name
	Name_sufx_cd                        //			char  4         Voter name suffix
	House_num                           //			char 10         Residential address street number
	Half_code                           //			char  1         Residential address street number half code
	Street_dir                          //			char  2         Residential address street direction (N,S,E,W,NE,SW, etc.)
	Street_name                         //			char 30         Residential address street name
	Street_type_cd                      //			char  4         Residential address street type (RD, ST, DR, BLVD, etc.)
	Street_sufx_cd                      //			char  4         Residential address street suffix (BUS, EXT, and directional)
	Unit_designator                     //			char  4         <not used>
	Unit_num                            //			char  7         Residential address unit number
	Res_city_desc                       //			char 20         Residential address city name
	State_cd                            //			char  2         Residential address state code
	Zip_code                            //			char  9         Residential address zip code
	Mail_addr1                          //			char 40         Mailing street address
	Mail_addr2                          //			char 40         Mailing address line two
	Mail_addr3                          //			char 40         Mailing address line three
	Mail_addr4                          //			char 40         Mailing address line four
	Mail_city                           //			char 30         Mailing address city name
	Mail_state                          //			char  2         Mailing address state code
	Mail_zipcode                        //			char  9         Mailing address zip code
	Area_cd                             //				char  3         Area code for phone number
	Phone_num                           //			char  7         Telephone number
	Race_code                           //			char  3         Race code
	Race_desc                           //			char 35         Race description
	Ethnic_code                         //			char  2         Ethnicity code
	Ethnic_desc                         //			char 30         Ethnicity description
	Party_cd                            //			char  3         Party affiliation code
	Party_desc                          //			char 12         Party affiliation description
	Sex_code                            //			char  1         Gender code
	Sex                                 //				char  6         Gender description
	Age                                 //				char  3         Age
	Birth_place                         //			char  2         Birth place
	Registr_dt                          //			char 10         Voter registration date
	Precinct_abbrv                      //			char  6         Precinct abbreviation
	Precinct_desc                       //			char 30         Precinct name
	Municipality_abbrv                  //		char  4         Municipality abbreviation
	Municipality_desc                   //		char 30         Municipality name
	Ward_abbrv                          //			char  4         Ward abbreviation
	Ward_desc                           //			char 30         Ward name
	Cong_dist_abbrv                     //			char  4         Congressional district abbreviation
	Cong_dist_desc                      //			char 30         Congressional district name
	Super_court_abbrv                   //		char  4         Supreme Court abbreviation
	Super_court_desc                    //		char 30         Supreme Court name
	Judic_dist_abbrv                    //		char  4         Judicial district abbreviation
	Judic_dist_desc                     //			char 30         Judicial district name
	NC_senate_abbrv                     //			char  4         NC Senate district abbreviation
	NC_senate_desc                      //			char 30         NC Senate district name
	NC_house_abbrv                      //			char  4         NC House district abbreviation
	NC_house_desc                       //			char 30         NC House district name
	County_commiss_abbrv                //		char  4         County Commissioner district abbreviation
	County_commiss_desc                 //		char 30         County Commissioner district name
	Township_abbrv                      //			char  6         Township district abbreviation
	Township_desc                       //			char 30         Township district name
	School_dist_abbrv                   //		char  6         School district abbreviation
	School_dist_desc                    //		char 30         School district name
	Fire_dist_abbrv                     //			char  4         Fire district abbreviation
	Fire_dist_desc                      //			char 30         Fire district name
	Water_dist_abbrv                    //		char  4         Water district abbreviation
	Water_dist_desc                     //			char 30         Water district name
	Sewer_dist_abbrv                    //		char  4         Sewer district abbreviation
	Sewer_dist_desc                     //			char 30         Sewer district name
	Sanit_dist_abbrv                    //		char  4         Sanitation district abbreviation
	Sanit_dist_desc                     //			char 30         Sanitation district name
	Rescue_dist_abbrv                   //		char  4         Rescue district abbreviation
	Rescue_dist_desc                    //		char 30         Rescue district name
	Munic_dist_abbrv                    //		char  4         Municipal district abbreviation
	Munic_dist_desc                     //			char 30         Municipal district name
	Dist_1_abbrv                        //			char  4         Prosecutorial district abbreviation
	Dist_1_desc                         //			char 30         Prosecutorial district name
	Dist_2_abbrv                        //			char  4         <not used>
	Dist_2_desc                         //			char 30         <not used>
	Confidential_ind                    //		char  1         Confidential indicator
	Cancellation_dt                     //			char 10         Cancellation date
	Vtd_abbrv                           //			char  6         Voter tabuluation district abbreviation
	Vtd_desc                            //			char 30         Voter tabuluation district name
	Load_dt                             //				char 10         Data load date
	Age_group                           //			char 35         Age group range
)

var NC HciConfig = HciConfig{
	MaxLineLength:  1500,
	CITY:           Res_city_desc,
	STATE:          State_cd,
	ZIP:            Zip_code,
	STATE_VOTER_ID: Ncid,
	Road:           []int{House_num, Half_code, Street_dir, Street_name, Street_type_cd, Street_sufx_cd, Unit_num},
	RoadNoUnit:     []int{House_num, Half_code, Street_dir, Street_name, Street_type_cd, Street_sufx_cd},
	FilterStr: func(pieces []string) bool {
		// fmt.Println(pieces[Confidential_ind])
		return pieces[Status_cd] != "R" && pieces[Confidential_ind] != "Y" // ignore all the REMOVED and CONFIDENTIAL users
	},
}

const VoterIDLength = 12
