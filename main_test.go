package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPHandlers(t *testing.T) {
	testQueries := []struct {
		testName    string
		endpoint    string
		endpointURL string
		success     bool
		msg         string
		statusCode  int
	}{
		{"newAdd(https)", "add", "/add/?url=https://www.google.com", true, "", 200},
		{"newAdd(http)", "add", "/add/?url=http://www.google.com", true, "", 200},
		{"emptyQuery", "add", "/add/?", false, "URL Missing from request", 200},
		{"invalidURL", "add", "/add/?url=notreal", false, "Unable to get requested url", 200},
		{"existingAdd", "add", "/add/?url=https://www.google.com", true, "", 200},
		{"getValid", "get", "/u/", true, "", 301},
		{"getInvalid", "get", "/u/", true, "", 404},
		{"duelParamaters", "remove", "/remove/?url=https://www.google.com&purl=false", false, "Expected purl or url got both", 200},
		{"removeInvalidURL", "remove", "/remove/?url=invalid", false, "URL not found", 200},
		{"removeInvalidPURL", "remove", "/remove/?purl=invalid", false, "Pocket URL not found", 200},
		{"RemoveValidURL", "remove", "/remove/?url=http://www.google.com", true, "", 200},
		{"RemoveValidPURL", "remove", "/remove/?purl=", true, "", 200},
	}
	var validURL string
	for _, v := range testQueries {
		t.Run(v.testName, func(t *testing.T) {
			switch v.endpoint {
			case "add":
				var resp response
				req, err := http.NewRequest("GET", v.endpointURL, nil)
				if err != nil {
					t.Fatal(err)
				}
				rr := httptest.NewRecorder()
				handler := http.HandlerFunc(addURL)

				handler.ServeHTTP(rr, req)

				if rr.Code != v.statusCode {
					t.Errorf("Endpoint: %v returned status code %v expected %v", v.endpointURL, rr.Code, v.statusCode)
				}
				err = json.Unmarshal([]byte(rr.Body.String()), &resp)
				if err != nil {
					t.Fatal(err)
				}
				if resp.Success != v.success {
					t.Errorf("Expected Success value %v got %v", v.success, resp.Success)
				}
				if resp.Success && resp.PocketURL == "" {
					t.Errorf("Success true but PocketURL contains no data")
				}
				if resp.Msg != v.msg {
					t.Errorf("Expected msg value %v got %v", v.msg, resp.Msg)
				}
				if v.testName == "newAdd(https)" {
					validURL = resp.PocketURL
				}
				if v.testName == "getValid" && resp.PocketURL != validURL {
					t.Errorf("PocketURL should be the same as newAdd expected %v got %v", validURL, resp.PocketURL)
				}
			case "get":
				if v.testName == "getValid" {
					v.endpointURL = v.endpointURL + validURL
				}
				req, err := http.NewRequest("GET", v.endpointURL, nil)
				if err != nil {
					t.Fatal(err)
				}
				rr := httptest.NewRecorder()
				handler := http.HandlerFunc(getURL)
				handler.ServeHTTP(rr, req)

				if rr.Code != v.statusCode {
					t.Errorf("Endpoint: %v returned status code %v expected %v", v.endpointURL, rr.Code, v.statusCode)
				}
			case "remove":
				var resp response
				if v.testName == "RemoveValidPURL" {
					v.endpointURL = v.endpointURL + validURL
				}
				req, err := http.NewRequest("GET", v.endpointURL, nil)
				if err != nil {
					t.Fatal(err)
				}
				rr := httptest.NewRecorder()
				handler := http.HandlerFunc(removeURL)

				handler.ServeHTTP(rr, req)

				if rr.Code != http.StatusOK {
					t.Errorf("Endpoint: %v returned status code %v expected 200", v.endpointURL, rr.Code)
				}
				err = json.Unmarshal([]byte(rr.Body.String()), &resp)
				if err != nil {
					t.Fatal(err)
				}
				if resp.Success != v.success {
					t.Errorf("Expected Success value %v got %v", v.success, resp.Success)
				}
				if resp.Success && resp.PocketURL == "" {
					t.Errorf("Success true but PocketURL contains no data")
				}
				if resp.Msg != v.msg {
					t.Errorf("Expected msg value %v got %v", v.msg, resp.Msg)
				}
				if v.testName == "newAdd" {
					validURL = resp.PocketURL
				}
				if v.testName == "getValid" && resp.PocketURL != validURL {
					t.Errorf("PocketURL should be the same as newAdd expected %v got %v", validURL, resp.PocketURL)
				}
			default:
				t.Errorf("Unexpected endpoint type %v", v.endpoint)
			}
		})
	}
}
