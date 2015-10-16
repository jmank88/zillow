// Package zillow implements a client for the Zillow api
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
	//GetChart()
	//GetComps()

	// Property Details
	//GetDeepComps()
	//GetDeepSearchResults()
	//GetUpdatedPropertyDetails()

	// Neighborhood Data
	//GetRegionChildren()
	//GetRegionChart()

	// Mortgage Rates
	//GetRateSummary()

	// Mortgage Calculators
	//GetMonthlyPayments()
	//CalculateMonthlyPaymentsAdvanced()
	//CalculateAffordability()
}

// Creates a new zillow client.
func NewZillow(zwsId string) Zillow {
	return &zillow{zwsId, baseUrl}
}

type Message struct {
	Text string `xml:"text"`
	Code int    `xml:"code"`
	//TODO limit?
}

type Address struct {
	Street    string  `xml:"street"`
	Zipcode   string  `xml:"zipcode"`
	City      string  `xml:"city"`
	State     string  `xml:"state"`
	Latitude  float64 `xml:"latitude"`
	Longitude float64 `xml:"longitude"`
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
	Percentile  int         `xml:"percentile"`
}

type ZestimateRequest struct {
	Zpid          string `xml:"zpid"`
	Rentzestimate bool   `xml:"rentzestimate"`
}

type Region struct {
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
	Comparables   string `xml:"comparables"`
}

type ZestimateResult struct {
	XMLName xml.Name `xml:"zestimate"`

	Request ZestimateRequest `xml:"request"`
	Message Message          `xml:"message"`

	Links           Links     `xml:"response>links"`
	Address         Address   `xml:"response>address"`
	Zestimate       Zestimate `xml:"response>zestimate"`
	LocalRealEstate []Region  `xml:"response>localRealEstate>region"`

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

	Links           Links     `xml:"links"`
	Address         Address   `xml:"address"`
	Zestimate       Zestimate `xml:"zestimate"`
	LocalRealEstate []Region  `xml:"localRealEstate>region"`
}

const baseUrl = "http://www.zillow.com/webservice/"

const (
	zwsIdParam         = "zws-Id"
	zpidParam          = "zpid"
	rentzestimateParam = "rentzestimate"
	addressParam       = "address"
	cityStateZipParam  = "citystatezip"
)

const (
	getZestimatePath = "GetZestimate"
	getSearchResults = "GetSearchResults"
	//TODO other services
)

type zillow struct {
	zwsId string
	url   string
}

func (z *zillow) get(servicePath string, values url.Values, result interface{}) error {
	if resp, err := http.Get(z.url + "/" + servicePath + ".htm?" + values.Encode()); err != nil {
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
	if err := z.get(getZestimatePath, values, &result); err != nil {
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
	if err := z.get(getSearchResults, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}
