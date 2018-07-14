// main_test.go
// a simple test, to check all our endpoint
package main

import (
	"os"
	"log"
	"testing"
	"net/http"
	"net/http/httptest"
	"github.com/dieehard/forex-be/lib"
	"encoding/json"
	"bytes"
	"fmt"
)

var a lib.AppHandler

func TestMain(m *testing.M) {
	a = lib.AppHandler{}

	// check db connection
	a.Initialize(
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_DATABASE"))

	checkTableExist()

	code := m.Run()

	os.Exit(code)
}

func TestTruncateTable(t *testing.T) {
	truncateTable()

	req, _ := http.NewRequest("GET", "/rates", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetRateNotFound(t *testing.T) {
	truncateTable()

	req, _ := http.NewRequest("GET", "/rate/9999", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Rate not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'User not found'. Got '%s'", m["error"])
	}
}

func TestCreateRate(t *testing.T) {
	truncateTable()
	var from = "USD"
	var to = "IDR"

	payload := []byte(fmt.Sprintf(`{"from":"%s","to":"%s"}`, from , to))

	req, _ := http.NewRequest("POST", "/rate", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["from"] != from {
		t.Errorf("Expected 'from' = %s. Got '%v'", from, m["from"])
	}

	if m["to"] != to {
		t.Errorf("Expected 'to' = %s. Got '%v'", to, m["to"])
	}

	if m["id"] != 1.0 {
		t.Errorf("Expected rate id = '1'. Got '%v'", m["id"])
	}
}

func TestGetRate(t *testing.T) {
	var from = "USD"
	var to = "IDR"
	truncateTable()
	addNewRate(from, to)

	req, _ := http.NewRequest("GET", "/rate/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["from"] != from {
		t.Errorf("Expected 'from' = %s. Got '%v'", from, m["from"])
	}

	if m["to"] != to {
		t.Errorf("Expected 'to' = %s. Got '%v'", to, m["to"])
	}

	if m["id"] != 1.0 {
		t.Errorf("Expected rate id = '1'. Got '%v'", m["id"])
	}

}

func TestDeleteRateById(t *testing.T) {
	var from = "USD"
	var to = "IDR"

	truncateTable()
	addNewRate(from, to)

	req, _ := http.NewRequest("GET", "/rate/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/rate/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/rate/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestDeleteRateByFromTo(t *testing.T) {
	var from = "USD"
	var to = "IDR"

	truncateTable()
	addNewRate(from, to)

	req, _ := http.NewRequest("GET", "/rate/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	payload := []byte(fmt.Sprintf(`{"from":"%s","to":"%s"}`, from , to))

	req, _ = http.NewRequest("DELETE", "/rate", bytes.NewBuffer(payload))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/rate/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestDailyRate(t *testing.T) {
	truncateTable()

	var date = "2108-07-01"
	var from = "IDR"
	var to = "USD"
	var value = 0.000070
	addNewRate(from, to)

	payload := []byte(fmt.Sprintf(`{"date":"%s","from":"%s","to":"%s","rate":"%f"}`, date, from , to, value))

	req, _ := http.NewRequest("POST", "/rate/daily", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["date"] != date {
		t.Errorf("Expected 'date' = %s. Got '%v'", date, m["date"])
	}

	if m["from"] != from {
		t.Errorf("Expected 'from' = %s. Got '%v'", from, m["from"])
	}

	if m["to"] != to {
		t.Errorf("Expected 'to' = %s. Got '%v'", to, m["to"])
	}

	if m["rate"] != value {
		t.Errorf("Expected 'value' = '%f'. Got '%v'", value, m["rate"])
	}
}

func TestTrackRate(t *testing.T) {
	truncateTable()

	var date = "2108-07-01"
	var from = "IDR"
	var to = "USD"
	var value = 0.000070
	addNewRate(from, to)

	payload := []byte(fmt.Sprintf(`{"date":"%s","from":"%s","to":"%s","rate":"%f"}`, date, from , to, value))

	req, _ := http.NewRequest("POST", "/rate/daily", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	payload = []byte(fmt.Sprintf(`{ "date":"%s" }`, date))

	req, _ = http.NewRequest("POST", "/rate/track", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func addNewRate(from , to string) {
	statement := fmt.Sprintf("INSERT INTO rate(`from`, `to`) VALUES('%s', '%s')", from, to)
	a.DB.Exec(statement)
}

func checkTableExist() {
	if _, err := a.DB.Exec("SELECT * FROM rate"); err != nil {
		log.Fatal(err)
	}
}

func truncateTable() {
	a.DB.Exec("SET FOREIGN_KEY_CHECKS=0")
	a.DB.Exec("TRUNCATE TABLE rate")

	_, err := a.DB.Exec("TRUNCATE TABLE rate_data")
	if err != nil {
		log.Fatal(err)
	}
}
