package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetPrinterUsage(t *testing.T) {
	tests := []struct {
		Filename string
		Expected map[string]int
	}{
		{"testdata/jobs_response.json",
			map[string]int{
				"alice@example.com": 1,
				"bob@example.com":   7,
			},
		},
	}

	for _, test := range tests {
		jobsResponse, err := ioutil.ReadFile(test.Filename)
		if err != nil {
			t.Fatal(err)
		}
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, string(jobsResponse))
		}))
		defer ts.Close()

		JobsPage = ts.URL
		resp, err := GetPrinterUsage(ts.Client())
		if err != nil {
			t.Fatal(err)
		}
		if eq := reflect.DeepEqual(test.Expected, resp); !eq {
			t.Fatal(fmt.Sprintf("%s: want: %+v, got: %+v", test.Filename, test.Expected, resp))
		}
	}
}
