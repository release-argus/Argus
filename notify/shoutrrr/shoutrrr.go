// Copyright [2024] [Argus]
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

// Package shoutrrr provides the shoutrrr notification service to services.
package shoutrrr

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	shoutrrr_lib "github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/router"
	"github.com/containrrr/shoutrrr/pkg/types"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

// Send sends a notification with the given title and message to all Shoutrrrs in the Slice.
// It attempts to send each message up to max_tries times until they succeed or fail.
func (s *Slice) Send(
	title, message string,
	serviceInfo util.ServiceInfo,
	useDelay bool,
) error {
	if s == nil {
		return nil
	}

	errChan := make(chan error, len(*s))
	for _, shoutrrr := range *s {
		// Send each message up to max_tries amount of times until they don't err.
		go func(shoutrrr *Shoutrrr) {
			errChan <- shoutrrr.Send(title, message, serviceInfo, useDelay, true)
		}(shoutrrr)

		// Space out Shoutrrr send starts.
		//#nosec G404 -- sleep does not need cryptographic security.
		time.Sleep(time.Duration(100+rand.Intn(150)) * time.Millisecond)
	}

	// Collect the errors.
	var errs []error
	for range *s {
		if err := <-errChan; err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// Send sends a notification with the given title and message.
// It attempts to send the message up to max_tries times until it succeeds.
func (s *Shoutrrr) Send(
	title, msg string,
	serviceInfo util.ServiceInfo,
	useDelay bool,
	useMetrics bool,
) error {
	logFrom := util.LogFrom{Primary: s.ID, Secondary: serviceInfo.ID} // For logging.

	if useDelay && s.GetDelay() != "0s" {
		// Delay sending the Shoutrrr message by the defined interval.
		msg := fmt.Sprintf("%s, Sleeping for %s before sending the Shoutrrr message", s.ID, s.GetDelay())
		jLog.Info(msg, logFrom, s.GetDelay() != "0s")
		time.Sleep(s.GetDelayDuration())
	}

	sender, message, params, url, err := s.getSender(title, msg, serviceInfo)
	if err != nil {
		return err
	}

	// Try sending the message.
	if jLog.IsLevel("VERBOSE") || jLog.IsLevel("DEBUG") {
		msg := fmt.Sprintf("Sending %q to %q", message, url)
		jLog.Verbose(msg, logFrom, !jLog.IsLevel("DEBUG"))
		jLog.Debug(
			fmt.Sprintf("%s with params=%q", msg, *params),
			logFrom, true)
	}
	serviceName := serviceInfo.ID
	if !useMetrics {
		serviceName = ""
	}
	return s.send(
		sender,
		message,
		params,
		serviceName,
		logFrom)
}

// getSender returns the Shoutrrr sender, message, params, and url.
func (s *Shoutrrr) getSender(
	title, msg string,
	serviceInfo util.ServiceInfo,
) (*router.ServiceRouter, string, *types.Params, string, error) {
	// Build the URL.
	url := s.BuildURL()

	// Check the URL provides a valid sender.
	sender, err := shoutrrr_lib.CreateSender(url)
	if err != nil {
		err = fmt.Errorf("failed to create Shoutrrr sender: %w", err)
		return nil, "", nil, "", err
	}

	// Build the params.
	params := s.BuildParams(serviceInfo)
	if title != "" {
		(*params)["title"] = title
	}

	// Build the message.
	message := msg
	if message == "" {
		message = s.Message(serviceInfo)
	}

	return sender, message, params, url, nil
}

// BuildURL returns the URL for this Shoutrrr notification.
func (s *Shoutrrr) BuildURL() (url string) {
	switch s.GetType() {
	case "bark":
		// bark://:devicekey@host:port/[path]
		url = fmt.Sprintf("bark://:%s@%s:%s%s",
			s.GetURLField("devicekey"),
			s.GetURLField("host"),
			s.GetURLField("port"),
			util.ValueUnlessDefault(s.GetURLField("path"), "/"+s.GetURLField("path")))
	case "discord":
		// discord://token@webhookid
		url = fmt.Sprintf("discord://%s@%s",
			s.GetURLField("token"),
			s.GetURLField("webhookid"))
	case "smtp":
		// smtp://username:password@host:port/?fromaddress=X&toaddresses=Y
		login := s.GetURLField("password")
		login = s.GetURLField("username") + util.ValueUnlessDefault(login, ":"+login)
		port := s.GetURLField("port")
		url = fmt.Sprintf("smtp://%s%s%s/?fromaddress=%s&toaddresses=%s",
			util.ValueUnlessDefault(login, login+"@"),
			s.GetURLField("host"),
			util.ValueUnlessDefault(port, ":"+port),
			s.GetParam("fromaddress"),
			s.GetParam("toaddresses"))
	case "gotify":
		// gotify://host:port/path/token
		port := s.GetURLField("port")
		path := s.GetURLField("path")
		url = fmt.Sprintf("gotify://%s%s%s/%s",
			s.GetURLField("host"),
			util.ValueUnlessDefault(port, ":"+port),
			util.ValueUnlessDefault(path, "/"+path),
			s.GetURLField("token"))
	case "googlechat":
		url = s.GetURLField("raw")
		// googlechat://url
		url = fmt.Sprintf("googlechat://%s",
			url)
	case "ifttt":
		// ifttt://webhookid/?events=event1,event2
		url = fmt.Sprintf("ifttt://%s/?events=%s",
			s.GetURLField("webhookid"),
			s.GetParam("events"))
	case "join":
		// join://shoutrrr:apiKey@join/?devices=X
		url = fmt.Sprintf("join://shoutrrr:%s@join/?devices=%s",
			s.GetURLField("apikey"),
			s.GetParam("devices"))
	case "mattermost":
		// mattermost://[username@]host[:port][/path]/token[/channel]
		username := s.GetURLField("username")
		port := s.GetURLField("port")
		path := s.GetURLField("path")
		channel := s.GetURLField("channel")
		url = fmt.Sprintf("mattermost://%s%s%s%s/%s%s",
			util.ValueUnlessDefault(username, username+"@"),
			s.GetURLField("host"),
			util.ValueUnlessDefault(port, ":"+port),
			util.ValueUnlessDefault(path, "/"+path),
			s.GetURLField("token"),
			util.ValueUnlessDefault(channel, "/"+channel))
	case "matrix":
		// matrix://user:password@host[:port][/port]/[?rooms=!roomID1[,roomAlias2]][&disableTLS=yes]
		port := s.GetURLField("port")
		path := s.GetURLField("path")
		rooms := s.GetParam("rooms")
		rooms = util.ValueUnlessDefault(rooms, "?rooms="+rooms)
		disableTLS := s.GetParam("disabletls")
		disableTLS = util.ValueUnlessDefault(disableTLS, "disableTLS="+disableTLS)
		if disableTLS != "" {
			if rooms != "" {
				disableTLS = "&" + disableTLS
			} else {
				disableTLS = "?" + disableTLS
			}
		}
		url = fmt.Sprintf("matrix://%s:%s@%s%s%s/%s%s",
			s.GetURLField("user"),
			s.GetURLField("password"),
			s.GetURLField("host"),
			util.ValueUnlessDefault(port, ":"+port),
			util.ValueUnlessDefault(path, "/"+path),
			rooms,
			disableTLS,
		)
	case "ntfy":
		// ntfy://[username]:[password]@[host][:port]/topic
		url = fmt.Sprintf("ntfy://%s:%s@%s%s/%s",
			s.GetURLField("username"),
			s.GetURLField("password"),
			s.GetURLField("host"),
			util.ValueUnlessDefault(s.GetURLField("port"), ":"+s.GetURLField("port")),
			s.GetURLField("topic"))
	case "opsgenie":
		// opsgenie://host[:port][/path]/apikey
		port := s.GetURLField("port")
		path := s.GetURLField("path")
		url = fmt.Sprintf("opsgenie://%s%s%s/%s",
			s.GetURLField("host"),
			util.ValueUnlessDefault(port, ":"+port),
			util.ValueUnlessDefault(path, "/"+path),
			s.GetURLField("apikey"))
	case "pushbullet":
		// pushbullet://token/targets
		url = fmt.Sprintf("pushbullet://%s/%s",
			s.GetURLField("token"),
			s.GetURLField("targets"))
	case "pushover":
		// pushover://shoutrrr:token@user/[?devices=device1,device2]
		devices := s.GetParam("devices")
		url = fmt.Sprintf("pushover://shoutrrr:%s@%s/%s",
			s.GetURLField("token"),
			s.GetURLField("user"),
			util.ValueUnlessDefault(devices, "?devices="+devices))
	case "rocketchat":
		// rocketchat://[username@]host:port[/port]/tokenA/tokenB/channel
		username := s.GetURLField("username")
		port := s.GetURLField("port")
		path := s.GetURLField("path")
		url = fmt.Sprintf("rocketchat://%s%s%s%s/%s/%s/%s",
			util.ValueUnlessDefault(username, username+"@"),
			s.GetURLField("host"),
			util.ValueUnlessDefault(port, ":"+port),
			util.ValueUnlessDefault(path, "/"+path),
			s.GetURLField("tokena"),
			s.GetURLField("tokenb"),
			s.GetURLField("channel"))
	case "slack":
		// slack://token:token@channel
		url = fmt.Sprintf("slack://%s@%s",
			s.GetURLField("token"),
			s.GetURLField("channel"))
	case "teams":
		// teams://[group@][tenant][/altid][/groupowner]?host=host.example.com
		group := s.GetURLField("group")
		altid := strings.TrimPrefix(s.GetURLField("altid"), "/")
		groupowner := strings.TrimPrefix(s.GetURLField("groupowner"), "/")
		url = fmt.Sprintf("teams://%s%s%s%s?host=%s",
			util.ValueUnlessDefault(group, group+"@"),
			s.GetURLField("tenant"),
			util.ValueUnlessDefault(altid, "/"+altid),
			util.ValueUnlessDefault(groupowner, "/"+groupowner),
			s.GetParam("host"))
		url = strings.Replace(url, "///", "//", 1)
	case "telegram":
		// telegram://token@telegram?chats=@chat1,@chat2
		url = fmt.Sprintf("telegram://%s@telegram?chats=%s",
			s.GetURLField("token"),
			s.GetParam("chats"))
	case "zulip":
		// zulip://botMail:botKey@host?stream=STREAM&topic=TOPIC
		stream := s.GetParam("stream")
		stream = util.ValueUnlessDefault(stream, "?stream="+stream)
		topic := s.GetParam("topic")
		topic = util.ValueUnlessDefault(topic, "&topic="+topic)
		if stream == "" {
			topic = strings.Replace(topic, "&", "?", 1)
		}
		url = fmt.Sprintf("zulip://%s:%s@%s%s%s",
			s.GetURLField("botmail"),
			s.GetURLField("botkey"),
			s.GetURLField("host"),
			stream,
			topic)
	case "generic":
		// generic://example.com:123/api/v1/postStuff
		port := s.GetURLField("port")
		path := s.GetURLField("path")

		// Add the json payload vars, custom headers, and query vars to the url.
		var urlParamsBuilder strings.Builder
		// Separate vars to preserve order.
		jsonMaps := []string{"custom_headers", "json_payload_vars", "query_vars"}
		prefixes := []string{"@", "$", ""}

		first := true // flag to track first entry.
		for i := range jsonMaps {
			urlField := s.GetURLField(jsonMaps[i])
			// Skip non-empty values.
			if urlField == "" {
				continue
			}

			if !first {
				// Add separator before entries after the first.
				urlParamsBuilder.WriteString("&")
			} else {
				// Start the string with '?'.
				urlParamsBuilder.WriteString("?")
				first = false
			}
			urlParamsBuilder.WriteString(jsonMapToString(urlField, prefixes[i]))
		}
		urlParams := urlParamsBuilder.String()

		url = fmt.Sprintf("generic://%s%s%s%s",
			s.GetURLField("host"),
			util.ValueUnlessDefault(port, ":"+port),
			util.ValueUnlessDefault(path, "/"+path),
			urlParams)
	case "shoutrrr":
		// Raw
		url = s.GetURLField("raw")
	}
	return
}

// jsonMapToString returns the JSON param map as an '&' joined list of strings with the prefix added to each key
//
// e.g.
//
//	{"key1": "val1", "key2": "val2"} with prefix '@'
//	returns:
//	@key1=val1&@key2=val2
func jsonMapToString(param string, prefix string) string {
	if param == "" {
		return ""
	}

	// Convert the json string to a map.
	var jsonMap map[string]string
	err := json.Unmarshal([]byte(param), &jsonMap)
	if err != nil {
		return ""
	}

	// Extract and sort keys from the map.
	keys := util.SortedKeys(jsonMap)

	// Build the string.
	var builder strings.Builder
	for i, key := range keys {
		// Add separator before entries after the first.
		if i != 0 {
			builder.WriteString("&")
		}

		builder.WriteString(fmt.Sprintf("%s%s=%s",
			prefix, key, jsonMap[key]))
	}
	return builder.String()
}

// BuildParams returns the params using everything from master>main>defaults>hardDefaults when
// the key is not defined in the lower level.
func (s *Shoutrrr) BuildParams(context util.ServiceInfo) *types.Params {
	params := make(types.Params, len(s.Params)+len(s.Main.Params))

	// Service Params.
	for key, value := range s.Params {
		params[key] = value
	}

	// Main Params.
	for key, value := range s.Main.Params {
		_, exist := s.Params[key]
		// Only overwrite if it doesn't exist in the level below.
		if !exist {
			params[key] = value
		}
	}

	// Default Params.
	for key, value := range s.Defaults.Params {
		_, exist := params[key]
		// Only overwrite if it doesn't exist in the levels below.
		if !exist {
			params[key] = value
		}
	}

	// HardDefault Params.
	for key, value := range s.HardDefaults.Params {
		_, exist := params[key]
		// Only overwrite if it doesn't exist in the levels below.
		if !exist {
			params[key] = value
		}
	}

	// Apply Jinja templating.
	for key, value := range params {
		params[key] = util.TemplateString(value, context)
	}

	return &params
}

// send the message to the Shoutrrr service using the given sender and params.
// It attempts to send the message up to max_tries times until it succeeds.
func (s *Shoutrrr) send(
	sender *router.ServiceRouter,
	message string,
	params *types.Params,
	serviceName string,
	logFrom util.LogFrom,
) error {
	combinedErrs := make(map[string]int)

	if err := util.RetryWithBackoff(
		func() error {
			err := sender.Send(message, params)
			if failed := s.parseSend(err, combinedErrs, serviceName, logFrom); failed {
				return fmt.Errorf("send failed")
			}
			return nil
		},
		s.GetMaxTries(), //#nosec G115 -- Validated in CheckValues
		1*time.Second,
		30*time.Second,
		s.ServiceStatus.Deleting,
	); err == nil {
		return nil
	}

	msg := fmt.Sprintf("failed %d times to send a %s message for %q to %q",
		s.GetMaxTries(), s.GetType(), *s.ServiceStatus.ServiceID, s.BuildURL())
	jLog.Error(msg, logFrom, true)
	failed := true
	s.Failed.Set(s.ID, &failed)
	errs := make([]error, 0, len(combinedErrs))
	for key := range combinedErrs {
		errs = append(errs, fmt.Errorf("%s x %d", key, combinedErrs[key]))
	}
	return errors.Join(errs...)
}

// parseSend processes the errors encountered during the send operation,
// updates the combined error counts, logs the errors, and updates the
// Prometheus metrics based on the success or failure of the operation.
//
// Returns true if the send operation failed over all attempts.
func (s *Shoutrrr) parseSend(
	err []error,
	combinedErrs map[string]int,
	serviceName string,
	logFrom util.LogFrom,
) (failed bool) {
	if s.ServiceStatus.Deleting() {
		return
	}

	for i := range err {
		if err[i] != nil {
			failed = true
			jLog.Error(err[i], logFrom, true)
			combinedErrs[err[i].Error()]++
		}
	}
	// No serviceName, no metrics.
	if serviceName == "" {
		return
	}

	// SUCCESS!
	if !failed {
		metric.IncreasePrometheusCounter(metric.NotifyMetric,
			s.ID,
			serviceName,
			s.GetType(),
			"SUCCESS")
		s.Failed.Set(s.ID, &failed)
		return
	}

	// FAIL.
	metric.IncreasePrometheusCounter(metric.NotifyMetric,
		s.ID,
		serviceName,
		s.GetType(),
		"FAIL")
	return
}
