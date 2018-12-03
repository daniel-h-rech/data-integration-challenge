package api

import (
	"context"
	"encoding/json"
	"github.com/daniel-h-rech/data-integration-challenge/data"
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

	http.HandleFunc("/companies", func(w http.ResponseWriter, r *http.Request) {

		if r.Body != nil {
			defer r.Body.Close()
		}

		switch r.Method {

		case http.MethodGet:

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

		case http.MethodPost:
			err := data.MergeCompanies(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				// TODO write StatusBadRequest message
			}

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	server := &http.Server{Addr: apiAddress}

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
