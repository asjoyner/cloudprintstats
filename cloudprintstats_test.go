package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetPrinterUsage(t *testing.T) {
	tests := []struct {
		Filename string
		Want     map[string]int
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
		got, err := GetPrinterUsage(ts.Client(), "", 0, int64(math.MaxInt64))
		if err != nil {
			t.Fatal(err)
		}
		if eq := reflect.DeepEqual(test.Want, got); !eq {
			t.Fatal(fmt.Sprintf("%s: got: %+v, want: %+v", test.Filename, got, test.Want))
		}
	}
}

func TestSummarize(t *testing.T) {
	tests := []struct {
		Jobs   []Job
		After  int64
		Before int64
		Want   map[string]int
	}{
		{ // test 0
			[]Job{
				{
					OwnerID:       "alice@example.com",
					NumberOfPages: 1,
					CreateTime:    "500",
					SemanticState: SemanticState{
						State: State{
							Type: "DONE",
						},
					},
				},
				{
					OwnerID:       "bob@example.com",
					NumberOfPages: 7,
					CreateTime:    "1000",
					SemanticState: SemanticState{
						State: State{
							Type: "DONE",
						},
					},
				},
			},
			0,
			2000,
			map[string]int{
				"alice@example.com": 1,
				"bob@example.com":   7,
			},
		},
		{ // test 1
			[]Job{
				{
					OwnerID:       "alice@example.com",
					NumberOfPages: 1,
					CreateTime:    "500",
					SemanticState: SemanticState{
						State: State{
							Type: "DONE",
						},
					},
				},
				{
					OwnerID:       "bob@example.com",
					NumberOfPages: 7,
					CreateTime:    "1000",
					SemanticState: SemanticState{
						State: State{
							Type: "DONE",
						},
					},
				},
			},
			0,
			700,
			map[string]int{
				"alice@example.com": 1,
			},
		},
	}

	for n, test := range tests {
		got := make(map[string]int)
		summarize(got, test.Jobs, test.After, test.Before)

		if eq := reflect.DeepEqual(test.Want, got); !eq {
			t.Fatal(fmt.Sprintf("Test %d: got: %+v, want: %+v", n, got, test.Want))
		}
	}
}
