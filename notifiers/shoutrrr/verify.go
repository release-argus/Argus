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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	shoutrrr_lib "github.com/containrrr/shoutrrr"
	"github.com/release-argus/Argus/util"
)

// CheckValues of this SliceDefaults.
func (s *SliceDefaults) CheckValues(prefix string) (errs error) {
	if s == nil {
		return
	}

	keys := util.SortedKeys(*s)
	itemPrefix := prefix + "    "
	for _, key := range keys {
		if err := (*s)[key].CheckValues(itemPrefix); err != nil {
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

// CheckValues of this Slice.
func (s *Slice) CheckValues(prefix string) (errs error) {
	if s == nil {
		return
	}

	keys := util.SortedKeys(*s)
	itemPrefix := prefix + "    "
	for _, key := range keys {
		if err := (*s)[key].CheckValues(itemPrefix); err != nil {
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
func (s *ShoutrrrBase) CheckValues(prefix string) (errs error) {
	var (
		errsOptions error
		errsParams  error
	)
	s.InitMaps()

	// Delay
	if delay := s.GetOption("delay"); delay != "" {
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

	if !util.CheckTemplate(s.GetOption("message")) {
		errsOptions = fmt.Errorf("%s%s  message: %q <invalid> (didn't pass templating)\\",
			util.ErrorToString(errsOptions), prefix, s.GetOption("message"))
	}
	for key, value := range s.Params {
		if !util.CheckTemplate(value) {
			errsParams = fmt.Errorf("%s%s  %s: %q <invalid> (didn't pass templating)\\",
				util.ErrorToString(errsParams), prefix, key, value)
		}
	}
	if errsOptions != nil {
		errs = fmt.Errorf("%s%soptions:\\%w",
			util.ErrorToString(errs), prefix, errsOptions)
	}
	if errsParams != nil {
		errs = fmt.Errorf("%s%sparams:\\%w",
			util.ErrorToString(errs), prefix, errsParams)
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

	baseErrs := s.ShoutrrrBase.CheckValues(prefix)
	// split option and param errs
	if baseErrs != nil {
		strErr := baseErrs.Error()
		paramsStr := fmt.Sprintf("%sparams:\\", prefix)
		optionsStr := fmt.Sprintf("%soptions:\\", prefix)
		// Has errParams
		if strings.Contains(strErr, paramsStr) {
			splitStrErr := strings.Split(strErr, paramsStr)
			errsParams = errors.New(splitStrErr[1])
			// Has errOptions too
			if strings.Contains(splitStrErr[0], optionsStr) {
				errsOptions = errors.New(
					strings.TrimPrefix(splitStrErr[0], optionsStr))
			}
			// only errOptions
		} else {
			errsOptions = errors.New(
				strings.TrimPrefix(strErr, optionsStr))
		}
	}

	s.checkValuesForType(prefix, &errs, &errsURLFields, &errsParams)

	// Exclude matrix since it logs in, so may run into a rate-limit
	if errsParams == nil && errsURLFields == nil && s.GetType() != "matrix" {
		//#nosec G104 -- Disregard as we're not giving any rawURLs
		sender, _ := shoutrrr_lib.CreateSender()
		if _, err := sender.Locate(s.BuildURL()); err != nil {
			errsLocate = fmt.Errorf("%s%s^ %w\\",
				util.ErrorToString(errsLocate), prefix, err)
		}
	}

	// options
	if errsOptions != nil {
		errs = fmt.Errorf("%s%soptions:\\%w",
			util.ErrorToString(errs), prefix, errsOptions)
	}
	// params
	if errsParams != nil {
		errs = fmt.Errorf("%s%sparams:\\%w",
			util.ErrorToString(errs), prefix, errsParams)
	}
	// url_fields
	if errsURLFields != nil {
		errs = fmt.Errorf("%s%surl_fields:\\%w",
			util.ErrorToString(errs), prefix, errsURLFields)
	}
	// locate
	if errsLocate != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), errsLocate)
	}
	return
}

// correctSelf will do a few corrections to user provided vars
// e.g. slack color wants $23 instead of #
func (s *ShoutrrrBase) correctSelf() {
	// Port, strip leading :
	if port := strings.TrimPrefix(s.GetURLField("port"), ":"); port != "" {
		s.SetURLField("port", port)
	}

	// Path, strip leading /
	if path := strings.TrimPrefix(s.GetURLField("path"), "/"); path != "" {
		s.SetURLField("path", path)
	}

	// Host
	host := s.GetURLField("host")
	if strings.Contains(host, ":") {
		// Trim leading http(s)://
		host = strings.TrimPrefix(host, "http://")
		host = strings.TrimPrefix(host, "https://")
		s.SetURLField("host", host)

		// Move port from "host" to "port"
		split := strings.Split(host, ":")
		if len(split) > 1 {
			s.SetURLField("host", split[0])
			s.SetURLField("port", split[1])
		}
	}

	switch s.Type {
	case "matrix":
		// Remove #'s in channel aliases
		if rooms := strings.ReplaceAll(s.GetParam("rooms"), "#", ""); rooms != "" {
			s.SetParam("rooms", rooms)
		}
	case "mattermost":
		// Channel, strip leading /
		if channel := strings.TrimPrefix(s.GetURLField("channel"), "/"); channel != "" {
			s.SetURLField("channel", channel)
		}
	case "slack":
		// # -> %23
		// https://containrrr.dev/shoutrrr/v0.5/services/slack/
		// The format for the Color prop follows the slack docs but # needs to be escaped as %23 when passed in a URL.
		// So #ff8000 would be %23ff8000 etc.
		key := "color"
		if s.GetParam(key) != "" {
			s.SetParam(key, strings.Replace(s.GetParam(key), "#", "%23", 1))
		}
	case "teams":
		// AltID, strip leading /
		if altid := strings.TrimPrefix(s.GetURLField("altid"), "/"); altid != "" {
			s.SetURLField("altid", altid)
		}
		// GroupOwner, strip leading /
		if groupowner := strings.TrimPrefix(s.GetURLField("groupowner"), "/"); groupowner != "" {
			s.SetURLField("groupowner", groupowner)
		}
	case "zulip":
		// BotMail, replace the @ with a %40 - https://containrrr.dev/shoutrrr/v0.5/services/zulip/
		if botmail := s.GetURLField("botmail"); botmail != "" {
			s.SetURLField("botmail", strings.ReplaceAll(botmail, "@", "%40"))
		}
	}
}

// checkValuesForType will check that this Shoutrrr can access all vars required
// for its Type and that this Type is valid.
func (s *Shoutrrr) checkValuesForType(
	prefix string,
	errs *error,
	errsURLFields *error,
	errsParams *error,
) {
	// Check that the Type is valid
	sType := s.GetType()
	if !util.Contains(supportedTypes, sType) {
		sTypeWithoutID := util.FirstNonDefault(s.Type, s.Main.Type)
		if sTypeWithoutID == "" {
			*errs = fmt.Errorf("%s%stype: <required> e.g. 'slack', see the docs for possible types - https://release-argus.io/docs/config/notify\\",
				util.ErrorToString(*errs), prefix)
			return
		}
	}
	// Check that the Type doesn't differ in the Main
	if s.Main.Type != "" && sType != s.Main.Type {
		*errs = fmt.Errorf("%s%stype: %q != %q <invalid> (omit the type, or make it the same as the root notify.%s.type)\\",
			util.ErrorToString(*errs), prefix, sType, s.Main.Type, s.ID)
	}

	switch sType {
	case "bark":
		// bark://:devicekey@host:port/[path]
		if s.GetURLField("devicekey") == "" {
			*errsURLFields = fmt.Errorf("%s%s  devicekey: <required>\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		if s.GetURLField("host") == "" {
			*errsURLFields = fmt.Errorf("%s%s  host: <required>\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
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
	case "ntfy":
		// ntfy://[username]:[password]@[host][:port][/path]/topic
		if s.GetURLField("topic") == "" {
			*errsURLFields = fmt.Errorf("%s%s  topic: <required>\\",
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
	case "zulip":
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
	case "generic":
		// generic://host[:port][/path]
		if s.GetURLField("host") == "" {
			*errsURLFields = fmt.Errorf("%s%s  host: <required> e.g. 'example.com'\\",
				util.ErrorToString(*errsURLFields), prefix)
		}
		jsonMaps := []string{"custom_headers", "json_payload_vars", "query_vars"}
		for _, jsonMap := range jsonMaps {
			value := s.GetURLField(jsonMap)
			if value != "" {
				converted := jsonMapToString(s.GetURLField(jsonMap), "-")
				if converted == "" {
					*errsURLFields = fmt.Errorf("%s%s  %s: %q <invalid> (must be a JSON map)\\",
						util.ErrorToString(*errsParams), prefix, jsonMap, value)
				}
			}
		}
	default:
		// Invalid/Unknown type
		if s.Type != "" {
			*errs = fmt.Errorf("%s%stype: %q <invalid> (supported types = [%s])\\",
				util.ErrorToString(*errs), prefix, sType, strings.Join(supportedTypes, ","))
		}
	}
}

// TestSend will test the Shoutrrr by sending a test message.
func (s *Shoutrrr) TestSend(serviceURL string) (err error) {
	if s == nil {
		err = fmt.Errorf("Shoutrrr is nil")
		return
	}

	s.SetOption("max_tries", "1")

	testServiceInfo := &util.ServiceInfo{
		ID:            *s.ServiceStatus.ServiceID,
		URL:           serviceURL,
		WebURL:        *s.ServiceStatus.WebURL,
		LatestVersion: s.ServiceStatus.LatestVersion()}
	if testServiceInfo.LatestVersion == "" {
		testServiceInfo.LatestVersion = "MAJOR.MINOR.PATCH"
	}

	title := "TEST - " + s.Title(testServiceInfo)
	message := "TEST - " + s.Message(testServiceInfo)
	err = s.Send(
		title,
		message,
		testServiceInfo,
		false,
		false)

	return
}

// Print the SliceDefaults.
func (s *SliceDefaults) Print(prefix string) {
	if s == nil || len(*s) == 0 {
		return
	}

	str := s.String(prefix + "  ")
	fmt.Printf("%snotify:\n%s", prefix, str)
}
