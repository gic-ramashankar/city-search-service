package service

import (
	"city/model"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/signintech/gopdf"
	"github.com/xuri/excelize/v2"
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

func (e *Connection) SearchData(searchBoth model.SearchBoth, option string) ([]byte, string, error) {
	var data []*model.CityData
	var cursor *mongo.Cursor
	var err error
	var dataty []byte
	str := "please provide value of either city or category"

	os.MkdirAll("data/download", os.ModePerm)
	dir := "data/download/"
	file := "searchResult" + fmt.Sprintf("%v", time.Now().Format("3_4_5_pm"))
	//	xlsFile, err := os.Create(dir + file + ".xls")
	if err != nil {
		fmt.Println(err)
	}

	if (searchBoth.City != "") && (searchBoth.Category != "") {
		categoryData, error := e.SearchDataInCategories(searchBoth.Category)

		if error != nil {
			return dataty, file, err
		}

		id := categoryData[0].ID
		cursor, err = Collection.Find(ctx, bson.D{primitive.E{Key: "categories_id", Value: id}, primitive.E{Key: "city", Value: searchBoth.City}})

		if err != nil {
			return dataty, file, err
		}
		str = "No data present in city data db for given category or city"
	} else if searchBoth.City != "" {
		cursor, err = Collection.Find(ctx, bson.D{primitive.E{Key: "city", Value: searchBoth.City}})

		if err != nil {
			return dataty, file, err
		}
		str = "No data present in db for given city name"
	} else if searchBoth.Category != "" {
		categoryData, error := e.SearchDataInCategories(searchBoth.Category)

		if error != nil {
			return dataty, file, err
		}

		id := categoryData[0].ID
		cursor, err = Collection.Find(ctx, bson.D{primitive.E{Key: "categories_id", Value: id}})

		if err != nil {
			return dataty, file, err
		}
		str = "No data present in city data db for given category"
	}

	for cursor.Next(ctx) {
		var e model.CityData
		err := cursor.Decode(&e)
		if err != nil {
			return dataty, file, err
		}
		data = append(data, &e)
	}

	if data == nil {
		return dataty, file, errors.New(str)
	}

	if option == "Excel" {
		log.Println("Excel")
		errExcel := writeDataIntoExcel(dir, file, data)
		if errExcel != nil {
			return dataty, file, err
		}
		dataty, err = ioutil.ReadFile(dir + file + ".xlsx")
		if err != nil {
			return dataty, file, err
		}
	}

	if option == "Pdf" {
		log.Println("Pdf")
		errExcel := writeDataIntoPdf(dir, file, data)
		if errExcel != nil {
			return dataty, file, err
		}
		// dataty, err = ioutil.ReadFile(dir + file + ".xlsx")
		// if err != nil {
		// 	return dataty, file, err
		// }
	}
	return dataty, file, nil
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

func writeDataIntoExcel(dir, file string, data []*model.CityData) error {

	f := excelize.NewFile()
	f.SetSheetName("Sheet1", "SearchData")

	f.SetCellValue("SearchData", "A1", "ID")
	f.SetCellValue("SearchData", "B1", "Title")
	f.SetCellValue("SearchData", "C1", "Name")
	f.SetCellValue("SearchData", "D1", "Address")
	f.SetCellValue("SearchData", "E1", "Latitude")
	f.SetCellValue("SearchData", "F1", "Longitude")
	f.SetCellValue("SearchData", "G1", "Website")
	f.SetCellValue("SearchData", "H1", "ContactNumber")
	f.SetCellValue("SearchData", "I1", "User")
	f.SetCellValue("SearchData", "J1", "City")
	f.SetCellValue("SearchData", "K1", "Country")
	f.SetCellValue("SearchData", "L1", "PinCode")
	f.SetCellValue("SearchData", "M1", "UpdatedBy")
	f.SetCellValue("SearchData", "N1", "CategoriesId")

	for i := range data {
		f.SetCellValue("SearchData", "A"+fmt.Sprintf("%v", i+2), data[i].ID)
		f.SetCellValue("SearchData", "B"+fmt.Sprintf("%v", i+2), data[i].Title)
		f.SetCellValue("SearchData", "C"+fmt.Sprintf("%v", i+2), data[i].Name)
		f.SetCellValue("SearchData", "D"+fmt.Sprintf("%v", i+2), data[i].Address)
		f.SetCellValue("SearchData", "E"+fmt.Sprintf("%v", i+2), data[i].Latitude)
		f.SetCellValue("SearchData", "F"+fmt.Sprintf("%v", i+2), data[i].Longitude)
		f.SetCellValue("SearchData", "G"+fmt.Sprintf("%v", i+2), data[i].Website)
		f.SetCellValue("SearchData", "H"+fmt.Sprintf("%v", i+2), data[i].ContactNumber)
		f.SetCellValue("SearchData", "I"+fmt.Sprintf("%v", i+2), data[i].User)
		f.SetCellValue("SearchData", "J"+fmt.Sprintf("%v", i+2), data[i].City)
		f.SetCellValue("SearchData", "K"+fmt.Sprintf("%v", i+2), data[i].Country)
		f.SetCellValue("SearchData", "L"+fmt.Sprintf("%v", i+2), data[i].PinCode)
		f.SetCellValue("SearchData", "M"+fmt.Sprintf("%v", i+2), data[i].UpdatedBy)
		f.SetCellValue("SearchData", "N"+fmt.Sprintf("%v", i+2), data[i].CategoriesId)
	}

	if err := f.SaveAs(dir + file + ".xlsx"); err != nil {
		return err
	}
	return nil
}

func writeDataIntoPdf(dir, file string, data []*model.CityData) error {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	err := pdf.AddTTFFont("wts11", "./font/Lato-Light.ttf")
	if err != nil {
		log.Print(err.Error())
		fmt.Println(err)
		return err
	}

	err = pdf.SetFont("wts11", "", 10)
	if err != nil {
		log.Print(err.Error())
		return err
	}
	pdf.SetXY(50, 50)
	x := 10.0
	y := 10.0

	for i := range data {
		pdf.SetXY(50, 50+y)
		pdf.Cell(nil, fmt.Sprintf("%v", data[i].ID))
		pdf.Cell(nil, data[i].Title)
		pdf.Cell(nil, data[i].Name)
		pdf.Cell(nil, data[i].Address)
		pdf.Cell(nil, fmt.Sprintf("%v", data[i].Latitude))
		pdf.Cell(nil, fmt.Sprintf("%v", data[i].Longitude))
		pdf.Cell(nil, data[i].Website)
		pdf.Cell(nil, fmt.Sprintf("%v", data[i].ContactNumber))
		pdf.Cell(nil, data[i].User)
		pdf.Cell(nil, data[i].City)
		pdf.Cell(nil, data[i].Country)
		pdf.Cell(nil, fmt.Sprintf("%v", data[i].PinCode))
		pdf.Cell(nil, data[i].UpdatedBy)
		pdf.Cell(nil, fmt.Sprintf("%v", data[i].CategoriesId))
		pdf.Next
		x = x + 50.0
		y = y + 50.0
	}
	pdf.WritePdf(dir + file + ".pdf")
	fmt.Printf("Completed")
	return nil
}
