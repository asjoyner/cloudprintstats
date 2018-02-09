package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/asjoyner/googoauth"
)

var (
	clientID     = "623072786125-6lfie1tgra46origrb350qsom20qj3oa.apps.googleusercontent.com"
	clientSecret = "C6v5lsX9CLLx7ibxlyzOfmq8"
	clientScopes = []string{"https://www.googleapis.com/auth/cloudprint"}
	JobsPage     = "https://www.google.com/cloudprint/jobs"
)

var (
	lastDuration = flag.Duration("last", 0, "Limit to print jobs in the last <period>.  0 shows all jobs.")
	printerId    = flag.String("printerId", "", "Limit to print jobs from a specific printer.  An empty string will show jobs printed to all printers the authenticated user has access to.")
)

type Job struct {
	ID            string
	Title         string
	OwnerID       string
	NumberOfPages int
	CreateTime    string
	ContentType   string
	PrinterName   string
	SemanticState SemanticState
}

type SemanticState struct {
	State State
}

type State struct {
	Type string
}

func GetPrinterUsage(client *http.Client, printerId string, after, before int64) (map[string]int, error) {
	// map of username to number of pages printed
	usage := make(map[string]int)

	jobsCounted := 0
	offset := 0
	limit := 100
	for {
		// fetch the Jobs page
		form := url.Values{}
		form.Set("sortorder", "CREATE_TIME_DESC")
		form.Set("status", "DONE")
		form.Set("limit", fmt.Sprintf("%d", limit))   // number of jobs per page
		form.Set("offset", fmt.Sprintf("%d", offset)) // zero-ordered
		if printerId != "" {
			form.Set("printerid", printerId)
		}

		response, err := client.PostForm(JobsPage, form)
		if err != nil {
			return nil, fmt.Errorf("Could not retrieve the list of print jobs: %s", err)
		}
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Error retrieving the list of print jobs: %s", response.Status)
		}

		// decode JSON response into struct
		var jobsResponse struct {
			Jobs  []Job
			Range struct {
				JobsTotal string
				JobsCount int
			}
		}
		if err := json.NewDecoder(response.Body).Decode(&jobsResponse); err != nil {
			return nil, fmt.Errorf("could not decode Jobs page response: %s", err)
		}

		summarize(usage, jobsResponse.Jobs, after, before)

		offset += limit
		total, err := strconv.Atoi(jobsResponse.Range.JobsTotal)
		if err != nil {
			return nil, err
		}
		jobsCounted += jobsResponse.Range.JobsCount
		if jobsCounted < total {
			continue
		} else if jobsCounted == total {
			break
		} else if jobsCounted > total {
			fmt.Printf("Counting may be inflated: counted %d jobs of %d total.\n", jobsCounted, total)
			break
		}
	}

	return usage, nil
}

// process statistics
func summarize(usage map[string]int, jobs []Job, after, before int64) {
	for _, job := range jobs {
		ctime, err := strconv.ParseInt(job.CreateTime, 10, 64)
		if err != nil {
			continue
		}
		if ctime < after || ctime > before {
			continue
		}

		if job.SemanticState.State.Type == "DONE" {
			usage[strings.ToLower(job.OwnerID)] += job.NumberOfPages
		}
	}
}

func main() {
	flag.Parse()

	after := time.Now().Add(-*lastDuration).Unix()
	before := int64(math.MaxInt64)

	// get oauth credentials
	client := googoauth.Client(clientID, clientSecret, clientScopes)

	usage, err := GetPrinterUsage(client, *printerId, after, before)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// print the usage in a sorted order
	var sortedUsernames []string
	for username := range usage {
		sortedUsernames = append(sortedUsernames, username)
	}
	sort.Strings(sortedUsernames)

	fmt.Println("Pages printed by each user:")
	for _, username := range sortedUsernames {
		fmt.Println(" ", username, ":", usage[username])
	}
}
