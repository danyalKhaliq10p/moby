package authorization

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/docker/docker/pkg/plugingetter"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"

)

func TestMiddleware(t *testing.T) {
	pluginNames := []string{"testPlugin1", "testPlugin2"}
	var pluginGetter plugingetter.PluginGetter
	m := NewMiddleware(pluginNames, pluginGetter)
	authPlugins := m.GetAuthzPlugins()
	require.Equal(t, 2, len(authPlugins))
	require.EqualValues(t, pluginNames[0], authPlugins[0].Name())
	require.EqualValues(t, pluginNames[1], authPlugins[1].Name())

}

func TestMiddleware_WrapHandler(t *testing.T) {
	server := authZPluginTestServer{t: t}
	server.start()
	defer server.stop()

	authZPlugin := createTestPlugin(t)
	pluginNames := []string{authZPlugin.name}

	var pluginGetter plugingetter.PluginGetter
	m := NewMiddleware(pluginNames, pluginGetter)
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {

		return nil
	}

	authList := []Plugin{authZPlugin}
	m.SetPlugins([]string{"My Test Plugin"})
	m.SetAuthzPlugins(authList)
	h := m.WrapHandler(handler)
	require.NotNil(t, h)

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
		if err := h(ctx, resp, req, map[string]string{}); err != nil {
			t.Log(err.Error())
		}

	})

	t.Run("Positive Test Case :", func(t *testing.T) {
		server.replayResponse = Response{
			Allow: true,
			Msg:   "Server Auth Allowed",
		}
		if err := h(ctx, resp, req, map[string]string{}); err != nil {
			t.Log(err.Error())
		}

	})

}

func TestNewResponseModifier(t *testing.T) {
	r := httptest.NewRecorder()

	m := NewResponseModifier(r)
	m.Header().Set("h1", "v1")
	m.Write([]byte("body"))

	require.False(t, m.Hijacked())

	m.WriteHeader(http.StatusInternalServerError)
	require.NotNil(t, m.RawBody())

	x, err := m.RawHeaders()
	require.NotNil(t, x)
	require.Nil(t, err)

	m.Flush()
	m.FlushAll()

	if r.Header().Get("h1") != "v1" {
		t.Fatalf("Header value must exists %s", r.Header().Get("h1"), x)
	}

}
