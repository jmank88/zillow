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
	count        = 5
	city         = "lacey"
	state        = "WA"
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
	server, zillow := testFixtures(t, zestimatePath, func(values url.Values) {
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
			Latitude:  "47.63793",
			Longitude: "-122.347936",
		},
		Zestimate: Zestimate{
			Amount:      Value{Currency: "USD", Value: 1219500},
			LastUpdated: "11/03/2009",
			ValueChange: ValueChange{Duration: 30, Currency: "USD", Value: -41500},
			Percentile:  "95",
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
		t.Fatalf("expected:\n %#v\n\n but got:\n %#v\n\n diff:\n %s\n",
			pretty.Formatter(expected), pretty.Formatter(result), pretty.Diff(expected, result))
	}
}

func TestGetSearchResults(t *testing.T) {
	server, zillow := testFixtures(t, searchResultsPath, func(values url.Values) {
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
					Latitude:  "47.63793",
					Longitude: "-122.347936",
				},
				Zestimate: Zestimate{
					Amount:      Value{Currency: "USD", Value: 1219500},
					LastUpdated: "11/03/2009",
					ValueChange: ValueChange{Duration: 30, Currency: "USD", Value: -41500},
					Low:         Value{Currency: "USD", Value: 1024380},
					High:        Value{Currency: "USD", Value: 1378035},
					Percentile:  "0",
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
	server, zillow := testFixtures(t, chartPath, func(values url.Values) {
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
		t.Fatalf("expected:\n %#v\n\n but got:\n %#v\n\n diff:\n %s\n",
			pretty.Formatter(expected), pretty.Formatter(result), pretty.Diff(expected, result))
	}
}

func TestGetComps(t *testing.T) {
	server, zillow := testFixtures(t, compsPath, func(values url.Values) {
		assertOnlyParam(t, values, zpidParam, zpid)
		assertOnlyParam(t, values, countParam, strconv.Itoa(count))
		assertOnlyParam(t, values, rentzestimateParam, "false")
	})
	defer server.Close()

	request := CompsRequest{Zpid: zpid, Count: count}
	result, err := zillow.GetComps(request)
	if err != nil {
		t.Fatal(err)
	}
	expected := &CompsResult{
		XMLName: xml.Name{Space: "http://www.zillowstatic.com/vstatic/8d9b5f1/static/xsd/Comps.xsd", Local: "comps"},
		Request: request,
		Message: Message{
			Text: "Request successfully processed",
			Code: 0,
		},
		Principal: Principal{
			Zpid: zpid,
			Links: Links{
				XMLName:       xml.Name{Local: "links"},
				HomeDetails:   "http://www.zillow.com/HomeDetails.htm?city=SEATTLE+&state=WA&zprop=48749425&partner=<ZWSID>",
				GraphsAndData: "http://www.zillow.com/Charts.htm?chartDuration=1year&zpid=48749425&cbt=7604042719451599549%7E5%7E3H0JLxtdY3zX%2F2rM093I6LYKRS2%2FYJQyYaLUNkW54os%3D&partner=<ZWSID>",
				MapThisHome:   "http://www.zillow.com/homes/48749425_zpid&partner=<ZWSID>",
				Comparables:   "http://www.zillow.com/comps/48749425_zpid&partner=<ZWSID>",
			},
			Address: Address{
				Street:    "2114 Bigelow Ave N",
				Zipcode:   "98109",
				City:      "SEATTLE",
				State:     "WA",
				Latitude:  "47.637934",
				Longitude: "-122.347936",
			},
			Zestimate: Zestimate{
				Amount:      Value{Currency: "USD", Value: 1124072},
				LastUpdated: "09/01/2006",
				Low:         Value{Currency: "USD", Value: 966702},
				High:        Value{Currency: "USD", Value: 1236479},
				Percentile:  "93",
			},
		},
		Comparables: []Comp{
			Comp{
				Score: 0.257106811263241,
				Zpid:  "48749459",
				Links: Links{
					XMLName:       xml.Name{Local: "links"},
					HomeDetails:   "http://www.zillow.com/HomeDetails.htm?city=SEATTLE+&state=WA&zprop=48749459&partner=<ZWSID>",
					GraphsAndData: "http://www.zillow.com/Charts.htm?chartDuration=1year&zpid=48749459&cbt=7604042719451599549%7E5%7E3H0JLxtdY3zX%2F2rM093I6LYKRS2%2FYJQyYaLUNkW54os%3D&partner=<ZWSID>",
					MapThisHome:   "http://www.zillow.com/homes/48749459_zpid&partner=<ZWSID>",
					MyZestimator:  "http://www.zillow.com/myzestimator/MyZestimatorHomeFactsPage.htm?context=1158087975250&zprop=48749459&partner=<ZWSID>",
					Comparables:   "http://www.zillow.com/comps/48749459_zpid&partner=<ZWSID>",
				},
				Address: Address{
					Street:    "2021 5th Ave N",
					Zipcode:   "98109",
					City:      "SEATTLE",
					State:     "WA",
					Latitude:  "47.637253",
					Longitude: "-122.347385",
				},
				Zestimate: Zestimate{
					Amount:      Value{Currency: "USD", Value: 985000},
					LastUpdated: "09/01/2006",
					Low:         Value{Currency: "USD", Value: 847100},
					High:        Value{Currency: "USD", Value: 1083500},
				},
			},
			Comp{
				Score: 0.31179534464349695,
				Zpid:  "0.31179534464349695",
				Links: Links{
					XMLName:       xml.Name{Local: "links"},
					HomeDetails:   "http://www.zillow.com/HomeDetails.htm?city=SEATTLE+&state=WA&zprop=48749409&partner=<ZWSID>",
					GraphsAndData: "http://www.zillow.com/Charts.htm?chartDuration=1year&zpid=48749409&cbt=7604042719451599549%7E5%7E3H0JLxtdY3zX%2F2rM093I6LYKRS2%2FYJQyYaLUNkW54os%3D&partner=<ZWSID>",
					MapThisHome:   "http://www.zillow.com/homes/48749409_zpid&partner=<ZWSID>",
					MyZestimator:  "http://www.zillow.com/myzestimator/MyZestimatorHomeFactsPage.htm?context=1158087975250&zprop=48749409&partner=<ZWSID>",
					Comparables:   "http://www.zillow.com/comps/48749409_zpid&partner=<ZWSID>",
				},
				Address: Address{
					Street:    "2208 Bigelow Ave N",
					Zipcode:   "98109",
					City:      "SEATTLE",
					State:     "WA",
					Latitude:  "47.638543",
					Longitude: "-122.348008",
				},
				Zestimate: Zestimate{
					Amount:      Value{Currency: "USD", Value: 1326256},
					LastUpdated: "09/01/2006",
					Low:         Value{Currency: "USD", Value: 1140580},
					High:        Value{Currency: "USD", Value: 1458882},
				},
			},
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected:\n %#v\n\n but got:\n %#v\n\n diff:\n %s\n",
			pretty.Formatter(expected), pretty.Formatter(result), pretty.Diff(expected, result))
	}
}

func TestGetDeepComp(t *testing.T) {
	server, zillow := testFixtures(t, deepCompsPath, func(values url.Values) {
		assertOnlyParam(t, values, zpidParam, zpid)
		assertOnlyParam(t, values, countParam, strconv.Itoa(count))
		assertOnlyParam(t, values, rentzestimateParam, "false")
	})
	defer server.Close()

	request := CompsRequest{Zpid: zpid, Count: count}
	result, err := zillow.GetDeepComps(request)
	if err != nil {
		t.Fatal(err)
	}
	expected := &DeepCompsResult{
		XMLName: xml.Name{Space: "Comps", Local: "comps"},
		Request: request,
		Message: Message{
			Text: "Request successfully processed",
			Code: 0,
		},

		Principal: DeepPrincipal{
			Zpid: "lastSoldPrice",
			Links: Links{
				XMLName:       xml.Name{Local: "links"},
				HomeDetails:   "http://www.zillow.com/homedetails/2114-Bigelow-Ave-N-Seattle-WA-98109/48749425_zpid/",
				GraphsAndData: "http://www.zillow.com/homedetails/charts/48749425_zpid,1year_chartDuration/?cbt=8860375400203215891%7E4%7E4rtHGS99FewWZQdZkxwcJh2zVPQgG28TgCLWpvfp18j0KOoW_noNWg**",
				MapThisHome:   "http://www.zillow.com/homes/map/48749425_zpid/",
				Comparables:   "http://www.zillow.com/homes/comps/48749425_zpid/",
			},
			Address: Address{
				Street:    "2114 Bigelow Ave N",
				Zipcode:   "98109",
				City:      "Seattle",
				State:     "WA",
				Latitude:  "47.63793",
				Longitude: "-122.347936",
			},
			TaxAssesmentYear: 2008,
			TaxAssesment:     1054000.0,
			YearBuilt:        1924,
			LotSizeSqFt:      4680,
			FinishedSqFt:     3470,
			Bathrooms:        3.0,
			Bedrooms:         4,
			LastSoldDate:     "11/26/2008",
			LastSoldPrice:    Value{Currency: "USD", Value: 995000},
			Zestimate: Zestimate{
				Amount:      Value{Currency: "USD", Value: 1219500},
				LastUpdated: "12/31/1969",
				ValueChange: ValueChange{Duration: 30, Currency: "USD", Value: -41500},
				Low:         Value{Currency: "USD", Value: 1024380},
				High:        Value{Currency: "USD", Value: 1378035},
				Percentile:  "95",
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
		Comparables: []DeepComp{
			{
				Score: 0.156502,
				Zpid:  "89210365",
				Links: Links{
					XMLName:       xml.Name{Space: "", Local: "links"},
					HomeDetails:   "http://www.zillow.com/homedetails/1511-10th-Ave-W-Seattle-WA-98119/89210365_zpid/",
					GraphsAndData: "http://www.zillow.com/homedetails/charts/89210365_zpid,1year_chartDuration/?cbt=8860375400203215891%7E4%7E4rtHGS99FewWZQdZkxwcJh2zVPQgG28TgCLWpvfp18j0KOoW_noNWg**",
					MapThisHome:   "http://www.zillow.com/homes/map/89210365_zpid/",
					Comparables:   "http://www.zillow.com/homes/comps/89210365_zpid/",
				},
				Address: Address{
					Street:    "1511 10th Ave W",
					Zipcode:   "98119",
					City:      "Seattle",
					State:     "WA",
					Latitude:  "",
					Longitude: "",
				},
				TaxAssesmentYear: 2008,
				TaxAssesment:     804000,
				YearBuilt:        2006,
				LotSizeSqFt:      3750,
				FinishedSqFt:     2520,
				Bathrooms:        4,
				Bedrooms:         4,
				LastSoldDate:     "09/24/2009",
				LastSoldPrice:    Value{Currency: "USD", Value: 832500},
				Zestimate: Zestimate{
					Amount:      Value{Currency: "USD", Value: 836500},
					LastUpdated: "11/03/2009",
					ValueChange: ValueChange{Duration: 30, Currency: "USD", Value: -220500},
					Low:         Value{Currency: "USD", Value: 777945},
					High:        Value{Currency: "USD", Value: 886690},
					Percentile:  "83",
				},
			},
			{
				Score: 0.156114,
				Zpid:  "49009208",
				Links: Links{
					XMLName:       xml.Name{Space: "", Local: "links"},
					HomeDetails:   "http://www.zillow.com/homedetails/2928-Queen-Anne-Ave-N-Seattle-WA-98109/49009208_zpid/",
					GraphsAndData: "http://www.zillow.com/homedetails/charts/49009208_zpid,1year_chartDuration/?cbt=8860375400203215891%7E4%7E4rtHGS99FewWZQdZkxwcJh2zVPQgG28TgCLWpvfp18j0KOoW_noNWg**",
					MapThisHome:   "http://www.zillow.com/homes/map/49009208_zpid/",
					MyZestimator:  "",
					Comparables:   "http://www.zillow.com/homes/comps/49009208_zpid/",
				},
				Address: Address{
					Street:    "2928 Queen Anne Ave N",
					Zipcode:   "98109",
					City:      "Seattle",
					State:     "WA",
					Latitude:  "47.646643",
					Longitude: "-122.356534",
				},
				TaxAssesmentYear: 2008,
				TaxAssesment:     633000,
				YearBuilt:        1927,
				LotSizeSqFt:      3240,
				FinishedSqFt:     1920,
				Bathrooms:        2,
				Bedrooms:         2,
				LastSoldDate:     "08/20/2009",
				LastSoldPrice:    Value{Currency: "USD", Value: 595000},
				Zestimate: Zestimate{
					Amount:      Value{Currency: "USD", Value: 608000},
					LastUpdated: "11/03/2009",
					ValueChange: ValueChange{Duration: 30, Currency: "USD", Value: 11000},
					Low:         Value{Currency: "USD", Value: 559360},
					High:        Value{Currency: "USD", Value: 656640},
					Percentile:  "68",
				},
			},
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected:\n %#v\n\n but got:\n %#v\n\n diff:\n %s\n",
			pretty.Formatter(expected), pretty.Formatter(result), pretty.Diff(expected, result))
	}
}

func TestGetDeepSearchResults(t *testing.T) {
	server, zillow := testFixtures(t, deepSearchPath, func(values url.Values) {
		assertOnlyParam(t, values, addressParam, address)
		assertOnlyParam(t, values, cityStateZipParam, citystatezip)
		assertOnlyParam(t, values, rentzestimateParam, "false")
	})
	defer server.Close()

	request := SearchRequest{Address: address, CityStateZip: citystatezip}
	result, err := zillow.GetDeepSearchResults(request)
	if err != nil {
		t.Fatal(err)
	}
	expected := &DeepSearchResults{
		XMLName: xml.Name{Space: "SearchResults", Local: "searchresults"},
		Request: request,
		Message: Message{
			Text: "Request successfully processed",
			Code: 0,
		},
		Results: []DeepSearchResult{
			{
				XMLName: xml.Name{Local: "result"},
				Zpid:    "48749425",
				Links: Links{
					XMLName:       xml.Name{Local: "links"},
					HomeDetails:   "http://www.zillow.com/homedetails/2114-Bigelow-Ave-N-Seattle-WA-98109/48749425_zpid/",
					GraphsAndData: "http://www.zillow.com/homedetails/charts/48749425_zpid,1year_chartDuration/?cbt=8224687894635748395%7E7%7EjS-H-hFDCRzaVl6bMy4IjMErWd4OhP23IK8vmp4_m9u_SO1ruBhoCA**",
					MapThisHome:   "http://www.zillow.com/homes/map/48749425_zpid/",
					Comparables:   "http://www.zillow.com/homes/comps/48749425_zpid/",
				},
				Address: Address{
					Street:    "2114 Bigelow Ave N",
					Zipcode:   "98109",
					City:      "Seattle",
					State:     "WA",
					Latitude:  "47.63793",
					Longitude: "-122.347936",
				},
				FIPSCounty:        "33",
				UseCode:           "SingleFamily",
				TaxAssessmentYear: 2008,
				TaxAssessment:     1054000.0,
				YearBuilt:         1924,
				LotSizeSqFt:       4680,
				FinishedSqFt:      3470,
				Bathrooms:         3.0,
				Bedrooms:          4,
				LastSoldDate:      "11/26/2008",
				LastSoldPrice:     Value{Currency: "USD", Value: 995000},
				Zestimate: Zestimate{
					Amount:      Value{Currency: "USD", Value: 1219500},
					LastUpdated: "12/31/1969",
					ValueChange: ValueChange{Duration: 30, Currency: "USD", Value: -41500},
					Low:         Value{Currency: "USD", Value: 1024380},
					High:        Value{Currency: "USD", Value: 1378035},
					Percentile:  "0",
				},
				LocalRealEstate: []Region{
					Region{
						XMLName:        xml.Name{Local: "region"},
						ID:             "271856",
						Type:           "neighborhood",
						Name:           "East Queen Anne",
						ZIndex:         "525,397",
						Overview:       "http://www.zillow.com/local-info/WA-Seattle/East-Queen-Anne/r_271856/",
						ForSaleByOwner: "http://www.zillow.com/homes/fsbo/East-Queen-Anne-Seattle-WA/",
						ForSale:        "http://www.zillow.com/east-queen-anne-seattle-wa/",
					},
					Region{
						XMLName:        xml.Name{Local: "region"},
						ID:             "16037",
						Type:           "city",
						Name:           "Seattle",
						ZIndex:         "381,764",
						Overview:       "http://www.zillow.com/local-info/WA-Seattle/r_16037/",
						ForSaleByOwner: "http://www.zillow.com/homes/fsbo/Seattle-WA/",
						ForSale:        "http://www.zillow.com/seattle-wa/",
					},
					Region{
						XMLName:        xml.Name{Local: "region"},
						ID:             "59",
						Type:           "state",
						Name:           "Washington",
						ZIndex:         "263,278",
						Overview:       "http://www.zillow.com/local-info/WA-home-value/r_59/",
						ForSaleByOwner: "http://www.zillow.com/homes/fsbo/WA/",
						ForSale:        "http://www.zillow.com/wa/",
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

func TestGetRegionChart(t *testing.T) {
	server, zillow := testFixtures(t, regionChartPath, func(values url.Values) {
		assertOnlyParam(t, values, cityParam, city)
		assertOnlyParam(t, values, stateParam, state)
		assertOnlyParam(t, values, unitTypeParam, unitType)
		assertOnlyParam(t, values, widthParam, strconv.Itoa(width))
		assertOnlyParam(t, values, heightParam, strconv.Itoa(height))
	})
	defer server.Close()

	request := RegionChartRequest{
		City:     city,
		State:    state,
		UnitType: unitType,
		Width:    width,
		Height:   height,
	}
	result, err := zillow.GetRegionChart(request)
	if err != nil {
		t.Fatal(err)
	}
	expected := &RegionChartResult{
		XMLName: xml.Name{Space: "http://www.zillow.com/static/xsd/RegionChart.xsd", Local: "regionchart"},
		Request: request,
		Message: Message{
			Text: "Request successfully processed",
			Code: 0,
		},
		Url:    "http://localhost:8080/app?chartDuration=1year&chartType=partner&cityRegionId=5470&countyRegionId=0&height=150&nationRegionId=0&page=webservice%2FGetRegionChart&service=chart&showCity=true&showPercent=true&stateRegionId=0&width=300&zipRegionId=0",
		Zindex: Value{Currency: "USD", Value: 463115},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected:\n %#v\n\n but got:\n %#v\n\n diff:\n %s\n",
			pretty.Formatter(expected), pretty.Formatter(result), pretty.Diff(expected, result))
	}
}

func TestGetUpdatedPropertyDetails(t *testing.T) {
	server, zillow := testFixtures(t, updatedPropertyDetailsPath, func(values url.Values) {
		assertOnlyParam(t, values, zpidParam, zpid)
	})
	defer server.Close()

	request := UpdatedPropertyDetailsRequest{
		Zpid: zpid,
	}
	result, err := zillow.GetUpdatedPropertyDetails(request)
	if err != nil {
		t.Fatal(err)
	}
	expected := &UpdatedPropertyDetails{
		XMLName: xml.Name{Space: "http://www.zillow.com/static/xsd/UpdatedPropertyDetails.xsd", Local: "updatedPropertyDetails"},
		Request: request,
		Message: Message{
			Text: "Request successfully processed",
			Code: 0,
		},
		PageViewCountMonth: 172,
		PageViewCountTotal: 4149,
		Address: Address{
			Street:    "2114 Bigelow Ave N",
			Zipcode:   "98109",
			City:      "Seattle",
			State:     "WA",
			Latitude:  "47.637924",
			Longitude: "-122.347929",
		},
		Price: Value{Currency: "USD", Value: 1290000},
		Posting: Posting{
			Status:          "Active",
			AgentName:       "John Blacksmith",
			AgentProfileUrl: "/profile/John.Blacksmith",
			Brokerage:       "Lake and Company Real Estate",
			Type:            "For sale by agent",
			LastUpdatedDate: "2008-06-05 10:28:00.0",
			ExternalUrl:     "http://mls.lakere.com/srch_mls/detail.php?mode=ag&LN=28097669&t=listings&l=",
			MLS:             "28097669",
		},
		HomeDetailsLink:  "http://www.zillow.com/HomeDetails.htm?zprop=48749425",
		PhotoGalleryLink: "http://www.zillow.com/Gallery.htm?zpid=48749425",
		HomeInfoLink:     "http://www.zillow.com/HomeInfo.htm?zprop=48749425",
		Images: Images{
			Count: 17,
			Urls: []string{
				"http://images3.zillow.com/is/image/i0/i0/i64/ISz23uixze1pr7.jpg?op_sharpen=1&qlt=90&size=400,400",
				"http://images1.zillow.com/is/image/i0/i0/i64/ISz23uj5vihxnn.jpg?op_sharpen=1&qlt=90&size=400,400",
				"http://images1.zillow.com/is/image/i0/i0/i64/ISz0l5yjj5pajn.jpg?op_sharpen=1&qlt=90&size=400,400",
				"http://images2.zillow.com/is/image/i0/i0/i64/ISz23ukda6z543.jpg?op_sharpen=1&qlt=90&size=400,400",
				"http://images1.zillow.com/is/image/i0/i0/i64/ISz0l5xk0loazn.jpg?op_sharpen=1&qlt=90&size=400,400",
			},
		},
		EditedFacts: EditedFacts{
			UseCode:        "Single family",
			Bedrooms:       4,
			Bathrooms:      3,
			FinishedSqFt:   3470,
			LotSizeSqFt:    4680,
			YearBuilt:      1924,
			YearUpdated:    2003,
			NumFloors:      2,
			Basement:       "Finished",
			Roof:           "Composition",
			View:           "Water, City, Mountain",
			ParkingType:    "Off-street",
			HeatingSources: "Gas",
			HeatingSystem:  "Forced air",
			Appliances:     "Dishwasher, Dryer, Freezer, Garbage disposal, Microwave, Range / Oven, Refrigerator, Washer",
			FloorCovering:  "Hardwood, Carpet, Tile",
			Rooms:          "Laundry room, Walk-in closet, Master bath, Office, Dining room, Family room, Breakfast nook",
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected:\n %#v\n\n but got:\n %#v\n\n diff:\n %s\n",
			pretty.Formatter(expected), pretty.Formatter(result), pretty.Diff(expected, result))
	}
}
