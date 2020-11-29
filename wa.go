package hcip2

// StateVoterID|FName|MName|LName|NameSuffix|birthdate|Gender|RegStNum|RegStFrac|RegStName|RegStType|RegUnitType|RegStPreDirection|RegStPostDirection|RegStUnitNum|RegCity|RegState|RegZipCode|CountyCode|PrecinctCode|PrecinctPart|LegislativeDistrict|CongressionalDistrict|Mail1|Mail2|Mail3|Mail4|MailCity|MailZip|MailState|MailCountry|Registrationdate|AbsenteeType|LastVoted|StatusCode

// type Column int

const (
	StateVoterID int = iota
	FirstName
	MiddleName
	LastName
	NameSuffix
	Birthdate
	Gender
	StreetNum
	StreetFrac
	StreetName
	StreetType
	UnitType
	PreDirection
	PostDirection
	UnitNum
	City
	State
	Zip
	County
	PrecinctCode
	PrecinctPart
	LegDistrict
	CongDistrict
	Mail1
	Mail2
	Mail3
	Mail4
	MailCity
	MailZip
	MailState
	MailCountry
	RegDate
	AbsenteeType
	LastVoted
	StatusCode
)

var WA HciConfig = HciConfig{
	MaxLineLength:  1000,
	CITY:           City,
	STATE:          State,
	ZIP:            Zip,
	STATE_VOTER_ID: StateVoterID,
	Road:           []int{StreetNum, StreetFrac, PreDirection, StreetName, StreetType, PostDirection, UnitType, UnitNum},
	RoadNoUnit:     []int{StreetNum, StreetFrac, PreDirection, StreetName, StreetType, PostDirection},
	FilterBytes:    NopFilterBytes,
}
