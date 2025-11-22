package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// код писать тут

const testToken = "test-token"

type TestCase struct {
	NameCase    string
	AccessToken string
	Request     *SearchRequest
	Result      *SearchResponse
	IsError     bool
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("AccessToken") != testToken {
		http.Error(w, "Bad AccessToken", http.StatusUnauthorized)
		return
	}

	r.URL.Query().Get("param")
}

func TestFindUsers(t *testing.T) {
	cases := []TestCase{
		TestCase{
			NameCase:    "Limit less than zero",
			AccessToken: testToken,
			Request: &SearchRequest{
				Limit: -1,
			},
			Result:  nil,
			IsError: true,
		},
		TestCase{
			NameCase:    "Offset less than zero",
			AccessToken: testToken,
			Request: &SearchRequest{
				Offset: -1,
			},
			Result:  nil,
			IsError: true,
		},
		TestCase{
			NameCase:    "Unauthorized",
			AccessToken: "invalid-token",
			Request:     &SearchRequest{},
			Result:      nil,
			IsError:     true,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	for caseNum, item := range cases {
		c := &SearchClient{
			AccessToken: item.AccessToken,
			URL:         ts.URL,
		}
		result, err := c.FindUsers(*item.Request)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if !reflect.DeepEqual(item.Result, result) {
			t.Errorf("[%d, %s] wrong result, expected %#v, got %#v", caseNum, item.NameCase, item.Result, result)
		}
	}

}
