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
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/web/metrics"
)

// GetParams returns the params using everything from master>main>defaults>hardDefaults when
// the key is not defined in the lower level
func (s *Shoutrrr) GetParams(context *utils.ServiceInfo) (params *shoutrrr_types.Params) {
	p := make(shoutrrr_types.Params)
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
		(*params)[key] = utils.TemplateString((*params)[key], *context)
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
		login = s.GetURLField("username") + utils.ValueIfNotDefault(login, ":"+login)
		url = fmt.Sprintf("smtp://%s%s:%s/?fromaddress=%s&toaddresses=%s",
			utils.ValueIfNotDefault(login, login+"@"),
			s.GetURLField("host"),
			s.GetURLField("port"),
			s.GetParam("fromaddress"),
			s.GetParam("toaddresses"))
	case "gotify":
		// gotify://host:port/path/token
		port := s.GetURLField("port")
		path := s.GetURLField("path")
		url = fmt.Sprintf("gotify://%s%s%s/%s",
			s.GetURLField("host"),
			utils.ValueIfNotDefault(port, ":"+port),
			utils.ValueIfNotDefault(path, "/"+path),
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
			utils.ValueIfNotDefault(username, username+"@"),
			s.GetURLField("host"),
			utils.ValueIfNotDefault(port, ":"+port),
			utils.ValueIfNotDefault(path, "/"+path),
			s.GetURLField("token"),
			utils.ValueIfNotDefault(channel, "/"+channel))
	case "matrix":
		// matrix://user:password@host[:port][/port]/[?rooms=!roomID1[,roomAlias2]][&disableTLS=yes]
		port := s.GetURLField("port")
		path := s.GetURLField("path")
		rooms := s.GetParam("rooms")
		rooms = utils.ValueIfNotDefault(rooms, "?rooms="+rooms)
		disableTLS := s.GetParam("disabletls")
		disableTLS = utils.ValueIfNotDefault(disableTLS, "disableTLS="+disableTLS)
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
			utils.ValueIfNotDefault(port, ":"+port),
			utils.ValueIfNotDefault(path, "/"+path),
			rooms,
			disableTLS,
		)
	case "opsgenie":
		// opsgenie://host[:port][/path]/apikey
		port := s.GetURLField("port")
		path := s.GetURLField("path")
		url = fmt.Sprintf("opsgenie://%s%s%s/%s",
			s.GetURLField("host"),
			utils.ValueIfNotDefault(port, ":"+port),
			utils.ValueIfNotDefault(path, ""+path),
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
			utils.ValueIfNotDefault(devices, "?devices="+devices))
	case "rocketchat":
		// rocketchat://[username@]host:port[/port]/tokenA/tokenB/channel
		username := s.GetURLField("username")
		port := s.GetURLField("port")
		path := s.GetURLField("path")
		url = fmt.Sprintf("rocketchat://%s%s%s%s/%s/%s/%s",
			utils.ValueIfNotDefault(username, username+"@"),
			s.GetURLField("host"),
			utils.ValueIfNotDefault(port, ":"+port),
			utils.ValueIfNotDefault(path, "/"+path),
			s.GetURLField("tokena"),
			s.GetURLField("tokena"),
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
			utils.ValueIfNotDefault(group, group+"@"),
			s.GetURLField("tenant"),
			utils.ValueIfNotDefault(altid, "/"+altid),
			utils.ValueIfNotDefault(groupowner, "/"+groupowner),
			s.GetParam("host"))
		url = strings.Replace(url, "///", "//", 1)
	case "telegram":
		// telegram://token@telegram?chats=@chat1,@chat2
		url = fmt.Sprintf("telegram://%s@telegram?chats=%s",
			s.GetURLField("token"),
			s.GetParam("chats"))
	case "zulipchat":
		// zulip://botMail:botKey@host:port
		port := s.GetURLField("port")
		path := s.GetURLField("path")
		url = fmt.Sprintf("zulip://%s:%s@%s%s%s",
			s.GetURLField("botmail"),
			s.GetURLField("botkey"),
			s.GetURLField("host"),
			utils.ValueIfNotDefault(port, ":"+port),
			utils.ValueIfNotDefault(path, "/"+path))
	case "shoutrrr":
		// Raw
		url = s.GetURLField("raw")
	}
	return
}

func (s *Slice) Send(
	title string,
	message string,
	serviceInfo *utils.ServiceInfo,
) error {
	if s == nil {
		return nil
	}
	if serviceInfo == nil {
		serviceInfo = &utils.ServiceInfo{}
	}

	errs := make(chan error)
	for key := range *s {
		// Send each message up to s.MaxTries number of times until they don't err.
		go func(shoutrrr *Shoutrrr) {
			logFrom := utils.LogFrom{Primary: *shoutrrr.ID, Secondary: serviceInfo.ID} // For logging
			triesLeft := shoutrrr.GetMaxTries()                                        // Number of times to send Shoutrrr (until 200 received).

			// Delay sending the Shoutrrr message by the defined interval.
			msg := fmt.Sprintf("%s, Sleeping for %s before sending the Shoutrrr message", *shoutrrr.ID, shoutrrr.GetDelay())
			jLog.Info(msg, logFrom, shoutrrr.GetDelay() != "0s")
			time.Sleep(shoutrrr.GetDelayDuration())

			url := shoutrrr.GetURL()
			sender, err := shoutrrr_lib.CreateSender(url)
			if err != nil {
				errs <- err
				return
			}
			params := shoutrrr.GetParams(serviceInfo)
			if title != "" {
				(*params)["title"] = title
			}
			toSend := message
			if message == "" {
				toSend = shoutrrr.GetMessage(serviceInfo)
			}
			combinedErrs := make(map[string]int)
			for {
				msg := fmt.Sprintf("Sending %q to %q", toSend, url)
				jLog.Verbose(msg, logFrom, !jLog.IsLevel("debug"))
				jLog.Debug(msg+fmt.Sprintf(" with params=%q", *params), logFrom, jLog.IsLevel("debug"))
				err := sender.Send(toSend, params)

				failed := false
				for i := range err {
					if err[i] != nil {
						failed = true
						break
					}
				}

				// SUCCESS!
				if !failed {
					metrics.InitPrometheusCounterActions(metrics.NotifyMetric, *shoutrrr.ID, serviceInfo.ID, shoutrrr.GetType(), "SUCCESS")
					failed := false
					shoutrrr.Failed = &failed
					errs <- nil
					return
				}

				// FAIL
				for new := range err {
					jLog.Error(err[new].Error(), logFrom, true)

					combinedErrs[err[new].Error()]++
					// errs = fmt.Errorf("%s%s  host: <required> e.g. 'mattermost.example.io'\\", utils.ErrorToString(errs), prefix)
				}
				metrics.InitPrometheusCounterActions(metrics.NotifyMetric, *shoutrrr.ID, serviceInfo.ID, shoutrrr.GetType(), "FAIL")
				triesLeft--

				// Give up after MaxTries.
				if triesLeft == 0 {
					msg = fmt.Sprintf("failed %d times to send a %s message to %s", shoutrrr.GetMaxTries(), shoutrrr.GetType(), shoutrrr.GetURL())
					jLog.Error(msg, logFrom, true)
					failed := true
					shoutrrr.Failed = &failed
					var err error
					for key := range combinedErrs {
						err = fmt.Errorf("%s%s x %d", utils.ErrorToString(err), key, combinedErrs[key])
					}
					errs <- err
					return
				}

				// Space out retries.
				time.Sleep(10 * time.Second)
			}
		}((*s)[key])
		// Space out Shoutrrr messages.const.
		time.Sleep(3 * time.Second)
	}

	var err error
	for range *s {
		errFound := <-errs
		if errFound != nil {
			if err == nil {
				err = errFound
			} else {
				err = fmt.Errorf("%s\\%s\\", err.Error(), errFound.Error())
			}
		}
	}
	return err
}
