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
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	_ "modernc.org/sqlite"

	"github.com/release-argus/Argus/auth"
	"github.com/release-argus/Argus/auth/password"
	"github.com/release-argus/Argus/auth/provider"
	"github.com/release-argus/Argus/auth/provider/local"
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
	"golang.org/x/sync/errgroup"
)

var (
	packageName          = "api_v1"
	secretValueMarshaled string
	testTempDir          string // Process-lifetime home of test config files.
)

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	ctx, cancel := context.WithCancel(context.Background())
	g, _ := errgroup.WithContext(ctx)

	// Shorten the WebSocket ping interval so writePump tests don't wait out the production default.
	pingPeriod = 10 * time.Millisecond

	config.DebounceDuration = 500 * time.Millisecond
	flags := make(map[string]bool)
	// Sweep strays from earlier killed runs.
	strays, _ := filepath.Glob(filepath.Join(os.TempDir(), "argus-api-v1-test*"))
	for _, stray := range strays {
		_ = os.RemoveAll(stray)
	}
	tempDir, err := os.MkdirTemp("", "argus-api-v1-test")
	if err != nil {
		fmt.Printf("%s\ncreate temp dir: %v", packageName, err)
		os.Exit(1)
	}
	testTempDir = tempDir
	path := filepath.Join(tempDir, "TestWebAPIv1Main.yml")
	testYAML_Argus(path)
	var cfg config.Config
	cfg.Load(ctx, g, path, &flags)

	// Marshal the secret value '<secret>' -> '\u003csecret\u003e'.
	secretValueMarshaledBytes, _ := decode.Marshal("json", util.SecretValue)
	secretValueMarshaled = string(secretValueMarshaledBytes)

	// Run other tests.
	exitCode := m.Run()
	_ = os.RemoveAll(tempDir)
	_ = os.Remove(cfg.Settings.DataDatabaseFile())
	cancel()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty", packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
}

type failWriter struct {
	header http.Header
	code   int
	body   string
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

func (f *failWriter) Write([]byte) (int, error) {
	if f.code == 0 {
		f.code = http.StatusOK
	}
	return 0, errors.New("write failed")
}

// PlainDefaults returns plain defaults and hardDefaults for testing.
func plainDefaults(t *testing.T) (*config.Defaults, *config.Defaults) {
	t.Helper()

	dockerDefaults, _ := docker.DecodeDefaults(
		"yaml", nil,
		nil,
	)
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
	hardDefaults.Service.LatestVersion.GitHub.AccessToken = test.GitHubToken(t)
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
	var cfg config.Config
	g, _ := errgroup.WithContext(t.Context())

	flags := make(map[string]bool)
	cfg.Load(t.Context(), g, file, &flags)
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
		_ = os.RemoveAll(cfg.Settings.Data.DatabaseFile)
	})

	return API{Config: cfg}
}

func testService(
	t *testing.T,
	id string,
	lvType, dvType string,
	semVer bool,
) *service.Service {
	if t != nil {
		t.Helper()
	}
	svcCfg := svctest.PlainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	svc := test.Must(t, func() (*service.Service, error) {
		return service.DecodeService(
			"yaml", []byte(test.TrimYAML(`
			options:
				semantic_versioning: `+fmt.Sprint(semVer)+`
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

	// Status channels.
	svc.Status.AnnounceChannel = announceChannel
	svc.Status.DatabaseChannel = databaseChannel

	return svc
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
func testAuthServerPendingSetup(t *testing.T, path string) (*API, *AuthDeps, *sql.DB) {
	t.Helper()

	if !filepath.IsAbs(path) {
		path = filepath.Join(testTempDir, path)
	}
	testYAML_Argus(path)
	cfg := testLoad(t, path)
	t.Cleanup(func() { _ = os.RemoveAll(cfg.Settings.DataDatabaseFile()) })

	dbConn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf(
			"%s\nopen auth db: %v",
			packageName, err,
		)
	}
	dbConn.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = dbConn.Close() })

	authStore, err := store.New(t.Context(), dbConn)
	if err != nil {
		t.Fatalf(
			"%s\nauth store: %v",
			packageName, err,
		)
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
	go hub.Run()
	api.SetupWebSocket(hub, wsRoute)
	api.SetupRoutesAPI()

	return api, deps, dbConn
}

// testAuthServer builds a routed API with session/RBAC auth enabled,
// backed by an in-memory auth store (bootstrap admin created).
// The returned *sql.DB is the auth store's handle (close to simulate
// infrastructure failure).
func testAuthServer(t *testing.T, path string) (*API, *AuthDeps, *sql.DB) {
	t.Helper()

	api, deps, dbConn := testAuthServerPendingSetup(t, path)
	if _, err := deps.Store.CreateFirstAdmin(
		t.Context(),
		"admin",
		"Administrator",
		"admin-password",
	); err != nil {
		t.Fatalf(
			"%s\nbootstrap admin: %v",
			packageName, err,
		)
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
		t.Fatalf(
			"%s\nhash password: %v",
			packageName, err,
		)
	}

	user, err := deps.Store.CreateUser(
		t.Context(),
		username, "", "", hash, groups,
	)
	if err != nil {
		t.Fatalf(
			"%s\ncreate user %q: %v",
			packageName, username, err,
		)
	}

	return user
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

	body := fmt.Sprintf(`{"username":%q,"password":%q}`, username, plaintext)
	w := serveAuth(api,
		httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Fatalf(
			"%s\nlogin as %q failed\ngot:  %d - %s",
			packageName, username, w.Code, w.Body.String(),
		)
	}
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == authCookieName {
			return cookie
		}
	}

	t.Fatalf("%s\nlogin response missing the session cookie", packageName)
	return nil
}

// authedRequest builds a request carrying the session cookie.
func authedRequest(
	method, target string,
	body string,
	cookie *http.Cookie,
) *http.Request {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, target, reader)
	if cookie != nil {
		req.AddCookie(cookie)
	}

	return req
}

// adminContext resolves the bootstrap admin's [auth.Context].
func adminContext(t *testing.T, api *API, deps *AuthDeps) *auth.Context {
	t.Helper()

	creds, err := deps.Store.LocalCredentials(t.Context(), "admin")
	if err != nil || creds == nil {
		t.Fatalf(
			"%s\nadmin lookup failed: %v",
			packageName, err,
		)
	}

	authCtx, err := api.contextForUser(t.Context(), creds.UserID, "local")
	if err != nil {
		t.Fatalf(
			"%s\nadmin context failed: %v",
			packageName, err,
		)
	}

	return authCtx
}

// withAuthCtx attaches an [auth.Context] to a request, as the middleware would.
func withAuthCtx(r *http.Request, authCtx *auth.Context) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), authCtxKey{}, authCtx))
}

// wireHub attaches a running Hub to api, returning a connect function that
// registers a fake WebSocket client with the given user/session identity
// (and, optionally, an allowed-service restriction).
func wireHub(t *testing.T, api *API) func(userID, sessionHash string, allowedServices ...map[string]bool) *Client {
	t.Helper()

	api.hub = NewHub()
	go api.hub.Run()
	return func(userID, sessionHash string, allowedServices ...map[string]bool) *Client {
		var services map[string]bool
		if len(allowedServices) > 0 {
			services = allowedServices[0]
		}
		client := &Client{
			hub:             api.hub,
			send:            make(chan []byte, 8),
			userID:          userID,
			sessionHash:     sessionHash,
			allowedServices: services,
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
