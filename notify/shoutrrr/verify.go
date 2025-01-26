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

// Package shoutrrr provides the shoutrrr notification service to services.
package shoutrrr

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	goshoutrrr "github.com/containrrr/shoutrrr"

	"github.com/release-argus/Argus/util"
)

// CheckValues validates the fields of the SliceDefaults struct.
func (s *SliceDefaults) CheckValues(prefix string) error {
	if s == nil {
		return nil
	}

	var errs []error
	keys := util.SortedKeys(*s)
	itemPrefix := prefix + "  "
	for _, key := range keys {
		util.AppendCheckError(&errs, prefix, key, (*s)[key].CheckValues(itemPrefix, key))
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// CheckValues validates the fields of the Base struct.
func (b *Base) CheckValues(prefix string, id string) error {
	if b == nil {
		return nil
	}
	b.InitMaps()
	b.correctSelf(util.FirstNonDefault(b.Type, id))

	var errs []error
	itemPrefix := prefix + "  "
	util.AppendCheckError(&errs, prefix, "options", b.checkValuesOptions(itemPrefix))
	util.AppendCheckError(&errs, prefix, "params", b.checkValuesParams(itemPrefix))

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// CheckValues validates the fields of each member of the Slice.
func (s *Slice) CheckValues(prefix string) error {
	if s == nil {
		return nil
	}

	var errs []error
	keys := util.SortedKeys(*s)
	itemPrefix := prefix + "  "
	for _, key := range keys {
		util.AppendCheckError(&errs, prefix, key, (*s)[key].CheckValues(itemPrefix))
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// CheckValues validates the fields of the Shoutrrr struct.
func (s *Shoutrrr) CheckValues(prefix string) error {
	if s == nil {
		return nil
	}
	s.InitMaps()
	s.correctSelf(s.GetType())

	var errs []error
	// Type.
	if errsType := s.checkValuesType(prefix); errsType != nil {
		errs = append(errs, errsType)
	}
	itemPrefix := prefix + "  "
	util.AppendCheckError(&errs, prefix, "options", s.checkValuesOptions(itemPrefix))
	util.AppendCheckError(&errs, prefix, "url_fields", s.checkValuesURLFields(itemPrefix))
	util.AppendCheckError(&errs, prefix, "params", s.checkValuesParams(itemPrefix))

	// Exclude matrix since it logs in, so may run into a rate-limit.
	if len(errs) == 0 && s.GetType() != "matrix" {
		//#nosec G104 -- Disregard as we are not giving any rawURLs.
		sender, _ := goshoutrrr.CreateSender()
		if _, err := sender.Locate(s.BuildURL()); err != nil {
			errs = append(errs, fmt.Errorf("%s^ %w",
				prefix, err))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// correctSelf will do a few corrections to user provided vars.
//
//	e.g. slack color wants $23 instead of #.
func (b *Base) correctSelf(shoutrrrType string) {
	// Port, strip leading :
	port := b.GetURLField("port")
	if strings.HasPrefix(port, ":") {
		b.SetURLField("port", strings.TrimPrefix(port, ":"))
	}

	// Path, strip leading /
	path := b.GetURLField("path")
	if strings.HasPrefix(path, "/") {
		b.SetURLField("path", strings.TrimPrefix(path, "/"))
	}

	// Host.
	host := b.GetURLField("host")
	// Check if host contains a scheme and/or port.
	if util.RegexCheck(`^https?:\/\/.*:?`, host) {
		// Trim leading http(s)://
		host = strings.TrimPrefix(host, "http://")
		host = strings.TrimPrefix(host, "https://")

		// Move port from "host" to "port".
		split := strings.Split(host, ":")
		if len(split) > 1 {
			host = split[0]
			b.SetURLField("port", split[1])
		}
		b.SetURLField("host", host)
	}

	switch shoutrrrType {
	case "matrix":
		// Remove #'s in channel aliases.
		if rooms := strings.ReplaceAll(b.GetParam("rooms"), "#", ""); rooms != "" {
			b.SetParam("rooms", rooms)
		}
	case "mattermost":
		// Channel, strip leading /
		if channel := strings.TrimPrefix(b.GetURLField("channel"), "/"); channel != "" {
			b.SetURLField("channel", channel)
		}
	case "slack":
		// # -> %23
		// https://containrrr.dev/shoutrrr/v0.5/services/slack/
		// The format for the Color prop follows the slack docs but # needs to be escaped as %23 when passed in a URL.
		// So #ff8000 would be %23ff8000 etc.
		key := "color"
		if b.GetParam(key) != "" {
			b.SetParam(key, strings.Replace(b.GetParam(key), "#", "%23", 1))
		}
	case "teams":
		// AltID, strip leading /
		if altid := strings.TrimPrefix(b.GetURLField("altid"), "/"); altid != "" {
			b.SetURLField("altid", altid)
		}
		// GroupOwner, strip leading /
		if groupowner := strings.TrimPrefix(b.GetURLField("groupowner"), "/"); groupowner != "" {
			b.SetURLField("groupowner", groupowner)
		}
	case "zulip":
		// BotMail, replace the @ with a %40 - https://containrrr.dev/shoutrrr/v0.5/services/zulip/
		if botmail := b.GetURLField("botmail"); botmail != "" {
			b.SetURLField("botmail", strings.ReplaceAll(botmail, "@", "%40"))
		}
	}
}

// checkValuesType validates that fields of this Shoutrrr struct are valid for `Type`.
func (s *Shoutrrr) checkValuesType(prefix string) error {
	// Check we have a Type.
	sType := s.GetType()
	if !util.Contains(supportedTypes, sType) {
		sTypeWithoutID := util.FirstNonDefault(s.Type, s.Main.Type)
		if sTypeWithoutID == "" {
			return fmt.Errorf("%stype: <required> e.g. 'slack', see the docs for possible types - https://release-argus.io/docs/config/notify",
				prefix)
		}
	}

	// Check the Type doesn't differ in the Main.
	if s.Main.Type != "" && sType != s.Main.Type {
		return fmt.Errorf("%stype: %q != %q <invalid> (must be the same as the root notify.%s.type)",
			prefix, sType, s.Main.Type, s.ID)
	}

	// Invalid/Unknown type.
	if !util.Contains(supportedTypes, sType) {
		return fmt.Errorf("%stype: %q <invalid> (supported types = [%s])",
			prefix, sType, strings.Join(supportedTypes, ","))
	}

	// Pass.
	return nil
}

// CheckValues validates the `Options` of the Shoutrrr struct.
func (b *Base) checkValuesOptions(prefix string) error {
	var errs []error
	// Options.Delay.
	if optionDelay := b.GetOption("delay"); optionDelay != "" {
		// Treat integers as seconds by default.
		if _, err := strconv.Atoi(optionDelay); err == nil {
			b.Options["delay"] += "s"
		}
		if _, err := time.ParseDuration(b.Options["delay"]); err != nil {
			errs = append(errs, fmt.Errorf("%sdelay: %q <invalid> (Use 'AhBmCs' duration format)",
				prefix, optionDelay))
		}
	}

	// Options.MaxTries.
	if maxTriesStr := b.GetOption("max_tries"); maxTriesStr != "" {
		if maxTries, err := strconv.ParseUint(maxTriesStr, 10, 64); err == nil {
			// Too large.
			if maxTries > math.MaxUint8 {
				errs = append(errs, fmt.Errorf("%smax_tries: %q <invalid> (too large, max = %d)",
					prefix, maxTriesStr, math.MaxUint8))
			}
		} else {
			// Too large?
			if util.RegexCheck(`^-?\d+$`, maxTriesStr) {
				// Negative.
				if strings.HasPrefix(maxTriesStr, "-") {
					errs = append(errs, fmt.Errorf("%smax_tries: %q <invalid> (must be positive)",
						prefix, maxTriesStr))
					// Positive.
				} else {
					errs = append(errs, fmt.Errorf("%smax_tries: %q <invalid> (too large, max = %d)",
						prefix, maxTriesStr, uint(math.MaxUint8)))
				}
			} else {
				// Not an integer.
				errs = append(errs, fmt.Errorf("%smax_tries: %q <invalid> (must be an integer)",
					prefix, maxTriesStr))
			}
		}
	}

	// Options.Message.
	optionMessage := b.GetOption("message")
	if !util.CheckTemplate(optionMessage) {
		errs = append(errs, fmt.Errorf("%smessage: %q <invalid> (didn't pass templating)",
			prefix, optionMessage))
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// checkValuesURLFields validates the `URLFields` of the Shoutrrr struct.
func (s *Shoutrrr) checkValuesURLFields(prefix string) error {
	var errs []error

	switch s.GetType() {
	case "bark":
		// bark://:devicekey@host:port/[path]
		if s.GetURLField("devicekey") == "" {
			errs = append(errs, fmt.Errorf("%sdevicekey: <required>",
				prefix))
		}
		if s.GetURLField("host") == "" {
			errs = append(errs, fmt.Errorf("%shost: <required>",
				prefix))
		}
	case "discord":
		// discord://token@webhookid
		if s.GetURLField("token") == "" {
			errs = append(errs, fmt.Errorf("%stoken: <required> e.g. 'https://discord.com/api/webhooks/[ 975870285909737583 <- webhookid ]/[ QEdyk-Qi5AiMXoZdxQFpWNcwEfmz5oOm_1Rni9DnjQAUap4zWcurM4IquamVrDIyNgBG <- TOKEN ]'",
				prefix))
		}
		if s.GetURLField("webhookid") == "" {
			errs = append(errs, fmt.Errorf("%swebhookid: <required> e.g. 'https://discord.com/api/webhooks/[ 975870285909737583 <- WEBHOOKID ]/[ QEdyk-Qi5AiMXoZdxQFpWNcwEfmz5oOm_1Rni9DnjQAUap4zWcurM4IquamVrDIyNgBG <- token ]'",
				prefix))
		}
	case "smtp":
		// smtp://username:password@host:port[/path]
		if s.GetURLField("host") == "" {
			errs = append(errs, fmt.Errorf("%shost: <required> e.g. 'smtp.example.io'",
				prefix))
		}
	case "gotify":
		// gotify://host:port/path/token
		if s.GetURLField("host") == "" {
			errs = append(errs, fmt.Errorf("%shost: <required> e.g. 'gotify.example.io'",
				prefix))
		}
		if s.GetURLField("token") == "" {
			errs = append(errs, fmt.Errorf("%stoken: <required> e.g. 'Aod9Cb0zXCeOrnD'",
				prefix))
		}
	case "googlechat":
		// googlechat://url
		if s.GetURLField("raw") == "" {
			errs = append(errs, fmt.Errorf("%sraw: <required> e.g. 'https://chat.googleapis.com/v1/spaces/FOO/messages?key=bar&token=baz'",
				prefix))
		}
	case "ifttt":
		// ifttt://webhookid
		if s.GetURLField("webhookid") == "" {
			errs = append(errs, fmt.Errorf("%swebhookid: <required> e.g. 'h1fyLh42h7lDI2L11T-bv'",
				prefix))
		}
	case "join":
		// join://apiKey@join
		if s.GetURLField("apikey") == "" {
			errs = append(errs, fmt.Errorf("%sapikey: <required> e.g. 'f8eae56127864015b0d2f4d8db6ff53f'",
				prefix))
		}
	case "mattermost":
		// mattermost://[username@]host[:port][/path]/token[/channel]
		if s.GetURLField("host") == "" {
			errs = append(errs, fmt.Errorf("%shost: <required> e.g. 'mattermost.example.io'",
				prefix))
		}
		if s.GetURLField("token") == "" {
			errs = append(errs, fmt.Errorf("%stoken: <required> e.g. 'Aod9Cb0zXCeOrnD'",
				prefix))
		}
	case "matrix":
		// matrix://user:password@host
		if s.GetURLField("host") == "" {
			errs = append(errs, fmt.Errorf("%shost: <required> e.g. 'matrix.example.io'",
				prefix))
		}
		if s.GetURLField("password") == "" {
			errs = append(errs, fmt.Errorf("%spassword: <required> e.g. 'pass123' (with user) OR 'access_token' (no user)",
				prefix))
		}
	case "ntfy":
		// ntfy://[username]:[password]@[host][:port][/path]/topic
		if s.GetURLField("topic") == "" {
			errs = append(errs, fmt.Errorf("%stopic: <required>",
				prefix))
		}
	case "opsgenie":
		// opsgenie://host[:port][/path]/apiKey
		if s.GetURLField("apikey") == "" {
			errs = append(errs, fmt.Errorf("%sapikey: <required> e.g. 'xxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx'",
				prefix))
		}
	case "pushbullet":
		// pushbullet://token/targets
		if s.GetURLField("token") == "" {
			errs = append(errs, fmt.Errorf("%stoken: <required> e.g. 'o.5NfxzU9yH4xBZlEXZArRtyUB4S4Ua8Hd'",
				prefix))
		}
		if s.GetURLField("targets") == "" {
			errs = append(errs, fmt.Errorf("%stargets: <required> e.g. 'fpwfXzDCYsTxw4VfAAoHiR,5eAzVLKp5VRUMJeYehwbzv,XR7VKoK5b2MYWDpstD3Hfq'",
				prefix))
		}
	case "pushover":
		// pushover://token@user
		if s.GetURLField("token") == "" {
			errs = append(errs, fmt.Errorf("%stoken: <required> e.g. 'aayohdg8gqjj3ssszuqwwmuipt5gcd'",
				prefix))
		}
		if s.GetURLField("user") == "" {
			errs = append(errs, fmt.Errorf("%suser: <required> e.g. '2QypyiVSnURsw72cpnXCuVAQMJpKKY'",
				prefix))
		}
	case "rocketchat":
		// rocketchat://[username@]host:port[/port]/tokenA/tokenB/channel
		if s.GetURLField("host") == "" {
			errs = append(errs, fmt.Errorf("%shost: <required> e.g. 'rocket-chat.example.io'",
				prefix))
		}
		if s.GetURLField("tokena") == "" {
			errs = append(errs, fmt.Errorf("%stokena: <required> e.g. '8eGdRzc9r4YYNyvge'",
				prefix))
		}
		if s.GetURLField("tokenb") == "" {
			errs = append(errs, fmt.Errorf("%stokenb: <required> e.g. '2XYQcX9NBwJBKfQnphpebPcnXZcPEi32Nt4NKJfrnbhsbRfX'",
				prefix))
		}
		if s.GetURLField("channel") == "" {
			errs = append(errs, fmt.Errorf("%schannel: <required> e.g. 'argusChannel' or '@user'",
				prefix))
		}
	case "slack":
		// slack://token:token@channel
		if s.GetURLField("token") == "" {
			errs = append(errs, fmt.Errorf("%stoken: <required> e.g. '123456789012-1234567890123-4mt0t4l1YL3g1T5L4cK70k3N'",
				prefix))
		}
		if s.GetURLField("channel") == "" {
			errs = append(errs, fmt.Errorf("%schannel: <required> e.g. 'C001CH4NN3L' or 'webhook'",
				prefix))
		}
	case "teams":
		// teams://[group@][tenant][/altid][/groupowner]
		if s.GetURLField("group") == "" {
			errs = append(errs, fmt.Errorf("%sgroup: <required> e.g. '<host>/webhookb2/<GROUP>@<tenant>/IncomingWebhook/<altId>/<groupOwner>'",
				prefix))
		}
		if s.GetURLField("tenant") == "" {
			errs = append(errs, fmt.Errorf("%stenant: <required> e.g. '<host>/webhookb2/<group>@<TENANT>/IncomingWebhook/<altId>/<groupOwner>'",
				prefix))
		}
		if s.GetURLField("altid") == "" {
			errs = append(errs, fmt.Errorf("%saltid: <required> e.g. '<host>/webhookb2/<group>@<tenant>/IncomingWebhook/<ALT-ID>/<groupOwner>'",
				prefix))
		}
		if s.GetURLField("groupowner") == "" {
			errs = append(errs, fmt.Errorf("%sgroupowner: <required> e.g. '<host>/webhookb2/<group>@<tenant>/IncomingWebhook/<altId>/<GROUP-OWNER>'",
				prefix))
		}
	case "telegram":
		// telegram://token@telegram
		if s.GetURLField("token") == "" {
			errs = append(errs, fmt.Errorf("%stoken: <required> e.g. '110201543:AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsaw'",
				prefix))
		}
	case "zulip":
		// zulip://botMail:botKey@host:port
		if s.GetURLField("host") == "" {
			errs = append(errs, fmt.Errorf("%shost: <required> e.g. 'example.zulipchat.com'",
				prefix))
		}
		if s.GetURLField("botmail") == "" {
			errs = append(errs, fmt.Errorf("%sbotmail: <required> e.g. 'my-bot@zulipchat.com'",
				prefix))
		}
		if s.GetURLField("botkey") == "" {
			errs = append(errs, fmt.Errorf("%sbotkey: <required> e.g. 'correcthorsebatterystable'",
				prefix))
		}
	case "shoutrrr":
		// Raw
		if s.GetURLField("raw") == "" {
			errs = append(errs, fmt.Errorf("%sraw: <required> e.g. 'service://foo:bar@something'",
				prefix))
		}
	case "generic":
		// generic://host[:port][/path]
		if s.GetURLField("host") == "" {
			errs = append(errs, fmt.Errorf("%shost: <required> e.g. 'example.com'",
				prefix))
		}
		jsonMaps := []string{"custom_headers", "json_payload_vars", "query_vars"}
		for _, jsonMap := range jsonMaps {
			value := s.GetURLField(jsonMap)
			if value != "" {
				converted := jsonMapToString(s.GetURLField(jsonMap), "-")
				if converted == "" {
					errs = append(errs, fmt.Errorf("%s%s: %q <invalid> (must be a JSON map)",
						prefix, jsonMap, value))
				}
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

// checkValuesParams validates the `Params` of the Base struct.
func (b *Base) checkValuesParams(prefix string) error {
	var errs []error

	// Params.*
	for key, value := range b.Params {
		if !util.CheckTemplate(value) {
			errs = append(errs,
				fmt.Errorf("%s%s: %q <invalid> (didn't pass templating)",
					prefix, key, value))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// checkValuesParams validates the `Params` of the Shoutrrr struct.
func (s *Shoutrrr) checkValuesParams(prefix string) error {
	var errs []error
	if baseErrs := s.Base.checkValuesParams(prefix); baseErrs != nil {
		errs = append(errs, baseErrs)
	}

	switch s.GetType() {
	case "smtp":
		// smtp://username:password@host:port[/path]/?from=fromAddress&to=recipient1[,recipient2,...]
		if s.GetParam("fromaddress") == "" {
			errs = append(errs, fmt.Errorf("%sfromaddress: <required> e.g. 'service@gmail.com'",
				prefix))
		}
		if s.GetParam("toaddresses") == "" {
			errs = append(errs, fmt.Errorf("%stoaddresses: <required> e.g. 'name@gmail.com'",
				prefix))
		}
	case "ifttt":
		// ifttt://webhookid/?events=event1[,event2,...]&value1=value1&value2=value2&value3=value3
		if s.GetParam("events") == "" {
			errs = append(errs, fmt.Errorf("%sevents: <required> e.g. 'event1,event2'",
				prefix))
		}
	case "join":
		// join://apiKey@join/?devices=device1[,device2, ...][&icon=icon][&title=title]
		if s.GetParam("devices") == "" {
			errs = append(errs, fmt.Errorf("%sdevices: <required> e.g. '550ddc132c2b4fd28b8b89f735860db1,7294feb73974e5c99d7479ab7b73ba39,d2d775a2f453237d733aa2b7ea2c3ecd'",
				prefix))
		}
	case "teams":
		// teams://group@tenant/altId/groupOwner?host=organization.webhook.office.com
		if s.GetParam("host") == "" {
			errs = append(errs, fmt.Errorf("%shost: <required> e.g. 'example.webhook.office.com'",
				prefix))
		}
	case "telegram":
		// telegram://token@telegram?chats=channel-1[,chat-id-1,...]
		if s.GetParam("chats") == "" {
			errs = append(errs, fmt.Errorf("%schats: <required> e.g. '@channelName' or 'chatID'",
				prefix))
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

// TestSend will test the Shoutrrr by sending a test message.
func (s *Shoutrrr) TestSend(serviceURL string) error {
	if s == nil {
		return errors.New("shoutrrr is nil")
	}

	// Ensure Options is not nil.
	util.InitMap(&s.Options)
	// Default delay to 0s and max_tries to 1 for the test.
	s.SetOption("delay", "0s")
	s.SetOption("max_tries", "1")

	latestVersion := s.ServiceStatus.LatestVersion()
	if latestVersion == "" {
		latestVersion = "MAJOR.MINOR.PATCH"
	}

	webURL := util.TemplateString(
		util.DereferenceOrDefault(s.ServiceStatus.WebURL),
		util.ServiceInfo{LatestVersion: latestVersion})

	testServiceInfo := util.ServiceInfo{
		ID:            util.DereferenceOrDefault(s.ServiceStatus.ServiceID),
		URL:           serviceURL,
		WebURL:        &webURL,
		LatestVersion: latestVersion}

	// Prefix 'TEST - ' if non-empty.
	title := s.Title(testServiceInfo)
	title = util.ValueUnlessDefault(
		title, "TEST - "+title)
	message := s.Message(testServiceInfo)
	message = "TEST" + util.ValueUnlessDefault(
		message, " - "+message)

	return s.Send(
		title,
		message,
		testServiceInfo,
		false,
		false)
}

// Print the SliceDefaults.
func (s *SliceDefaults) Print(prefix string) {
	if s == nil || len(*s) == 0 {
		return
	}

	str := s.String(prefix + "  ")
	fmt.Printf("%snotify:\n%s", prefix, str)
}
