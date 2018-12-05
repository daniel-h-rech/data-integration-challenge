/*
Package api data-integration-challenge - Daniel Haeser Rech

    Schemes:
      http
    Host: localhost:8080
    BasePath: /
    Version: 1.0.0

swagger:meta
*/
package api

import (
	"context"
	"encoding/json"
	"github.com/daniel-h-rech/data-integration-challenge/data"
	"github.com/go-chi/chi"
	. "log"
	"net/http"
	"os"
	"os/signal"
	. "strings"
	"time"
)

func Start(apiAddress string, mongoDBAddress string) {

	Println("Starting the API service...")

	defer data.Connect(mongoDBAddress)()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	r := chi.NewRouter()
	r.Get("/companies", getCompany)
	r.Post("/companies", postCompanyWebsites)

	server := &http.Server{Addr: apiAddress, Handler: r}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			Fatal(err)
		}
	}()

	<-stop

	Println("Stopping the API service...")

	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFunc()
	err := server.Shutdown(timeout)
	if err != nil {
		Println(err)
	}

	Println("API service Stopped")
}

// swagger:operation GET /companies getCompany
//
// Gets the company for the given name and zip query parameters.
// The search uses an inverted index, so you may use a single word for searching a multi-worded company name.
//
// ---
// produces:
// - application/json
// parameters:
// - name: name
//   in: query
//   description: company name
//   type: string
// - name: zip
//   in: query
//   description: company zip code
//   type: string
// responses:
//  200:
//    description: Successful operation
//    $ref: #/responses/Company
//  400:
//    description: Invalid parameters
//  404:
//    description: Company not found
func getCompany(w http.ResponseWriter, r *http.Request) {

	companyKey := data.Company{
		Name: TrimSpace(r.URL.Query().Get("name")),
		Zip:  TrimSpace(r.URL.Query().Get("zip")),
	}

	if companyKey.Name == "" || companyKey.Zip == "" {
		w.WriteHeader(http.StatusBadRequest)
		// TODO write StatusBadRequest message
		return
	}

	company, err := data.FindCompany(companyKey)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Println(err)
	}

	if company == nil {
		w.WriteHeader(http.StatusNotFound)
		// TODO write StatusNotFound message
	} else {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(company)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			Println(err)
		}
	}
}

// swagger:operation POST /companies postCompanyWebsites
//
// Merges the website link for the list of companies given in CSV format.
//
// ---
// consumes:
// - text/csv
// parameters:
//   - name: csvContent
//     description: CSV content in the the format name;addresszip;website. The header line is mandatory.
//     in: body
//     required: false
//     type: string
// responses:
//  204:
//    description: Successful operation
//  400:
//    description: Bad CSV format
func postCompanyWebsites(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}
	err := data.MergeCompanies(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		// TODO write StatusBadRequest message
	}
	w.WriteHeader(http.StatusNoContent)
}
