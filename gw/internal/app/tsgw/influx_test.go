package tsgw

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewInfluxGw(t *testing.T) {
	params := AppParams{
		InfluxUrl:   "http://localhost:8086",
		InfluxToken: "my-token",
		AppPort:     8080,
		AppUser:     "admin",
		AppPass:     "password",
	}

	gw, err := NewInfluxGw(params)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Assert that the InfluxGw struct is initialized correctly
	if gw.server == nil {
		t.Error("expected server to be initialized")
	}
	if gw.ctx == nil {
		t.Error("expected ctx to be initialized")
	}
	if gw.influx == nil {
		t.Error("expected influx to be initialized")
	}
	if gw.InfluxUrl != params.InfluxUrl {
		t.Errorf("expected InfluxUrl to be %s, got %s", params.InfluxUrl, gw.InfluxUrl)
	}
	if gw.InfluxToken != params.InfluxToken {
		t.Errorf("expected InfluxToken to be %s, got %s", params.InfluxToken, gw.InfluxToken)
	}
	if gw.AppPort != params.AppPort {
		t.Errorf("expected AppPort to be %d, got %d", params.AppPort, gw.AppPort)
	}
	if gw.AppUser != params.AppUser {
		t.Errorf("expected AppUser to be %s, got %s", params.AppUser, gw.AppUser)
	}
	if gw.AppPass != params.AppPass {
		t.Errorf("expected AppPass to be %s, got %s", params.AppPass, gw.AppPass)
	}
}

func TestInfluxGw_Run(t *testing.T) {
	params := AppParams{
		InfluxUrl:   "http://localhost:8086",
		InfluxToken: "my-token",
		AppPort:     8080,
		AppUser:     "admin",
		AppPass:     "password",
	}

	gw, err := NewInfluxGw(params)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Start the server in a separate goroutine
	go func() {
		err := gw.Run()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	// Send a request to the server
	req, err := http.NewRequest("GET", "http://localhost:8080", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	// Assert the response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Assert the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedBody := "Hello, World!"
	if string(body) != expectedBody {
		t.Errorf("expected response body %q, got %q", expectedBody, string(body))
	}

	// Stop the server
	err = gw.server.Shutdown(gw.ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInfluxGw_handlePost(t *testing.T) {
	params := AppParams{
		InfluxUrl:   "http://localhost:8086",
		InfluxToken: "my-token",
		AppPort:     8080,
		AppUser:     "admin",
		AppPass:     "password",
	}

	gw, err := NewInfluxGw(params)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Create a test request with a JSON payload
	payload := InfluxMsg{
		Measurement: "test_measurement",
		Tags: map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		},
		Fields: map[string]interface{}{
			"field1": 1,
			"field2": 2,
		},
		Ts: 1234567890,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:8080", bytes.NewReader(payloadBytes))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Call the handlePost function
	gw.handlePost(rr, req)

	// Assert the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// Assert the response body
	expectedBody := "OK"
	if rr.Body.String() != expectedBody {
		t.Errorf("expected response body %q, got %q", expectedBody, rr.Body.String())
	}
}

func TestInfluxGw_isAuthenticated(t *testing.T) {
	params := AppParams{
		InfluxUrl:   "http://localhost:8086",
		InfluxToken: "my-token",
		AppPort:     8080,
		AppUser:     "admin",
		AppPass:     "password",
	}

	gw, err := NewInfluxGw(params)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Create a test request with valid authentication header
	req, err := http.NewRequest("GET", "http://localhost:8080", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	req.SetBasicAuth(params.AppUser, params.AppPass)

	// Call the isAuthenticated function
	isAuth := gw.isAuthenticated(req)

	// Assert that the request is authenticated
	if !isAuth {
		t.Error("expected request to be authenticated")
	}

	// Create a test request without authentication header
	req, err = http.NewRequest("GET", "http://localhost:8080", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Call the isAuthenticated function
	isAuth = gw.isAuthenticated(req)

	// Assert that the request is not authenticated
	if isAuth {
		t.Error("expected request to be not authenticated")
	}
}

func TestPayload(t *testing.T) {
	// Create a test request with a JSON payload
	payload := InfluxMsg{
		Measurement: "test_measurement",
		Tags: map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		},
		Fields: map[string]interface{}{
			"field1": 1,
			"field2": 2,
		},
		Ts: 1234567890,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:8080", bytes.NewReader(payloadBytes))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Call the payload function
	msg, err := readPayload(req.Body)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Assert the parsed message
	if msg.Measurement != payload.Measurement {
		t.Errorf("expected measurement %q, got %q", payload.Measurement, msg.Measurement)
	}
	if len(msg.Tags) != len(payload.Tags) {
		t.Errorf("expected %d tags, got %d", len(payload.Tags), len(msg.Tags))
	}
	for key, value := range payload.Tags {
		if msg.Tags[key] != value {
			t.Errorf("expected tag %q to have value %q, got %q", key, value, msg.Tags[key])
		}
	}
	if len(msg.Fields) != len(payload.Fields) {
		t.Errorf("expected %d fields, got %d", len(payload.Fields), len(msg.Fields))
	}

	if msg.Ts != payload.Ts {
		t.Errorf("expected timestamp %d, got %d", payload.Ts, msg.Ts)
	}
}
