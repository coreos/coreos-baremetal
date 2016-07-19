package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	logtest "github.com/Sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	"github.com/mikeynap/coreos-baremetal/bootcfg/storage/storagepb"
	fake "github.com/mikeynap/coreos-baremetal/bootcfg/storage/testfakes"
)

func TestIPXEInspect(t *testing.T) {
	h := ipxeInspect()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, ipxeBootstrap, w.Body.String())
}

func TestIPXEHandler(t *testing.T) {
	logger, _ := logtest.NewNullLogger()
	srv := NewServer(&Config{Logger: logger})
	h := srv.ipxeHandler()
	ctx := withProfile(context.Background(), fake.Profile)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(ctx, w, req)
	// assert that:
	// - the Profile's NetBoot config is rendered as an iPXE script
	expectedScript := `#!ipxe
kernel /image/kernel a=b c
initrd /image/initrd_a /image/initrd_b 
boot
`
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, expectedScript, w.Body.String())
}

func TestIPXEHandler_MissingCtxProfile(t *testing.T) {
	logger, _ := logtest.NewNullLogger()
	srv := NewServer(&Config{Logger: logger})
	h := srv.ipxeHandler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(context.Background(), w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestIPXEHandler_RenderTemplateError(t *testing.T) {
	logger, _ := logtest.NewNullLogger()
	srv := NewServer(&Config{Logger: logger})
	h := srv.ipxeHandler()
	// a Profile with nil NetBoot forces a template.Execute error
	ctx := withProfile(context.Background(), &storagepb.Profile{Boot: nil})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(ctx, w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestIPXEHandler_WriteError(t *testing.T) {
	logger, _ := logtest.NewNullLogger()
	srv := NewServer(&Config{Logger: logger})
	h := srv.ipxeHandler()
	ctx := withProfile(context.Background(), fake.Profile)
	w := NewUnwriteableResponseWriter()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(ctx, w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Empty(t, w.Body.String())
}
