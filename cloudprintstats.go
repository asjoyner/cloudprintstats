package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/asjoyner/googoauth"
)

var (
	clientID     = "623072786125-6lfie1tgra46origrb350qsom20qj3oa.apps.googleusercontent.com"
	clientSecret = "C6v5lsX9CLLx7ibxlyzOfmq8"
	clientScopes = []string{"https://www.googleapis.com/auth/cloudprint"}
	JobsPage     = "https://www.google.com/cloudprint/jobs"
)

func GetPrinterUsage() (map[string]int, error) {
	// get oauth credentials
	client := googoauth.Client(clientID, clientSecret, clientScopes)

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
		// TODO: optionally set the printerid from a flag to restrict the output
		form.Set("printerid", "")

		response, err := client.PostForm(JobsPage, form)
		if err != nil {
			return nil, fmt.Errorf("Could not retrieve the list of print jobs: %s", err)
		}

		// decode JSON response into struct
		var jobsResponse struct {
			Jobs []struct {
				ID            string
				Title         string
				OwnerID       string
				NumberOfPages int
				CreateTime    string
				ContentType   string
				PrinterName   string
				SemanticState struct {
					State struct {
						Type string
					}
				}
			}
			Range struct {
				JobsTotal string
				JobsCount int
			}
		}
		json.NewDecoder(response.Body).Decode(&jobsResponse)

		// process statistics
		for _, job := range jobsResponse.Jobs {
			if job.SemanticState.State.Type == "DONE" {
				usage[strings.ToLower(job.OwnerID)] += job.NumberOfPages
			}
		}

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

func main() {
	usage, err := GetPrinterUsage()
	var sortedUsernames []string
	for username := range usage {
		sortedUsernames = append(sortedUsernames, username)
	}
	sort.Strings(sortedUsernames)

	fmt.Println("Pages printed by each user:")
	for _, username := range sortedUsernames {
		fmt.Println(" ", username, ":", usage[username])
	}
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
