package service

import (
	"city/model"
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Connection struct {
	Server       string
	Database     string
	Collection   string
	Colllection2 string
}

var Collection *mongo.Collection
var CategoryCollection *mongo.Collection
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
	CategoryCollection = client.Database(e.Database).Collection(e.Colllection2)
}

func (e *Connection) InsertAllData(cityData []model.CityData, field string) (int, error) {

	data, err := e.SearchDataInCategories(field)
	if err != nil {
		return 0, err
	}
	id := data[0].ID

	for i := range cityData {
		cityData[i].CategoriesId = id
		_, err := Collection.InsertOne(ctx, cityData[i])

		if err != nil {
			return 0, errors.New("Unable To Insert New Record")
		}
		insertDocs = i + 1
	}
	return insertDocs, nil
}

func (e *Connection) DeleteData(cityData string) (string, error) {

	id, err := primitive.ObjectIDFromHex(cityData)

	if err != nil {
		return "", err
	}

	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	cur, err := Collection.DeleteOne(ctx, filter)

	if err != nil {
		return "", err
	}

	if cur.DeletedCount == 0 {
		return "", errors.New("Unable To Delete Data")
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

func (e *Connection) UpdateData(cityData model.CityData, field string) (string, error) {

	id, err := primitive.ObjectIDFromHex(field)

	if err != nil {
		return "", err
	}

	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	update := bson.D{primitive.E{Key: "$set", Value: cityData}}

	err2 := Collection.FindOneAndUpdate(ctx, filter, update).Decode(e)

	if err2 != nil {
		return "", err2
	}
	return "Data Updated Successfully", nil
}

func (e *Connection) InsertAllDataInCategories(categoryData []model.Categories) (int, error) {
	for i := range categoryData {
		_, err := CategoryCollection.InsertOne(ctx, categoryData[i])

		if err != nil {
			return 0, errors.New("Unable To Insert New Record")
		}
		insertDocs = i + 1
	}
	return insertDocs, nil
}

func (e *Connection) DeleteDataInCategories(categoryId string) (string, error) {

	id, err := primitive.ObjectIDFromHex(categoryId)

	if err != nil {
		return "", err
	}

	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	cur, err := CategoryCollection.DeleteOne(ctx, filter)

	if err != nil {
		return "", err
	}

	if cur.DeletedCount == 0 {
		return "", errors.New("Unable To Delete Data")
	}

	return "Deleted Successfully", nil
}

func (e *Connection) SearchDataInCategories(name string) ([]*model.Categories, error) {
	var data []*model.Categories

	cursor, err := CategoryCollection.Find(ctx, bson.D{primitive.E{Key: "category", Value: name}})

	if err != nil {
		return data, err
	}

	for cursor.Next(ctx) {
		var e model.Categories
		err := cursor.Decode(&e)
		if err != nil {
			return data, err
		}
		data = append(data, &e)
	}

	if data == nil {
		return data, errors.New("No data present in db for given category")
	}
	return data, nil
}

func (e *Connection) UpdateDataInCategories(cityData model.Categories, field string) (string, error) {

	id, err := primitive.ObjectIDFromHex(field)

	if err != nil {
		return "", err
	}

	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	update := bson.D{primitive.E{Key: "$set", Value: cityData}}

	err2 := CategoryCollection.FindOneAndUpdate(ctx, filter, update).Decode(e)

	if err2 != nil {
		return "", err2
	}
	return "Data Updated Successfully", nil
}

func (e *Connection) SearchUsingBothTables(field string) ([]*model.CityData, error) {
	var finalData []*model.CityData

	data, err := e.SearchDataInCategories(field)
	if err != nil {
		return finalData, err
	}

	id := data[0].ID
	cursor, err := Collection.Find(ctx, bson.D{primitive.E{Key: "categories_id", Value: id}})

	if err != nil {
		return finalData, err
	}

	for cursor.Next(ctx) {
		var e model.CityData
		err := cursor.Decode(&e)
		if err != nil {
			return finalData, err
		}
		finalData = append(finalData, &e)
	}

	if finalData == nil {
		return finalData, errors.New("No data present in city data db for given category")
	}
	return finalData, nil

}