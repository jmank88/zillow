// Package zillow implements a client for the Zillow api
// http://www.zillow.com/howto/api/APIOverview.htm
package zillow

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"strconv"
)

type Zillow interface {
	// Home Valuation
	GetZestimate(ZestimateRequest) (*ZestimateResult, error)
	GetSearchResults(SearchRequest) (*SearchResults, error)
	GetChart(ChartRequest) (*ChartResult, error)
	GetComps(CompsRequest) (*CompsResult, error)

	// Property Details
	GetDeepComps(CompsRequest) (*DeepCompsResult, error)
	GetDeepSearchResults(SearchRequest) (*DeepSearchResults, error)
	GetUpdatedPropertyDetails(request UpdatedPropertyDetailsRequest) (*UpdatedPropertyDetails, error)

	// Neighborhood Data
	GetRegionChildren(RegionChildrenRequest) (*RegionChildren, error)
	GetRegionChart(RegionChartRequest) (*RegionChartResult, error)

	// Mortgage Rates
	GetRateSummary(RateSummaryRequest) (*RateSummary, error)

	// Mortgage Calculators
	GetMonthlyPayments(MonthlyPaymentsRequest) (*MonthlyPayments, error)
	//CalculateMonthlyPaymentsAdvanced()
	//CalculateAffordability()
}

// Creates a new zillow client.
func NewZillow(zwsId string) Zillow {
	return &zillow{zwsId, baseUrl}
}

type Message struct {
	Text         string `xml:"text"`
	Code         int    `xml:"code"`
	LimitWarning bool   `xml:"limit-warning"`
}

type Address struct {
	Street    string `xml:"street"`
	Zipcode   string `xml:"zipcode"`
	City      string `xml:"city"`
	State     string `xml:"state"`
	Latitude  string `xml:"latitude"`
	Longitude string `xml:"longitude"`
}

type Value struct {
	Currency string `xml:"currency,attr"`
	Value    int    `xml:",chardata"`
}

type ValueChange struct {
	Duration int    `xml:"duration,attr"`
	Currency string `xml:"currency,attr"`
	Value    int    `xml:",chardata"`
}

type Zestimate struct {
	Amount      Value       `xml:"amount"`
	LastUpdated string      `xml:"last-updated"`
	ValueChange ValueChange `xml:"valueChange"`
	Low         Value       `xml:"valuationRange>low"`
	High        Value       `xml:"valuationRange>high"`
	Percentile  string      `xml:"percentile"`
}

type ZestimateRequest struct {
	Zpid          string `xml:"zpid"`
	Rentzestimate bool   `xml:"rentzestimate"`
}

type RealEstateRegion struct {
	XMLName xml.Name `xml:"region"`

	ID                  string  `xml:"id,attr"`
	Type                string  `xml:"type,attr"`
	Name                string  `xml:"name,attr"`
	ZIndex              string  `xml:"zindexValue"`
	ZIndexOneYearChange float64 `xml:"zindexOneYearChange"`
	// Links
	Overview       string `xml:"links>overview"`
	ForSaleByOwner string `xml:"links>forSaleByOwner"`
	ForSale        string `xml:"links>forSale"`
}

type Links struct {
	XMLName xml.Name `xml:"links"`

	HomeDetails   string `xml:"homedetails"`
	GraphsAndData string `xml:"graphsanddata"`
	MapThisHome   string `xml:"mapthishome"`
	MyZestimator  string `xml:"myzestimator"`
	Comparables   string `xml:"comparables"`
}

type ZestimateResult struct {
	XMLName xml.Name `xml:"zestimate"`

	Request ZestimateRequest `xml:"request"`
	Message Message          `xml:"message"`

	Links           Links              `xml:"response>links"`
	Address         Address            `xml:"response>address"`
	Zestimate       Zestimate          `xml:"response>zestimate"`
	LocalRealEstate []RealEstateRegion `xml:"response>localRealEstate>region"`

	// Regions
	ZipcodeID string `xml:"response>regions>zipcode-id"`
	CityID    string `xml:"response>regions>city-id"`
	CountyID  string `xml:"response>regions>county-id"`
	StateID   string `xml:"response>regions>state-id"`
}

type SearchRequest struct {
	Address       string `xml:"address"`
	CityStateZip  string `xml:"citystatezip"`
	Rentzestimate bool   `xml:"rentzestimate"`
}

type SearchResults struct {
	XMLName xml.Name `xml:"searchresults"`

	Request SearchRequest `xml:"request"`
	Message Message       `xml:"message"`

	Results []SearchResult `xml:"response>results>result"`
}

type SearchResult struct {
	XMLName xml.Name `xml:"result"`

	Zpid string `xml:"zpid"`

	Links           Links              `xml:"links"`
	Address         Address            `xml:"address"`
	Zestimate       Zestimate          `xml:"zestimate"`
	LocalRealEstate []RealEstateRegion `xml:"localRealEstate>region"`
}

type ChartRequest struct {
	Zpid     string `xml:"zpid"`
	UnitType string `xml:"unit-type"`
	Width    int    `xml:"width"`
	Height   int    `xml:"height"`
	Duration string `xml:"chartDuration"`
}

type ChartResult struct {
	XMLName xml.Name `xml:"chart"`

	Request ChartRequest `xml:"request"`
	Message Message      `xml:"message"`
	Url     string       `xml:"response>url"`
}

type CompsRequest struct {
	Zpid          string `xml:"zpid"`
	Count         int    `xml:"count"`
	Rentzestimate bool   `xml:"rentzestimate"`
}

type Principal struct {
	Zpid      string    `xml:"zpid"`
	Links     Links     `xml:"links"`
	Address   Address   `xml:"address"`
	Zestimate Zestimate `xml:"zestimate"`
}

type Comp struct {
	Score     float64   `xml:"score,attr"`
	Zpid      string    `xml:"zpid"`
	Links     Links     `xml:"links"`
	Address   Address   `xml:"address"`
	Zestimate Zestimate `xml:"zestimate"`
}

type CompsResult struct {
	XMLName xml.Name `xml:"comps"`

	Request CompsRequest `xml:"request"`
	Message Message      `xml:"message"`

	Principal   Principal `xml:"response>properties>principal"`
	Comparables []Comp    `xml:"response>properties>comparables>comp"`
}

type DeepPrincipal struct {
	Zpid             string             `xml:"zpid"`
	Links            Links              `xml:"links"`
	Address          Address            `xml:"address"`
	TaxAssesmentYear int                `xml:"taxAssessmentYear"`
	TaxAssesment     float64            `xml:"taxAssessment"`
	YearBuilt        int                `xml:"yearBuilt"`
	LotSizeSqFt      int                `xml:"lotSizeSqFt"`
	FinishedSqFt     int                `xml:"finishedSqFt"`
	Bathrooms        float64            `xml:"bathrooms"`
	Bedrooms         int                `xml:"bedrooms"`
	LastSoldDate     string             `xml:"lastSoldDate"`
	LastSoldPrice    Value              `xml:"lastSoldPrice"`
	Zestimate        Zestimate          `xml:"zestimate"`
	LocalRealEstate  []RealEstateRegion `xml:"localRealEstate>region"`
}

type DeepComp struct {
	Score            float64   `xml:"score,attr"`
	Zpid             string    `xml:"zpid"`
	Links            Links     `xml:"links"`
	Address          Address   `xml:"address"`
	TaxAssesmentYear int       `xml:"taxAssessmentYear"`
	TaxAssesment     float64   `xml:"taxAssessment"`
	YearBuilt        int       `xml:"yearBuilt"`
	LotSizeSqFt      int       `xml:"lotSizeSqFt"`
	FinishedSqFt     int       `xml:"finishedSqFt"`
	Bathrooms        float64   `xml:"bathrooms"`
	Bedrooms         int       `xml:"bedrooms"`
	LastSoldDate     string    `xml:"lastSoldDate"`
	LastSoldPrice    Value     `xml:"lastSoldPrice"`
	Zestimate        Zestimate `xml:"zestimate"`
}

type DeepCompsResult struct {
	XMLName xml.Name `xml:"comps"`

	Request CompsRequest `xml:"request"`
	Message Message      `xml:"message"`

	Principal   DeepPrincipal `xml:"response>properties>principal"`
	Comparables []DeepComp    `xml:"response>properties>comparables>comp"`
}

type DeepSearchResult struct {
	XMLName xml.Name `xml:"result"`

	Zpid              string             `xml:"zpid"`
	Links             Links              `xml:"links"`
	Address           Address            `xml:"address"`
	FIPSCounty        string             `xml:"FIPScounty"`
	UseCode           string             `xml:"useCode"`
	TaxAssessmentYear int                `xml:"taxAssessmentYear"`
	TaxAssessment     float64            `xml:"taxAssessment"`
	YearBuilt         int                `xml:"yearBuilt"`
	LotSizeSqFt       int                `xml:"lotSizeSqFt"`
	FinishedSqFt      int                `xml:"finishedSqFt"`
	Bathrooms         float64            `xml:"bathrooms"`
	Bedrooms          int                `xml:"bedrooms"`
	LastSoldDate      string             `xml:"lastSoldDate"`
	LastSoldPrice     Value              `xml:"lastSoldPrice"`
	Zestimate         Zestimate          `xml:"zestimate"`
	LocalRealEstate   []RealEstateRegion `xml:"localRealEstate>region"`
}

type DeepSearchResults struct {
	XMLName xml.Name `xml:"searchresults"`

	Request SearchRequest `xml:"request"`
	Message Message       `xml:"message"`

	Results []DeepSearchResult `xml:"response>results>result"`
}

type RegionChartRequest struct {
	City          string `xml:"city"`
	State         string `xml:"state"`
	Neighborhood  string `xml:"neighborhood"`
	Zipcode       string `xml:"zip"`
	UnitType      string `xml:"unit-type"`
	Width         int    `xml:"width"`
	Height        int    `xml:"height"`
	ChartDuration string `xml:"chartDuration"`
}

type RegionChartResult struct {
	XMLName xml.Name `xml:"regionchart"`

	Request RegionChartRequest `xml:"request"`
	Message Message            `xml:"message"`

	Url    string `xml:"response>url"`
	Zindex Value  `xml:"response>zindex"`
}

type UpdatedPropertyDetailsRequest struct {
	Zpid string `xml:"zpid"`
}

type Posting struct {
	Status          string `xml:"status"`
	AgentName       string `xml:"agentName"`
	AgentProfileUrl string `xml:"agentProfileUrl"`
	Brokerage       string `xml:"brokerage"`
	Type            string `xml:"type"`
	LastUpdatedDate string `xml:"lastUpdatedDate"`
	ExternalUrl     string `xml:"externalUrl"`
	MLS             string `xml:"mls"`
}

type Images struct {
	Count int      `xml:"count"`
	Urls  []string `xml:"image>url"`
}

type EditedFacts struct {
	UseCode        string  `xml:"useCode"`
	Bedrooms       int     `xml:"bedrooms"`
	Bathrooms      float64 `xml:"bathrooms"`
	FinishedSqFt   int     `xml:"finishedSqFt"`
	LotSizeSqFt    int     `xml:"lotSizeSqFt"`
	YearBuilt      int     `xml:"yearBuilt"`
	YearUpdated    int     `xml:"yearUpdated"`
	NumFloors      int     `xml:"numFloors"`
	Basement       string  `xml:"basement"`
	Roof           string  `xml:"roof"`
	View           string  `xml:"view"`
	ParkingType    string  `xml:"parkingType"`
	HeatingSources string  `xml:"heatingSources"`
	HeatingSystem  string  `xml:"heatingSystem"`
	Appliances     string  `xml:"appliances"`
	FloorCovering  string  `xml:"floorCovering"`
	Rooms          string  `xml:"rooms"`
}

type UpdatedPropertyDetails struct {
	XMLName xml.Name `xml:"updatedPropertyDetails"`

	Request UpdatedPropertyDetailsRequest `xml:"request"`
	Message Message                       `xml:"message"`

	PageViewCountMonth int `xml:"response>pageViewCount>currentMonth"`
	PageViewCountTotal int `xml:"response>pageViewCount>total"`

	Address Address `xml:"response>address"`

	Posting          Posting `xml:"response>posting"`
	Price            Value   `xml:"response>price"`
	HomeDetailsLink  string  `xml:"response>links>homeDetails"`
	PhotoGalleryLink string  `xml:"response>links>photoGallery"`
	HomeInfoLink     string  `xml:"response>links>homeInfo"`

	Images           Images      `xml:"response>images"`
	EditedFacts      EditedFacts `xml:"response>editedFacts"`
	HomeDescriptions string      `xml:"homeDesription"`
	Neighborhood     string      `xml:"neighborhood"`
	SchoolDistrict   string      `xml:"schoolDistrict"`
	ElementarySchool string      `xml:"elementarySchool"`
	MiddleSchool     string      `xml:"middleSchool"`
}

type RegionChildrenRequest struct {
	RegionId  string `xml:"regionId"`
	State     string `xml:"state"`
	Country   string `xml:"country"`
	City      string `xml:"city"`
	ChildType string `xml:"childtype"`
}

type Region struct {
	Id        string `xml:"id"`
	Name      string `xml:"name"`
	Country   string `xml:"country"`
	State     string `xml:"state"`
	County    string `xml:"county"`
	City      string `xml:"city"`
	CityUrl   string `xml:"cityurl"`
	Latitude  string `xml:"latitude"`
	Longitude string `xml:"longitude"`
	ZIndex    Value  `xml:"zindex"`
	Url       string `xml:"url"`
}

type RegionChildren struct {
	XMLName xml.Name `xml:"regionchildren"`

	Request RegionChildrenRequest `xml:"request"`
	Message Message               `xml:"message"`

	Region        Region   `xml:"response>region"`
	SubRegionType string   `xml:"response>subregiontype"`
	Regions       []Region `xml:"response>list>region"`
}

type RateSummaryRequest struct {
	State string `xml:"state"`
}

type Rate struct {
	LoanType string  `xml:"loanType,attr"`
	Count    int     `xml:"count,attr"`
	Value    float64 `xml:",chardata"`
}

type RateSummary struct {
	XMLName xml.Name `xml:"rateSummary"`

	Request RateSummaryRequest `xml:"request"`
	Message Message            `xml:"message"`

	Today    []Rate `xml:"response>today>rate"`
	LastWeek []Rate `xml:"response>lastWeek>rate"`
}

type MonthlyPaymentsRequest struct {
	Price int `xml:"price"`
	Down int `xml:"down"`
	DollarsDown int `xml:"dollarsdown"`
	Zip string `xml:"zip"`
}

type Payment struct {
	LoanType string `xml:"loanType,attr"`
	Rate float64 `xml:"rate"`
	MonthlyPrincipalAndInterest int `xml:"monthlyPrincipalAndInterest"`
	MonthlyMortgageInsurance int `xml:"monthlyMortgageInsurance"`
}

type MonthlyPayments struct {
	XMLName xml.Name `xml:"paymentsSummary"`

	Request MonthlyPaymentsRequest `xml:"request"`
	Message Message            `xml:"message"`

	Payments []Payment `xml:"response>payment"`
	DownPayment int `xml:"response>downPayment"`
	MonthlyPropertyTaxes int `xml:"response>monthlyPropertyTaxes"`
	MonthlyHazardInsurance int `xml:"response>monthlyHazardInsurance"`

}

const baseUrl = "http://www.zillow.com/webservice/"

const (
	zwsIdParam         = "zws-Id"
	zpidParam          = "zpid"
	rentzestimateParam = "rentzestimate"
	addressParam       = "address"
	cityStateZipParam  = "citystatezip"
	unitTypeParam      = "unit-type"
	widthParam         = "width"
	heightParam        = "height"
	chartDurationParam = "chartDuration"
	countParam         = "count"
	cityParam          = "city"
	stateParam         = "state"
	neighboorhoodParam = "neightborhood"
	zipParam           = "zip"
	countryParam       = "country"
	childTypeParam     = "childtype"
	regionIdParam      = "regionId"
	priceParam = "price"
	downParam = "down"
	dollarsDownParam = "dollarsdown"
)

const (
	zestimatePath              = "Zestimate"
	searchResultsPath          = "SearchResults"
	chartPath                  = "Chart"
	compsPath                  = "Comps"
	deepCompsPath              = "DeepComps"
	deepSearchPath             = "DeepSearchResults"
	updatedPropertyDetailsPath = "UpdatedPropertyDetails"
	regionChildrenPath         = "RegionChildren"
	regionChartPath            = "RegionChart"
	rateSummaryPath            = "RateSummary"
	monthlyPaymentsPath            = "MonthlyPayments"
	//TODO other services
)

type zillow struct {
	zwsId string
	url   string
}

func (z *zillow) get(path string, values url.Values, result interface{}) error {
	if resp, err := http.Get(z.url + "/Get" + path + ".htm?" + values.Encode()); err != nil {
		return err
	} else if err = xml.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}
	return nil
}

func (z *zillow) GetZestimate(request ZestimateRequest) (*ZestimateResult, error) {
	values := url.Values{
		zwsIdParam:         {z.zwsId},
		zpidParam:          {request.Zpid},
		rentzestimateParam: {strconv.FormatBool(request.Rentzestimate)},
	}
	var result ZestimateResult
	if err := z.get(zestimatePath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetSearchResults(request SearchRequest) (*SearchResults, error) {
	values := url.Values{
		zwsIdParam:         {z.zwsId},
		addressParam:       {request.Address},
		cityStateZipParam:  {request.CityStateZip},
		rentzestimateParam: {strconv.FormatBool(request.Rentzestimate)},
	}
	var result SearchResults
	if err := z.get(searchResultsPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetChart(request ChartRequest) (*ChartResult, error) {
	values := url.Values{
		zwsIdParam:         {z.zwsId},
		zpidParam:          {request.Zpid},
		unitTypeParam:      {request.UnitType},
		widthParam:         {strconv.Itoa(request.Width)},
		heightParam:        {strconv.Itoa(request.Height)},
		chartDurationParam: {request.Duration},
	}
	var result ChartResult
	if err := z.get(chartPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetComps(request CompsRequest) (*CompsResult, error) {
	values := url.Values{
		zwsIdParam:         {z.zwsId},
		zpidParam:          {request.Zpid},
		countParam:         {strconv.Itoa(request.Count)},
		rentzestimateParam: {strconv.FormatBool(request.Rentzestimate)},
	}
	var result CompsResult
	if err := z.get(compsPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetDeepComps(request CompsRequest) (*DeepCompsResult, error) {
	values := url.Values{
		zwsIdParam:         {z.zwsId},
		zpidParam:          {request.Zpid},
		countParam:         {strconv.Itoa(request.Count)},
		rentzestimateParam: {strconv.FormatBool(request.Rentzestimate)},
	}
	var result DeepCompsResult
	if err := z.get(deepCompsPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetDeepSearchResults(request SearchRequest) (*DeepSearchResults, error) {
	values := url.Values{
		zwsIdParam:         {z.zwsId},
		addressParam:       {request.Address},
		cityStateZipParam:  {request.CityStateZip},
		rentzestimateParam: {strconv.FormatBool(request.Rentzestimate)},
	}
	var result DeepSearchResults
	if err := z.get(deepSearchPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetUpdatedPropertyDetails(request UpdatedPropertyDetailsRequest) (*UpdatedPropertyDetails, error) {
	values := url.Values{
		zwsIdParam: {z.zwsId},
		zpidParam:  {request.Zpid},
	}
	var result UpdatedPropertyDetails
	if err := z.get(updatedPropertyDetailsPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetRegionChildren(request RegionChildrenRequest) (*RegionChildren, error) {
	values := url.Values{
		zwsIdParam:     {z.zwsId},
		regionIdParam:  {request.RegionId},
		stateParam:     {request.State},
		countryParam:   {request.Country},
		cityParam:      {request.City},
		childTypeParam: {request.ChildType},
	}
	var result RegionChildren
	if err := z.get(regionChildrenPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetRegionChart(request RegionChartRequest) (*RegionChartResult, error) {
	values := url.Values{
		zwsIdParam:         {z.zwsId},
		cityParam:          {request.City},
		stateParam:         {request.State},
		neighboorhoodParam: {request.Neighborhood},
		zipParam:           {request.Zipcode},
		unitTypeParam:      {request.UnitType},
		widthParam:         {strconv.Itoa(request.Width)},
		heightParam:        {strconv.Itoa(request.Height)},
		chartDurationParam: {request.ChartDuration},
	}
	var result RegionChartResult
	if err := z.get(regionChartPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetRateSummary(request RateSummaryRequest) (*RateSummary, error) {
	values := url.Values{
		zwsIdParam: {z.zwsId},
		stateParam: {request.State},
	}
	var result RateSummary
	if err := z.get(rateSummaryPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetMonthlyPayments(request MonthlyPaymentsRequest) (*MonthlyPayments, error) {
	values := url.Values{
		zwsIdParam: {z.zwsId},
		priceParam: {strconv.Itoa(request.Price)},
		downParam: {strconv.Itoa(request.Down)},
		dollarsDownParam: {strconv.Itoa(request.DollarsDown)},
		zipParam: {request.Zip},
	}
	var result MonthlyPayments
	if err := z.get(monthlyPaymentsPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}
