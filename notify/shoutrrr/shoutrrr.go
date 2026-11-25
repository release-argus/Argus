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

// Package shoutrrr provides the shoutrrr notification service to services.
package shoutrrr

import (
	"errors"
	"fmt"
	"maps"
	"math/rand"
	net_url "net/url"
	"slices"
	"strings"
	"time"

	"github.com/nicholas-fedor/shoutrrr"
	"github.com/nicholas-fedor/shoutrrr/pkg/format"
	"github.com/nicholas-fedor/shoutrrr/pkg/router"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

// BuildURL returns the URL for this Shoutrrr notification.
func (s *Shoutrrr) BuildURL() (url string) {
	switch s.GetType() {
	case "bark":
		// bark://:devicekey@host[:port][/path]
		path := s.GetURLField("path")

		url = fmt.Sprintf(
			"bark://:%s@%s:%s%s",
			s.GetURLField("devicekey"),
			s.GetURLField("host"),
			s.GetURLField("port"),
			util.ValueUnlessZero(path, "/"+path),
		)
	case "discord":
		// discord://token@webhookid[?thread_id=X]
		threadID := s.GetParam("thread_id")
		query := buildQuery(
			queryParam("thread_id", threadID),
		)

		url = fmt.Sprintf(
			"discord://%s@%s%s",
			s.GetURLField("token"),
			s.GetURLField("webhookid"),
			query,
		)
	case "smtp":
		// smtp://username:password@host[:port]/?fromaddress=X&toaddresses=Y[&fromname=X]
		login := s.GetURLField("password")
		login = s.GetURLField("username") + util.ValueUnlessZero(login, ":"+login)
		port := s.GetURLField("port")
		fromAddress := s.GetParam("fromaddress")
		fromName := s.GetParam("fromname")
		toAddresses := s.GetParam("toaddresses")
		query := buildQuery(
			queryParam("fromaddress", fromAddress),
			queryParam("fromname", fromName),
			queryParam("toaddresses", toAddresses),
		)

		url = fmt.Sprintf(
			"smtp://%s%s%s/%s",
			util.ValueUnlessZero(login, login+"@"),
			s.GetURLField("host"),
			util.ValueUnlessZero(port, ":"+port),
			query,
		)
	case "gotify":
		// gotify://host[:port][/path]/token
		port := s.GetURLField("port")
		path := s.GetURLField("path")

		url = fmt.Sprintf(
			"gotify://%s%s%s/%s",
			s.GetURLField("host"),
			util.ValueUnlessZero(port, ":"+port),
			util.ValueUnlessZero(path, "/"+path),
			s.GetURLField("token"),
		)
	case "googlechat":
		// googlechat://url
		url = s.GetURLField("raw")

		url = "googlechat://" + url
	case "homeassistant":
		// homeassistant://token@host[:port][/path]/[?disabletls=yes][&service=X][&skiptlsverify=yes]
		port := s.GetURLField("port")
		path := s.GetURLField("path")
		// Sent in the URL as Shoutrrr resolves them when it parses it,
		// not when the message is sent.
		query := buildQuery(
			queryParam("disabletls", s.GetParam("disabletls")),
			queryParam("service", s.GetParam("service")),
			queryParam("skiptlsverify", s.GetParam("skiptlsverify")),
		)

		url = fmt.Sprintf(
			"homeassistant://%s@%s%s%s/%s",
			s.GetURLField("token"),
			s.GetURLField("host"),
			util.ValueUnlessZero(port, ":"+port),
			util.ValueUnlessZero(path, "/"+path),
			query,
		)
	case "ifttt":
		// ifttt://webhookid/?events=event1,event2
		url = fmt.Sprintf(
			"ifttt://%s/?events=%s",
			s.GetURLField("webhookid"),
			s.GetParam("events"),
		)
	case "join":
		// join://shoutrrr:apiKey@join/?devices=X
		url = fmt.Sprintf(
			"join://shoutrrr:%s@join/?devices=%s",
			s.GetURLField("apikey"),
			s.GetParam("devices"),
		)
	case "lark":
		// lark://host/token
		url = fmt.Sprintf(
			"lark://%s/%s",
			s.GetURLField("host"),
			s.GetURLField("token"),
		)
	case "matrix":
		// matrix://user:password@host[:port]/[?disabletls=yes][&rooms=!roomID1[,roomAlias2]]
		port := s.GetURLField("port")
		disableTLS := s.GetParam("disabletls")
		rooms := s.GetParam("rooms")
		query := buildQuery(
			queryParam("disabletls", disableTLS),
			queryParam("rooms", rooms),
		)

		url = fmt.Sprintf(
			"matrix://%s:%s@%s%s/%s",
			s.GetURLField("user"),
			s.GetURLField("password"),
			s.GetURLField("host"),
			util.ValueUnlessZero(port, ":"+port),
			query,
		)
	case "mattermost":
		// mattermost://[username@]host[:port]/token[/channel]
		username := s.GetURLField("username")
		port := s.GetURLField("port")
		channel := s.GetURLField("channel")

		url = fmt.Sprintf(
			"mattermost://%s%s%s/%s%s",
			util.ValueUnlessZero(username, username+"@"),
			s.GetURLField("host"),
			util.ValueUnlessZero(port, ":"+port),
			s.GetURLField("token"),
			util.ValueUnlessZero(channel, "/"+channel),
		)
	case "mqtt":
		// mqtt://[username[:password]@]host[:port]/topic[?disabletls=X]
		port := s.GetURLField("port")

		url = fmt.Sprintf(
			"mqtt://%s%s%s/%s%s",
			userInfo(s.GetURLField("username"), s.GetURLField("password")),
			s.GetURLField("host"),
			util.ValueUnlessZero(port, ":"+port),
			s.GetURLField("topic"),
			buildQuery(queryParam("disabletls", s.GetParam("disabletls"))),
		)
	case "notifiarr":
		// notifiarr://apikey[?channel=X&color=X&image=X&name=X&thumbnail=X]
		channel := s.GetParam("channel")
		color := s.GetParam("color")
		image := s.GetParam("image")
		name := s.GetParam("name")
		thumbnail := s.GetParam("thumbnail")
		query := buildQuery(
			queryParam("channel", channel),
			queryParam("color", color),
			queryParam("image", image),
			queryParam("name", name),
			queryParam("thumbnail", thumbnail),
		)

		url = fmt.Sprintf("notifiarr://%s%s",
			s.GetURLField("apikey"),
			query,
		)
	case "ntfy":
		// ntfy://[username]:[password]@[host][:port]/topic
		port := s.GetURLField("port")

		url = fmt.Sprintf(
			"ntfy://%s:%s@%s%s/%s",
			s.GetURLField("username"),
			s.GetURLField("password"),
			s.GetURLField("host"),
			util.ValueUnlessZero(port, ":"+port),
			s.GetURLField("topic"),
		)
	case "opsgenie": // TODO: OpsGenie permanently shut down April 5, 2027
		// opsgenie://host[:port]/apikey
		port := s.GetURLField("port")

		url = fmt.Sprintf(
			"opsgenie://%s%s/%s",
			s.GetURLField("host"),
			util.ValueUnlessZero(port, ":"+port),
			s.GetURLField("apikey"),
		)
	case "pagerduty":
		// pagerduty://[host[:port]]/integrationkey
		port := s.GetURLField("port")

		url = fmt.Sprintf(
			"pagerduty://%s%s/%s",
			s.GetURLField("host"),
			util.ValueUnlessZero(port, ":"+port),
			s.GetURLField("integrationkey"),
		)
	case "pushbullet":
		// pushbullet://token/targets
		url = fmt.Sprintf(
			"pushbullet://%s/%s",
			s.GetURLField("token"),
			s.GetURLField("targets"),
		)
	case "pushover":
		// pushover://shoutrrr:token@user/[?devices=device1[,device2]]
		devices := s.GetParam("devices")

		url = fmt.Sprintf(
			"pushover://shoutrrr:%s@%s/%s",
			s.GetURLField("token"),
			s.GetURLField("user"),
			util.ValueUnlessZero(devices, "?devices="+devices),
		)
	case "rocketchat":
		// rocketchat://[username@]host[:port]/tokenA/tokenB[/channel]
		username := s.GetURLField("username")
		port := s.GetURLField("port")
		channel := s.GetURLField("channel")

		url = fmt.Sprintf(
			"rocketchat://%s%s%s/%s/%s%s",
			util.ValueUnlessZero(username, username+"@"),
			s.GetURLField("host"),
			util.ValueUnlessZero(port, ":"+port),
			s.GetURLField("tokena"),
			s.GetURLField("tokenb"),
			util.ValueUnlessZero(channel, "/"+channel),
		)
	case "signal":
		// signal://[user:password@]host[:port]/source/recipient1[/recipient2...]
		port := s.GetURLField("port")

		url = fmt.Sprintf(
			"signal://%s%s%s/%s/%s",
			userInfo(s.GetURLField("user"), s.GetURLField("password")),
			s.GetURLField("host"),
			util.ValueUnlessZero(port, ":"+port),
			s.GetURLField("source"),
			strings.ReplaceAll(s.GetURLField("recipients"), ",", "/"),
		)
	case "signalgrid":
		// signalgrid://clientkey@channel
		url = fmt.Sprintf(
			"signalgrid://%s@%s",
			s.GetURLField("clientkey"),
			s.GetURLField("channel"),
		)
	case "slack":
		// slack://token:token@channel[?botname=X][&color=X][&icon=X][&thread_ts=X][&title=X]
		botName := s.GetParam("botname")
		color := s.GetParam("color")
		icon := s.GetParam("icon")
		threadTS := s.GetParam("thread_ts")
		title := s.GetParam("title")
		query := buildQuery(
			queryParam("botname", botName),
			queryParam("color", color),
			queryParam("icon", icon),
			queryParam("thread_ts", threadTS),
			queryParam("title", title),
		)

		url = fmt.Sprintf(
			"slack://%s@%s%s",
			s.GetURLField("token"),
			s.GetURLField("channel"),
			query,
		)
	case "teams":
		// teams://?host=fullPowerAutomateURL[&color=X][&title=X]
		host := s.GetParam("host")
		color := s.GetParam("color")
		title := s.GetParam("title")
		query := buildQuery(
			queryParam("host", host),
			queryParam("color", color),
			queryParam("title", title),
		)

		url = "teams://" + query
	case "telegram":
		// telegram://token@telegram?chats=@chat1,@chat2
		url = fmt.Sprintf(
			"telegram://%s@telegram?chats=%s",
			s.GetURLField("token"),
			s.GetParam("chats"),
		)
	case "twilio":
		// twilio://accountsid:authtoken@fromnumber/tonumber1[/tonumber2...]
		url = fmt.Sprintf(
			"twilio://%s%s/%s",
			userInfo(s.GetURLField("accountsid"), s.GetURLField("authtoken")),
			s.GetURLField("fromnumber"),
			strings.ReplaceAll(s.GetURLField("tonumbers"), ",", "/"),
		)
	case "wecom":
		// wecom://key
		url = "wecom://" + s.GetURLField("key")
	case "xmpp":
		// xmpp://user:password@host[:port]/[?rooms=X][&to=X]
		port := s.GetURLField("port")
		query := buildQuery(
			queryParam("rooms", s.GetParam("rooms")),
			queryParam("to", s.GetParam("to")),
		)

		url = fmt.Sprintf(
			"xmpp://%s%s%s/%s",
			userInfo(s.GetURLField("user"), s.GetURLField("password")),
			s.GetURLField("host"),
			util.ValueUnlessZero(port, ":"+port),
			query,
		)
	case "zulip":
		// zulip://botMail:botKey@host[:port][?read_by_sender=true&stream=X&title=X&to=X&topic=X&type=X]
		port := s.GetURLField("port")
		readBySender := s.GetParam("read_by_sender")
		stream := s.GetParam("stream")
		title := s.GetParam("title")
		to := s.GetParam("to")
		topic := s.GetParam("topic")
		msgType := s.GetParam("type")
		query := buildQuery(
			queryParam("read_by_sender", readBySender),
			queryParam("stream", stream),
			queryParam("title", title),
			queryParam("to", to),
			queryParam("topic", topic),
			queryParam("type", msgType),
		)

		url = fmt.Sprintf(
			"zulip://%s:%s@%s%s%s",
			s.GetURLField("botmail"),
			s.GetURLField("botkey"),
			s.GetURLField("host"),
			util.ValueUnlessZero(port, ":"+port),
			query,
		)
	case "generic":
		// generic://[username[:password]@]example.com[:port][/path]
		port := s.GetURLField("port")
		path := s.GetURLField("path")

		// Add the JSON payload vars, custom headers, and query vars to the url.
		// Separate vars to preserve order.
		jsonMaps := []string{"headers", "json_payload_vars", "query_vars"}
		prefixes := []string{"@", "$", ""}

		parts := make([]string, len(jsonMaps))
		for i := range jsonMaps {
			parts[i] = jsonMapToString(s.GetURLField(jsonMaps[i]), prefixes[i])
		}
		urlParams := buildQuery(parts...)

		url = fmt.Sprintf(
			"generic://%s%s%s%s%s",
			userInfo(s.GetURLField("username"), s.GetURLField("password")),
			s.GetURLField("host"),
			util.ValueUnlessZero(port, ":"+port),
			util.ValueUnlessZero(path, "/"+path),
			urlParams,
		)
	case "shoutrrr":
		// Raw
		url = s.GetURLField("raw")
	}
	return
}

// BuildParams returns the merged Params, resolving each key from instance, Main, Defaults, and HardDefaults in order.
func (s *Shoutrrr) BuildParams(info serviceinfo.ServiceInfo) *types.Params {
	params := make(types.Params, len(s.Params)+len(s.Main.Params))

	// Service Params.
	maps.Copy(params, s.Params)

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

	// Apply django templating.
	for key, value := range params {
		value = util.EvalEnvVars(value)
		params[key] = util.TemplateString(value, info)
	}

	// 'Unknown' auth means "no method given". Let Shoutrrr handle it.
	if s.GetType() == "smtp" && strings.EqualFold(params["auth"], smtpAuthUnknown) {
		delete(params, "auth")
	}

	// ntfy 'disabletls_' holds Shoutrrr's 'disabletls' until the next major release.
	if disableTLS, ok := params["disabletls_"]; ok && s.GetType() == "ntfy" {
		params["disabletls"] = disableTLS
		delete(params, "disabletls_")
	}

	return &params
}

// Send sends a notification to every Shoutrrr in the map concurrently.
func (s *Shoutrrrs) Send(
	title, message string,
	serviceInfo serviceinfo.ServiceInfo,
	useDelay bool,
) error {
	if s == nil {
		return nil
	}

	errChan := make(chan error, len(*s))
	for _, shoutrrr := range *s {
		go func(shoutrrr *Shoutrrr) {
			errChan <- shoutrrr.Send(title, message, serviceInfo, useDelay, true)
		}(shoutrrr)

		// Space out sends.
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
	serviceInfo serviceinfo.ServiceInfo,
	useDelay bool,
	useMetrics bool,
) error {
	logFrom := logx.LogFrom{Primary: s.ID, Secondary: serviceInfo.ID} // For logging.

	if useDelay && s.GetDelay() != "0s" {
		// Delay sending the Shoutrrr message by the defined interval.
		logx.Info(
			fmt.Sprintf("Sleeping for %s before sending the Shoutrrr message", s.GetDelay()),
			logFrom,
			true,
		)
		time.Sleep(s.GetDelayDuration())
	}

	sender, message, params, url, err := s.getSender(title, msg, serviceInfo)
	if err != nil {
		return err
	}

	// Try sending the message.
	if logx.IsLevel("VERBOSE") || logx.IsLevel("DEBUG") {
		msg := fmt.Sprintf(
			"Sending %q to %q",
			message, url,
		)
		logx.Verbose(msg, logFrom, !logx.IsLevel("DEBUG"))
		logx.Debug(
			fmt.Sprintf(
				"%s with params=%q",
				msg, *params,
			),
			logFrom,
			true,
		)
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
		logFrom,
	)
}

// getSender returns the router, rendered message, params, and URL for the Shoutrrr.
func (s *Shoutrrr) getSender(
	title, msg string,
	serviceInfo serviceinfo.ServiceInfo,
) (*router.ServiceRouter, string, *types.Params, string, error) {
	// Build the URL.
	url := s.BuildURL()

	// Check the URL provides a valid sender.
	sender, err := shoutrrr.CreateSenderWithOptions(types.SenderOptions{}, url)
	if err != nil {
		err = fmt.Errorf("failed to create Shoutrrr sender: %w", err)
		return nil, "", nil, "", err
	}

	// Build the params.
	params := s.BuildParams(serviceInfo)
	for _, key := range unknownParams(sender, url, *params) {
		delete(*params, key)
	}
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

// payloadParams are the params a Shoutrrr service reads into its message rather than its config,
// keyed by URL scheme.
var payloadParams = map[string][]string{
	"gotify":     {"extras"},
	"mattermost": {"channel", "username"},
	"notifiarr":  {"content", "fields", "footer", "icon", "id", "update"},
	"rocketchat": {"channel", "username"},
}

// unknownParams returns the sorted keys of params that the service at url rejects at send.
func unknownParams(sender *router.ServiceRouter, url string, params types.Params) (unknown []string) {
	service, err := sender.Locate(url)
	if err != nil {
		return nil
	}
	resolver := format.NewPropKeyResolver(format.GetServiceConfig(service))

	scheme, _, _ := strings.Cut(url, "://")
	for key := range params {
		lower := strings.ToLower(key)
		if lower == types.TitleKey ||
			lower == types.MessageKey ||
			lower == types.LevelKey ||
			slices.Contains(payloadParams[scheme], lower) {
			continue
		}

		if _, err := resolver.Get(lower); errors.Is(err, format.ErrInvalidConfigKey) {
			unknown = append(unknown, key)
		}
	}
	slices.Sort(unknown)
	return unknown
}

// send delivers the message via the given sender and params, retrying up to max_tries times.
func (s *Shoutrrr) send(
	sender *router.ServiceRouter,
	message string,
	params *types.Params,
	serviceName string,
	logFrom logx.LogFrom,
) error {
	combinedErrs := make(map[string]int)

	if err := util.RetryWithBackoff(
		func() error {
			err := sender.Send(message, params)
			if failed := s.parseSend(err, combinedErrs, serviceName, logFrom); failed {
				return errors.New("send failed")
			}
			return nil
		},
		s.GetMaxTries(), //#nosec G115 -- Validated in CheckValues.
		1*time.Second,
		30*time.Second,
		s.ServiceStatus.Deleting,
	); err == nil {
		return nil
	}

	msg := fmt.Sprintf(
		"failed %d times to send a %s message for %q to %q",
		s.GetMaxTries(), s.GetType(), s.ServiceStatus.ServiceInfo.ID, s.BuildURL(),
	)
	logx.Error(msg, logFrom, true)
	failed := true
	s.Failed.Set(s.ID, &failed)
	errs := make([]error, 0, len(combinedErrs))
	for key := range combinedErrs {
		errs = append(
			errs,
			fmt.Errorf("%s x %d", key, combinedErrs[key]),
		)
	}
	return errors.Join(errs...)
}

// parseSend records send errors, updates Prometheus metrics, and reports whether the attempt failed.
func (s *Shoutrrr) parseSend(
	err []error,
	combinedErrs map[string]int,
	serviceName string,
	logFrom logx.LogFrom,
) (failed bool) {
	if s.ServiceStatus.Deleting() {
		return
	}

	for i := range err {
		if err[i] != nil {
			failed = true
			logx.Error(err[i], logFrom, true)
			combinedErrs[err[i].Error()]++
		}
	}
	// No serviceName, no metrics.
	if serviceName == "" {
		return
	}

	// SUCCESS!
	if !failed {
		metric.IncPrometheusCounter(
			metric.NotifyResultTotal,
			s.ID,
			serviceName,
			s.GetType(),
			metric.ActionResultSuccess,
		)
		s.Failed.Set(s.ID, &failed)
		return
	}

	// FAIL!
	metric.IncPrometheusCounter(
		metric.NotifyResultTotal,
		s.ID,
		serviceName,
		s.GetType(),
		metric.ActionResultFail,
	)
	return
}

// jsonMapToString returns the JSON param map as an '&' joined list of strings with the prefix added to each key.
//
//	e.g.
//		{"key1": "val1", "key2": "val2"} with prefix '@'
//	returns:
//		@key1=val1&@key2=val2
func jsonMapToString(param string, prefix string) string {
	if param == "" {
		return ""
	}

	// Convert the JSON string to a map.
	var jsonMap map[string]string
	err := decode.Unmarshal("json", []byte(param), &jsonMap)
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

		fmt.Fprintf(&builder,
			"%s%s=%s",
			prefix, key, jsonMap[key],
		)
	}
	return builder.String()
}

// queryParam returns "key=url-encoded-val" when val is non-empty, otherwise "".
func queryParam(key, val string) string {
	if val == "" {
		return ""
	}
	return key + "=" + net_url.QueryEscape(val)
}

// userInfo returns the escaped "[username][:password]@", or "" when both are empty.
func userInfo(username, password string) string {
	switch {
	case password != "":
		return net_url.UserPassword(username, password).String() + "@"
	case username != "":
		return net_url.User(username).String() + "@"
	default:
		return ""
	}
}

// buildQuery joins the non-empty "key=value" parts into a URL query string,
// prefixed with '?'.
//
//	e.g. buildQuery("", "foo=1", "bar=2") returns "?foo=1&bar=2"
//
// Returns "" if all parts are empty.
func buildQuery(parts ...string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			nonEmpty = append(nonEmpty, part)
		}
	}
	if len(nonEmpty) == 0 {
		return ""
	}

	return "?" + strings.Join(nonEmpty, "&")
}
