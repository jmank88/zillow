package zillow

import (
	"encoding/xml"
	"github.com/kr/pretty"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

const (
	testZwsId = "test-id"

	zpid         = "48749425"
	address      = "2114 Bigelow Ave"
	citystatezip = "Seattle, WA"
	unitType     = "percent"
	width        = 300
	height       = 150
)

func assertOnlyParam(t *testing.T, values url.Values, param, expected string) {
	if len(values[param]) != 1 {
		t.Fatalf("expected single %q param", param)
	}
	if actual := values.Get(param); actual != expected {
		t.Fatalf("expected %q %q param but got %q", expected, param, actual)
	}
}

func testFixtures(t *testing.T, expectedPath string, validateQuery func(url.Values)) (*httptest.Server, Zillow) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, expectedPath+".htm") {
			t.Errorf("expected path %q to end with %q", r.URL.Path, expectedPath)
		}
		values := r.URL.Query()
		assertOnlyParam(t, values, zwsIdParam, testZwsId)
		validateQuery(values)

		if f, err := os.Open("testdata/" + expectedPath + ".xml"); err != nil {
			t.Fatal(err)
		} else if _, err := io.Copy(w, f); err != nil {
			t.Fatal(err)
		}
	}))
	return ts, &zillow{zwsId: testZwsId, url: ts.URL}
}

func TestGetZestimate(t *testing.T) {
	server, zillow := testFixtures(t, getZestimatePath, func(values url.Values) {
		assertOnlyParam(t, values, zpidParam, zpid)
		assertOnlyParam(t, values, rentzestimateParam, "false")
	})
	defer server.Close()

	request := ZestimateRequest{Zpid: zpid}
	result, err := zillow.GetZestimate(request)
	if err != nil {
		t.Fatal(err)
	}
	expected := &ZestimateResult{
		XMLName: xml.Name{Space: "Zestimate", Local: "zestimate"},
		Request: request,
		Message: Message{
			Text: "Request successfully processed",
			Code: 0,
		},
		Links: Links{
			XMLName:       xml.Name{Local: "links"},
			HomeDetails:   "http://www.zillow.com/homedetails/2114-Bigelow-Ave-N-Seattle-WA-98109/48749425_zpid/",
			GraphsAndData: "http://www.zillow.com/homedetails/charts/48749425_zpid,1year_chartDuration/?cbt=2950402095890968938%7E4%7ECh-lwa20e2Scegkf_Ev1dsQ2hJD7f74f1dovt2o0BMi2IuvfsZN-sg**",
			MapThisHome:   "http://www.zillow.com/homes/map/48749425_zpid/",
			Comparables:   "http://www.zillow.com/homes/comps/48749425_zpid/",
		},
		Address: Address{
			Street:    "2114 Bigelow Ave N",
			Zipcode:   "98109",
			City:      "Seattle",
			State:     "WA",
			Latitude:  47.63793,
			Longitude: -122.347936,
		},
		Zestimate: Zestimate{
			Amount:      Value{Currency: "USD", Value: 1219500},
			LastUpdated: "11/03/2009",
			ValueChange: ValueChange{Duration: 30, Currency: "USD", Value: -41500},
			Percentile:  95,
			Low:         Value{Currency: "USD", Value: 1024380},
			High:        Value{Currency: "USD", Value: 1378035},
		},
		LocalRealEstate: []Region{
			Region{
				XMLName:             xml.Name{Local: "region"},
				ID:                  "271856",
				Type:                "neighborhood",
				Name:                "East Queen Anne",
				ZIndex:              "525,397",
				ZIndexOneYearChange: -0.144,
				Overview:            "http://www.zillow.com/local-info/WA-Seattle/East-Queen-Anne/r_271856/",
				ForSaleByOwner:      "http://www.zillow.com/homes/fsbo/East-Queen-Anne-Seattle-WA/",
				ForSale:             "http://www.zillow.com/east-queen-anne-seattle-wa/",
			},
			Region{
				XMLName:             xml.Name{Local: "region"},
				ID:                  "16037",
				Type:                "city",
				Name:                "Seattle",
				ZIndex:              "381,764",
				ZIndexOneYearChange: -0.074,
				Overview:            "http://www.zillow.com/local-info/WA-Seattle/r_16037/",
				ForSaleByOwner:      "http://www.zillow.com/homes/fsbo/Seattle-WA/",
				ForSale:             "http://www.zillow.com/seattle-wa/",
			},
			Region{
				XMLName:             xml.Name{Local: "region"},
				ID:                  "59",
				Type:                "state",
				Name:                "Washington",
				ZIndex:              "263,278",
				ZIndexOneYearChange: -0.066,
				Overview:            "http://www.zillow.com/local-info/WA-home-value/r_59/",
				ForSaleByOwner:      "http://www.zillow.com/homes/fsbo/WA/",
				ForSale:             "http://www.zillow.com/wa/",
			},
		},
		ZipcodeID: "99569",
		CityID:    "16037",
		CountyID:  "207",
		StateID:   "59",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected:\n %s\n\n but got:\n %s\n\n diff:\n %s\n",
			pretty.Formatter(expected), pretty.Formatter(result), pretty.Diff(expected, result))
	}
}

func TestGetSearchResults(t *testing.T) {
	server, zillow := testFixtures(t, getSearchResults, func(values url.Values) {
		assertOnlyParam(t, values, addressParam, address)
		assertOnlyParam(t, values, cityStateZipParam, citystatezip)
		assertOnlyParam(t, values, rentzestimateParam, "false")
	})
	defer server.Close()

	request := SearchRequest{Address: address, CityStateZip: citystatezip}
	result, err := zillow.GetSearchResults(request)
	if err != nil {
		t.Fatal(err)
	}
	expected := &SearchResults{
		XMLName: xml.Name{Space: "SearchResults", Local: "searchresults"},
		Request: request,
		Message: Message{
			Text: "Request successfully processed",
			Code: 0,
		},
		Results: []SearchResult{
			SearchResult{
				XMLName: xml.Name{Local: "result"},
				Zpid:    "48749425",
				Links: Links{
					XMLName:       xml.Name{Local: "links"},
					HomeDetails:   "http://www.zillow.com/homedetails/2114-Bigelow-Ave-N-Seattle-WA-98109/48749425_zpid/",
					GraphsAndData: "http://www.zillow.com/homedetails/charts/48749425_zpid,1year_chartDuration/?cbt=7522682882544325802%7E9%7EY2EzX18jtvYTCel5PgJtPY1pmDDLxGDZXzsfRy49lJvCnZ4bh7Fi9w**",
					MapThisHome:   "http://www.zillow.com/homes/map/48749425_zpid/",
					Comparables:   "http://www.zillow.com/homes/comps/48749425_zpid/",
				},
				Address: Address{
					Street:    "2114 Bigelow Ave N",
					Zipcode:   "98109",
					City:      "Seattle",
					State:     "WA",
					Latitude:  47.63793,
					Longitude: -122.347936,
				},
				Zestimate: Zestimate{
					Amount:      Value{Currency: "USD", Value: 1219500},
					LastUpdated: "11/03/2009",
					ValueChange: ValueChange{Duration: 30, Currency: "USD", Value: -41500},
					Low:         Value{Currency: "USD", Value: 1024380},
					High:        Value{Currency: "USD", Value: 1378035},
					Percentile:  0,
				},
				LocalRealEstate: []Region{
					Region{
						XMLName:             xml.Name{Local: "region"},
						ID:                  "271856",
						Type:                "neighborhood",
						Name:                "East Queen Anne",
						ZIndex:              "525,397",
						ZIndexOneYearChange: -0.144,
						Overview:            "http://www.zillow.com/local-info/WA-Seattle/East-Queen-Anne/r_271856/",
						ForSaleByOwner:      "http://www.zillow.com/homes/fsbo/East-Queen-Anne-Seattle-WA/",
						ForSale:             "http://www.zillow.com/east-queen-anne-seattle-wa/",
					},
					Region{
						XMLName:             xml.Name{Local: "region"},
						ID:                  "16037",
						Type:                "city",
						Name:                "Seattle",
						ZIndex:              "381,764",
						ZIndexOneYearChange: -0.074,
						Overview:            "http://www.zillow.com/local-info/WA-Seattle/r_16037/",
						ForSaleByOwner:      "http://www.zillow.com/homes/fsbo/Seattle-WA/",
						ForSale:             "http://www.zillow.com/seattle-wa/",
					},
					Region{
						XMLName:             xml.Name{Local: "region"},
						ID:                  "59",
						Type:                "state",
						Name:                "Washington",
						ZIndex:              "263,278",
						ZIndexOneYearChange: -0.066,
						Overview:            "http://www.zillow.com/local-info/WA-home-value/r_59/",
						ForSaleByOwner:      "http://www.zillow.com/homes/fsbo/WA/",
						ForSale:             "http://www.zillow.com/wa/",
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected:\n %s\n\n but got:\n %s\n\n diff:\n %s\n",
			pretty.Formatter(expected), pretty.Formatter(result), pretty.Diff(expected, result))
	}
}

func TestGetChart(t *testing.T) {
	server, zillow := testFixtures(t, getChart, func(values url.Values) {
		assertOnlyParam(t, values, zpidParam, zpid)
		assertOnlyParam(t, values, unitTypeParam, unitType)
		assertOnlyParam(t, values, widthParam, strconv.Itoa(width))
		assertOnlyParam(t, values, heightParam, strconv.Itoa(height))
	})
	defer server.Close()

	request := ChartRequest{Zpid: zpid, UnitType: unitType, Width: width, Height: height}
	result, err := zillow.GetChart(request)
	if err != nil {
		t.Fatal(err)
	}
	expected := &ChartResult{
		XMLName: xml.Name{Space: "http://www.zillowstatic.com/vstatic/8d9b5f1/static/xsd/Chart.xsd", Local: "chart"},
		Request: request,
		Message: Message{
			Text: "Request successfully processed",
			Code: 0,
		},
		Url: "http://www.zillow.com/app?chartDuration=1year&chartType=partner&height=150&page=webservice%2FGetChart&service=chart&showPercent=true&width=300&zpid=48749425",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected:\n %s\n\n but got:\n %s\n\n diff:\n %s\n",
			pretty.Formatter(expected), pretty.Formatter(result), pretty.Diff(expected, result))
	}
}
