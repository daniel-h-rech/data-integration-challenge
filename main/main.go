package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"io"
	. "log"
	"net/http"
	"os"
	. "strings"
	"time"
)

func main() {

	if len(os.Args) < 3 {
		Println("Usage:")
		Println("\t" + "integration <CSV file path> <MongoDB network address>")
		Println()
		Println("Example:")
		Println("\t" + "integration q1_catacsv 172.17.0.2:27017")
		return
	}

	csvFilepath := os.Args[1]
	mongoDBAddress := os.Args[2]

	loadCompanyData(csvFilepath, mongoDBAddress)

	err := startAPI(":8080", mongoDBAddress)
	if err != nil {
		Fatal(err)
	}
}

func readCSVStream(reader io.ReadCloser, processCSVRecord func(record []string) error) error {

	csvReader := csv.NewReader(reader)
	csvReader.Comma = ';'

	// skip header
	_, err := csvReader.Read()
	if err != nil {
		return err
	}

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		err = processCSVRecord(record)
		if err != nil {
			return err
		}
	}
}

func getMongoDBClient(mongoDBAddress string, timeoutInSeconds time.Duration) *mongo.Client {

	client, err := mongo.NewClient("mongodb://" + mongoDBAddress)
	if err != nil {
		Fatal("Invalid connection string")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutInSeconds)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		Fatalf("Could not connect to mongodb at %s ", mongoDBAddress)
	}

	return client
}

func loadCompanyData(filepath string, mongoDBAddress string) {

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		Fatalf("file %s does not exist", filepath)
	}

	client := getMongoDBClient(mongoDBAddress, 30*time.Second) // TODO configurable timeout

	collection := client.Database("rechHaeserDaniel").Collection("companies")

	file, err := os.Open(filepath)
	if err != nil {
		Fatal(err)
	}
	defer file.Close()

	err = readCSVStream(file, func(record []string) error {

		record[0] = ToUpper(record[0])

		count, err := collection.Count(context.Background(), bson.D{
			{"name", record[0]},
			{"zip", record[1]},
		})
		if err != nil {
			Fatal(err)
		}

		if count == 0 {
			_, err = collection.InsertOne(context.Background(), bson.M{
				"name": record[0],
				"zip":  record[1],
			})
			if err != nil {
				Fatal(err)
			}
		}

		return nil
	})

	if err != nil {
		Fatal(err)
	}
}

type Company struct {
	Id      primitive.ObjectID `json:"id" bson:"_id"`
	Name    string             `json:"name"`
	Zip     string             `json:"zip"`
	Website string             `json:"website"`
}

func mergeCompanies(collection *mongo.Collection, reader io.ReadCloser) error {

	return readCSVStream(reader, func(record []string) error {

		company := Company{}

		err := collection.FindOneAndUpdate(context.Background(),
			bson.D{
				{
					"name", ToUpper(record[0]),
				},
				{
					"zip", record[1],
				},
			},
			bson.D{
				{
					"$set", bson.M{"website": ToLower(record[2])},
				},
			}).Decode(&company)

		if err != nil {
			return err
		}

		Println(company)

		return nil
	})
}

func findCompany(collection *mongo.Collection, companyKey Company) (*Company, error) {

	company := Company{}

	err := collection.FindOne(context.Background(),
		bson.M{
			"$text": bson.M{"$search": ToUpper(companyKey.Name)},
			"zip":   companyKey.Zip,
		},
	).Decode(&company)

	if err != nil {
		return nil, err
	}

	return &company, nil
}

func startAPI(apiAddress string, mongoDBAddress string) error {

	client := getMongoDBClient(mongoDBAddress, 5*time.Second) // TODO configurable timeout

	collection := client.Database("rechHaeserDaniel").Collection("companies")

	// TODO add ensure indexes

	http.HandleFunc("/companies", func(w http.ResponseWriter, r *http.Request) {

		if r.Body != nil {
			defer r.Body.Close()
		}

		if r.Method == "POST" {
			err := mergeCompanies(collection, r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				// TODO write bad request message
			}
		} else if r.Method == "GET" {
			company, err := findCompany(collection, Company{
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

		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	return http.ListenAndServe(apiAddress, nil)
}
