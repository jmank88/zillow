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
	"strings"
	"testing"
)

const (
	testZwsId = "test-id"
)

const zpid = "48749425"

func testServer(t *testing.T, expectedPath string, validateQuery func(url.Values)) (*httptest.Server, Zillow) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, expectedPath+".htm") {
			t.Errorf("expected path %q to end with %s", r.URL.Path, expectedPath)
		}
		validateQuery(r.URL.Query())
		f, err := os.Open("testdata/" + getZestimatePath + ".xml")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.Copy(w, f); err != nil {
			t.Fatal(err)
		}
	}))
	return ts, &zillow{zwsId: testZwsId, url: ts.URL}
}

func TestGetZestimate(t *testing.T) {
	validateQuery := func(values url.Values) {
		if len(values[zwsIdParam]) != 1 {
			t.Fatalf("expected single %q param", zwsIdParam)
		}
		if zwsId := values.Get(zwsIdParam); zwsId != testZwsId {
			t.Fatalf("expected %q %q param but got %q", testZwsId, zwsIdParam, zwsId)
		}
		if len(values[zpidParam]) != 1 {
			t.Fatalf("expected single %q param", zpidParam)
		}
		if actualZpid := values.Get(zpidParam); actualZpid != zpid {
			t.Fatalf("expected %q %q param but got %q", zpid, zpidParam, actualZpid)
		}
	}
	ts, zillow := testServer(t, getZestimatePath, validateQuery)
	defer ts.Close()

	result, err := zillow.GetZestimate(ZestimateRequest{Zpid: zpid})
	if err != nil {
		t.Fatal(err)
	}
	expected := &ZestimateResult{
		XMLName: xml.Name{Space: "Zestimate", Local: "zestimate"},
		Request: ZestimateRequest{
			Zpid:          zpid,
			Rentzestimate: false,
		},
		Message: Message{
			Text: "Request successfully processed",
			Code: 0,
		},
		HomeDetails: `
                http://www.zillow.com/homedetails/2114-Bigelow-Ave-N-Seattle-WA-98109/48749425_zpid/
            `,
		GraphsAndData: `
                http://www.zillow.com/homedetails/charts/48749425_zpid,1year_chartDuration/?cbt=2950402095890968938%7E4%7ECh-lwa20e2Scegkf_Ev1dsQ2hJD7f74f1dovt2o0BMi2IuvfsZN-sg**
            `,
		MapThisHome: "http://www.zillow.com/homes/map/48749425_zpid/",
		Comparables: "http://www.zillow.com/homes/comps/48749425_zpid/",
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
				ID:                  "271856",
				Type:                "neighborhood",
				Name:                "East Queen Anne",
				ZIndex:              "525,397",
				ZIndexOneYearChange: -0.144,
				Overview: `
                        http://www.zillow.com/local-info/WA-Seattle/East-Queen-Anne/r_271856/
                    `,
				ForSaleByOwner: `
                        http://www.zillow.com/homes/fsbo/East-Queen-Anne-Seattle-WA/
                    `,
				ForSale: `
                        http://www.zillow.com/east-queen-anne-seattle-wa/
                    `,
			},
			Region{
				ID:                  "16037",
				Type:                "city",
				Name:                "Seattle",
				ZIndex:              "381,764",
				ZIndexOneYearChange: -0.074,
				Overview: `
                        http://www.zillow.com/local-info/WA-Seattle/r_16037/
                    `,
				ForSaleByOwner: `http://www.zillow.com/homes/fsbo/Seattle-WA/`,
				ForSale:        `http://www.zillow.com/seattle-wa/`,
			},
			Region{
				ID:                  "59",
				Type:                "state",
				Name:                "Washington",
				ZIndex:              "263,278",
				ZIndexOneYearChange: -0.066,
				Overview: `
                        http://www.zillow.com/local-info/WA-home-value/r_59/
                    `,
				ForSaleByOwner: `http://www.zillow.com/homes/fsbo/WA/`,
				ForSale:        `http://www.zillow.com/wa/`,
			},
		},
		ZipcodeID: "99569",
		CityID:    "16037",
		CountyID:  "207",
		StateID:   "59",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected:\n %s\n\n but got:\n %s\n\n diff: %s\n",
			pretty.Formatter(expected), pretty.Formatter(result), pretty.Diff(expected, result))
	}
}
