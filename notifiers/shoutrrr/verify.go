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
	"strconv"
	"strings"
	"time"

	shoutrrr_lib "github.com/containrrr/shoutrrr"
	"github.com/release-argus/Argus/util"
)

// CheckValues of this Slice.
func (s *Slice) CheckValues(prefix string) (errs error) {
	if s == nil {
		return
	}

	keys := util.SortedKeys(*s)
	for _, key := range keys {
		if err := (*s)[key].CheckValues(prefix + "    "); err != nil {
			errs = fmt.Errorf("%s%s  %s:\\%w",
				util.ErrorToString(errs), prefix, key, err)
		}
	}

	if errs != nil {
		errs = fmt.Errorf("%snotify:\\%w",
			prefix, errs)
	}
	return
}

// CheckValues of this Notification.
func (s *Shoutrrr) CheckValues(prefix string) (errs error) {
	if s == nil {
		return
	}

	var (
		errsOptions   error
		errsURLFields error
		errsParams    error
		errsLocate    error
	)
	s.InitMaps()

	// Delay
	if delay := s.GetSelfOption("delay"); delay != "" {
		// Default to seconds when an integer is provided
		if _, err := strconv.Atoi(delay); err == nil {
			s.Options["delay"] += "s"
		}
		if _, err := time.ParseDuration(s.Options["delay"]); err != nil {
			errsOptions = fmt.Errorf("%s%s  delay: %q <invalid> (Use 'AhBmCs' duration format)\\",
				util.ErrorToString(errsOptions), prefix, delay)
		}
	}

	s.correctSelf()

	if s.Main != nil {
		s.checkValuesMaster(prefix, &errs, &errsOptions, &errsURLFields, &errsParams)

		// Exclude matrix since it logs in, so may run into a rate-limit
		if s.GetType() != "matrix" {
			//#nosec G104 -- Disregard as we're not giving any rawURLs
			sender, _ := shoutrrr_lib.CreateSender()
			if _, err := sender.Locate(s.GetURL()); err != nil {
				errsLocate = fmt.Errorf("%s%s^ %w\\",
					util.ErrorToString(errsLocate), prefix, err)
			}
		}
	}

	if !util.CheckTemplate(s.GetSelfOption("message")) {
		errsOptions = fmt.Errorf("%s%s  message: %q <invalid> (didn't pass templating)\\",
			util.ErrorToString(errsOptions), prefix, s.GetSelfOption("message"))
	}
	for key := range s.Params {
		if !util.CheckTemplate(s.GetSelfParam(key)) {
			errsParams = fmt.Errorf("%s%s  %s: %q <invalid> (didn't pass templating)\\",
				util.ErrorToString(errsParams), prefix, key, s.GetSelfParam("title"))
		}
	}
	if errsOptions != nil {
		errs = fmt.Errorf("%s%soptions:\\%w",
			util.ErrorToString(errs), prefix, errsOptions)
	}
	if errsURLFields != nil {
		errs = fmt.Errorf("%s%surl_fields:\\%w",
			util.ErrorToString(errs), prefix, errsURLFields)
	}
	if errsParams != nil {
		errs = fmt.Errorf("%s%sparams:\\%w",
			util.ErrorToString(errs), prefix, errsParams)
	}
	if errsLocate != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), errsLocate)
	}
	return
}

// correctSelf will do a few corrections to user provided vars
// e.g. slack color wants $23 instead of #
func (s *Shoutrrr) correctSelf() {
	// Port, strip leading :
	if port := strings.TrimPrefix(s.GetSelfURLField("port"), ":"); port != "" {
		s.SetURLField("port", port)
	}

	// Path, strip leading /
	if path := strings.TrimPrefix(s.GetSelfURLField("path"), "/"); path != "" {
		s.SetURLField("path", path)
	}

	switch s.Type {
	case "matrix":
		// Remove #'s in channel aliases
		if rooms := strings.ReplaceAll(s.GetSelfParam("rooms"), "#", ""); rooms != "" {
			s.SetParam("rooms", rooms)
		}
	case "mattermost":
		// Channel, strip leading /
		if channel := strings.TrimPrefix(s.GetSelfURLField("channel"), "/"); channel != "" {
			s.SetURLField("channel", channel)
		}
	case "slack":
		// # -> %23
		// https://containrrr.dev/shoutrrr/v0.5/services/slack/
		// The format for the Color prop follows the slack docs but # needs to be escaped as %23 when passed in a URL.
		// So #ff8000 would be %23ff8000 etc.
		key := "color"
		if s.GetSelfParam(key) != "" {
			s.SetParam(key, strings.Replace(s.GetSelfParam(key), "#", "%23", 1))
		}
	case "teams":
		// AltID, strip leading /
		if altid := strings.TrimPrefix(s.GetSelfURLField("altid"), "/"); altid != "" {
			s.SetURLField("altid", altid)
		}
		// GroupOwner, strip leading /
		if groupowner := strings.TrimPrefix(s.GetSelfURLField("groupowner"), "/"); groupowner != "" {
			s.SetURLField("groupowner", groupowner)
		}
	case "zulip_chat":
		// BotMail, replace the @ with a %40 - https://containrrr.dev/shoutrrr/v0.5/services/zulip/
		if botmail := s.GetSelfURLField("botmail"); botmail != "" {
			s.SetURLField("botmail", strings.ReplaceAll(botmail, "@", "%40"))
		}
	}
}

// checkValuesMaster will check that the leading Shoutrrr can access all vars required
// for its Type
func (s *Shoutrrr) checkValuesMaster(prefix string, errs *error, errsOptions *error, errsURLFields *error, errsParams *error) {
	if util.GetFirstNonDefault(s.Type, s.Main.Type) == "" {
		*errs = fmt.Errorf("%s%stype: <required> e.g. 'slack', see the docs for possible types - https://release-argus.io/docs/config/notify\\",
			util.ErrorToString(*errs), prefix)
	}

	switch s.GetType() {
	case "discord":
		// discord://token@webhookid
		if s.GetURLField("token") == "" {
			*errsURLFields = fmt.Errorf("%s%s  token: <required> e.g. 'https://discord.com/api/webhooks/[ 975870285909737583 <- webhookid ]/[ QEdyk-Qi5AiMXoZdxQFpWNcwEfmz5oOm_1Rni9DnjQAUap4zWcurM4IquamVrDIyNgBG <- TOKEN ]'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetURLField("webhookid") == "" {
			*errsURLFields = fmt.Errorf("%s%s  webhookid: <required> e.g. 'https://discord.com/api/webhooks/[ 975870285909737583 <- WEBHOOKID ]/[ QEdyk-Qi5AiMXoZdxQFpWNcwEfmz5oOm_1Rni9DnjQAUap4zWcurM4IquamVrDIyNgBG <- token ]'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
	case "smtp":
		// smtp://username:password@host:port[/path]
		if s.GetURLField("host") == "" {
			*errsURLFields = fmt.Errorf("%s%s  host: <required> e.g. 'smtp.example.io'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetParam("fromaddress") == "" {
			*errsParams = fmt.Errorf("%s%s  fromaddress: <required> e.g. 'service@gmail.com'\\",
				util.ErrorToString(*errsParams), prefix)
		}
		if s.GetParam("toaddresses") == "" {
			*errsParams = fmt.Errorf("%s%s  toaddresses: <required> e.g. 'name@gmail.com'\\",
				util.ErrorToString(*errsParams), prefix)
		}
	case "gotify":
		// gotify://host:port/path/token
		if s.GetURLField("host") == "" {
			*errsURLFields = fmt.Errorf("%s%s  host: <required> e.g. 'gotify.example.io'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetURLField("token") == "" {
			*errsURLFields = fmt.Errorf("%s%s  token: <required> e.g. 'Aod9Cb0zXCeOrnD'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
	case "googlechat":
		// googlechat://url
		if s.GetURLField("raw") == "" {
			*errsURLFields = fmt.Errorf("%s%s  raw: <required> e.g. 'https://chat.googleapis.com/v1/spaces/FOO/messages?key=bar&token=baz'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
	case "ifttt":
		// ifttt://webhookid
		if s.GetURLField("webhookid") == "" {
			*errsURLFields = fmt.Errorf("%s%s  webhookid: <required> e.g. 'h1fyLh42h7lDI2L11T-bv'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetParam("events") == "" {
			*errsParams = fmt.Errorf("%s%s  events: <required> e.g. 'event1,event2'\\",
				util.ErrorToString(*errsParams), prefix)
		}
	case "join":
		// join://apiKey@join
		if s.GetURLField("apikey") == "" {
			*errsURLFields = fmt.Errorf("%s%s  apikey: <required> e.g. 'f8eae56127864015b0d2f4d8db6ff53f'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetParam("devices") == "" {
			*errsParams = fmt.Errorf("%s%s  devices: <required> e.g. '550ddc132c2b4fd28b8b89f735860db1,7294feb73974e5c99d7479ab7b73ba39,d2d775a2f453237d733aa2b7ea2c3ecd'\\",
				util.ErrorToString(*errsParams), prefix)
		}
	case "mattermost":
		// mattermost://[username@]host[:port][/path]/token[/channel]
		if s.GetURLField("host") == "" {
			*errsURLFields = fmt.Errorf("%s%s  host: <required> e.g. 'mattermost.example.io'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetURLField("token") == "" {
			*errsURLFields = fmt.Errorf("%s%s  token: <required> e.g. 'Aod9Cb0zXCeOrnD'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
	case "matrix":
		// matrix://user:password@host
		if s.GetURLField("host") == "" {
			*errsURLFields = fmt.Errorf("%s%s  host: <required> e.g. 'matrix.example.io'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetURLField("password") == "" {
			*errsURLFields = fmt.Errorf("%s%s  password: <required> e.g. 'pass123' (with user) OR 'access_token' (no user)\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
	case "opsgenie":
		// opsgenie://host[:port][/path]/apiKey
		if s.GetURLField("apikey") == "" {
			*errsURLFields = fmt.Errorf("%s%s  apikey: <required> e.g. 'xxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
	case "pushbullet":
		// pushbullet://token/targets
		if s.GetURLField("token") == "" {
			*errsURLFields = fmt.Errorf("%s%s  token: <required> e.g. 'o.5NfxzU9yH4xBZlEXZArRtyUB4S4Ua8Hd'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetURLField("targets") == "" {
			*errsURLFields = fmt.Errorf("%s%s  targets: <required> e.g. 'fpwfXzDCYsTxw4VfAAoHiR,5eAzVLKp5VRUMJeYehwbzv,XR7VKoK5b2MYWDpstD3Hfq'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
	case "pushover":
		// pushover://token@user
		if s.GetURLField("token") == "" {
			*errsURLFields = fmt.Errorf("%s%s  token: <required> e.g. 'aayohdg8gqjj3ssszuqwwmuipt5gcd'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetURLField("user") == "" {
			*errsURLFields = fmt.Errorf("%s%s  user: <required> e.g. '2QypyiVSnURsw72cpnXCuVAQMJpKKY'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
	case "rocketchat":
		// rocketchat://[username@]host:port[/port]/tokenA/tokenB/channel
		if s.GetURLField("host") == "" {
			*errsURLFields = fmt.Errorf("%s%s  host: <required> e.g. 'rocket-chat.example.io'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetURLField("tokena") == "" {
			*errsURLFields = fmt.Errorf("%s%s  tokena: <required> e.g. '8eGdRzc9r4YYNyvge'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetURLField("tokenb") == "" {
			*errsURLFields = fmt.Errorf("%s%s  tokenb: <required> e.g. '2XYQcX9NBwJBKfQnphpebPcnXZcPEi32Nt4NKJfrnbhsbRfX'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetURLField("channel") == "" {
			*errsURLFields = fmt.Errorf("%s%s  channel: <required> e.g. 'argusChannel' or '@user'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
	case "slack":
		// slack://token:token@channel
		if s.GetURLField("token") == "" {
			*errsURLFields = fmt.Errorf("%s%s  token: <required> e.g. '123456789012-1234567890123-4mt0t4l1YL3g1T5L4cK70k3N'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetURLField("channel") == "" {
			*errsURLFields = fmt.Errorf("%s%s  channel: <required> e.g. 'C001CH4NN3L' or 'webhook'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
	case "teams":
		// teams://[group@][tenant][/altid][/groupowner]
		if s.GetURLField("group") == "" {
			*errsURLFields = fmt.Errorf("%s%s  group: <required> e.g. '<host>/webhookb2/<GROUP>@<tenant>/IncomingWebhook/<altId>/<groupOwner>'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetURLField("tenant") == "" {
			*errsURLFields = fmt.Errorf("%s%s  tenant: <required> e.g. '<host>/webhookb2/<group>@<TENANT>/IncomingWebhook/<altId>/<groupOwner>'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetURLField("altid") == "" {
			*errsURLFields = fmt.Errorf("%s%s  altid: <required> e.g. '<host>/webhookb2/<group>@<tenant>/IncomingWebhook/<ALT-ID>/<groupOwner>'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetURLField("groupowner") == "" {
			*errsURLFields = fmt.Errorf("%s%s  groupowner: <required> e.g. '<host>/webhookb2/<group>@<tenant>/IncomingWebhook/<altId>/<GROUP-OWNER>'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetParam("host") == "" {
			*errsParams = fmt.Errorf("%s%s  host: <required> e.g. 'example.webhook.office.com'\\",
				util.ErrorToString(*errsParams), prefix)
		}
	case "telegram":
		// telegram://token@telegram
		if s.GetURLField("token") == "" {
			*errsURLFields = fmt.Errorf("%s%s  token: <required> e.g. '110201543:AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsaw'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetParam("chats") == "" {
			*errsParams = fmt.Errorf("%s%s  chats: <required> e.g. '@channelName' or 'chatID'\\",
				util.ErrorToString(*errsParams), prefix)
		}
	case "zulip_chat":
		// zulip://botMail:botKey@host:port
		if s.GetURLField("host") == "" {
			*errsURLFields = fmt.Errorf("%s%s  host: <required> e.g. 'example.zulipchat.com'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetURLField("botmail") == "" {
			*errsURLFields = fmt.Errorf("%s%s  botmail: <required> e.g. 'my-bot@zulipchat.com'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetURLField("botkey") == "" {
			*errsURLFields = fmt.Errorf("%s%s  botkey: <required> e.g. 'correcthorsebatterystable'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
	case "shoutrrr":
		// Raw
		if s.GetURLField("raw") == "" {
			*errsURLFields = fmt.Errorf("%s%s  raw: <required> e.g. 'service://foo:bar@something'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
	default:
		// Invalid/Unknown type
		if s.Type != "" {
			*errs = fmt.Errorf("%s%stype: %q <invalid> e.g. 'slack', see the docs for possible types - https://release-argus.io/docs/config/notify\\",
				util.ErrorToString(*errs), prefix, s.GetType())
		}
	}
}

// Print the Slice.
func (s *Slice) Print(prefix string) bool {
	if s == nil || len(*s) == 0 {
		return false
	}

	fmt.Printf("%snotify:\n", prefix)
	for key := range *s {
		fmt.Printf("%s  %s:\n", prefix, key)
		(*s)[key].Print(prefix + "    ")
	}
	return true
}

// Print the Shourrr Struct.
func (s *Shoutrrr) Print(prefix string) {
	util.PrintlnIfNotDefault(s.Type, fmt.Sprintf("%stype: %s", prefix, s.Type))

	if len(s.Options) != 0 {
		fmt.Printf("%soptions:\n", prefix)
		keys := util.SortedKeys(s.Options)
		for _, key := range keys {
			util.PrintlnIfNotDefault(s.GetSelfOption(key), fmt.Sprintf("%s  %s: %s", prefix, key, s.GetSelfOption(key)))
		}
	}

	if len(s.URLFields) != 0 {
		fmt.Printf("%surl_fields:\n", prefix)
		keys := util.SortedKeys(s.URLFields)
		for _, key := range keys {
			util.PrintlnIfNotDefault(s.GetSelfURLField(key), fmt.Sprintf("%s  %s: %s", prefix, key, s.GetSelfURLField(key)))
		}
	}

	if len(s.Params) != 0 {
		fmt.Printf("%sparams:\n", prefix)
		keys := util.SortedKeys(s.Params)
		for _, key := range keys {
			util.PrintlnIfNotDefault(s.GetSelfParam(key), fmt.Sprintf("%s  %s: %s", prefix, key, s.GetSelfParam(key)))
		}
	}
}
