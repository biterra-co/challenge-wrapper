package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestInjectsScriptIntoHTMLOnce(t *testing.T) {
	res := &http.Response{
		Header:        http.Header{"Content-Type": {"text/html; charset=utf-8"}, "Content-Length": {"40"}},
		Body:          io.NopCloser(strings.NewReader("<html><head></head><body>ok</body></html>")),
		ContentLength: 40,
	}
	if err := inject(res); err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	if strings.Count(string(body), "/__biterra__/challenge-wrapper.js") != 1 {
		t.Fatalf("unexpected body: %s", body)
	}
	if res.Header.Get("Content-Length") != "" || res.ContentLength != -1 {
		t.Fatal("content length was not cleared")
	}
	if err := inject(res); err != nil {
		t.Fatal(err)
	}
	body, _ = io.ReadAll(res.Body)
	if strings.Count(string(body), "/__biterra__/challenge-wrapper.js") != 1 {
		t.Fatalf("script injected twice: %s", body)
	}
}

func TestLeavesNonHTMLUntouched(t *testing.T) {
	res := &http.Response{Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"ok":true}`))}
	if err := inject(res); err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", body)
	}
}
