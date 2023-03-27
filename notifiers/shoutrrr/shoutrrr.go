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

package shoutrrr

import (
	"fmt"
	"strings"
	"time"

	shoutrrr_lib "github.com/containrrr/shoutrrr"
	shoutrrr_types "github.com/containrrr/shoutrrr/pkg/types"
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
)

// GetParams returns the params using everything from master>main>defaults>hardDefaults when
// the key is not defined in the lower level
func (s *Shoutrrr) GetParams(context *util.ServiceInfo) (params *shoutrrr_types.Params) {
	p := make(shoutrrr_types.Params, len(s.Params)+len(s.Main.Params))
	params = &p

	// Service Params
	for key := range s.Params {
		(*params)[key] = s.GetSelfParam(key)
	}

	// Main Params
	for key := range s.Main.Params {
		_, exist := s.Params[key]
		// Only overwrite if it doesn't exist in the level below
		if !exist {
			(*params)[key] = s.Main.GetSelfParam(key)
		}
	}

	// Default Params
	for key := range s.Defaults.Params {
		_, exist := (*params)[key]
		// Only overwrite if it doesn't exist in the level below
		if !exist {
			(*params)[key] = s.Defaults.GetSelfParam(key)
		}
	}

	// HardDefault Params
	for key := range s.HardDefaults.Params {
		_, exist := (*params)[key]
		// Only overwrite if it doesn't exist in the level below
		if !exist {
			(*params)[key] = s.HardDefaults.GetSelfParam(key)
		}
	}

	// Apply Jinja templating
	for key := range *params {
		(*params)[key] = util.TemplateString((*params)[key], *context)
	}

	return
}

func (s *Shoutrrr) GetURL() (url string) {
	switch s.GetType() {
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
	case "shoutrrr":
		// Raw
		url = s.GetURLField("raw")
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
			errChan <- shoutrrr.Send(title, message, serviceInfo, useDelay)
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

func (s *Shoutrrr) Send(
	title string,
	message string,
	serviceInfo *util.ServiceInfo,
	useDelay bool,
) (errs error) {
	logFrom := util.LogFrom{Primary: s.ID, Secondary: serviceInfo.ID} // For logging
	triesLeft := s.GetMaxTries()                                      // Number of times to send Shoutrrr (until 200 received).

	if useDelay && s.GetDelay() != "0s" {
		// Delay sending the Shoutrrr message by the defined interval.
		msg := fmt.Sprintf("%s, Sleeping for %s before sending the Shoutrrr message", s.ID, s.GetDelay())
		jLog.Info(msg, logFrom, s.GetDelay() != "0s")
		time.Sleep(s.GetDelayDuration())
	}

	url := s.GetURL()
	sender, err := shoutrrr_lib.CreateSender(url)
	if err != nil {
		return fmt.Errorf("failed to create Shoutrrr sender: %w", err)
	}
	params := s.GetParams(serviceInfo)
	if title != "" {
		(*params)["title"] = title
	}
	toSend := message
	if message == "" {
		toSend = s.GetMessage(serviceInfo)
	}

	combinedErrs := make(map[string]int)
	for {
		// Check if we're deleting the Service.
		if s.ServiceStatus.Deleting {
			return
		}

		// Try sending the message.
		msg := fmt.Sprintf("Sending %q to %q", toSend, url)
		jLog.Verbose(msg, logFrom, !jLog.IsLevel("debug"))
		jLog.Debug(fmt.Sprintf("%s with params=%q", msg, *params), logFrom, jLog.IsLevel("debug"))
		err := sender.Send(toSend, params)

		failed := false
		for i := range err {
			if err[i] != nil {
				failed = true
				jLog.Error(err[i].Error(), logFrom, true)
				combinedErrs[err[i].Error()]++
			}
		}

		// SUCCESS!
		if !failed {
			metric.IncreasePrometheusCounter(metric.NotifyMetric,
				s.ID,
				serviceInfo.ID,
				s.GetType(),
				"SUCCESS")
			failed := false
			(*s.Failed)[s.ID] = &failed
			return
		}

		// FAIL
		metric.IncreasePrometheusCounter(metric.NotifyMetric,
			s.ID,
			serviceInfo.ID,
			s.GetType(),
			"FAIL")
		triesLeft--

		// Give up after MaxTries.
		if triesLeft == 0 {
			msg = fmt.Sprintf("failed %d times to send a %s message for %q to %q", s.GetMaxTries(), s.GetType(), *s.ServiceStatus.ServiceID, s.GetURL())
			jLog.Error(msg, logFrom, true)
			failed := true
			(*s.Failed)[s.ID] = &failed
			for key := range combinedErrs {
				errs = fmt.Errorf("%s%s x %d",
					util.ErrorToString(errs), key, combinedErrs[key])
			}
			return
		}

		// Space out retries.
		time.Sleep(5 * time.Second)
	}
}
