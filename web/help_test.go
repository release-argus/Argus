// Copyright [2022] [Argus]
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
	"io/ioutil"
	"math/big"
	"net"
	"time"

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/config"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

func boolPtr(val bool) *bool {
	return &val
}
func intPtr(val int) *int {
	return &val
}
func stringPtr(val string) *string {
	return &val
}
func stringifyPointer[T comparable](ptr *T) string {
	str := "nil"
	if ptr != nil {
		str = fmt.Sprint(*ptr)
	}
	return str
}

func testLogging(level string, timestamps bool) {
	jLog = util.NewJLog(level, timestamps)
	service.LogInit(jLog)
}

func testConfig(path string) (cfg *config.Config) {
	testYAML_Argus(path)
	cfg = &config.Config{}

	// Settings.Log
	cfg.Settings.Log.Level = stringPtr("DEBUG")

	cfg.Load(
		path,
		&map[string]bool{},
		jLog)

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
	dvl := testDeployedVersion()
	svc.DeployedVersionLookup = &dvl
	svc.LatestVersion.URLCommands = filter.URLCommandSlice{testURLCommandRegex()}
	emptyNotify := shoutrrr.Shoutrrr{
		Options:   map[string]string{},
		Params:    map[string]string{},
		URLFields: map[string]string{},
	}
	notify := shoutrrr.Slice{
		"test": &shoutrrr.Shoutrrr{
			Options: map[string]string{
				"message": "{{ service_id }} release"},
			Params:       map[string]string{},
			URLFields:    map[string]string{},
			Main:         &emptyNotify,
			Defaults:     &emptyNotify,
			HardDefaults: &emptyNotify},
	}
	notify["test"].Params = map[string]string{}
	svc.Notify = notify
	svc.Comment = "test service's comment"
	cfg.Service = service.Slice{
		svc.ID: svc,
	}

	// Notify
	cfg.Notify = cfg.Defaults.Notify

	// WebHook
	whPass := testWebHook(false, "pass")
	whFail := testWebHook(true, "pass")
	cfg.WebHook = webhook.Slice{
		whPass.ID: whPass,
		whFail.ID: whFail,
	}

	// Order
	cfg.Order = []string{svc.ID}

	return
}

func getFreePort() (int, error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	ln.Close()
	if err != nil {
		return 0, err
	}
	return ln.Addr().(*net.TCPAddr).Port, nil
}

func testService(id string) (svc *service.Service) {
	var (
		sAnnounceChannel chan []byte         = make(chan []byte, 2)
		sDatabaseChannel chan dbtype.Message = make(chan dbtype.Message, 5)
		sSaveChannel     chan bool           = make(chan bool, 5)
	)
	svc = &service.Service{
		ID: id,
		LatestVersion: latestver.Lookup{
			URL:               "https://release-argus.io",
			AccessToken:       stringPtr(""),
			AllowInvalidCerts: boolPtr(false),
			UsePreRelease:     boolPtr(false),
			Require: &filter.Require{
				RegexContent: "content",
				RegexVersion: "version",
				Docker: &filter.DockerCheck{
					Type:  "ghcr",
					Image: "release-argus/argus",
					Tag:   "{{ version }}",
				},
			},
		},
		Options: opt.Options{
			SemanticVersioning: boolPtr(true),
			Interval:           "10m",
			Defaults:           &opt.Options{},
			HardDefaults:       &opt.Options{},
		},
		Dashboard: service.DashboardOptions{
			AutoApprove:  boolPtr(false),
			Icon:         "test",
			Defaults:     &service.DashboardOptions{},
			HardDefaults: &service.DashboardOptions{},
			WebURL:       "https://release-argus.io",
		},
		Defaults:          &service.Service{},
		HardDefaults:      &service.Service{},
		Command:           command.Slice{command.Command{"ls", "-lah"}},
		CommandController: &command.Controller{},
		WebHook:           webhook.Slice{"test": &webhook.WebHook{URL: "example.com"}},
		Status: svcstatus.Status{
			AnnounceChannel: &sAnnounceChannel,
			DatabaseChannel: &sDatabaseChannel,
			SaveChannel:     &sSaveChannel,
		},
	}
	svc.Status.Init(
		len(svc.Notify),
		len(svc.Command), len(svc.WebHook),
		&svc.ID,
		&svc.Dashboard.WebURL)
	svc.LatestVersion.Init(
		&latestver.Lookup{}, &latestver.Lookup{},
		&svc.Status,
		&svc.Options)
	svc.CommandController.Init(
		&svc.Status,
		&svc.Command,
		nil,
		nil)
	return
}

func testCommand(failing bool) command.Command {
	if failing {
		return command.Command{"ls", "-lah", "/root"}
	}
	return command.Command{"ls", "-lah"}
}

func testWebHook(failing bool, id string) *webhook.WebHook {
	var slice *webhook.Slice
	slice.Init(
		nil,
		nil, nil, nil,
		nil,
		nil)

	whDesiredStatusCode := 0
	whMaxTries := uint(1)
	wh := &webhook.WebHook{
		Type:              "github",
		URL:               "https://valid.release-argus.io/hooks/github-style",
		Secret:            "argus",
		AllowInvalidCerts: boolPtr(false),
		DesiredStatusCode: &whDesiredStatusCode,
		Delay:             "0s",
		SilentFails:       boolPtr(false),
		MaxTries:          &whMaxTries,
		ID:                "test",
		ParentInterval:    stringPtr("11m"),
		ServiceStatus:     &svcstatus.Status{},
		Main:              &webhook.WebHook{},
		Defaults:          &webhook.WebHook{},
		HardDefaults:      &webhook.WebHook{},
	}
	if failing {
		wh.Secret = "notArgus"
	}
	return wh
}

func testDeployedVersion() deployedver.Lookup {
	var (
		allowInvalidCerts = false
	)
	return deployedver.Lookup{
		URL:               "https://release-argus.io",
		AllowInvalidCerts: &allowInvalidCerts,
		Headers: []deployedver.Header{
			{Key: "foo", Value: "bar"},
		},
		JSON:  "something",
		Regex: "([0-9]+) The Argus Developers",
		BasicAuth: &deployedver.BasicAuth{
			Username: "fizz",
			Password: "buzz",
		},
		Defaults:     &deployedver.Lookup{},
		HardDefaults: &deployedver.Lookup{},
	}
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
	if err != nil {
		return err
	}

	// Write the certificate and private key to files
	if err := ioutil.WriteFile(certFile, certPEM, 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(keyFile, keyPEM, 0600); err != nil {
		return err
	}

	return nil
}
