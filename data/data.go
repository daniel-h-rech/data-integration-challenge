package data

import (
	"context"
	"encoding/csv"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"io"
	. "log"
	"os"
	. "strings"
	"time"
)

type Company struct {
	Id      primitive.ObjectID `json:"id" bson:"_id"`
	Name    string             `json:"name"`
	Zip     string             `json:"zip"`
	Website string             `json:"website"`
}

var companies *mongo.Collection

func Connect(mongoDBAddress string) {

	client, err := mongo.NewClient("mongodb://" + mongoDBAddress)
	if err != nil {
		Fatalf("Invalid address %s", mongoDBAddress)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		Fatalf("Could not connect to mongodb at %s ", mongoDBAddress)
	}

	// TODO add ensure indexes
	companies = client.Database("rechHaeserDaniel").Collection("companies")
}

func LoadCompanyData(filepath string, mongoDBAddress string) {

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		Fatalf("file %s does not exist", filepath)
	}

	Connect(mongoDBAddress)

	file, err := os.Open(filepath)
	if err != nil {
		Fatal(err)
	}
	defer file.Close()

	err = readCSVStream(file, func(record []string) error {

		record[0] = ToUpper(record[0])

		count, err := companies.Count(context.Background(), bson.D{
			{"name", record[0]},
			{"zip", record[1]},
		})
		if err != nil {
			Fatal(err)
		}

		if count == 0 {
			_, err = companies.InsertOne(context.Background(), bson.M{
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

func FindCompany(companyKey Company) (*Company, error) {

	company := Company{}

	err := companies.FindOne(context.Background(),
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

func MergeCompanies(reader io.ReadCloser) error {

	return readCSVStream(reader, func(record []string) error {

		company := Company{}

		err := companies.FindOneAndUpdate(context.Background(),
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

		return nil
	})
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
