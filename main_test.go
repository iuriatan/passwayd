package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMainHandler(t *testing.T) {
	var testCases = []struct {
		name    string
		method  string
		body    io.Reader
		url     string
		xStatus int
		xLogs   []string
	}{
		{
			name:    "fetch unknown key",
			method:  http.MethodGet,
			body:    nil,
			url:     "/unknown",
			xStatus: http.StatusNotFound,
			xLogs:   []string{"Request key `unknown` from", ": not found"},
		},
		{
			name:    "set key should fail if URL.path is not root",
			method:  http.MethodPost,
			body:    strings.NewReader(`{"name":"testkey","ip":"192.168.0.97"}`),
			url:     "/invalid",
			xStatus: http.StatusBadRequest,
			xLogs:   []string{},
		},
		{
			name:    "set key should fail when key is too long",
			method:  http.MethodPost,
			body:    strings.NewReader(`{"name":"testkeyisaveryveryveryveryveryveryveryveryveryveryverylongkey","ip":"192.168.0.97"}`),
			url:     "/",
			xStatus: http.StatusBadRequest,
			xLogs:   []string{},
		},
		{
			name:    "set key ip only",
			method:  http.MethodPost,
			body:    strings.NewReader(`{"name":"testkey","ip":"192.168.0.97"}`),
			url:     "/",
			xStatus: http.StatusOK,
			xLogs:   []string{"Update from ", ": (testkey) 192.168.0.97"},
		},
		{
			name:    "fetch test key",
			method:  http.MethodGet,
			body:    nil,
			url:     "/testkey",
			xStatus: http.StatusOK,
			xLogs:   []string{"Request key `testkey` from", ": 12 bytes"},
		},
		{
			name:    "set key ip and port",
			method:  http.MethodPost,
			body:    strings.NewReader(`{"name":"testkey2","ip":"192.168.0.97","port":"1984"}`),
			url:     "/",
			xStatus: http.StatusOK,
			xLogs:   []string{"Update from ", ": (testkey2) 192.168.0.97:1984"},
		},
		{
			name:    "fetch test key2",
			method:  http.MethodGet,
			body:    nil,
			url:     "/testkey2",
			xStatus: http.StatusOK,
			xLogs:   []string{"Request key `testkey2` from", ": 17 bytes"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.url, tc.body)
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(os.Stderr)

			handler(rr, req)

			require.EqualValues(t, tc.xStatus, rr.Result().StatusCode)
			logged := buf.String()
			for _, str := range tc.xLogs {
				require.Contains(t, logged, str)
			}
		})
	}
}
