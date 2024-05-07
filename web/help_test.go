// Copyright [2023] [Argus]
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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/config"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	test_shoutrrr "github.com/release-argus/Argus/notifiers/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

var mainCfg *config.Config
var port *string

func TestMain(m *testing.M) {
	// initialize jLog
	jLog := util.NewJLog("DEBUG", false)
	jLog.Testing = true

	// GIVEN a valid config with a Service
	file := "TestMain.yml"
	mainCfg = testConfig(file, jLog, nil)
	os.Remove(file)
	defer os.Remove(*mainCfg.Settings.Data.DatabaseFile)
	port = mainCfg.Settings.Web.ListenPort
	mainCfg.Settings.Web.ListenHost = test.StringPtr("localhost")

	// WHEN the Router is fetched for this Config
	router = newWebUI(mainCfg)
	go Run(mainCfg, jLog)

	// THEN Web UI is accessible for the tests
	code := m.Run()
	os.Exit(code)
}

func getFreePort() (int, error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	ln.Close()
	return ln.Addr().(*net.TCPAddr).Port, nil
}

func testConfig(path string, jLog *util.JLog, t *testing.T) (cfg *config.Config) {
	testYAML_Argus(path, t)
	cfg = &config.Config{}

	// Settings.Log
	cfg.Settings.Log.Level = test.StringPtr("DEBUG")

	cfg.Load(
		path,
		&map[string]bool{},
		jLog)
	if t != nil {
		t.Cleanup(func() { os.Remove(cfg.Settings.DataDatabaseFile()) })
	}

	cfg.Settings.NilUndefinedFlags(&map[string]bool{})

	// Settings.Web
	port, err := getFreePort()
	if err != nil {
		panic(err)
	}
	var (
		listenHost  string = "0.0.0.0"
		listenPort  string = fmt.Sprint(port)
		routePrefix string = "/"
	)
	cfg.Settings.Web = config.WebSettings{
		ListenHost:  &listenHost,
		ListenPort:  &listenPort,
		RoutePrefix: &routePrefix,
	}

	// Defaults
	cfg.Defaults.SetDefaults()

	// Service
	svc := testService("test")
	svc.DeployedVersionLookup = testDeployedVersion()
	svc.LatestVersion.URLCommands = filter.URLCommandSlice{testURLCommandRegex()}
	emptyNotify := shoutrrr.ShoutrrrDefaults{}
	emptyNotify.InitMaps()
	notify := shoutrrr.Slice{
		"test": test_shoutrrr.Shoutrrr(false, false)}
	notify["test"].Params = map[string]string{}
	svc.Notify = notify
	svc.Comment = "test service's comment"
	cfg.Service = service.Slice{
		svc.ID: svc,
	}

	// Notify
	cfg.Notify = cfg.Defaults.Notify

	// WebHook
	whPass := testWebHookDefaults(false)
	whFail := testWebHookDefaults(true)
	cfg.WebHook = webhook.SliceDefaults{
		"pass": whPass,
		"fail": whFail,
	}

	// Order
	cfg.Order = []string{svc.ID}

	return
}

func testService(id string) (svc *service.Service) {
	var (
		sAnnounceChannel chan []byte         = make(chan []byte, 2)
		sDatabaseChannel chan dbtype.Message = make(chan dbtype.Message, 5)
		sSaveChannel     chan bool           = make(chan bool, 5)
	)
	webhookDefaults := webhook.WebHookDefaults{}
	webhookDefaults.SetDefaults()
	svc = &service.Service{
		ID: id,
		LatestVersion: *latestver.New(
			test.StringPtr(""),
			test.BoolPtr(false),
			nil, nil,
			&filter.Require{
				RegexContent: "content",
				RegexVersion: "version",
				Docker: filter.NewDockerCheck(
					"ghcr",
					"release-argus/argus",
					"{{ version }}",
					"", "", "", time.Time{}, nil)},
			nil,
			"url",
			"https://release-argus.io",
			&filter.URLCommandSlice{testURLCommandRegex()},
			test.BoolPtr(false),
			nil, nil),
		DeployedVersionLookup: testDeployedVersion(),
		Options: *opt.New(
			nil,
			"10m",
			test.BoolPtr(true),
			&opt.OptionsDefaults{},
			&opt.OptionsDefaults{}),
		Dashboard: *service.NewDashboardOptions(
			test.BoolPtr(false), "test", "", "https://release-argus.io",
			&service.DashboardOptionsDefaults{}, &service.DashboardOptionsDefaults{}),
		Defaults:          &service.Defaults{},
		HardDefaults:      &service.Defaults{},
		Command:           command.Slice{command.Command{"ls", "-lah"}},
		CommandController: &command.Controller{},
		WebHook: webhook.Slice{
			"test": webhook.New(
				nil, nil, "", nil, nil, nil, nil, nil, "", nil, "",
				"example.com",
				nil, nil, nil)}}
	svc.Status.AnnounceChannel = &sAnnounceChannel
	svc.Status.DatabaseChannel = &sDatabaseChannel
	svc.Status.SaveChannel = &sSaveChannel
	svc.Status.Init(
		len(svc.Notify),
		len(svc.Command), len(svc.WebHook),
		&svc.ID,
		&svc.Dashboard.WebURL)
	svc.Status.SetApprovedVersion("2.0.0", false)
	svc.Status.SetDeployedVersion("2.0.0", false)
	svc.Status.SetDeployedVersionTimestamp(time.Now().UTC().Format(time.RFC3339))
	svc.Status.SetLatestVersion("3.0.0", true)
	svc.Status.SetDeployedVersionTimestamp(time.Now().UTC().Format(time.RFC3339))
	svc.LatestVersion.Init(
		&latestver.LookupDefaults{}, &latestver.LookupDefaults{},
		&svc.Status,
		&svc.Options)
	svc.DeployedVersionLookup.Init(
		&deployedver.LookupDefaults{}, &deployedver.LookupDefaults{},
		&svc.Status,
		&svc.Options)
	svc.CommandController.Init(
		&svc.Status,
		&svc.Command,
		&svc.Notify,
		&svc.Options.Interval)
	svc.WebHook.Init(
		&svc.Status,
		&webhook.SliceDefaults{}, &webhook.WebHookDefaults{}, &webhookDefaults,
		&svc.Notify,
		&svc.Options.Interval)

	return
}

func testWebHookDefaults(failing bool) *webhook.WebHookDefaults {
	whDesiredStatusCode := 0
	whMaxTries := uint(1)
	wh := webhook.NewDefaults(
		test.BoolPtr(false),
		nil,
		"0s",
		&whDesiredStatusCode,
		&whMaxTries,
		"argus",
		test.BoolPtr(false),
		"github",
		"https://valid.release-argus.io/hooks/github-style")
	if failing {
		wh.Secret = "notArgus"
	}
	return wh
}

func testDeployedVersion() *deployedver.Lookup {
	var (
		allowInvalidCerts = false
		json              = "something"
		regex             = `([0-9]+)\s<[^>]+>The Argus Developers`
		regexTemplate     = "v$1"
		url               = "https://release-argus.io"
	)
	return deployedver.New(
		&allowInvalidCerts,
		&deployedver.BasicAuth{
			Username: "fizz",
			Password: "buzz"},
		nil,
		&[]deployedver.Header{
			{Key: "foo", Value: "bar"}},
		json,
		"GET",
		nil,
		regex,
		&regexTemplate,
		&svcstatus.Status{},
		url,
		&deployedver.LookupDefaults{},
		&deployedver.LookupDefaults{})
}

func testURLCommandRegex() filter.URLCommand {
	regex := "-([0-9.]+)-"
	index := 0
	return filter.URLCommand{
		Type:  "regex",
		Regex: &regex,
		Index: index,
	}
}

func generateCertFiles(certFile, keyFile string) error {
	// Generate the certificate and private key
	// Generate a private key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Create a self-signed certificate
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

	// Convert the certificate and private key to PEM format
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	// Write the certificate and private key to files
	if err := os.WriteFile(certFile, certPEM, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		return err
	}

	return nil
}
