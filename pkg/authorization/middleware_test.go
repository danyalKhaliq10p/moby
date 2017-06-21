// +build !windows
package authorization

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/docker/docker/pkg/plugingetter"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestMiddleware(t *testing.T) {
	pluginNames := []string{"testPlugin1", "testPlugin2"}
	var pluginGetter plugingetter.PluginGetter
	m := NewMiddleware(pluginNames, pluginGetter)
	authPlugins := m.getAuthzPlugins()
	require.Equal(t, 2, len(authPlugins))
	require.EqualValues(t, pluginNames[0], authPlugins[0].Name())
	require.EqualValues(t, pluginNames[1], authPlugins[1].Name())
}

func TestMiddlewareWrapHandler(t *testing.T) {
	server := authZPluginTestServer{t: t}
	server.start()
	defer server.stop()

	authZPlugin := createTestPlugin(t)
	pluginNames := []string{authZPlugin.name}

	var pluginGetter plugingetter.PluginGetter
	middleWare := NewMiddleware(pluginNames, pluginGetter)
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {

		return nil
	}

	authList := []Plugin{authZPlugin}
	middleWare.SetPlugins([]string{"My Test Plugin"})
	middleWare.SetAuthzPlugins(authList)
	mdHandler := middleWare.WrapHandler(handler)
	require.NotNil(t, mdHandler)

	addr := "www.authz.com/auth"
	req, _ := http.NewRequest("GET", addr, nil)
	req.RequestURI = addr
	req.Header.Add("header", "value")

	resp := httptest.NewRecorder()
	ctx := context.Background()

	t.Run("Error Test Case :", func(t *testing.T) {
		server.replayResponse = Response{
			Allow: false,
			Msg:   "Server Auth Not Allowed",
		}
		if err := mdHandler(ctx, resp, req, map[string]string{}); err == nil {
			require.Error(t, err)
		}

	})

	t.Run("Positive Test Case :", func(t *testing.T) {
		server.replayResponse = Response{
			Allow: true,
			Msg:   "Server Auth Allowed",
		}
		if err := mdHandler(ctx, resp, req, map[string]string{}); err != nil {
			require.NoError(t, err)
		}

	})

}

func TestNewResponseModifier(t *testing.T) {
	recorder := httptest.NewRecorder()
	modifier := NewResponseModifier(recorder)
	modifier.Header().Set("H1", "V1")
	modifier.Write([]byte("body"))
	require.False(t, modifier.Hijacked())
	modifier.WriteHeader(http.StatusInternalServerError)
	require.NotNil(t, modifier.RawBody())

	raw, err := modifier.RawHeaders()
	require.NotNil(t, raw)
	require.Nil(t, err)

	headerData := strings.Split(strings.TrimSpace(string(raw)), ":")
	require.EqualValues(t, "H1", strings.TrimSpace(headerData[0]))
	require.EqualValues(t, "V1", strings.TrimSpace(headerData[1]))

	modifier.Flush()
	modifier.FlushAll()

	if recorder.Header().Get("H1") != "V1" {
		t.Fatalf("Header value must exists %s", recorder.Header().Get("H1"), raw)
	}

}

func (m *Middleware) SetAuthzPlugins(plugins []Plugin) {
	m.mu.Lock()
	m.plugins = plugins
	m.mu.Unlock()
}
