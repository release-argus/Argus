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

//go:build unit

package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"testing"
	"time"

	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
)

func TestApplyTemplate(t *testing.T) {
	// GIVEN various Command's
	tests := map[string]struct {
		input         Command
		want          Command
		serviceStatus *service_status.Status
	}{
		"command with no templating and non-nil service status": {
			input: Command{"ls", "-lah"}, want: Command{"ls", "-lah"},
			serviceStatus: &service_status.Status{LatestVersion: "1.2.3"},
		},
		"command with no templating and nil service status": {input: Command{"ls", "-lah"}, want: Command{"ls", "-lah"}},
		"command with templating and nil service status":    {input: Command{"ls", "-lah", "{{ version }}"}, want: Command{"ls", "-lah", "{{ version }}"}},
		"command with templating and non-nil service status": {input: Command{"ls", "-lah", "{{ version }}"}, want: Command{"ls", "-lah", "1.2.3"},
			serviceStatus: &service_status.Status{LatestVersion: "1.2.3"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN ApplyTemplate is called on the Command
			got := tc.input.ApplyTemplate(tc.serviceStatus)

			// THEN the result is expected
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("%s:\nwant: %v\ngot:  %v",
					name, tc.want, got)
			}
		})
	}
}

func TestCommandExec(t *testing.T) {
	// GIVEN different Command's to execute
	jLog = utils.NewJLog("INFO", false)
	tests := map[string]struct {
		cmd         Command
		err         error
		outputRegex string
	}{
		"command that will pass": {cmd: Command{"date", "+%m-%d-%Y"}, err: nil, outputRegex: `[0-9]{2}-[0-9]{2}-[0-9]{4}\s+$`},
		"command that will fail": {cmd: Command{"false"}, err: fmt.Errorf("exit status 1"), outputRegex: `exit status 1\s+$`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN Exec is called on it
			err := tc.cmd.Exec(&utils.LogFrom{})

			// THEN the output is expected
			if utils.ErrorToString(err) != utils.ErrorToString(tc.err) {
				t.Fatalf("%s:\nerr's differ\nwant: %s\ngot:  %s",
					name, tc.err, err)
			}
			w.Close()
			out, _ := ioutil.ReadAll(r)
			os.Stdout = stdout
			output := string(out)
			re := regexp.MustCompile(tc.outputRegex)
			match := re.MatchString(output)
			if !match {
				t.Errorf("%s:\nwant match for %q\nnot: %q",
					name, tc.outputRegex, output)
			}
		})
	}
}

func TestExecIndex(t *testing.T) {
	// GIVEN a Controller with different Command's to execute
	jLog = utils.NewJLog("INFO", false)
	announce := make(chan []byte, 8)
	controller := Controller{
		ServiceID: stringPtr("service_id"),
		Command: &Slice{
			{"date", "+%m-%d-%Y"},
			{"false"}},
		Failed:         &[]*bool{nil, nil},
		NextRunnable:   make([]time.Time, 2),
		ParentInterval: stringPtr("10m"),
		ServiceStatus:  &service_status.Status{AnnounceChannel: &announce},
	}
	tests := map[string]struct {
		index       int
		err         error
		outputRegex string
		noAnnounce  bool
	}{
		"command index out of range":   {index: 2, err: nil, outputRegex: `^$`, noAnnounce: true},
		"command index that will pass": {index: 0, err: nil, outputRegex: `[0-9]{2}-[0-9]{2}-[0-9]{4}\s+$`},
		"command index that will fail": {index: 1, err: fmt.Errorf("exit status 1"), outputRegex: `exit status 1\s+$`},
	}

	runNumber := 0
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN the Command @index is exectured
			err := controller.ExecIndex(&utils.LogFrom{}, tc.index)

			// THEN the output is expected
			// err
			if utils.ErrorToString(err) != utils.ErrorToString(tc.err) {
				t.Fatalf("%s:\nerr's differ\nwant: %s\ngot:  %s",
					name, tc.err, err)
			}
			// output
			w.Close()
			out, _ := ioutil.ReadAll(r)
			os.Stdout = stdout
			output := string(out)
			re := regexp.MustCompile(tc.outputRegex)
			match := re.MatchString(output)
			if !match {
				t.Fatalf("%s:\nwant match for %q\nnot: %q",
					name, tc.outputRegex, output)
			}
			// announced
			if !tc.noAnnounce {
				runNumber++
			}
			if len(announce) != runNumber {
				t.Fatalf("%s:\nCommand run not announced\nat %d, want %d",
					name, len(announce), runNumber)
			}
		})
	}
}

func TestControllerExec(t *testing.T) {
	// GIVEN a Controller
	jLog = utils.NewJLog("INFO", false)
	announce := make(chan []byte, 8)
	controller := Controller{
		ServiceID:      stringPtr("service_id"),
		Failed:         &[]*bool{nil, nil},
		NextRunnable:   make([]time.Time, 2),
		ParentInterval: stringPtr("10m"),
		ServiceStatus:  &service_status.Status{AnnounceChannel: &announce},
	}
	tests := map[string]struct {
		nilController bool
		commands      *Slice
		err           error
		outputRegex   string
		noAnnounce    bool
	}{
		"nil Controller": {nilController: true, err: nil, outputRegex: `^$`, noAnnounce: true},
		"nil Command":    {err: nil, outputRegex: `^$`, noAnnounce: true},
		"single Command": {err: nil, outputRegex: `[0-9]{2}-[0-9]{2}-[0-9]{4}\s+$`,
			commands: &Slice{
				{"date", "+%m-%d-%Y"}}},
		"multiple Command's": {err: fmt.Errorf("\nexit status 1"), outputRegex: `[0-9]{2}-[0-9]{2}-[0-9]{4}\s+.*'false'\s.*exit status 1\s+$`,
			commands: &Slice{
				{"date", "+%m-%d-%Y"},
				{"false"}}},
	}

	runNumber := 0
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN the Command @index is exectured
			var err error
			if tc.nilController {
				var controller *Controller
				err = controller.Exec(&utils.LogFrom{})
			} else {
				controller.Command = tc.commands
				err = controller.Exec(&utils.LogFrom{})
			}

			// THEN the output is expected
			// err
			if utils.ErrorToString(err) != utils.ErrorToString(tc.err) {
				t.Fatalf("%s:\nerr's differ\nwant: %q\ngot:  %q",
					name, utils.ErrorToString(tc.err), utils.ErrorToString(err))
			}
			// output
			w.Close()
			out, _ := ioutil.ReadAll(r)
			os.Stdout = stdout
			output := string(out)
			re := regexp.MustCompile(tc.outputRegex)
			match := re.MatchString(output)
			if !match {
				t.Fatalf("%s:\nwant match for %q\nnot: %q",
					name, tc.outputRegex, output)
			}
			// announced
			if !tc.noAnnounce {
				runNumber += len(*controller.Command)
			}
			if len(announce) != runNumber {
				t.Fatalf("%s:\nCommand run not announced\nat %d, want %d",
					name, len(announce), runNumber)
			}
		})
	}
}
