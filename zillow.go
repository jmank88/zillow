// Package zillow implements a client for the Zillow api
package zillow

import (
	"encoding/xml"
	"net/url"
	"net/http"
	"strconv"
)

type Zillow interface {
	// Home Valuation
	GetZestimate(request ZestimateRequest) (*ZestimateResult, error)
	//GetSearchResults()
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
	Code int	`xml:"code"`
}

type Address struct {
	Street string `xml:"street"`
	Zipcode string `xml:"zipcode"`
	City string `xml:"city"`
	State string `xml:"state"`
	Latitude float64 `xml:"latitude"`
	Longitude float64 `xml:"longitude"`
}

type Value struct {
	Currency string `xml:"currency,attr"`
	Value int `xml:",chardata"`
}

type ValueChange struct {
	Duration int `xml:"duration,attr"`
	Currency string `xml:"currency,attr"`
	Value int `xml:",chardata"`
}

type Zestimate struct {
	Amount Value `xml:"amount"`
	LastUpdated string `xml:"last-updated"`
	ValueChange ValueChange `xml:"valueChange"`
	Low Value `xml:"valuationRange>low"`
	High Value `xml:"valuationRange>high"`
	Percentile int `xml:"percentile"`
}

type ZestimateRequest struct {
	Zpid string	`xml:"zpid"`
	Rentzestimate bool `xml:"rentzestimate"`
}

type Region struct {
	ID string `xml:"id,attr"`
	Type string `xml:"type,attr"`
	Name string `xml:"name,attr"`
	ZIndex string `xml:"zindexValue"`
	ZIndexOneYearChange float64 `xml:"zindexOneYearChange"`
	// Links
	Overview string `xml:"links>overview"`
	ForSaleByOwner string `xml:"links>forSaleByOwner"`
	ForSale string `xml:"links>forSale"`
}

type ZestimateResult struct {
	XMLName xml.Name `xml:"zestimate"`

	Request ZestimateRequest `xml:"request"`
	Message Message	`xml:"message"`

	// Links
	HomeDetails string `xml:"response>links>homedetails"`
	GraphsAndData string `xml:"response>links>graphsanddata"`
	MapThisHome string `xml:"response>links>mapthishome"`
	Comparables string `xml:"response>links>comparables"`

	Address Address `xml:"response>address"`

	Zestimate Zestimate `xml:"response>zestimate"`

	LocalRealEstate []Region `xml:"response>localRealEstate>region"`

	// Regions
	ZipcodeID string `xml:"response>regions>zipcode-id"`
	CityID string `xml:"response>regions>city-id"`
	CountyID string `xml:"response>regions>county-id"`
	StateID string `xml:"response>regions>state-id"`
}

const baseUrl = "http://www.zillow.com/webservice/"

const (
	zwsIdParam = "zws-Id"
	zpidParam = "zpid"
	rentzestimateParam = "rentzestimate"
)

const (
	getZestimatePath = "GetZestimate"
	//TODO other services
)

type zillow struct {
	zwsId string
	url string
}

func (z *zillow) get(servicePath string, values url.Values, result interface{}) error {
	if resp, err := http.Get(z.url + "/" + getZestimatePath + ".htm?" + values.Encode()); err != nil {
		return err
	} else if err = xml.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}
	return nil
}

func (z *zillow) GetZestimate(request ZestimateRequest) (*ZestimateResult, error)  {
	values := url.Values{
		zwsIdParam:{z.zwsId},
		zpidParam:{request.Zpid},
		rentzestimateParam:{strconv.FormatBool(request.Rentzestimate)},
	}
	var result ZestimateResult
	if err := z.get(getZestimatePath, values, &result); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}