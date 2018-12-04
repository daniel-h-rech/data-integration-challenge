package api

import (
	"encoding/json"
	"github.com/daniel-h-rech/data-integration-challenge/data"
	"net/http"
	"net/url"
	"os"
	. "strings"
	"testing"
	"time"
)

const apiAddress = "localhost:8081"

const baseAPIAddress = "http://" + apiAddress

const mongoDBAddress = "localhost:27018"

const loadCSVFilePath = "../q2_clientData.csv"

const postCSVFilePath = "../q2_clientData.csv"

func init() {

	data.LoadCompanyData(loadCSVFilePath, mongoDBAddress)

	go Start(apiAddress, mongoDBAddress)
}

func TestPostCompanies(t *testing.T) {

	waitForServer()

	file, e := os.Open(postCSVFilePath)
	if e != nil {
		t.Fail()
	}
	defer file.Close()

	response, e := http.DefaultClient.Post(baseAPIAddress+"/companies", "text/csv", file)
	if e != nil {
		t.Error(e)
	}

	if response.StatusCode != http.StatusNoContent {
		t.Errorf("HTTP status returned %d, was expecting %d", response.StatusCode, http.StatusNoContent)
	}
}

func TestGetCompany(t *testing.T) {

	waitForServer()

	file, e := os.Open(postCSVFilePath)
	if e != nil {
		t.Fail()
	}
	defer file.Close()

	data.ReadCSVStream(file, func(record []string) error {

		response, e := http.DefaultClient.Get(baseAPIAddress + "/companies?name=" + url.QueryEscape(ToUpper(record[0])) + "&zip=" + record[1])
		if e != nil {
			t.Error(e)
		}
		defer response.Body.Close()

		if response.StatusCode == http.StatusOK {
			company := data.Company{}

			json.NewDecoder(response.Body).Decode(&company)

			record[0] = ToUpper(record[0])

			if record[0] != company.Name || record[1] != company.Zip || record[2] != company.Website {
				t.Errorf("returned %s, expected %s", company, record)
			}

		} else {
			t.Errorf("HTTP status returned %d, was expecting %d", response.StatusCode, http.StatusNoContent)
		}

		return nil
	})
}

func waitForServer() {
	// TODO find a less hackish way to wait for the server
	for {
		resp, err := http.Get(baseAPIAddress)

		if err != nil {
			time.Sleep(time.Millisecond * 10)
			continue
		}

		resp.Body.Close()
		if resp.StatusCode >= http.StatusInternalServerError {
			time.Sleep(time.Millisecond * 10)
			continue
		}
		break
	}
}
