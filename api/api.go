package api

import (
	"encoding/json"
	"github.com/daniel-h-rech/data-integration-challenge/data"
	"net/http"
)

func Start(apiAddress string, mongoDBAddress string) error {

	data.Connect(mongoDBAddress)

	http.HandleFunc("/companies", func(w http.ResponseWriter, r *http.Request) {

		if r.Body != nil {
			defer r.Body.Close()
		}

		switch r.Method {

		case http.MethodGet:
			company, err := data.FindCompany(data.Company{
				Name: r.URL.Query().Get("name"),
				Zip:  r.URL.Query().Get("zip"),
			})

			if err == nil {
				// TODO
			}

			if company == nil {
				w.WriteHeader(http.StatusNotFound)
				// TODO write not found message
			} else {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(company)
			}

		case http.MethodPost:
			err := data.MergeCompanies(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				// TODO write bad request message
			}

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	return http.ListenAndServe(apiAddress, nil)
}
