package main

import (
	"city/model"
	"city/service"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

const tokenIdForAdmin = "admin000000122343456"
const tokenIdForPosters = "posters000000122343456"

var mongoDetails = service.Connection{}

func init() {
	mongoDetails.Server = "mongodb://localhost:27017"
	mongoDetails.Database = "Dummy"
	mongoDetails.Collection = "test"

	mongoDetails.Connect()
}

func addData(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	token := r.Header.Get("tokenid")

	admin := token == tokenIdForAdmin
	poster := token == tokenIdForPosters

	if !(admin || poster) {
		respondWithError(w, http.StatusBadRequest, "Unauthorized")
		return
	}

	if r.Method != "POST" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
		return
	}

	var cityData []model.CityData

	if err := json.NewDecoder(r.Body).Decode(&cityData); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	if inserted, err := mongoDetails.InsertAllData(cityData); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		respondWithJson(w, http.StatusAccepted, map[string]string{
			"message": strconv.Itoa(inserted) + " Record Inserted Successfully",
		})
	}
}

func deleteData(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	token := r.Header.Get("tokenid")

	if token != tokenIdForAdmin {
		respondWithError(w, http.StatusBadRequest, "Unauthorized")
		return
	}

	if r.Method != "DELETE" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
		return
	}

	var reqBody map[string]string
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
	}

	name := reqBody["name"]

	if deleted, err := mongoDetails.DeleteData(name); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		respondWithJson(w, http.StatusAccepted, map[string]string{
			"message": deleted,
		})
	}
}

func searchByCity(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "GET" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
		return
	}

	var reqBody map[string]string

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
	}

	name := reqBody["city"]

	if searchData, err := mongoDetails.SearchData(name); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		respondWithJson(w, http.StatusAccepted, searchData)
	}
}

func search(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "GET" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
		return
	}

	var reqBody model.Search

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
	}

	if searchData, err := mongoDetails.SearchDataByKeyAndValue(reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		respondWithJson(w, http.StatusAccepted, searchData)
	}
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}

func main() {
	http.HandleFunc("/add-data", addData)
	http.HandleFunc("/delete-data", deleteData)
	http.HandleFunc("/search-by-city", searchByCity)
	http.HandleFunc("/search", search)
	fmt.Println("Service Started at 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
