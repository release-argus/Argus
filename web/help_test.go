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
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/logx"
	whtest "github.com/release-argus/Argus/webhook/test"
	"golang.org/x/sync/errgroup"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/webhook"
)

var (
	packageName = "web"
	mainCfg     *config.Config
	host        = "localhost"
	port        string
)

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// GIVEN: a valid config with a Service.
	file := "TestWebMain.yml"
	mainCfg = testConfig(nil, file)
	port = mainCfg.Settings.Web.ListenPort
	mainCfg.Settings.Web.ListenHost = host

	// Create a cancellable context for shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// WHEN: the Router is fetched for this Config.
	router = newWebUI(mainCfg)
	go Run(ctx, mainCfg)
	url := fmt.Sprintf(
		"http://%s:%s%s",
		host, port, mainCfg.Settings.Web.RoutePrefix,
	)
	if err := waitForServer(url, 1*time.Second); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// THEN: Web UI is accessible for the tests.
	exitCode := m.Run()
	_ = os.Remove(file)
	_ = os.Remove(mainCfg.Settings.Data.DatabaseFile)

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty", packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
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
	cfg.Load(ctx, g, path, &flags)
	if t != nil {
		t.Cleanup(func() { _ = os.Remove(cfg.Settings.DataDatabaseFile()) })
	}

	// Settings.Web.
	port, err := getFreePort()
	if err != nil {
		panic(err)
	}
	var (
		listenHost  = "0.0.0.0"
		listenPort  = fmt.Sprint(port)
		routePrefix = "/"
	)
	cfg.Settings.Web = config.WebSettings{
		ListenHost:  listenHost,
		ListenPort:  listenPort,
		RoutePrefix: routePrefix,
	}

	// Defaults.
	defaults, hardDefaults := plainDefaults()
	cfg.Defaults, cfg.HardDefaults = *defaults, *hardDefaults
	cfg.WebHook = webhook.WebHooksDefaults{}
	cfg.Notify = shoutrrr.ShoutrrrsDefaults{}
	svcCfg := service.DefaultsConfig{
		Soft: &cfg.Defaults.Service,
		Hard: &cfg.HardDefaults.Service,
	}

	// Service.
	svc := testService(t, "test", svcCfg)
	go func() {
		<-cfg.HardDefaults.Service.Status.SaveChannel
	}()
	_ = cfg.AddService(svc.ID, svc)

	// Notify.
	cfg.Notify = cfg.Defaults.Notify

	// WebHook.
	whPass := testWebHookDefaults(false)
	whFail := testWebHookDefaults(true)
	cfg.WebHook = webhook.WebHooksDefaults{
		"pass": whPass,
		"fail": whFail,
	}

	// Order.
	cfg.Order = []string{svc.ID}

	return
}

func testService(t *testing.T, id string, svcCfg service.DefaultsConfig) *service.Service {
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	// Service.
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
						url: `+test.WebHookGitHub["url_valid"]+`
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
	svc.Status.SetLatestVersion("3.0.0", "", false)
	svc.Status.SetDeployedVersion("3.0.0", "", false)
	svc.Status.SetApprovedVersion("3.0.0", false)

	return svc
}

func testWebHookDefaults(failing bool) *webhook.Defaults {
	secret := "argus"
	if failing {
		secret = "notArgus"
	}

	wh, _ := webhook.DecodeDefaults(
		"yaml", []byte(test.TrimYAML(`
			allow_invalid_certs: false
			delay: 0s
			desired_status_code: 0
			max_tries: 1
			secret: `+secret+`
			silent_fails: false
			type: github
			url: `+test.WebHookGitHub["url_valid"]+`
		`)),
	)
	return wh
}

func generateCertFiles(certFile, keyFile string) error {
	// Generate the certificate and private key:
	// 	Generate a private key.
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// 	Create a self-signed certificate.
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
		return err
	}

	// Convert the certificate and private key to PEM format.
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	// Write the certificate and private key to files.
	if err := os.WriteFile(certFile, certPEM, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		return err
	}

	return nil
}

// assertServerShutdown verifies that the server shuts down correctly when the context is cancelled.
func assertServerShutdown(
	t *testing.T,
	cancelFunc context.CancelFunc,
	errCh <-chan error,
	testURL string,
) {
	t.Helper()

	// Cancel the context to trigger shutdown.
	cancelFunc()

	// Wait for the server to return.
	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Errorf(
				"%s\nserver returned unexpected error on shutdown: %v",
				packageName, err,
			)
		}
	case <-time.After(time.Second):
		t.Errorf("%s\nserver did not shutdown in time", packageName)
	}

	// Test that the server no longer accepts requests.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 500 * time.Millisecond,
	}
	if _, err := client.Get(testURL); err == nil {
		t.Errorf("%s\nexpected request to fail after shutdown, but it succeeded", packageName)
	}
}

// waitForServer waits until the server at the given URL is ready, or the timeout is reached.
func waitForServer(url string, timeout time.Duration) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 500 * time.Millisecond,
	}
	deadline := time.Now().Add(timeout)
	for {
		resp, err := client.Get(url)

		// Server is ready.
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}

		// Timeout reached.
		if time.Now().After(deadline) {
			return fmt.Errorf(
				"%s\nserver did not start in time: %v",
				packageName, err,
			)
		}

		// Delay before retrying.
		time.Sleep(10 * time.Millisecond)
	}
}
