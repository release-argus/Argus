// Copyright [2025] [Argus]
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

	logutil "github.com/release-argus/Argus/util/log"
	"golang.org/x/sync/errgroup"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrr_test "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/dashboard"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	deployedver_base "github.com/release-argus/Argus/service/deployed_version/types/base"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	latestver_base "github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/latest_version/types/web"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	logtest "github.com/release-argus/Argus/test/log"
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

	// GIVEN a valid config with a Service.
	file := "TestWebMain.yml"
	mainCfg = testConfig(nil, file)
	port = mainCfg.Settings.Web.ListenPort
	mainCfg.Settings.Web.ListenHost = host

	// Create a cancellable context for shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// WHEN the Router is fetched for this Config.
	router = newWebUI(mainCfg)
	go Run(ctx, mainCfg)
	url := fmt.Sprintf("http://%s:%s%s",
		host, port, mainCfg.Settings.Web.RoutePrefix)
	if err := waitForServer(url, 1*time.Second); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// THEN Web UI is accessible for the tests.
	exitCode := m.Run()
	_ = os.Remove(file)
	_ = os.Remove(mainCfg.Settings.Data.DatabaseFile)

	if len(logutil.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty",
			packageName)
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
	cfg.Defaults.Default()
	cfg.HardDefaults.Default()

	// Service.
	svc := testService(t, "test")
	svc.DeployedVersionLookup = testDeployedVersion(t)
	svc.LatestVersion.(*web.Lookup).URLCommands = filter.URLCommands{testURLCommandRegex()}
	emptyNotify := shoutrrr.Defaults{}
	emptyNotify.InitMaps()
	notify := shoutrrr.Shoutrrrs{
		"test": shoutrrr_test.Shoutrrr(false, false)}
	notify["test"].Params = map[string]string{}
	svc.Notify = notify
	svc.Comment = "test services comment"
	svc.Init(
		&cfg.Defaults.Service, &cfg.HardDefaults.Service,
		&cfg.Notify, &cfg.Defaults.Notify, &cfg.HardDefaults.Notify,
		&cfg.WebHook, &cfg.Defaults.WebHook, &cfg.HardDefaults.WebHook)
	_ = cfg.AddService(svc.ID, svc)

	// Notify.
	cfg.Notify = cfg.Defaults.Notify

	// WebHook.
	whPass := testDefaults(false)
	whFail := testDefaults(true)
	cfg.WebHook = webhook.WebHooksDefaults{
		"pass": whPass,
		"fail": whFail,
	}

	// Order.
	cfg.Order = []string{svc.ID}

	return
}

func testService(t *testing.T, id string) (svc *service.Service) {
	hardDefaults := config.Defaults{}
	hardDefaults.Default()

	svc = &service.Service{
		ID:   id,
		Name: id,
		LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
			return latestver.New(
				"url",
				"yaml", test.TrimYAML(`
					url: release-argus/argus
					require:
						regex_content: content
						regex_version: version
						docker:
							type: ghcr
							image: release-argus/argus
							tag: "{{ version }}"
				`),
				nil,
				nil,
				&latestver_base.Defaults{}, &hardDefaults.Service.LatestVersion)
		}),
		DeployedVersionLookup: testDeployedVersion(t),
		Options: *opt.New(
			nil,
			"10m",
			test.BoolPtr(true),
			&opt.Defaults{}, &opt.Defaults{}),
		Dashboard: *dashboard.NewOptions(
			test.BoolPtr(false), "test", "", "https://release-argus.io", nil,
			&dashboard.OptionsDefaults{}, &dashboard.OptionsDefaults{}),
		Defaults:          &service.Defaults{},
		HardDefaults:      &service.Defaults{},
		Command:           command.Commands{command.Command{"ls", "-lah"}},
		CommandController: &command.Controller{},
		WebHook: webhook.WebHooks{
			"test": webhook.New(
				nil, nil, "", nil, nil, "test", nil, nil, nil, "", nil, "",
				"example.com",
				nil, nil, nil)}}

	// Status.
	var (
		sAnnounceChannel = make(chan []byte, 2)
		sDatabaseChannel = make(chan dbtype.Message, 5)
		sSaveChannel     = make(chan bool, 5)
	)
	svc.Status.AnnounceChannel = sAnnounceChannel
	svc.Status.DatabaseChannel = sDatabaseChannel
	svc.Status.SaveChannel = sSaveChannel
	svc.Status.Init(
		len(svc.Notify),
		len(svc.Command), len(svc.WebHook),
		svc.ID, svc.Name, "",
		&dashboard.Options{
			WebURL: svc.Dashboard.WebURL})
	svc.Status.SetApprovedVersion("2.0.0", false)
	svc.Status.SetDeployedVersion("2.0.0", "", false)
	svc.Status.SetLatestVersion("3.0.0", "", true)

	// LatestVersion.
	svc.LatestVersion.Init(
		&svc.Options,
		&svc.Status,
		&latestver_base.Defaults{}, &hardDefaults.Service.LatestVersion)

	// DeployedVersionLookup.
	svc.DeployedVersionLookup.Init(
		&svc.Options,
		&svc.Status,
		&deployedver_base.Defaults{}, &hardDefaults.Service.DeployedVersionLookup)

	// Command.
	svc.CommandController.Init(
		&svc.Status,
		&svc.Command,
		&svc.Notify,
		&svc.Options.Interval)

	// WebHook.
	svc.WebHook.Init(
		&svc.Status,
		&webhook.WebHooksDefaults{}, &webhook.Defaults{}, &hardDefaults.WebHook,
		&svc.Notify,
		&svc.Options.Interval)

	return
}

func testDefaults(failing bool) *webhook.Defaults {
	whDesiredStatusCode := uint16(0)
	whMaxTries := uint8(1)
	wh := webhook.NewDefaults(
		test.BoolPtr(false),
		nil,
		"0s",
		&whDesiredStatusCode,
		&whMaxTries,
		"argus",
		test.BoolPtr(false),
		"github",
		test.LookupGitHub["url_valid"])
	if failing {
		wh.Secret = "notArgus"
	}
	return wh
}

func testDeployedVersion(t *testing.T) deployedver.Lookup {
	defaults := &deployedver_base.Defaults{}
	hardDefaults := &deployedver_base.Defaults{}
	hardDefaults.Default()

	return test.IgnoreError(t, func() (deployedver.Lookup, error) {
		return deployedver.New(
			"url",
			"yaml", test.TrimYAML(`
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
			`),
			opt.New(
				nil, "",
				test.BoolPtr(false),
				defaults.Options, hardDefaults.Options),
			&status.Status{},
			defaults, hardDefaults)
	})
}

func testURLCommandRegex() filter.URLCommand {
	regex := "-([0-9.]+)-"
	index := 0
	return filter.URLCommand{
		Type:  "regex",
		Regex: regex,
		Index: &index,
	}
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
			t.Errorf("%s\nserver returned unexpected error on shutdown: %v",
				packageName, err)
		}
	case <-time.After(time.Second):
		t.Errorf("%s\nserver did not shutdown in time",
			packageName)
	}

	// Test that the server no longer accepts requests.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		Timeout: 500 * time.Millisecond}
	if _, err := client.Get(testURL); err == nil {
		t.Errorf("%s\nexpected request to fail after shutdown, but it succeeded",
			packageName)
	}
}

// waitForServer waits until the server at the given URL is ready, or the timeout is reached.
func waitForServer(url string, timeout time.Duration) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		Timeout: 500 * time.Millisecond}
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
			return fmt.Errorf("%s\nserver did not start in time: %v",
				packageName, err)
		}

		// Delay before retrying.
		time.Sleep(10 * time.Millisecond)
	}
}
