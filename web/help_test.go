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

package web

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"

	"github.com/release-argus/Argus/config"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/httpx"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/webhook"
	whtest "github.com/release-argus/Argus/webhook/test"
)

var (
	mainCfg *config.Config
	host    = "localhost"
	port    string
	router  *mux.Router
)

func TestMain(m *testing.M) {
	logtest.InitLog()

	file := "TestWebMain.yml"
	mainCfg = testConfig(nil, file)
	port = mainCfg.Settings.Web.ListenPort
	mainCfg.Settings.Web.ListenHost = host

	ctx, cancel := context.WithCancel(context.Background())

	router = newWebUI(ctx, mainCfg, nil)
	go func() { _ = Run(ctx, mainCfg, nil) }()
	url := fmt.Sprintf(
		"http://%s:%s%s",
		host, port, mainCfg.Settings.Web.RoutePrefix,
	)
	if err := waitForServer(url, 1*time.Second); err != nil {
		fmt.Println(err.Error())
		cancel()
		os.Exit(1)
	}

	exitCode := m.Run()
	cancel()
	_ = os.Remove(file)
	_ = os.Remove(mainCfg.Settings.Data.DatabaseFile)

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Println("exit code channel not empty")
		exitCode = 1
	}

	os.Exit(exitCode)
}

// hasRoute reports whether router serves a route ending in path,
// ignoring the configured route prefix.
func hasRoute(t *testing.T, router *mux.Router, path string) bool {
	t.Helper()

	found := false
	if err := router.Walk(
		func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
			if template, err := route.GetPathTemplate(); err == nil &&
				strings.HasSuffix(template, path) {
				found = true
			}
			return nil
		},
	); err != nil {
		t.Fatalf("Router.Walk() failed: %v", err)
	}
	return found
}

func getFreePort() (int, error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	_ = ln.Close()
	return ln.Addr().(*net.TCPAddr).Port, nil
}

func plainDefaults() (*config.Defaults, *config.Defaults) {
	defaults, _ := config.DecodeDefaults("yaml", nil)
	defaults.Notify = shoutrrr.ShoutrrrsDefaults{}

	hardDefaults, _ := config.DecodeDefaults("yaml", nil)
	hardDefaults.Default()
	defaults.SetDefaults(hardDefaults)

	hardDefaults.Service.Status.DatabaseChannel = make(chan dbtype.Message, 32)
	hardDefaults.Service.Status.SaveChannel = make(chan bool, 4)

	return defaults, hardDefaults
}

func testConfig(t *testing.T, path string) (cfg *config.Config) {
	var ctx context.Context
	if t == nil {
		ctx = context.Background()
	} else {
		ctx = t.Context()
	}
	g, _ := errgroup.WithContext(ctx)
	testYAML_Argus(path)
	cfg = &config.Config{}

	flags := make(map[string]bool)
	if !cfg.Load(ctx, g, path, &flags) {
		panic(fmt.Sprintf("Config.Load(%q) failed", path))
	}
	if t != nil {
		t.Cleanup(func() { _ = os.Remove(cfg.Settings.DataDatabaseFile()) })
	}

	freePort, err := getFreePort()
	if err != nil {
		panic(err)
	}
	var (
		listenHost  = "0.0.0.0"
		listenPort  = strconv.Itoa(freePort)
		routePrefix = "/"
	)
	cfg.Settings.Web = config.WebSettings{
		ListenHost:  listenHost,
		ListenPort:  listenPort,
		RoutePrefix: routePrefix,
	}

	defaults, hardDefaults := plainDefaults()
	cfg.Defaults, cfg.HardDefaults = *defaults, *hardDefaults
	cfg.Webhook = webhook.WebhooksDefaults{}
	cfg.Notify = shoutrrr.ShoutrrrsDefaults{}
	svcCfg := service.DefaultsConfig{
		Soft: &cfg.Defaults.Service,
		Hard: &cfg.HardDefaults.Service,
	}

	svc := testService(t, "test", svcCfg)
	go func() {
		<-cfg.HardDefaults.Service.Status.SaveChannel
	}()
	if err := cfg.AddService(svc.ID, svc); err != nil {
		panic(err)
	}

	cfg.Notify = cfg.Defaults.Notify

	whPass := testWebhookDefaults(false)
	whFail := testWebhookDefaults(true)
	cfg.Webhook = webhook.WebhooksDefaults{
		"pass": whPass,
		"fail": whFail,
	}

	cfg.Order = []string{svc.ID}

	return
}

func testService(t *testing.T, id string, svcCfg service.DefaultsConfig) *service.Service {
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	svc := test.Must(t, func() (*service.Service, error) {
		return service.DecodeService(
			"yaml", []byte(test.TrimYAML(`
				name: `+id+`
				comment: test services comment
				options:
					interval: 10m
					semantic_versioning: true
				latest_version:
					type: url
					url: `+test.ArgusGitHubRepo+`
					url_command:
						- type: regex
							regex: '-([0-9.]+)-'
							index: 0
					require:
						regex_content: content
						regex_version: version
						docker:
							type: ghcr
							image: `+test.ArgusDockerGHCRRepo+`
							tag: "{{ version }}"
				deployed_version:
					type: url
					method: GET
					url: https://release-argus.io
					basic_auth:
						username: fizz
						password: buzz
					headers:
						- key: foo
							value: bar
					json: something
					regex: '([0-9]+)\s<[^>]+>The Argus Developers'
					regex_template: 'v$1'
				command:
					- ["ls", "-lah"]
				notify:
					test:
				`+shoutrrrtest.Shoutrrr(false, false).String("    ")+`
				webhook:
					test:
						type: github
						url: `+test.WebhookGitHub.URLValid+`
						secret: argus
				dashboard:
					auto_approve: false
					icon: test
					web_url: https://release-argus.io
			`)),
			id,
			svcCfg, notifyCfg, whCfg,
		)
	})
	version := "3.0.0"
	svc.Status.SetLatestVersion(version, "", false)
	svc.Status.SetDeployedVersion(version, "", false)
	svc.Status.SetApprovedVersion(version, false)

	return svc
}

func testWebhookDefaults(failing bool) *webhook.Defaults {
	secret := "argus"
	if failing {
		secret = "notArgus"
	}

	wh, err := webhook.DecodeDefaults(
		"yaml", []byte(test.TrimYAML(`
			allow_invalid_certs: false
			delay: 0s
			desired_status_code: 0
			max_tries: 1
			secret: `+secret+`
			silent_fails: false
			type: github
			url: `+test.WebhookGitHub.URLValid+`
		`)),
	)
	if err != nil {
		panic(err)
	}
	return wh
}

func generateCertFiles(t *testing.T, certFile, keyFile string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey() failed: %v", err)
	}

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "localhost"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("x509.CreateCertificate() failed: %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	if err := os.WriteFile(certFile, certPEM, 0644); err != nil {
		t.Fatalf("os.WriteFile(%q) failed: %v", certFile, err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		t.Fatalf("os.WriteFile(%q) failed: %v", keyFile, err)
	}
}

// assertServerShutdown fails unless Run returns after cancel and testURL then refuses requests.
func assertServerShutdown(
	t *testing.T,
	cancelFunc context.CancelFunc,
	errCh <-chan error,
	testURL string,
) {
	t.Helper()

	cancelFunc()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Errorf(
				"Run() error type mismatch\ngot:  %v\nwant: nil or %v",
				err, http.ErrServerClosed,
			)
		}
	case <-time.After(time.Second):
		t.Error("Run() mismatch\ngot:  still running after 1s\nwant: returned")
	}

	if resp, err := insecureClient().Get(testURL); err == nil {
		_ = resp.Body.Close()
		t.Errorf("Client.Get(%q) error mismatch\ngot:  nil\nwant: non-nil (shut down)", testURL)
	}
}

// waitForServer waits until the server at the given URL is ready, or the timeout is reached.
func waitForServer(url string, timeout time.Duration) error {
	client := insecureClient()
	deadline := time.Now().Add(timeout)
	for {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("server did not start in time: %w", err)
		}

		time.Sleep(10 * time.Millisecond)
	}
}

// insecureClient returns an HTTP client that accepts the self-signed test certificate.
func insecureClient() *http.Client {
	return &http.Client{Transport: httpx.InsecureTransport, Timeout: 500 * time.Millisecond}
}
