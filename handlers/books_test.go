package handlers

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"io/ioutil"
	"github.com/go-kit/kit/log"
	"bytes"
	"encoding/json"
	"os"
)


type a struct {
	r *Router
	token string
}
var app a




func Initialize(t *testing.T) {
	app.r = NewMux()
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	Config.Logger = logger
}


func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	app.r.ServeHTTP(rr, req)

	return rr
}

func authInit(t *testing.T) {
	auth := authCredentials{Login: "admin", Password: "admin"}
	authData, err := json.Marshal(auth)
	if err != nil {
		t.Error(err)
	}
	req, err := http.NewRequest("POST", "/api/auth", bytes.NewBuffer(authData))

	if err != nil {
		t.Error(err)
	}

	rr := executeRequest(req)


	data, err := ioutil.ReadAll(rr.Body)

	if err != nil {
		t.Error(err)
	}

	token := jwtToken{}
	err = json.Unmarshal(data, &token)

	if err != nil {
		t.Error(err)
	}
	app.token = "Bearer " + token.Jwt
}


func TestGetBookInstances(t *testing.T) {

	Initialize(t)
	authInit(t)

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	Config.Logger = logger


	req, err := http.NewRequest("GET", "/api/books", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Authorization", app.token)
	recorder := executeRequest(req)

	if err != nil {
		t.Error(err)
	}
	data, err := ioutil.ReadAll(recorder.Body)

	wantString := `[{"instance_id":1,"name":"The C Programming Language","year":2017,"author":[{"first_name":"Dennis","last_name":"Ritchie"},{"first_name":"Brian","last_name":"Kernighan"}],"publisher":"dmk press"},` +
		`{"instance_id":2,"name":"The C Programming Language","year":2017,"author":[{"first_name":"Dennis","last_name":"Ritchie"},{"first_name":"Brian","last_name":"Kernighan"}],"publisher":"dmk press"},` +
		`{"instance_id":3,"name":"The C Programming Language","year":2017,"author":[{"first_name":"Dennis","last_name":"Ritchie"},{"first_name":"Brian","last_name":"Kernighan"}],"publisher":"dmk press"},` +
		`{"instance_id":4,"name":"Современное проектирование на C++","year":2015,"author":[{"first_name":"Андрей","last_name":"Александреску"}],"publisher":"williams"},` +
		`{"instance_id":5,"name":"бог как илюзия","year":2016,"author":[{"first_name":"Ричард","last_name":"Докинз"}],"publisher":"аст"},` +
		`{"instance_id":6,"name":"бог как илюзия","year":2016,"author":[{"first_name":"Ричард","last_name":"Докинз"}],"publisher":"аст"}]`

	resultString := string(data)

	if resultString != wantString {
		t.Errorf("invalid result data: get %s\n want %s", resultString, wantString)
	}
}


//func TestAddHandlerArgs(t *testing.T) {
//
//	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
//
//	Config.Logger = logger
//
//
//
//	ts := httptest.NewServer(http.HandlerFunc(cardAddHandler))
//
//	//cardJson := []byte(`[{"instance_id": 1}, {"instance_id": 2}, {"instance_id": 4}]`)
//	cardJson := []byte(`[]`)
//	resp, err := http.Post(ts.URL, "application/json", bytes.NewBuffer(cardJson))
//	if err != nil {
//		//log.Print(err)
//	}
//	data, err := ioutil.ReadAll(resp.Body)
//	defer resp.Body.Close()
//
//	if err != nil {
//		t.Error(err)
//	}
//
//	wantString := "{}"
//	resultString := string(data)
//	statusCode := resp.StatusCode
//
//	if resultString != wantString || statusCode != http.StatusBadRequest {
//		t.Errorf("cardAddHandler: get = $s, want = $s\n", resultString, wantString)
//	}
//
//
//	ts = httptest.NewServer(http.HandlerFunc(cardAddHandler))
//	cardJson = []byte(`[]`)
//	resp, err = http.Post(ts.URL + "?id=1", "application/json", bytes.NewBuffer(cardJson))
//	if err != nil {
//		t.Error(err)
//	}
//
//	data, err = ioutil.ReadAll(resp.Body)
//	defer resp.Body.Close()
//
//	if err != nil {
//		t.Error(err)
//	}
//
//
//	wantString = "{}"
//	resultString = string(data)
//	statusCode = resp.StatusCode
//
//	if resultString != wantString || statusCode != http.StatusBadRequest {
//		t.Errorf("cardAddHandler: get = $s, want = $s\n", resultString, wantString)
//	}
//
//
//
//
//
//}
//
//
//func TestCardQueryOperationArgs(t *testing.T) {
//	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
//
//	Config.Logger = logger
//	ts := httptest.NewServer(http.HandlerFunc(cardQueryOperation))
//
//	resp, err := http.Get(ts.URL)
//
//	if err != nil {
//		t.Error(err)
//	}
//
//	data, err := ioutil.ReadAll(resp.Body)
//	defer resp.Body.Close()
//
//	wantString := ""
//
//	resultString := string(data)
//
//	if resultString != wantString {
//		t.Errorf("invalid result data: get %s\n want %s", resultString, wantString)
//	}
//}



func TestAddHandlerData(t *testing.T) {
	//logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	Initialize(t)
	authInit(t)

	//Config.Logger = logger
	addHandlerTest(t)
	queryOperationTest(t)
	queryCardCheckoutTest(t)
	clientsReturnBook(t)


}

func addHandlerTest(t *testing.T) {

	cardJson := []byte(`[{"instance_id":1},{"instance_id":2},{"instance_id":4}]`)
	req := httptest.NewRequest("POST", "/api/card/1", bytes.NewReader(cardJson))
	req.Header.Set("Authorization", app.token)

	rr := httptest.NewRecorder()
	app.r.ServeHTTP(rr, req)


	data, err := ioutil.ReadAll(rr.Body)

	if err != nil {
		t.Error(err)
	}

	wantString := `{"message":"data added"}`
	resultString := string(data)
	statusCode := rr.Code
	//t.Error("status code ", statusCode)

	if resultString != wantString || statusCode != http.StatusOK {
		t.Errorf("cardAddHandler: get = $s, want = $s\n", resultString, wantString)
	}
}

func queryOperationTest(t *testing.T) {

	req, err := http.NewRequest("GET", "/api/card/1", nil)
	if err != nil {
		t.Error(err)
	}

	req.Header.Set("Authorization", app.token)
	rr := httptest.NewRecorder()
	app.r.ServeHTTP(rr, req)

	data, err := ioutil.ReadAll(rr.Body)
	statusCode := rr.Code


	wantString := `[{"instance_id":1,"name":"The C Programming Language","year":2017,"author":[{"first_name":"Dennis","last_name":"Ritchie"},{"first_name":"Brian","last_name":"Kernighan"}],"publisher":"dmk press"},` +
		`{"instance_id":2,"name":"The C Programming Language","year":2017,"author":[{"first_name":"Dennis","last_name":"Ritchie"},{"first_name":"Brian","last_name":"Kernighan"}],"publisher":"dmk press"},` +
		`{"instance_id":4,"name":"Современное проектирование на C++","year":2015,"author":[{"first_name":"Андрей","last_name":"Александреску"}],"publisher":"williams"}]`

	resultString := string(data)

	if resultString != wantString || statusCode != http.StatusOK {
		t.Errorf("invalid result data:\n get %s\n want %s\n", resultString, wantString)
	}
}



func queryCardCheckoutTest(t *testing.T) {

	cardJson := []byte(`[{"instance_id":1},{"instance_id":2},{"instance_id":4}]`)
	req, err := http.NewRequest("POST", "/api/card/1/1", bytes.NewReader(cardJson))
	if err != nil {
		t.Error(err)
	}

	req.Header.Set("Authorization", app.token)
	rr := httptest.NewRecorder()
	app.r.ServeHTTP(rr, req)

	statusCode := rr.Code

	if  statusCode != http.StatusOK {
		t.Errorf("invalid result data:\n get status Code %s\n want %s\n", rr.Code, statusCode)
	}
}




func clientsReturnBook(t *testing.T) {

	cardJson := []byte(`[{"instance_id":1},{"instance_id":2},{"instance_id":4}]`)
	req, err := http.NewRequest("POST", "/api/clients/1", bytes.NewReader(cardJson))
	if err != nil {
		t.Error(err)
	}

	req.Header.Set("Authorization", app.token)
	rr := httptest.NewRecorder()
	app.r.ServeHTTP(rr, req)

	statusCode := rr.Code

	if  statusCode != http.StatusOK {
		t.Errorf("invalid result data:\n get status Code %s\n want %s\n", rr.Code, statusCode)
	}
}
