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

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrr_test "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	latestver_base "github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/latest_version/types/web"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

var mainCfg *config.Config
var port string

func TestMain(m *testing.M) {
	// Initialise jLog.
	jLog = util.NewJLog("DEBUG", false)
	jLog.Testing = true
	// v1.LogInit(jLog)

	// GIVEN a valid config with a Service.
	file := "TestWebMain.yml"
	mainCfg = testConfig(file, jLog, nil)
	os.Remove(file)
	defer os.Remove(mainCfg.Settings.Data.DatabaseFile)
	port = mainCfg.Settings.Web.ListenPort
	mainCfg.Settings.Web.ListenHost = "localhost"

	// WHEN the Router is fetched for this Config.
	router = newWebUI(mainCfg)
	go Run(mainCfg, jLog)
	time.Sleep(250 * time.Millisecond)

	// THEN Web UI is accessible for the tests.
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
	cfg.Settings.Log.Level = "DEBUG"

	cfg.Load(path, &map[string]bool{}, jLog)
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
		ListenHost:  listenHost,
		ListenPort:  listenPort,
		RoutePrefix: routePrefix,
	}

	// Defaults
	cfg.Defaults.Default()
	cfg.HardDefaults.Default()

	// Service
	svc := testService(t, "test")
	svc.DeployedVersionLookup = testDeployedVersion(t)
	svc.LatestVersion.(*web.Lookup).URLCommands = filter.URLCommandSlice{testURLCommandRegex()}
	emptyNotify := shoutrrr.Defaults{}
	emptyNotify.InitMaps()
	notify := shoutrrr.Slice{
		"test": shoutrrr_test.Shoutrrr(false, false)}
	notify["test"].Params = map[string]string{}
	svc.Notify = notify
	svc.Comment = "test services comment"
	svc.Init(
		&cfg.Defaults.Service, &cfg.HardDefaults.Service,
		&cfg.Notify, &cfg.Defaults.Notify, &cfg.HardDefaults.Notify,
		&cfg.WebHook, &cfg.Defaults.WebHook, &cfg.HardDefaults.WebHook)
	cfg.AddService(svc.ID, svc)

	// Notify
	cfg.Notify = cfg.Defaults.Notify

	// WebHook
	whPass := testDefaults(false)
	whFail := testDefaults(true)
	cfg.WebHook = webhook.SliceDefaults{
		"pass": whPass,
		"fail": whFail,
	}

	// Order
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
		Dashboard: *service.NewDashboardOptions(
			test.BoolPtr(false), "test", "", "https://release-argus.io", nil,
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

	// Status
	var (
		sAnnounceChannel chan []byte         = make(chan []byte, 2)
		sDatabaseChannel chan dbtype.Message = make(chan dbtype.Message, 5)
		sSaveChannel     chan bool           = make(chan bool, 5)
	)
	svc.Status.AnnounceChannel = &sAnnounceChannel
	svc.Status.DatabaseChannel = &sDatabaseChannel
	svc.Status.SaveChannel = &sSaveChannel
	svc.Status.Init(
		len(svc.Notify),
		len(svc.Command), len(svc.WebHook),
		&svc.ID, &svc.Name,
		&svc.Dashboard.WebURL)
	svc.Status.SetApprovedVersion("2.0.0", false)
	svc.Status.SetDeployedVersion("2.0.0", "", false)
	svc.Status.SetLatestVersion("3.0.0", "", true)

	// LatestVersion
	svc.LatestVersion.Init(
		&svc.Options,
		&svc.Status,
		&latestver_base.Defaults{}, &hardDefaults.Service.LatestVersion)

	// DeployedVersionLookup
	svc.DeployedVersionLookup.Init(
		&svc.Options,
		&svc.Status,
		&deployedver.Defaults{}, &hardDefaults.Service.DeployedVersionLookup)

	// Command
	svc.CommandController.Init(
		&svc.Status,
		&svc.Command,
		&svc.Notify,
		&svc.Options.Interval)

	// WebHook
	svc.WebHook.Init(
		&svc.Status,
		&webhook.SliceDefaults{}, &webhook.Defaults{}, &hardDefaults.WebHook,
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
		"https://valid.release-argus.io/hooks/github-style")
	if failing {
		wh.Secret = "notArgus"
	}
	return wh
}

func testDeployedVersion(t *testing.T) *deployedver.Lookup {
	defaults := &deployedver.Defaults{}
	hardDefaults := &deployedver.Defaults{}
	hardDefaults.Default()

	return test.IgnoreError(t, func() (*deployedver.Lookup, error) {
		return deployedver.New(
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
