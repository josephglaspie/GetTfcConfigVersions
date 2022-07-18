package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/hashicorp/go-tfe"
	"log"
	"os"
	"strings"
	"time"
)

var (
	tfcToken                 = os.Getenv("TFC_TOKEN")
	fileName                 = "tfcConfigVersion"
	allConfigurationVersions []configVersion
)

type configVersion struct {
	Workspace    string
	Id           string
	Status       string
	ErrorMessage string
}

func main() {
	fullFileName := fileName + "_" + TimeStamp() + ".csv"
	ctx := context.Background()
	config := &tfe.Config{
		Token: tfcToken,
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	wslOpts := &tfe.WorkspaceListOptions{
		ListOptions: tfe.ListOptions{
			PageSize: 100,
		},
	}

	//allWorkspaces := []*tfe.Workspace
	// get all pages of results
	var allWorkspaces []*tfe.Workspace
	for {
		ws, err := client.Workspaces.List(ctx, "twilio-main", wslOpts)
		if err != nil {
			log.Fatal(err)
		}
		allWorkspaces = append(allWorkspaces, ws.Items...)
		if ws.NextPage == 0 {
			break
		}
		wslOpts.PageNumber = ws.NextPage
	}
	totalNumberOfWorkspaces := len(allWorkspaces)

	cvOpts := &tfe.ConfigurationVersionListOptions{
		ListOptions: tfe.ListOptions{
			PageSize: 100,
		},
		Include: nil,
	}

	for i := 0; i < len(allWorkspaces); i++ {
		cvs, err := client.ConfigurationVersions.List(ctx, allWorkspaces[i].ID, cvOpts)
		if err != nil {
			log.Fatal(err)
		}
		configVersions := cvs.Items
		for ii := 0; ii < len(configVersions); ii++ {
			ConfVersions := configVersion{
				Workspace:    allWorkspaces[i].Name,
				Id:           configVersions[ii].ID,
				Status:       string(configVersions[ii].Status),
				ErrorMessage: configVersions[ii].ErrorMessage,
				//Started:    configVersions[ii].StatusTimestamps.StartedAt.Format(time.RFC3339),
				//Finished:   configVersions[ii].StatusTimestamps.FinishedAt.Format(time.RFC3339),
				//Archived:   configVersions[ii].StatusTimestamps.ArchivedAt.Format(time.RFC3339),
				//FetchingAt: configVersions[ii].StatusTimestamps.FetchingAt.Format(time.RFC3339),
			}

			allConfigurationVersions = append(allConfigurationVersions, ConfVersions)
		}
	}

	csvFile, err := os.Create(fullFileName)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	defer csvFile.Close()
	csvwriter := csv.NewWriter(csvFile)

	// Using WriteAll
	var data [][]string
	for _, record := range allConfigurationVersions {
		row := []string{record.Workspace, record.Id, record.Status}
		data = append(data, row)
	}

	err = csvwriter.WriteAll(data)
	if err != nil {
		log.Fatalf("WriteAll to csv failed %s", err)
	}

	fmt.Printf("Number of workspaces: %d\n", totalNumberOfWorkspaces)
	fmt.Printf("Your file %s is ready", fullFileName)
}

func TimeStamp() string {
	ts := time.Now().UTC().Format(time.RFC3339)
	return strings.Replace(ts, ":", "", -1) // get rid of offensive colons
}
