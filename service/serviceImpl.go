package service

import (
	"city/model"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

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

func (e *Connection) SearchData(searchBoth model.SearchBoth) ([]*model.CityData, error) {
	var data []*model.CityData
	var cursor *mongo.Cursor
	var err error
	str := "please provide value of either city or category"

	if searchBoth.City != "" {
		cursor, err = Collection.Find(ctx, bson.D{primitive.E{Key: "city", Value: searchBoth.City}})

		if err != nil {
			return data, err
		}
		str = "No data present in db for given city name"
	} else if searchBoth.Category != "" {
		categoryData, error := e.SearchDataInCategories(searchBoth.Category)

		if error != nil {
			return data, err
		}

		id := categoryData[0].ID
		cursor, err = Collection.Find(ctx, bson.D{primitive.E{Key: "categories_id", Value: id}})

		if err != nil {
			return data, err
		}
		str = "No data present in city data db for given category"
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
		return data, errors.New(str)
	}
	os.MkdirAll("data/download", os.ModePerm)
	file := "data/download/searchResult" + fmt.Sprintf("%v", time.Now().Format("3_4_5_pm")) + ".csv"
	csvFile, err := os.Create(file)

	if err != nil {
		fmt.Println(err)
	}
	defer csvFile.Close()
	writer := csv.NewWriter(csvFile)

	header := []string{"ID", "Title", "Name", "Address", "Latitude", "Longitude", "Website", "ContactNumber", "User", "City", "Country", "PinCode", "UpdatedBy", "CategoriesId"}
	if err := writer.Write(header); err != nil {
		return data, err
	}

	for _, r := range data {
		var csvRow []string
		csvRow = append(csvRow, fmt.Sprintf("%v", r.ID), r.Title, r.Name, r.Address, fmt.Sprintf("%f", r.Latitude), fmt.Sprintf("%f", r.Longitude), r.Website, fmt.Sprintf("%v", r.ContactNumber),
			r.User, r.City, r.Country, fmt.Sprintf("%v", r.PinCode), r.UpdatedBy, fmt.Sprintf("%v", r.CategoriesId))
		if err := writer.Write(csvRow); err != nil {
			return data, err
		}
	}

	// remember to flush!
	writer.Flush()
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
