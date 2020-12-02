// Package zillow implements a client for the Zillow api
// http://www.zillow.com/howto/api/APIOverview.htm
package zillow

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/url"
	"strconv"

	"golang.org/x/net/context/ctxhttp"
)

type Zillow interface {
	// Home Valuation
	GetZestimate(context.Context, ZestimateRequest) (*ZestimateResult, error)
	GetSearchResults(context.Context, SearchRequest) (*SearchResults, error)
	GetChart(context.Context, ChartRequest) (*ChartResult, error)
	GetComps(context.Context, CompsRequest) (*CompsResult, error)

	// Property Details
	GetDeepComps(context.Context, CompsRequest) (*DeepCompsResult, error)
	GetDeepSearchResults(context.Context, SearchRequest) (*DeepSearchResults, error)
	GetUpdatedPropertyDetails(ctx context.Context, request UpdatedPropertyDetailsRequest) (*UpdatedPropertyDetails, error)

	// Neighborhood Data
	GetRegionChildren(context.Context, RegionChildrenRequest) (*RegionChildren, error)
	GetRegionChart(context.Context, RegionChartRequest) (*RegionChartResult, error)

	// Mortgage Rates
	GetRateSummary(context.Context, RateSummaryRequest) (*RateSummary, error)

	// Mortgage Calculators
	GetMonthlyPayments(context.Context, MonthlyPaymentsRequest) (*MonthlyPayments, error)
	CalculateMonthlyPaymentsAdvanced(context.Context, MonthlyPaymentsAdvancedRequest) (*MonthlyPaymentsAdvanced, error)
	CalculateAffordability(context.Context, AffordabilityRequest) (*Affordability, error)
}

// New creates a new zillow client.
func New(zwsId string) Zillow {
	return NewExt(zwsId, baseUrl)
}

// NewExt creates a new zillow client.
// It's like New but accepts more options.
func NewExt(zwsId, baseUrl string) Zillow {
	return &zillow{zwsId, baseUrl, http.DefaultClient}
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
	Amount      Value  `xml:"amount"`
	LastUpdated string `xml:"last-updated"`
	// TODO(pedge): fix
	//ValueChange ValueChange `xml:"valueChange"`
	Low        Value  `xml:"valuationRange>low"`
	High       Value  `xml:"valuationRange>high"`
	Percentile string `xml:"percentile"`
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
	RentZestimate   *Zestimate         `xml:"rentzestimate"`
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
	Price       int    `xml:"price"`
	Down        int    `xml:"down"`
	DollarsDown int    `xml:"dollarsdown"`
	Zip         string `xml:"zip"`
}

type Payment struct {
	LoanType                    string  `xml:"loanType,attr"`
	Rate                        float64 `xml:"rate"`
	MonthlyPrincipalAndInterest int     `xml:"monthlyPrincipalAndInterest"`
	MonthlyMortgageInsurance    int     `xml:"monthlyMortgageInsurance"`
}

type MonthlyPayments struct {
	XMLName xml.Name `xml:"paymentsSummary"`

	Request MonthlyPaymentsRequest `xml:"request"`
	Message Message                `xml:"message"`

	Payments               []Payment `xml:"response>payment"`
	DownPayment            int       `xml:"response>downPayment"`
	MonthlyPropertyTaxes   int       `xml:"response>monthlyPropertyTaxes"`
	MonthlyHazardInsurance int       `xml:"response>monthlyHazardInsurance"`
}

type MonthlyPaymentsAdvancedRequest struct {
	Price        int     `xml:"price"`
	Down         int     `xml:"down"`
	Amount       int     `xml:"amount"`
	Rate         float32 `xml:"rate"`
	Schedule     string  `xml:"schedule"`
	TermInMonths int     `xml:"terminmonths"`
	PropertyTax  int     `xml:"propertytax"`
	Hazard       int     `xml:"hazard"`
	PMI          int     `xml:"pmi"`
	HOA          int     `xml:"hoa"`
	Zip          string  `xml:"zip"`
}

type AdvancedPayment struct {
	BeginningBalance int `xml:"beginningbalance"`
	Amount           int `xml:"amount"`
	Principal        int `xml:"principal"`
	Interest         int `xml:"interest"`
	EndingBalance    int `xml:"endingbalance"`
}

type AmortizationSchedule struct {
	Frequency string            `xml:"frequency,attr"`
	Payments  []AdvancedPayment `xml:"payment"`
}

type MonthlyPaymentsAdvanced struct {
	XMLName xml.Name `xml:"paymentsdetails"`

	Request MonthlyPaymentsAdvancedRequest `xml:"request"`
	Message Message                        `xml:"message"`

	MonthlyPrincipalAndInterest int                  `xml:"response>monthlyprincipalandinterest"`
	MonthlyPropertyTaxes        int                  `xml:"response>monthlypropertytaxes"`
	MonthlyHazardInsurance      int                  `xml:"response>monthlyhazardinsurance"`
	MonthlyPMI                  int                  `xml:"response>monthlypmi"`
	MonthlyHOADues              int                  `xml:"response>monthlyhoadues"`
	TotalMonthlyPayment         int                  `xml:"response>totalmonthlypayment"`
	TotalPayments               int                  `xml:"response>totalpayments"`
	TotalInterest               int                  `xml:"response>totalinterest"`
	TotalPrincipal              int                  `xml:"response>totalprincipal"`
	TotalTaxesFeesAndInsurance  int                  `xml:"response>totaltaxesfeesandinsurance"`
	AmortizationSchedule        AmortizationSchedule `xml:"response>amortizationschedule"`
}

type AffordabilityRequest struct {
	AnnualIncome   int     `xml:"annualincome"`
	MonthlyPayment int     `xml:"monthlypayment"`
	Down           int     `xml:"down"`
	MonthlyDebts   int     `xml:"monthlydebts"`
	Rate           float32 `xml:"rate"`
	Schedule       string  `xml:"schedule"`
	TermInMonths   int     `xml:"terminmonths"`
	DebtToIncome   float32 `xml:"debttoincome"`
	IncomeTax      float32 `xml:"incometax"`
	Estimate       bool    `xml:"estimate"`
	PropertyTax    float32 `xml:"propertytax"`
	Hazard         int     `xml:"hazard"`
	PMI            int     `xml:"pmi"`
	HOA            int     `xml:"hoa"`
	Zip            string  `xml:"zip"`
}

type AffordabilityPayment struct {
	Period           int `xml:"period"`
	BeginningBalance int `xml:"beginningbalance"`
	Payment          int `xml:"payment"`
	Principal        int `xml:"principal"`
	Interest         int `xml:"interest"`
	EndingBalance    int `xml:"endingbalance"`
}

type AffordabilityAmortizationSchedule struct {
	Type     string                 `xml:"type,attr"`
	Payments []AffordabilityPayment `xml:"payment"`
}

type Affordability struct {
	XMLName xml.Name `xml:"affordabilitydetails"`

	Request AffordabilityRequest `xml:"request"`
	Message Message              `xml:"message"`

	AffordabilityAmount         int                               `xml:"response>affordabilityamount"`
	MonthlyPrincipalAndInterest int                               `xml:"response>monthlyprincipalandinterest"`
	MonthlyPropertyTaxes        int                               `xml:"response>monthlypropertytaxes"`
	MonthlyHazardInsurance      int                               `xml:"response>monthlyhazardinsurance"`
	MonthlyPMI                  int                               `xml:"response>monthlypmi"`
	MonthlyHOADues              int                               `xml:"response>monthlyhoadues"`
	TotalMonthlyPayment         int                               `xml:"response>totalmonthlypayment"`
	TotalPayments               int                               `xml:"response>totalpayments"`
	TotalInterestPayments       int                               `xml:"response>totalinterestpayments"`
	TotalPrincipal              int                               `xml:"response>totalprincipal"`
	TotalTaxesFeesAndInsurance  int                               `xml:"response>totaltaxesfeesandinsurance"`
	MonthlyIncome               int                               `xml:"response>monthlyincome"`
	MonthlyDebts                int                               `xml:"response>monthlydebts"`
	MonthlyIncomeTax            int                               `xml:"response>monthlyincometax"`
	MonthlyRemainingBudget      int                               `xml:"response>monthlyremainingbudget"`
	AmortizationSchedule        AffordabilityAmortizationSchedule `xml:"response>amortizationschedule"`
}

const baseUrl = "https://www.zillow.com/webservice/"

const (
	zwsIdParam          = "zws-id"
	zpidParam           = "zpid"
	rentzestimateParam  = "rentzestimate"
	addressParam        = "address"
	cityStateZipParam   = "citystatezip"
	unitTypeParam       = "unit-type"
	widthParam          = "width"
	heightParam         = "height"
	chartDurationParam  = "chartDuration"
	countParam          = "count"
	cityParam           = "city"
	stateParam          = "state"
	neighboorhoodParam  = "neightborhood"
	zipParam            = "zip"
	countryParam        = "country"
	childTypeParam      = "childtype"
	regionIdParam       = "regionId"
	priceParam          = "price"
	downParam           = "down"
	dollarsDownParam    = "dollarsdown"
	amountParam         = "amount"
	rateParam           = "rate"
	scheduleParam       = "schedule"
	termInMonthsParam   = "terminmonths"
	propertyTaxParam    = "propertytax"
	hazardParam         = "hazardparam"
	pmiParam            = "pmi"
	hoaParam            = "hoa"
	annualIncomeParam   = "annualincome"
	monthlyPaymentParam = "monthlypayments"
	monthlyDebtsParam   = "monthlydebts"
	debtToIncomeParam   = "debtsinincome"
	incomeTaxParam      = "incometax"
	estimateParam       = "estimate"
)

const (
	zestimatePath               = "GetZestimate"
	searchResultsPath           = "GetSearchResults"
	chartPath                   = "GetChart"
	compsPath                   = "GetComps"
	deepCompsPath               = "GetDeepComps"
	deepSearchPath              = "GetDeepSearchResults"
	updatedPropertyDetailsPath  = "GetUpdatedPropertyDetails"
	regionChildrenPath          = "GetRegionChildren"
	regionChartPath             = "GetRegionChart"
	rateSummaryPath             = "GetRateSummary"
	monthlyPaymentsPath         = "GetMonthlyPayments"
	monthlyPaymentsAdvancedPath = "CalculateMonthlyPaymentsAdvanced"
	affordabilityPath           = "CalculateAffordability"
)

type zillow struct {
	zwsId string
	url   string

	client *http.Client
}

func (z *zillow) get(ctx context.Context, path string, values url.Values, result interface{}) error {
	if resp, err := ctxhttp.Get(ctx, z.client, z.url+"/"+path+".htm?"+values.Encode()); err != nil {
		return err
	} else if err = xml.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}
	return nil
}

func (z *zillow) GetZestimate(ctx context.Context, request ZestimateRequest) (*ZestimateResult, error) {
	values := url.Values{
		zwsIdParam:         {z.zwsId},
		zpidParam:          {request.Zpid},
		rentzestimateParam: {strconv.FormatBool(request.Rentzestimate)},
	}
	var result ZestimateResult
	if err := z.get(ctx, zestimatePath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetSearchResults(ctx context.Context, request SearchRequest) (*SearchResults, error) {
	values := url.Values{
		zwsIdParam:         {z.zwsId},
		addressParam:       {request.Address},
		cityStateZipParam:  {request.CityStateZip},
		rentzestimateParam: {strconv.FormatBool(request.Rentzestimate)},
	}
	var result SearchResults
	if err := z.get(ctx, searchResultsPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetChart(ctx context.Context, request ChartRequest) (*ChartResult, error) {
	values := url.Values{
		zwsIdParam:         {z.zwsId},
		zpidParam:          {request.Zpid},
		unitTypeParam:      {request.UnitType},
		widthParam:         {strconv.Itoa(request.Width)},
		heightParam:        {strconv.Itoa(request.Height)},
		chartDurationParam: {request.Duration},
	}
	var result ChartResult
	if err := z.get(ctx, chartPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetComps(ctx context.Context, request CompsRequest) (*CompsResult, error) {
	values := url.Values{
		zwsIdParam:         {z.zwsId},
		zpidParam:          {request.Zpid},
		countParam:         {strconv.Itoa(request.Count)},
		rentzestimateParam: {strconv.FormatBool(request.Rentzestimate)},
	}
	var result CompsResult
	if err := z.get(ctx, compsPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetDeepComps(ctx context.Context, request CompsRequest) (*DeepCompsResult, error) {
	values := url.Values{
		zwsIdParam:         {z.zwsId},
		zpidParam:          {request.Zpid},
		countParam:         {strconv.Itoa(request.Count)},
		rentzestimateParam: {strconv.FormatBool(request.Rentzestimate)},
	}
	var result DeepCompsResult
	if err := z.get(ctx, deepCompsPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetDeepSearchResults(ctx context.Context, request SearchRequest) (*DeepSearchResults, error) {
	values := url.Values{
		zwsIdParam:         {z.zwsId},
		addressParam:       {request.Address},
		cityStateZipParam:  {request.CityStateZip},
		rentzestimateParam: {strconv.FormatBool(request.Rentzestimate)},
	}
	var result DeepSearchResults
	if err := z.get(ctx, deepSearchPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetUpdatedPropertyDetails(ctx context.Context, request UpdatedPropertyDetailsRequest) (*UpdatedPropertyDetails, error) {
	values := url.Values{
		zwsIdParam: {z.zwsId},
		zpidParam:  {request.Zpid},
	}
	var result UpdatedPropertyDetails
	if err := z.get(ctx, updatedPropertyDetailsPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetRegionChildren(ctx context.Context, request RegionChildrenRequest) (*RegionChildren, error) {
	values := url.Values{
		zwsIdParam:     {z.zwsId},
		regionIdParam:  {request.RegionId},
		stateParam:     {request.State},
		countryParam:   {request.Country},
		cityParam:      {request.City},
		childTypeParam: {request.ChildType},
	}
	var result RegionChildren
	if err := z.get(ctx, regionChildrenPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetRegionChart(ctx context.Context, request RegionChartRequest) (*RegionChartResult, error) {
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
	if err := z.get(ctx, regionChartPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetRateSummary(ctx context.Context, request RateSummaryRequest) (*RateSummary, error) {
	values := url.Values{
		zwsIdParam: {z.zwsId},
		stateParam: {request.State},
	}
	var result RateSummary
	if err := z.get(ctx, rateSummaryPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) GetMonthlyPayments(ctx context.Context, request MonthlyPaymentsRequest) (*MonthlyPayments, error) {
	values := url.Values{
		zwsIdParam:       {z.zwsId},
		priceParam:       {strconv.Itoa(request.Price)},
		downParam:        {strconv.Itoa(request.Down)},
		dollarsDownParam: {strconv.Itoa(request.DollarsDown)},
		zipParam:         {request.Zip},
	}
	var result MonthlyPayments
	if err := z.get(ctx, monthlyPaymentsPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) CalculateMonthlyPaymentsAdvanced(ctx context.Context, request MonthlyPaymentsAdvancedRequest) (*MonthlyPaymentsAdvanced, error) {
	values := url.Values{
		zwsIdParam:        {z.zwsId},
		priceParam:        {strconv.Itoa(request.Price)},
		downParam:         {strconv.Itoa(request.Down)},
		amountParam:       {strconv.Itoa(request.Amount)},
		rateParam:         {strconv.FormatFloat(float64(request.Rate), 'f', -1, 32)},
		scheduleParam:     {request.Schedule},
		termInMonthsParam: {strconv.Itoa(request.TermInMonths)},
		propertyTaxParam:  {strconv.Itoa(request.PropertyTax)},
		hazardParam:       {strconv.Itoa(request.Hazard)},
		pmiParam:          {strconv.Itoa(request.PMI)},
		hoaParam:          {strconv.Itoa(request.HOA)},
		zipParam:          {request.Zip},
	}
	var result MonthlyPaymentsAdvanced
	if err := z.get(ctx, monthlyPaymentsAdvancedPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func (z *zillow) CalculateAffordability(ctx context.Context, request AffordabilityRequest) (*Affordability, error) {
	values := url.Values{
		zwsIdParam:          {z.zwsId},
		annualIncomeParam:   {strconv.Itoa(request.AnnualIncome)},
		monthlyPaymentParam: {strconv.Itoa(request.MonthlyPayment)},
		downParam:           {strconv.Itoa(request.Down)},
		monthlyDebtsParam:   {strconv.Itoa(request.MonthlyDebts)},
		rateParam:           {strconv.FormatFloat(float64(request.Rate), 'f', -1, 32)},
		scheduleParam:       {request.Schedule},
		termInMonthsParam:   {strconv.Itoa(request.TermInMonths)},
		debtToIncomeParam:   {strconv.FormatFloat(float64(request.DebtToIncome), 'f', -1, 32)},
		incomeTaxParam:      {strconv.FormatFloat(float64(request.IncomeTax), 'f', -1, 32)},
		estimateParam:       {strconv.FormatBool(request.Estimate)},
		propertyTaxParam:    {strconv.FormatFloat(float64(request.PropertyTax), 'f', -1, 32)},
		hazardParam:         {strconv.Itoa(request.Hazard)},
		pmiParam:            {strconv.Itoa(request.PMI)},
		hoaParam:            {strconv.Itoa(request.HOA)},
		zipParam:            {request.Zip},
	}
	var result Affordability
	if err := z.get(ctx, affordabilityPath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}
