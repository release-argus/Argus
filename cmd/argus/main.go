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

/*
Argus monitors GitHub and/or other URLs for version changes.
On a version change, send notifications and/or webhooks.
*/
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	_ "modernc.org/sqlite"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/db"
	"github.com/release-argus/Argus/testing"
	logutil "github.com/release-argus/Argus/util/log"
	"github.com/release-argus/Argus/web"
)

var (
	configFile = flag.String(
		"config.file",
		"config.yml",
		"Argus configuration file path.")
	configCheckFlag = flag.Bool(
		"config.check",
		false,
		"Print the fully-parsed config.")
	testCommandsFlag = flag.String(
		"test.commands",
		"",
		"Put the name of the Service to test the `commands` of.")
	testNotifyFlag = flag.String(
		"test.notify",
		"",
		"Put the name of the Notify service to send a test message.")
	testServiceFlag = flag.String(
		"test.service",
		"",
		"Put the name of the Service to test the version query.")
)

// run loads the config and then calls `service.Track` to monitor
// each Service of the config for version changes and acts on
// them as defined. It also sets up the DB, Web UI and SaveHandler.
func run() (exitCode int) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	g, gctx := errgroup.WithContext(ctx)

	flag.Parse()
	flags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flags[f.Name] = true })

	// Initialise the Log.
	exitCodeChannel := logutil.Init("ERROR", false)

	var cfg config.Config
	_ = cfg.Load(ctx, g, *configFile, &flags)

	// config.check
	cfg.Print(configCheckFlag)
	// test.*
	testing.RunAndExit(testing.CommandTest(testCommandsFlag, &cfg), testCommandsFlag)
	testing.RunAndExit(testing.NotifyTest(testNotifyFlag, &cfg), testNotifyFlag)
	testing.RunAndExit(testing.ServiceTest(testServiceFlag, &cfg), testServiceFlag)

	// Count of active services to monitor (if log level INFO or above).
	if logutil.Log.Level > 1 {
		// Count active services.
		serviceCount := len(cfg.Order)
		for _, key := range cfg.Order {
			if !cfg.Service[key].Options.GetActive() {
				serviceCount--
			}
		}

		// Log active count.
		msg := fmt.Sprintf("Found %d services to monitor:", serviceCount)
		logutil.Log.Info(msg, logutil.LogFrom{}, true)

		// Log names of active services.
		for _, key := range cfg.Order {
			if cfg.Service[key].Options.GetActive() {
				fmt.Printf("  - %s\n", cfg.Service[key].Name)
			}
		}
	}

	// Setup DB and last known service versions.
	g.Go(func() error {
		if ok := db.Run(gctx, &cfg); !ok {
			return errors.New("db.Run failed")
		}
		return nil
	})

	// Track all targets for changes in version and act on any found changes.
	go cfg.Service.Track(&cfg.Order, &cfg.OrderMutex)

	// Web server.
	g.Go(func() error {
		return web.Run(gctx, &cfg)
	})

	// Wait for cancellation.
	select {
	case <-exitCodeChannel:
		exitCode = 1
		cancel()
	case <-ctx.Done():
		// OS signal cancel received.
	}

	// Begin shutdown.
	logutil.Log.Info("Shutting down...",
		logutil.LogFrom{}, true)
	// Give goroutines time to finish.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		_ = g.Wait()
		close(done)
	}()

	select {
	case <-done:
		logutil.Log.Info("Shutdown complete",
			logutil.LogFrom{}, true)
	case <-shutdownCtx.Done():
		logutil.Log.Error(shutdownCtx.Err(),
			logutil.LogFrom{}, true,
		)
		exitCode = 1
	}

	return
}

func main() {
	os.Exit(run())
}
