package service

import (
	"city/pojo"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/creator"
	"github.com/unidoc/unipdf/v3/model"
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
	err = license.SetMeteredKey("72c4ab06d023bbc8b2e186d089f9e052654afea32b75141f39c7dc1ab3b108ca")
	if err != nil {
		log.Fatal(err)
	}

	Collection = client.Database(e.Database).Collection(e.Collection)
	CategoryCollection = client.Database(e.Database).Collection(e.Colllection2)
}

func (e *Connection) InsertAllData(cityData []pojo.CityData, field string) (int, error) {

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

func (e *Connection) SearchData(searchBoth pojo.SearchBoth, option string) ([]byte, string, error) {
	var data []*pojo.CityData
	var cursor *mongo.Cursor
	var err error
	var dataty []byte
	str := "please provide value of either city or category"

	os.MkdirAll("data/download", os.ModePerm)
	dir := "data/download/"
	file := "searchResult" + fmt.Sprintf("%v", time.Now().Format("3_4_5_pm"))
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
		var e pojo.CityData
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
		_, errPdf := writeDataIntoPDFTable(dir, file, data)
		if errPdf != nil {
			fmt.Println(errPdf)
			return dataty, file, err
		}
		dataty, err2 := ioutil.ReadFile(dir + file + ".pdf")
		fmt.Println("Data length", len(dataty))
		if err2 != nil {
			return dataty, file, err
		}
	}

	return dataty, file, nil
}

func (e *Connection) SearchDataByKeyAndValue(reqBody pojo.Search) ([]*pojo.CityData, error) {
	var data []*pojo.CityData

	cursor, err := Collection.Find(ctx, bson.D{primitive.E{Key: reqBody.Key, Value: reqBody.Value}})

	if err != nil {
		return data, err
	}

	for cursor.Next(ctx) {
		var e pojo.CityData
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

func (e *Connection) UpdateData(cityData pojo.CityData, field string) (string, error) {

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

func (e *Connection) InsertAllDataInCategories(categoryData []pojo.Categories) (int, error) {
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

func (e *Connection) SearchDataInCategories(name string) ([]*pojo.Categories, error) {
	var data []*pojo.Categories

	cursor, err := CategoryCollection.Find(ctx, bson.D{primitive.E{Key: "category", Value: name}})

	if err != nil {
		return data, err
	}

	for cursor.Next(ctx) {
		var e pojo.Categories
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

func (e *Connection) UpdateDataInCategories(cityData pojo.Categories, field string) (string, error) {

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

func writeDataIntoExcel(dir, file string, data []*pojo.CityData) error {

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

func writeDataIntoPDFTable(dir, file string, data []*pojo.CityData) (*creator.Creator, error) {

	c := creator.New()
	c.SetPageMargins(20, 20, 20, 20)

	// Create report fonts.
	// UniPDF supports a number of font-families, which can be accessed using model.
	// Here we are creating two fonts, a normal one and its bold version
	font, err := model.NewStandard14Font(model.HelveticaName)
	if err != nil {
		return c, err
	}

	// Bold font
	fontBold, err := model.NewStandard14Font(model.HelveticaBoldName)
	if err != nil {
		return c, err
	}

	// Generate basic usage chapter.
	if err := basicUsage(c, font, fontBold, data); err != nil {
		return c, err
	}

	err = c.WriteToFile(dir + file + ".pdf")
	if err != nil {
		return c, err
	}
	return c, nil
}

func basicUsage(c *creator.Creator, font, fontBold *model.PdfFont, data []*pojo.CityData) error {
	// Create chapter.
	ch := c.NewChapter("Search Data")
	ch.SetMargins(0, 0, 10, 0)
	ch.GetHeading().SetFont(font)
	ch.GetHeading().SetFontSize(18)
	ch.GetHeading().SetColor(creator.ColorRGBFrom8bit(72, 86, 95))
	// You can also set inbuilt colors using creator
	// ch.GetHeading().SetColor(creator.ColorBlack)

	// Draw subchapters. Here we are only create horizontally aligned chapter.
	// You can also vertically align and perform other optimizations as well.
	// Check GitHub example for more.
	contentAlignH(c, ch, font, fontBold, data)

	// Draw chapter.
	if err := c.Draw(ch); err != nil {
		return err
	}

	return nil
}

func contentAlignH(c *creator.Creator, ch *creator.Chapter, font, fontBold *model.PdfFont, data []*pojo.CityData) {
	// Create subchapter.
	// sc := ch.NewSubchapter("Content horizontal alignment")
	// sc.GetHeading().SetFontSize(10)
	// sc.GetHeading().SetColor(creator.ColorBlue)

	// Create table.
	table := c.NewTable(14)
	table.SetMargins(0, 0, 15, 0)

	drawCell := func(text string, font *model.PdfFont, align creator.CellHorizontalAlignment) {
		p := c.NewStyledParagraph()
		p.Append(text).Style.Font = font

		cell := table.NewCell()
		cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
		cell.SetHorizontalAlignment(align)
		cell.SetContent(p)
	}
	// Draw table header.
	drawCell("ID", fontBold, creator.CellHorizontalAlignmentLeft)
	drawCell("Title", fontBold, creator.CellHorizontalAlignmentCenter)
	drawCell("Name", fontBold, creator.CellHorizontalAlignmentRight)
	drawCell("Address", fontBold, creator.CellHorizontalAlignmentLeft)
	drawCell("Latitude", fontBold, creator.CellHorizontalAlignmentRight)
	drawCell("Longitude", fontBold, creator.CellHorizontalAlignmentLeft)
	drawCell("Website", fontBold, creator.CellHorizontalAlignmentCenter)
	drawCell("ContactNumber", fontBold, creator.CellHorizontalAlignmentRight)
	drawCell("User", fontBold, creator.CellHorizontalAlignmentLeft)
	drawCell("City", fontBold, creator.CellHorizontalAlignmentCenter)
	drawCell("Country", fontBold, creator.CellHorizontalAlignmentRight)
	drawCell("PinCode", fontBold, creator.CellHorizontalAlignmentLeft)
	drawCell("UpdatedBy", fontBold, creator.CellHorizontalAlignmentCenter)
	drawCell("CategoriesId", fontBold, creator.CellHorizontalAlignmentRight)

	// Draw table content.
	for i := range data {

		drawCell(fmt.Sprintf("%v", data[i].ID), font, creator.CellHorizontalAlignmentLeft)
		drawCell(data[i].Title, font, creator.CellHorizontalAlignmentCenter)
		drawCell(data[i].Name, font, creator.CellHorizontalAlignmentCenter)
		drawCell(data[i].Address, font, creator.CellHorizontalAlignmentCenter)
		drawCell(fmt.Sprintf("%v", data[i].Latitude), font, creator.CellHorizontalAlignmentCenter)
		drawCell(fmt.Sprintf("%v", data[i].Longitude), font, creator.CellHorizontalAlignmentCenter)
		drawCell(data[i].Website, font, creator.CellHorizontalAlignmentCenter)
		drawCell(fmt.Sprintf("%v", data[i].ContactNumber), font, creator.CellHorizontalAlignmentCenter)
		drawCell(data[i].User, font, creator.CellHorizontalAlignmentCenter)
		drawCell(data[i].City, font, creator.CellHorizontalAlignmentCenter)
		drawCell(data[i].Country, font, creator.CellHorizontalAlignmentCenter)
		drawCell(fmt.Sprintf("%v", data[i].PinCode), font, creator.CellHorizontalAlignmentCenter)
		drawCell(data[i].UpdatedBy, font, creator.CellHorizontalAlignmentCenter)
		drawCell(fmt.Sprintf("%v", data[i].CategoriesId), font, creator.CellHorizontalAlignmentCenter)
	}

	ch.Add(table)
}
