package data

import (
	"context"
	"encoding/csv"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/x/bsonx"
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

type Close func()

type index struct {
	Name string
}

var companies *mongo.Collection // TODO nil checks

func ensureIndex(context context.Context) {

	cursor, err := companies.Indexes().List(context)
	if err != nil {
		Fatal("error listing indexes")
	}

	const indexName = "nameTextIndex"

	exists := false
	for cursor.Next(context) {

		i := index{}
		err := cursor.Decode(&i)
		if err != nil {
			Fatal("error listing indexes")
		}

		if i.Name == indexName {
			exists = true
			break
		}
	}

	if !exists {
		Printf("index %s does not exist, will create", indexName)

		_, err = companies.Indexes().CreateOne(context, mongo.IndexModel{
			Keys:    bsonx.Doc{{"name", bsonx.String("text")}},
			Options: mongo.NewIndexOptionsBuilder().Name(indexName).Build(),
		})

		if err != nil {
			Fatalf("Could not create index %s", indexName)
		}

		Printf("created index %s", indexName)
	}
}

func Connect(mongoDBAddress string) Close {

	client, err := mongo.NewClient("mongodb://" + mongoDBAddress)
	if err != nil {
		Fatalf("invalid address %s", mongoDBAddress)
	}

	Printf("connecting to %s", client.ConnectionString())

	connectTimeout, cancel := context.WithTimeout(context.Background(), time.Minute) // TODO configurable timeout
	defer cancel()

	err = client.Connect(connectTimeout)
	if err != nil {
		Fatalf("could not connect to mongodb at %s ", mongoDBAddress)
	}

	companies = client.Database("rechHaeserDaniel").Collection("companies")

	ensureIndex(connectTimeout)

	Println("connected")

	return func() {
		Println("closing mongodb connection...")

		disconnectTimeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second) // TODO configurable timeout
		defer cancelFunc()

		err := client.Disconnect(disconnectTimeout)
		if err != nil {
			Println(err)
		}

		Println("mongodb connection closed")
	}
}

func LoadCompanyData(filepath string, mongoDBAddress string) {

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		Fatalf("file %s does not exist", filepath)
	}

	close := Connect(mongoDBAddress)
	defer close()

	file, err := os.Open(filepath)
	if err != nil {
		Fatal(err)
	}
	defer file.Close()

	Printf("Loading file %s...", filepath)

	err = readCSVStream(file, func(record []string) error {

		record[0] = ToUpper(record[0])

		recordTimeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second) // TODO configurable timeout
		defer cancelFunc()

		count, err := companies.Count(recordTimeout, bson.D{
			{"name", record[0]},
			{"zip", record[1]},
		})
		if err != nil {
			Fatal(err)
		}

		if count == 0 {
			_, err = companies.InsertOne(recordTimeout, bson.M{
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

	Println("Done loading")
}

func FindCompany(companyKey Company) (*Company, error) {

	timeout, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	company := Company{}

	err := companies.FindOne(timeout,
		bson.M{
			"$text": bson.M{"$search": ToUpper(companyKey.Name)},
			"zip":   companyKey.Zip,
		},
	).Decode(&company)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &company, nil
}

func MergeCompanies(reader io.ReadCloser) error {

	return readCSVStream(reader, func(record []string) error {

		timeout, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		result := companies.FindOneAndUpdate(timeout,
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
			},
		)

		if result.Err() != nil {
			return result.Err()
		}

		company := Company{}

		err := result.Decode(&company)
		if err == nil || err == mongo.ErrNoDocuments {
			return nil
		}

		return err
	})
}

func readCSVStream(reader io.Reader, processCSVRecord func(record []string) error) error {

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
