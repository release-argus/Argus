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

package shoutrrr

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	shoutrrr_lib "github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/router"
	shoutrrr_types "github.com/containrrr/shoutrrr/pkg/types"
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
)

// BuildParams returns the params using everything from master>main>defaults>hardDefaults when
// the key is not defined in the lower level
func (s *Shoutrrr) BuildParams(context *util.ServiceInfo) (params *shoutrrr_types.Params) {
	p := make(shoutrrr_types.Params, len(s.Params)+len(s.Main.Params))
	params = &p

	// Service Params
	for key := range s.Params {
		(*params)[key] = s.GetParam(key)
	}

	// Main Params
	for key := range s.Main.Params {
		_, exist := s.Params[key]
		// Only overwrite if it doesn't exist in the level below
		if !exist {
			(*params)[key] = s.Main.GetParam(key)
		}
	}

	// Default Params
	for key := range s.Defaults.Params {
		_, exist := (*params)[key]
		// Only overwrite if it doesn't exist in the levels below
		if !exist {
			(*params)[key] = s.Defaults.GetParam(key)
		}
	}

	// HardDefault Params
	for key := range s.HardDefaults.Params {
		_, exist := (*params)[key]
		// Only overwrite if it doesn't exist in the levels below
		if !exist {
			(*params)[key] = s.HardDefaults.GetParam(key)
		}
	}

	// Apply Jinja templating
	for key := range *params {
		(*params)[key] = util.TemplateString((*params)[key], *context)
	}

	return
}

func (s *Shoutrrr) BuildURL() (url string) {
	switch s.GetType() {
	case "bark":
		// bark://:devicekey@host:port/[path]
		url = fmt.Sprintf("bark://:%s@%s:%s%s",
			s.GetURLField("devicekey"),
			s.GetURLField("host"),
			s.GetURLField("port"),
			util.ValueIfNotDefault(s.GetURLField("path"), "/"+s.GetURLField("path")))
	case "discord":
		// discord://token@webhookid
		url = fmt.Sprintf("discord://%s@%s",
			s.GetURLField("token"),
			s.GetURLField("webhookid"))
	case "smtp":
		// smtp://username:password@host:port/?fromaddress=X&toaddresses=Y
		login := s.GetURLField("password")
		login = s.GetURLField("username") + util.ValueIfNotDefault(login, ":"+login)
		port := s.GetURLField("port")
		url = fmt.Sprintf("smtp://%s%s%s/?fromaddress=%s&toaddresses=%s",
			util.ValueIfNotDefault(login, login+"@"),
			s.GetURLField("host"),
			util.ValueIfNotDefault(port, ":"+port),
			s.GetParam("fromaddress"),
			s.GetParam("toaddresses"))
	case "gotify":
		// gotify://host:port/path/token
		port := s.GetURLField("port")
		path := s.GetURLField("path")
		url = fmt.Sprintf("gotify://%s%s%s/%s",
			s.GetURLField("host"),
			util.ValueIfNotDefault(port, ":"+port),
			util.ValueIfNotDefault(path, "/"+path),
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
			util.ValueIfNotDefault(username, username+"@"),
			s.GetURLField("host"),
			util.ValueIfNotDefault(port, ":"+port),
			util.ValueIfNotDefault(path, "/"+path),
			s.GetURLField("token"),
			util.ValueIfNotDefault(channel, "/"+channel))
	case "matrix":
		// matrix://user:password@host[:port][/port]/[?rooms=!roomID1[,roomAlias2]][&disableTLS=yes]
		port := s.GetURLField("port")
		path := s.GetURLField("path")
		rooms := s.GetParam("rooms")
		rooms = util.ValueIfNotDefault(rooms, "?rooms="+rooms)
		disableTLS := s.GetParam("disabletls")
		disableTLS = util.ValueIfNotDefault(disableTLS, "disableTLS="+disableTLS)
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
			util.ValueIfNotDefault(port, ":"+port),
			util.ValueIfNotDefault(path, "/"+path),
			rooms,
			disableTLS,
		)
	case "ntfy":
		// ntfy://[username]:[password]@[host][:port]/topic
		url = fmt.Sprintf("ntfy://%s:%s@%s%s/%s",
			s.GetURLField("username"),
			s.GetURLField("password"),
			s.GetURLField("host"),
			util.ValueIfNotDefault(s.GetURLField("port"), ":"+s.GetURLField("port")),
			s.GetURLField("topic"))
	case "opsgenie":
		// opsgenie://host[:port][/path]/apikey
		port := s.GetURLField("port")
		path := s.GetURLField("path")
		url = fmt.Sprintf("opsgenie://%s%s%s/%s",
			s.GetURLField("host"),
			util.ValueIfNotDefault(port, ":"+port),
			util.ValueIfNotDefault(path, "/"+path),
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
			util.ValueIfNotDefault(devices, "?devices="+devices))
	case "rocketchat":
		// rocketchat://[username@]host:port[/port]/tokenA/tokenB/channel
		username := s.GetURLField("username")
		port := s.GetURLField("port")
		path := s.GetURLField("path")
		url = fmt.Sprintf("rocketchat://%s%s%s%s/%s/%s/%s",
			util.ValueIfNotDefault(username, username+"@"),
			s.GetURLField("host"),
			util.ValueIfNotDefault(port, ":"+port),
			util.ValueIfNotDefault(path, "/"+path),
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
			util.ValueIfNotDefault(group, group+"@"),
			s.GetURLField("tenant"),
			util.ValueIfNotDefault(altid, "/"+altid),
			util.ValueIfNotDefault(groupowner, "/"+groupowner),
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
		stream = util.ValueIfNotDefault(stream, "?stream="+stream)
		topic := s.GetParam("topic")
		topic = util.ValueIfNotDefault(topic, "&topic="+topic)
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
		// Add the json payload vars, custom headers, and query vars to the url
		urlParams := "?"
		// Separate vars to preserve order
		jsonMaps := []string{"custom_headers", "json_payload_vars", "query_vars"}
		prefixes := []string{"@", "$", ""}
		for index := range jsonMaps {
			urlParams += jsonMapToString(s.GetURLField(jsonMaps[index]), prefixes[index])
		}
		if len(urlParams) > 1 {
			urlParams = strings.TrimSuffix(urlParams, "&")
		} else {
			urlParams = ""
		}
		url = fmt.Sprintf("generic://%s%s%s%s",
			s.GetURLField("host"),
			util.ValueIfNotDefault(port, ":"+port),
			util.ValueIfNotDefault(path, "/"+path),
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
// {"key1": "val1", "key2": "val2"} with prefix '@' returns:
//
// @key1=val1&@key2=val2&
func jsonMapToString(param string, prefix string) (converted string) {
	if param == "" {
		return
	}

	// Convert the json string to a map
	jsonMap := make(map[string]interface{}, 0)
	err := json.Unmarshal([]byte(param), &jsonMap)
	if err != nil {
		return
	}
	// Extract and sort keys from the map
	keys := util.SortedKeys(jsonMap)
	// Build the string
	for _, key := range keys {
		// Formatted as '{prefix}{key}={value}&'
		converted += fmt.Sprintf("%s%s=%v&", prefix, key, jsonMap[key])
	}
	return
}

func (s *Slice) Send(
	title string,
	message string,
	serviceInfo *util.ServiceInfo,
	useDelay bool,
) (errs error) {
	if s == nil {
		return nil
	}
	if serviceInfo == nil {
		serviceInfo = &util.ServiceInfo{}
	}

	errChan := make(chan error)
	for key := range *s {
		// Send each message up to s.MaxTries number of times until they don't err.
		go func(shoutrrr *Shoutrrr) {
			errChan <- shoutrrr.Send(title, message, serviceInfo, useDelay, true)
		}((*s)[key])

		// Space out Shoutrrr send starts.
		time.Sleep(200 * time.Millisecond)
	}

	for range *s {
		err := <-errChan
		if err != nil {
			errs = fmt.Errorf("%s\n%w",
				util.ErrorToString(errs), err)
		}
	}
	return
}

// getSender returns the Shoutrrr sender, message, params, and url.
func (s *Shoutrrr) getSender(
	title string,
	msg string,
	serviceInfo *util.ServiceInfo,
) (sender *router.ServiceRouter, message string, params *shoutrrr_types.Params, url string, err error) {
	url = s.BuildURL()
	sender, err = shoutrrr_lib.CreateSender(url)
	if err != nil {
		err = fmt.Errorf("failed to create Shoutrrr sender: %w", err)
		return
	}
	params = s.BuildParams(serviceInfo)
	if title != "" {
		(*params)["title"] = title
	}
	message = msg
	if message == "" {
		message = s.Message(serviceInfo)
	}
	return
}

// Send the Shoutrrr message with the given title and message (if non-empty).
func (s *Shoutrrr) Send(
	title string,
	msg string,
	serviceInfo *util.ServiceInfo,
	useDelay bool,
	useMetrics bool,
) (errs error) {
	logFrom := util.LogFrom{Primary: s.ID, Secondary: serviceInfo.ID} // For logging

	if useDelay && s.GetDelay() != "0s" {
		// Delay sending the Shoutrrr message by the defined interval.
		msg := fmt.Sprintf("%s, Sleeping for %s before sending the Shoutrrr message", s.ID, s.GetDelay())
		jLog.Info(msg, logFrom, s.GetDelay() != "0s")
		time.Sleep(s.GetDelayDuration())
	}

	sender, message, params, url, errs := s.getSender(title, msg, serviceInfo)
	if errs != nil {
		return
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
	errs = s.send(
		sender,
		message,
		params,
		serviceName,
		logFrom)
	return
}

// parseSend logs and counts the errors from a Shoutrrr send, returning true if the send failed.
//
// If the serviceName is empty, no metrics are recorded.
func (s *Shoutrrr) parseSend(
	err []error,
	combinedErrs map[string]int,
	serviceName string,
	logFrom util.LogFrom,
) (failed bool) {
	for i := range err {
		if err[i] != nil {
			failed = true
			jLog.Error(err[i].Error(), logFrom, true)
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

	// FAIL
	metric.IncreasePrometheusCounter(metric.NotifyMetric,
		s.ID,
		serviceName,
		s.GetType(),
		"FAIL")
	return
}

func (s *Shoutrrr) send(
	sender *router.ServiceRouter,
	message string,
	params *shoutrrr_types.Params,
	serviceName string,
	logFrom util.LogFrom,
) (errs error) {
	combinedErrs := make(map[string]int)
	triesLeft := s.GetMaxTries() // Number of times to send Shoutrrr (until 200 received).

	for triesLeft > 0 {
		// Check if we're deleting the Service.
		if s.ServiceStatus.Deleting() {
			return
		}
		err := sender.Send(message, params)

		failed := s.parseSend(err, combinedErrs, serviceName, logFrom)
		if !failed {
			return
		}
		triesLeft--

		// Space out retries.
		if triesLeft > 0 {
			time.Sleep(5 * time.Second)
		}
	}

	// Give up after MaxTries.
	msg := fmt.Sprintf("failed %d times to send a %s message for %q to %q",
		s.GetMaxTries(), s.GetType(), *s.ServiceStatus.ServiceID, s.BuildURL())
	jLog.Error(msg, logFrom, true)
	failed := true
	s.Failed.Set(s.ID, &failed)
	for key := range combinedErrs {
		errs = fmt.Errorf("%s%s x %d",
			util.ErrorToString(errs), key, combinedErrs[key])
	}
	return
}
