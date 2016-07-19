package http

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	logtest "github.com/Sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	"github.com/mikeynap/coreos-baremetal/bootcfg/storage/storagepb"
	fake "github.com/mikeynap/coreos-baremetal/bootcfg/storage/testfakes"
)

func TestMetadataHandler(t *testing.T) {
	logger, _ := logtest.NewNullLogger()
	srv := NewServer(&Config{Logger: logger})
	h := srv.metadataHandler()
	ctx := withGroup(context.Background(), fake.Group)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/?uuid=a1b2c3d4", nil)
	h.ServeHTTP(ctx, w, req)
	// assert that:
	// - the Group's custom metadata and selectors are served
	// - key names are upper case
	expectedData := map[string]string{
		"POD_NETWORK":  "10.2.0.0/16",
		"SERVICE_NAME": "etcd2",
		"UUID":         "a1b2c3d4",
	}
	assert.Equal(t, http.StatusOK, w.Code)
	// convert response (random order) to map (tests compare in order)
	assert.Equal(t, expectedData, metadataToMap(w.Body.String()))
	assert.Equal(t, plainContentType, w.HeaderMap.Get(contentType))
}

func TestMetadataHandler_MetadataEdgeCases(t *testing.T) {
	logger, _ := logtest.NewNullLogger()
	srv := NewServer(&Config{Logger: logger})
	h := srv.metadataHandler()
	// groups with different metadata
	cases := []struct {
		group    *storagepb.Group
		expected string
	}{
		{&storagepb.Group{Metadata: []byte(`{"num":3}`)}, "NUM=3\n"},
		{&storagepb.Group{Metadata: []byte(`{"yes":true}`)}, "YES=true\n"},
		{&storagepb.Group{Metadata: []byte(`{"no":false}`)}, "NO=false\n"},
		// Issue #84 - improve list and map printouts
		{&storagepb.Group{Metadata: []byte(`{"list":["3","d"]}`)}, "LIST=[3 d]\n"},
	}
	for _, c := range cases {
		ctx := withGroup(context.Background(), c.group)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		h.ServeHTTP(ctx, w, req)
		// assert that:
		// - the Group's custom metadata is served
		// - key names are upper case
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, c.expected, w.Body.String())
		assert.Equal(t, plainContentType, w.HeaderMap.Get(contentType))
	}
}

func TestMetadataHandler_MissingCtxGroup(t *testing.T) {
	logger, _ := logtest.NewNullLogger()
	srv := NewServer(&Config{Logger: logger})
	h := srv.metadataHandler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(context.Background(), w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// metadataToMap converts a KEY=val\nKEY=val ResponseWriter body to a map for
// testing purposes.
func metadataToMap(metadata string) map[string]string {
	scanner := bufio.NewScanner(strings.NewReader(metadata))
	data := make(map[string]string)
	for scanner.Scan() {
		token := scanner.Text()
		pair := strings.SplitN(token, "=", 2)
		if len(pair) != 2 {
			continue
		}
		data[pair[0]] = pair[1]
	}
	return data
}
