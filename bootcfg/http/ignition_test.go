package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	logtest "github.com/Sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	"github.com/mikeynap/coreos-baremetal/bootcfg/server"
	"github.com/mikeynap/coreos-baremetal/bootcfg/storage/storagepb"
	fake "github.com/mikeynap/coreos-baremetal/bootcfg/storage/testfakes"
)

var (
	expectedIgnitionV1 = `{"ignitionVersion":1,"storage":{},"systemd":{"units":[{"name":"etcd2.service","enable":true},{"name":"a1b2c3d4.service","enable":true}]},"networkd":{},"passwd":{}}`
	expectedIgnitionV2 = `{"ignition":{"version":"2.0.0","config":{}},"storage":{},"systemd":{"units":[{"name":"etcd2.service","enable":true},{"name":"a1b2c3d4.service","enable":true}]},"networkd":{},"passwd":{}}`
)

func TestIgnitionHandler_V2JSON(t *testing.T) {
	content := `{"ignition":{"version":"2.0.0","config":{}},"storage":{},"systemd":{"units":[{"name":"etcd2.service","enable":true},{"name":"a1b2c3d4.service","enable":true}]},"networkd":{},"passwd":{}}`
	profile := &storagepb.Profile{
		Id:         fake.Group.Profile,
		IgnitionId: "file.ign",
	}
	store := &fake.FixedStore{
		Profiles:        map[string]*storagepb.Profile{fake.Group.Profile: profile},
		IgnitionConfigs: map[string]string{"file.ign": content},
	}
	logger, _ := logtest.NewNullLogger()
	srv := NewServer(&Config{Logger: logger})
	c := server.NewServer(&server.Config{Store: store})
	h := srv.ignitionHandler(c)
	ctx := withGroup(context.Background(), fake.Group)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(ctx, w, req)
	// assert that:
	// - Ignition template is rendered with Group metadata and selectors
	// - Rendered Ignition template is parsed as JSON
	// - Ignition Config served as JSON
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, jsonContentType, w.HeaderMap.Get(contentType))
	assert.Equal(t, expectedIgnitionV2, w.Body.String())
}

func TestIgnitionHandler_V2YAML(t *testing.T) {
	content := `
systemd:
  units:
    - name: {{.service_name}}.service
      enable: true
    - name: {{.uuid}}.service
      enable: true
`
	store := &fake.FixedStore{
		Profiles:        map[string]*storagepb.Profile{fake.Group.Profile: testProfileIgnitionYAML},
		IgnitionConfigs: map[string]string{testProfileIgnitionYAML.IgnitionId: content},
	}
	logger, _ := logtest.NewNullLogger()
	srv := NewServer(&Config{Logger: logger})
	c := server.NewServer(&server.Config{Store: store})
	h := srv.ignitionHandler(c)
	ctx := withGroup(context.Background(), fake.Group)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(ctx, w, req)
	// assert that:
	// - Ignition template is rendered with Group metadata and selectors
	// - Rendered Ignition template is parsed as YAML
	// - Ignition Config served as JSON
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, jsonContentType, w.HeaderMap.Get(contentType))
	assert.Equal(t, expectedIgnitionV2, w.Body.String())
}

func TestIgnitionHandler_MissingCtxProfile(t *testing.T) {
	logger, _ := logtest.NewNullLogger()
	srv := NewServer(&Config{Logger: logger})
	c := server.NewServer(&server.Config{Store: &fake.EmptyStore{}})
	h := srv.ignitionHandler(c)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(context.Background(), w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestIgnitionHandler_MissingIgnitionConfig(t *testing.T) {
	logger, _ := logtest.NewNullLogger()
	srv := NewServer(&Config{Logger: logger})
	c := server.NewServer(&server.Config{Store: &fake.EmptyStore{}})
	h := srv.ignitionHandler(c)
	ctx := withProfile(context.Background(), fake.Profile)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(ctx, w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestIgnitionHandler_MissingTemplateMetadata(t *testing.T) {
	content := `
ignition_version: 1
systemd:
  units:
    - name: {{.missing_key}}
      enable: true
`
	store := &fake.FixedStore{
		Profiles:        map[string]*storagepb.Profile{fake.Group.Profile: fake.Profile},
		IgnitionConfigs: map[string]string{fake.Profile.IgnitionId: content},
	}
	logger, _ := logtest.NewNullLogger()
	srv := NewServer(&Config{Logger: logger})
	c := server.NewServer(&server.Config{Store: store})
	h := srv.ignitionHandler(c)
	ctx := withGroup(context.Background(), fake.Group)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(ctx, w, req)
	// assert that:
	// - Ignition template rendering errors because "missing_key" is not
	// present in the Group metadata
	assert.Equal(t, http.StatusNotFound, w.Code)
}
