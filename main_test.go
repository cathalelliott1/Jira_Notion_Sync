package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var baseURL string

func TestFetchIssues(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := IssueResponse{Issues: []Issue{{Key: "TU-1"}, {Key: "TU-2"}}}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Fatalf("Error encoding response: %v", err)
		}
	}))

	defer ts.Close()

	baseURL = ts.URL

	issues, _, err := fetchIssues("encodedCredentials", baseURL)
	if err != nil {
		t.Fatalf("Error fetching issues: %v", err)
	}

	expectedIssues := []Issue{{Key: "TU-1"}, {Key: "TU-2"}}
	if len(issues) != len(expectedIssues) {
		t.Errorf("Expected %d issues, got %d", len(expectedIssues), len(issues))
	}

	for i, issue := range issues {
		if issue.Key != expectedIssues[i].Key {
			t.Errorf("Expected issue %d key to be %s, got %s", i, expectedIssues[i].Key, issue.Key)
		}
	}
}

func TestUpdateCustomField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	defer ts.Close()

	statusCode, _, err := updateCustomField("TU-1", "customfield_10506", map[string]interface{}{"id": "11755"}, "encodedCredentials", ts.URL)
	if err != nil {
		t.Fatalf("Error updating custom field: %v", err)
	}

	if statusCode != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, statusCode)
	}
}

func TestMain(m *testing.M) {
	os.Setenv("CI", "false")

	exitVal := m.Run()

	os.Unsetenv("CI")

	os.Exit(exitVal)
}

func TestFetchIssuesWithMockJira(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := IssueResponse{Issues: []Issue{{Key: "JIRA-1"}, {Key: "JIRA-2"}}}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Fatalf("Error encoding response: %v", err)
		}
	}))

	defer ts.Close()

	issues, _, err := fetchIssues("encodedCredentials", ts.URL)
	if err != nil {
		t.Fatalf("Error fetching issues: %v", err)
	}

	expectedIssues := []Issue{{Key: "JIRA-1"}, {Key: "JIRA-2"}}
	if len(issues) != len(expectedIssues) {
		t.Errorf("Expected %d issues, got %d", len(expectedIssues), len(issues))
	}

	for i, issue := range issues {
		if issue.Key != expectedIssues[i].Key {
			t.Errorf("Expected issue %d key to be %s, got %s", i, expectedIssues[i].Key, issue.Key)
		}
	}
}

func TestFetchIssuesErrorHandling(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte{})
		if err != nil {
			t.Fatalf("Error writing response: %v", err)
		}
	}))

	defer ts.Close()

	baseURL = ts.URL

	_, _, err := fetchIssues("encodedCredentials", baseURL)
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	expectedError := "HTTP error! Status: 404, Body: "
	if got := err.Error(); got != expectedError {
		t.Errorf("Expected %q, got %q", expectedError, got)
	}
}

func TestSetCommonHeaders(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	encodedCredentials := "encodedCredentials"
	setCommonHeaders(req, encodedCredentials)

	if req.Header.Get("Authorization") != "Basic "+encodedCredentials {
		t.Errorf("Expected Authorization header to be %s, got %s", "Basic "+encodedCredentials, req.Header.Get("Authorization"))
	}

	if req.Header.Get("Accept") != "application/json" {
		t.Errorf("Expected Accept header to be %s, got %s", "application/json", req.Header.Get("Accept"))
	}

	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type header to be %s, got %s", "application/json", req.Header.Get("Content-Type"))
	}
}

func TestUpdateCustomFieldErrorHandling(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte{})
		if err != nil {
			t.Fatalf("Error writing response: %v", err)
		}
	}))

	defer ts.Close()

	statusCode, _, err := updateCustomField("TU-1", "customfield_10506", map[string]interface{}{"id": "11755"}, "encodedCredentials", ts.URL)
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	if statusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, statusCode)
	}

	expectedError := "HTTP error! Status: 500, Body: "
	if got := err.Error(); got != expectedError {
		t.Errorf("Expected %q, got %q", expectedError, got)
	}
}
