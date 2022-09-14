package service

import (
	"city/model"
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Connection struct {
	Server     string
	Database   string
	Collection string
}

var Collection *mongo.Collection
var ctx = context.TODO()
var insertDocs int

func (e *Connection) Connect() {
	clientOptions := options.Client().ApplyURI(e.Server)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	Collection = client.Database(e.Database).Collection(e.Collection)
}

func (e *Connection) InsertAllData(cityData []model.CityData) (int, error) {
	for i := range cityData {
		_, err := Collection.InsertOne(ctx, cityData[i])

		if err != nil {
			return 0, errors.New("Unable To Insert New Record")
		}
		insertDocs = i + 1
	}
	return insertDocs, nil
}

func (e *Connection) DeleteData(cityData string) (string, error) {

	filter := bson.D{primitive.E{Key: "name", Value: cityData}}

	cur, err := Collection.DeleteOne(ctx, filter)

	if err != nil {
		return "", err
	}

	if cur.DeletedCount == 0 {
		return "", errors.New("Unable To Delete Employee")
	}

	return "Deleted Successfully", nil
}

func (e *Connection) SearchData(cityName string) ([]*model.CityData, error) {
	var data []*model.CityData

	cursor, err := Collection.Find(ctx, bson.D{primitive.E{Key: "city", Value: cityName}})

	if err != nil {
		return data, err
	}

	for cursor.Next(ctx) {
		var e model.CityData
		err := cursor.Decode(&e)
		if err != nil {
			return data, err
		}
		data = append(data, &e)
	}

	if data == nil {
		return data, errors.New("No data present in db for given city name")
	}
	return data, nil
}

func (e *Connection) SearchDataByKeyAndValue(reqBody model.Search) ([]*model.CityData, error) {
	var data []*model.CityData

	fmt.Println(reqBody.Key)
	fmt.Println(reqBody.Value)
	cursor, err := Collection.Find(ctx, bson.D{primitive.E{Key: reqBody.Key, Value: reqBody.Value}})

	if err != nil {
		return data, err
	}

	for cursor.Next(ctx) {
		var e model.CityData
		err := cursor.Decode(&e)
		if err != nil {
			return data, err
		}
		data = append(data, &e)
	}

	if data == nil {
		return data, errors.New("No data present in db for given city name")
	}
	return data, nil
}
