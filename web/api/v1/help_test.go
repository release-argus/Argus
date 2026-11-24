// Copyright [2026] [Argus]
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build unit || integration

package v1

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	_ "modernc.org/sqlite"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/auth/password"
	"github.com/release-argus/Argus/auth/provider"
	"github.com/release-argus/Argus/auth/provider/local"
	"github.com/release-argus/Argus/auth/rbac"
	"github.com/release-argus/Argus/auth/session"
	"github.com/release-argus/Argus/auth/store"
	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/config/decode"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	dvtest "github.com/release-argus/Argus/service/deployed_version/test"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	lvtest "github.com/release-argus/Argus/service/latest_version/test"
	lvbase "github.com/release-argus/Argus/service/latest_version/types/base"
	svctest "github.com/release-argus/Argus/service/test"
	"github.com/release-argus/Argus/util"
	whtest "github.com/release-argus/Argus/webhook/test"
)

// dummyGitHubToken stands in for a real token where no request is made.
const dummyGitHubToken = "dummy-github-token"

var (
	secretValueMarshaled string
	testTempDir          string // Process-lifetime home of test config files.
)

func TestMain(m *testing.M) {
	logtest.InitLog()

	ctx, cancel := context.WithCancel(context.Background())
	g, _ := errgroup.WithContext(ctx)

	// Shorten the WebSocket ping interval so writePump tests don't wait out the production default.
	pingPeriod = 10 * time.Millisecond
	// Shorten the pong wait so a missed pong ends readPump within a test.
	pongWait = 500 * time.Millisecond

	config.DebounceDuration = 500 * time.Millisecond
	flags := make(map[string]bool)
	tempDir, err := os.MkdirTemp("", "argus-api-v1-test")
	if err != nil {
		fmt.Printf("os.MkdirTemp() failed: %v\n", err)
		os.Exit(1)
	}
	testTempDir = tempDir
	path := filepath.Join(tempDir, "TestWebAPIv1Main.yml")
	testYAML_Argus(path)
	var cfg config.Config
	if !cfg.Load(ctx, g, path, &flags) {
		fmt.Printf("Config.Load(%q) failed\n", path)
		os.Exit(1)
	}

	// Marshal the secret value '<secret>' -> '\u003csecret\u003e'.
	secretValueMarshaledBytes, err := decode.Marshal("json", util.SecretValue)
	if err != nil {
		fmt.Printf("decode.Marshal() failed: %v\n", err)
		os.Exit(1)
	}
	secretValueMarshaled = string(secretValueMarshaledBytes)

	exitCode := m.Run()
	_ = os.RemoveAll(tempDir)
	_ = os.Remove(cfg.Settings.DataDatabaseFile())
	cancel()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Println("exit code channel not empty")
		exitCode = 1
	}

	os.Exit(exitCode)
}

type failWriter struct {
	header http.Header
	code   int
	body   string // What was handed to Write, though the write fails.
}

func (f *failWriter) Header() http.Header {
	if f.header == nil {
		f.header = make(http.Header)
	}
	return f.header
}

func (f *failWriter) WriteHeader(code int) {
	f.code = code
}

func (f *failWriter) Write(p []byte) (int, error) {
	if f.code == 0 {
		f.code = http.StatusOK
	}
	f.body += string(p)
	return 0, errors.New("write failed")
}

// plainDefaults returns plain defaults and hardDefaults for testing.
func plainDefaults() (*config.Defaults, *config.Defaults) {
	dockerDefaults, _ := docker.DecodeDefaults("yaml", nil, nil)
	defaults := config.Defaults{
		Service: service.Defaults{
			LatestVersion: latestver.Defaults{
				Common: lvbase.Defaults{
					Require: filter.RequireDefaults{
						Docker: *dockerDefaults,
					},
				},
			},
		},
	}
	hardDefaults := config.Defaults{}
	hardDefaults.Default()
	hardDefaults.Service.LatestVersion.GitHub.AccessToken = dummyGitHubToken
	defaults.SetDefaults(&hardDefaults)

	return &defaults, &hardDefaults
}

func testClient() Client {
	hub := NewHub()
	return Client{
		hub:  hub,
		ip:   "1.1.1.1",
		conn: &websocket.Conn{},
		send: make(chan []byte, 5),
	}
}

func testLoad(t *testing.T, file string) *config.Config {
	t.Helper()

	var cfg config.Config
	g, _ := errgroup.WithContext(t.Context())

	flags := make(map[string]bool)
	if !cfg.Load(t.Context(), g, file, &flags) {
		t.Fatalf("Config.Load(%q) failed", file)
	}
	// The SaveHandler flushes a pending save once the test ends, so wait for it.
	t.Cleanup(func() { _ = g.Wait() })
	announceChannel := make(chan []byte, 8)
	cfg.HardDefaults.Service.Status.AnnounceChannel = announceChannel

	return &cfg
}

func testAPI(t *testing.T, path string) API {
	t.Helper()
	if !filepath.IsAbs(path) {
		path = filepath.Join(testTempDir, path)
	}
	testYAML_Argus(path)

	cfg := testLoad(t, path)

	t.Cleanup(func() {
		_ = os.RemoveAll(cfg.Settings.DataDatabaseFile())
	})

	return API{Config: cfg}
}

func testService(
	t *testing.T,
	id string,
	lvType, dvType string,
	semVer bool,
) *service.Service {
	t.Helper()

	svcCfg := svctest.PlainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	svc := test.Must(t, func() (*service.Service, error) {
		return service.DecodeService(
			"yaml", []byte(test.TrimYAML(`
			options:
				semantic_versioning: `+strconv.FormatBool(semVer)+`
			latest_version:
			`+lvtest.Lookup(t, lvType, false).String("  ")+`
			deployed_version:
			`+dvtest.Lookup(t, dvType, false, "").String("  ")+`
			dashboard:
				icon: https://example.com/icon.png
				icon_link_to: https://example.com/icon-{{ version }}.png
				web_url: https://example.com/{{ version }}
		`)),
			id,
			svcCfg, notifyCfg, whCfg,
		)
	})

	announceChannel := make(chan []byte, 8)
	databaseChannel := make(chan dbtype.Message, 8)

	svc.Status.AnnounceChannel = announceChannel
	svc.Status.DatabaseChannel = databaseChannel

	return svc
}

// metricSeriesDeleter is a Prometheus vector whose series can be deleted by label.
type metricSeriesDeleter interface {
	DeletePartialMatch(labels prometheus.Labels) int
}

// assertNoMetricSeries fails for each of vecs holding a series for id, deleting any it finds.
func assertNoMetricSeries(t *testing.T, id string, vecs map[string]metricSeriesDeleter) {
	t.Helper()

	for name, vec := range vecs {
		if got := vec.DeletePartialMatch(prometheus.Labels{"id": id}); got != 0 {
			t.Errorf("metric.%s series count mismatch\ngot:  %d\nwant: 0", name, got)
		}
	}
}

func testCommand(failing bool) command.Command {
	if failing {
		return command.Command{"ls", "-lah", "/root"}
	}
	return command.Command{"ls", "-lah"}
}

func testFaviconSettings(png string, svg string) *config.FaviconSettings {
	if svg == "" && png == "" {
		return nil
	}

	return &config.FaviconSettings{
		SVG: svg,
		PNG: png,
	}
}

// testAuthServerPendingSetup builds a routed API with session/RBAC auth
// enabled, backed by an in-memory auth store with NO users yet (first-run
// setup pending). The returned *sql.DB is the auth store's handle (close to
// simulate infrastructure failure).
// opts mutate the config before the API is built, for settings the router
// reads at setup (route prefix, demo group).
func testAuthServerPendingSetup(
	t *testing.T,
	path string,
	opts ...func(*config.Config),
) (*API, *AuthDeps, *sql.DB) {
	t.Helper()

	if !filepath.IsAbs(path) {
		path = filepath.Join(testTempDir, path)
	}
	testYAML_Argus(path)
	cfg := testLoad(t, path)
	for _, opt := range opts {
		opt(cfg)
	}
	t.Cleanup(func() { _ = os.RemoveAll(cfg.Settings.DataDatabaseFile()) })

	dbConn := test.MemoryDB(t)

	authStore, err := store.New(t.Context(), dbConn)
	if err != nil {
		t.Fatalf("store.New() failed: %v", err)
	}

	sessions := session.New(
		authStore,
		session.Config{
			Lifetime:    time.Hour,
			IdleTimeout: time.Hour,
		},
	)
	registry := provider.NewRegistry()
	registry.Register(local.New(authStore))

	deps := &AuthDeps{Store: authStore, Sessions: sessions, Providers: registry}
	api, wsRoute := NewAPI(cfg)
	api.EnableAuth(deps)
	hub := NewHub()
	go hub.Run(t.Context())
	api.SetupWebSocket(hub, wsRoute)
	api.SetupRoutesAPI()

	return api, deps, dbConn
}

// testAuthServer builds a routed API with session/RBAC auth enabled,
// backed by an in-memory auth store (bootstrap admin created).
// The returned *sql.DB is the auth store's handle (close to simulate
// infrastructure failure).
// opts mutate the config before the API is built, for settings the router
// reads at setup (route prefix, demo group).
func testAuthServer(
	t *testing.T,
	path string,
	opts ...func(*config.Config),
) (*API, *AuthDeps, *sql.DB) {
	t.Helper()

	api, deps, dbConn := testAuthServerPendingSetup(t, path, opts...)
	if _, err := deps.Store.CreateFirstAdmin(
		t.Context(),
		"admin",
		"Administrator",
		"admin-password",
	); err != nil {
		t.Fatalf("Store.CreateFirstAdmin() failed: %v", err)
	}

	return api, deps, dbConn
}

// createAuthUser creates a user with the given password and groups.
func createAuthUser(
	t *testing.T,
	deps *AuthDeps,
	username, plaintext string,
	groups ...string,
) *auth.User {
	t.Helper()

	hash, err := password.Hash(plaintext)
	if err != nil {
		t.Fatalf("password.Hash() failed: %v", err)
	}

	user, err := deps.Store.CreateUser(
		t.Context(),
		username, "", "", hash, groups,
	)
	if err != nil {
		t.Fatalf("Store.CreateUser(%q) failed: %v", username, err)
	}
	t.Logf("created user %q (id=%q)", username, user.ID)

	return user
}

// createAuthGroup creates a group holding grants.
func createAuthGroup(
	t *testing.T,
	deps *AuthDeps,
	name, description string,
	grants []rbac.Grant,
) *store.Group {
	t.Helper()

	group, err := deps.Store.CreateGroup(t.Context(), name, description, grants)
	if err != nil {
		t.Fatalf("Store.CreateGroup(%q) failed: %v", name, err)
	}
	t.Logf("created group %q (id=%q)", name, group.ID)

	return group
}

// createAPIToken creates an API token for userID, returning its plaintext and record.
func createAPIToken(
	t *testing.T,
	deps *AuthDeps,
	userID, name string,
	expiresAt *time.Time,
) (string, *store.APIToken) {
	t.Helper()

	plaintext, token, err := deps.Store.CreateAPIToken(t.Context(), userID, name, expiresAt)
	if err != nil {
		t.Fatalf("Store.CreateAPIToken(%q) failed: %v", name, err)
	}
	t.Logf("created API token %q (id=%q)", name, token.ID)

	return plaintext, token
}

// withRoutePrefix serves the API under prefix.
func withRoutePrefix(prefix string) func(*config.Config) {
	return func(cfg *config.Config) { cfg.Settings.Web.RoutePrefix = prefix }
}

// withDemoGroup makes the API a demo instance holding group read-only.
func withDemoGroup(group string) func(*config.Config) {
	return func(cfg *config.Config) {
		cfg.Settings.Web.Demo = &config.WebSettingsDemo{Group: group}
	}
}

// apiPath returns the "/api/v1"+path route of api, under its route prefix.
func apiPath(api *API, path string) string {
	return strings.TrimSuffix(api.RoutePrefix, "/") + "/api/v1" + path
}

// serveAuth routes a request through the API's base router.
func serveAuth(api *API, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	api.BaseRouter.ServeHTTP(w, r)
	return w
}

// loginCookie logs username in, returning the session cookie.
func loginCookie(t *testing.T, api *API, username, plaintext string) *http.Cookie {
	t.Helper()

	req := httptest.NewRequest(
		http.MethodPost, apiPath(api, "/auth/login"),
		strings.NewReader(fmt.Sprintf(`{"username":%q,"password":%q}`, username, plaintext)),
	)
	w := serveAuth(api, req)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf(
			"login as %q: status code mismatch\ngot:  %d - %q\nwant: %d",
			username, got, w.Body.String(), want,
		)
	}
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == authCookieName {
			return cookie
		}
	}

	t.Fatalf("login as %q: session cookie mismatch\ngot:  absent\nwant: present", username)
	return nil
}

// assertSessionCookieCleared fails unless the logout response w expires the session cookie.
func assertSessionCookieCleared(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()

	cleared := false
	for _, c := range w.Result().Cookies() {
		if c.Name == authCookieName {
			cleared = true
			if c.MaxAge >= 0 {
				t.Errorf("logout: Cookie.MaxAge mismatch\ngot:  %d\nwant: <0", c.MaxAge)
			}
		}
	}
	if !cleared {
		t.Errorf("logout: %s cookie mismatch\ngot:  absent\nwant: present", authCookieName)
	}
}

// authedRequest builds a request carrying the session cookie.
func authedRequest(method, target, body string, cookie *http.Cookie) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if cookie != nil {
		req.AddCookie(cookie)
	}

	return req
}

// adminContext resolves the bootstrap admin's [auth.Context].
func adminContext(t *testing.T, api *API, deps *AuthDeps) *auth.Context {
	t.Helper()

	creds, err := deps.Store.LocalCredentials(t.Context(), "admin")
	if err != nil {
		t.Fatalf("Store.LocalCredentials() failed: %v", err)
	}
	if creds == nil {
		t.Fatal("Store.LocalCredentials() mismatch\ngot:  nil\nwant: non-nil")
	}

	authCtx, err := api.contextForUser(t.Context(), creds.UserID, "local")
	if err != nil {
		t.Fatalf("API.contextForUser(%q) failed: %v", creds.UserID, err)
	}
	t.Logf("resolved the admin's auth context (userID=%q)", creds.UserID)

	return authCtx
}

// withAuthCtx attaches an [auth.Context] to a request, as the middleware would.
func withAuthCtx(r *http.Request, authCtx *auth.Context) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), authCtxKey{}, authCtx))
}

// wireHub attaches a running Hub to api, returning a connect function that
// registers a fake WebSocket client with the given user/session identity
// (and, optionally, an allowed-service restriction).
func wireHub(t *testing.T, api *API) func(userID, sessionHash string, readableServices ...map[string]bool) *Client {
	t.Helper()

	api.hub = NewHub()
	go api.hub.Run(t.Context())
	return func(userID, sessionHash string, readableServices ...map[string]bool) *Client {
		var services map[string]bool
		if len(readableServices) > 0 {
			services = readableServices[0]
		}
		client := &Client{
			hub:              api.hub,
			send:             make(chan []byte, 8),
			userID:           userID,
			sessionHash:      sessionHash,
			readableServices: services,
		}
		api.hub.register <- client
		return client
	}
}

// failingSessionStore is a [session.Store] whose every method fails.
type failingSessionStore struct{}

var errSessionStore = errors.New("session store broke")

func (failingSessionStore) InsertSession(_ context.Context, _ store.Session) error {
	return errSessionStore
}

func (failingSessionStore) SessionByTokenHash(_ context.Context, _ string) (*store.Session, error) {
	return nil, errSessionStore
}

func (failingSessionStore) TouchSession(_ context.Context, _ string, _ time.Time) error {
	return errSessionStore
}

func (failingSessionStore) DeleteSession(_ context.Context, _ string) error {
	return errSessionStore
}

func (failingSessionStore) DeleteSessionsForUser(_ context.Context, _ string) error {
	return errSessionStore
}

func (failingSessionStore) DeleteExpiredSessions(_ context.Context, _ time.Time) error {
	return errSessionStore
}

func (failingSessionStore) TrimSessionsForUser(_ context.Context, _ string, _ int) ([]string, error) {
	return nil, errSessionStore
}
