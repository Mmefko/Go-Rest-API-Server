package backend_test

import (
	"bytes"
	"os"
	"strconv"
	"testing"

	"encoding/json"
	"linkedin/internal/backend"
	"log"
	"net/http"
	"net/http/httptest"
)

var a backend.App

const tableProductCreationQuery = `CREATE TABLE IF NOT EXISTS product
(
	ID INTEGER PRIMARY KEY,
	productCode TEXT NOT NULL,
	name TEXT NOT NULL,
	inventory INTEGER NOT NULL,
	price INTEGER NOT NULL,
	status TEXT NOT NULL
)`

const tableOrderCreationQuery = `CREATE TABLE IF NOT EXISTS orders
(
	id INTEGER PRIMARY KEY,
	customerName VARCHAR(256) NOT NULL,
	total INTEGER NOT NULL,
	status VARCHAR(64) NOT NULL
)`

const tableOrderItemCreationQuery = `CREATE TABLE IF NOT EXISTS order_items
(
	order_id INTEGER,
	product_id INTEGER,
	quantity INTEGER NOT NULL,
	FOREIGN KEY (order_id) REFERENCES orders (id),
	FOREIGN KEY (product_id) REFERENCES products (id),
	PRIMARY KEY (order_id, product_id)
)`

func TestMain(m *testing.M) {
	a = backend.App{}
	a.Initialize()
	ensureTableExists()
	code := m.Run()

	clearProductTable()
	clearOrderTable()
	clearOrderItemTable()
	os.Exit(code)
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableProductCreationQuery); err != nil {
		log.Fatal(err)
	}
	if _, err := a.DB.Exec(tableOrderCreationQuery); err != nil {
		log.Fatal(err)
	}
	if _, err := a.DB.Exec(tableOrderItemCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearProductTable() {
	a.DB.Exec("DELETE FROM product")
	a.DB.Exec("DELETE FROM sqlite_sequence WHERE name = 'product'")
}

func clearOrderTable() {
	a.DB.Exec("DELETE FROM orders")
	a.DB.Exec("DELETE FROM sqlite_sequence WHERE name = 'orders'")
}

func clearOrderItemTable() {
	a.DB.Exec("DELETE FROM order_items")
}

func TestGetNonExistentProduct(t *testing.T) {
	clearProductTable()

	req, _ := http.NewRequest("GET", "/product/11", nil)
	responce := executeRequest(req)

	checkResponseCode(t, http.StatusInternalServerError, responce.Code)

	var m map[string]string
	json.Unmarshal(responce.Body.Bytes(), &m)
	if m["error"] != "sql: no rows in result set" {
		t.Errorf("Expected the 'error' key of the response to be set to 'sql: no rows in result set'. Got '[%s]'", m["error"])
	}
}

func TestCreateProduct(t *testing.T) {
	clearProductTable()

	payload := []byte(`{"productCode":"TEST12345","name":"ProductTest","inventory":1,"price":1,"status":"testing"}`)

	req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["productCode"] != "TEST12345" {
		t.Errorf("Expected productCode to be 'TEST12345'. Got '%v'", m["productCode"])
	}
	if m["name"] != "ProductTest" {
		t.Errorf("Expected name to be 'ProductTest'. Got '%v'", m["name"])
	}
	if m["inventory"] != 1.0 {
		t.Errorf("Expected inventory to be '1'. Got '%v'", m["inventory"])
	}
	if m["price"] != 1.0 {
		t.Errorf("Expected price to be '1'. Got '%v'", m["price"])
	}
	if m["status"] != "testing" {
		t.Errorf("Expected status to be 'testing'. Got '%v'", m["status"])
	}
	if m["ID"] != 1.0 {
		t.Errorf("Expected id to be '1'. Got '%v'", m["ID"])
	}

}

func TestGetProduct(t *testing.T) {
	clearProductTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func addProducts(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO product (productCode, name, inventory, price, status) VALUES(?, ?, ?, ?, ?)", "ABC123"+strconv.Itoa(i), "Product"+strconv.Itoa(i), i, i, "test"+strconv.Itoa(i))
	}
}

func TestGetNonExistentOrder(t *testing.T) {
	clearOrderTable()

	req, _ := http.NewRequest("GET", "/order/11", nil)
	responce := executeRequest(req)

	checkResponseCode(t, http.StatusInternalServerError, responce.Code)

	var m map[string]string
	json.Unmarshal(responce.Body.Bytes(), &m)
	if m["error"] != "sql: no rows in result set" {
		t.Errorf("Expected the 'error' key of the response to be set to 'sql: no rows in result set'. Got '[%s]'", m["error"])
	}
}

func TestCreateOrder(t *testing.T) {
	clearOrderTable()

	payload := []byte(`{"customerName":"customerTest","total":1,"status":"testing"}`)

	req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["customerName"] != "customerTest" {
		t.Errorf("Expected productCode to be 'customerTest'. Got '%v'", m["customerName"])
	}
	if m["total"] != 1.0 {
		t.Errorf("Expected total to be '1'. Got '%v'", m["total"])
	}
	if m["status"] != "testing" {
		t.Errorf("Expected status to be 'testing'. Got '%v'", m["status"])
	}
	if m["id"] != 1.0 {
		t.Errorf("Expected id to be '1'. Got '%v'", m["id"])
	}
}

func TestGetOrder(t *testing.T) {
	clearOrderTable()
	addOrder(1)

	req, _ := http.NewRequest("GET", "/order/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func addOrder(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO orders (customerName, total, status) VALUES(?, ?, ?)", "Customer"+strconv.Itoa(i), i, "test"+strconv.Itoa(i))
	}
}

func TestCreateOrderItem(t *testing.T) {
	clearOrderItemTable()

	addProducts(1)
	addOrder(1)

	payload := []byte(`[{"order_id":1,"product_id":1,"quantity":1}]`)

	req, _ := http.NewRequest("POST", "/orderitems", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m [](map[string]interface{})
	json.Unmarshal(response.Body.Bytes(), &m)

	if m[0]["order_id"] != 1.0 {
		t.Errorf("Expected order_id to be '1'. Got '%v'", m[0]["order_id"])
	}
	if m[0]["product_id"] != 1.0 {
		t.Errorf("Expected product_id to be '1'. Got '%v'", m[0]["product_id"])
	}
	if m[0]["quantity"] != 1.0 {
		t.Errorf("Expected status to be '1'. Got '%v'", m[0]["quantity"])
	}
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
