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
	"github.com/release-argus/Argus/utils"
)

// CheckValues of this Slice.
func (s *Slice) CheckValues(prefix string) (errs error) {
	if s == nil || len(*s) == 0 {
		return
	}

	for key := range *s {
		if err := (*s)[key].CheckValues(prefix + "    "); err != nil {
			errs = fmt.Errorf("%s%s  %s:\\%w", utils.ErrorToString(errs), prefix, key, err)
		}
	}

	if errs != nil {
		errs = fmt.Errorf("%snotify:\\%w", prefix, errs)
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
	delay := s.GetSelfOption("delay")
	if delay != "" {
		// Default to seconds when an integer is provided
		if _, err := strconv.Atoi(delay); err == nil {
			(*s.Options)["delay"] += "s"
		}
		if _, err := time.ParseDuration(delay); err != nil {
			errsOptions = fmt.Errorf("%s%s  delay: <invalid> %q (Use 'AhBmCs' duration format)\\", utils.ErrorToString(errsOptions), prefix, delay)
		}
	}

	if s.Main != nil {
		if utils.GetFirstNonDefault(s.Type, s.Main.Type) == "" {
			errs = fmt.Errorf("%s%stype: <required> e.g. 'slack', see the docs for possible types\\", utils.ErrorToString(errs), prefix)
			tmp := ""
			s.Type = tmp
		}
		switch s.Type {
		case "discord":
			// discord://token@webhook_id
			if s.GetURLField("token") == "" {
				errsURLFields = fmt.Errorf("%s%s  token: <required> e.g. 'https://discord.com/api/webhooks/[ 975870285909737583 <- webhook_id ]/[ QEdyk-Qi5AiMXoZdxQFpWNcwEfmz5oOm_1Rni9DnjQAUap4zWcurM4IquamVrDIyNgBG <- TOKEN ]'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("webhook_id") == "" {
				errsURLFields = fmt.Errorf("%s%s  webhook_id: <required> e.g. 'https://discord.com/api/webhooks/[ 975870285909737583 <- WEBHOOK_ID ]/[ QEdyk-Qi5AiMXoZdxQFpWNcwEfmz5oOm_1Rni9DnjQAUap4zWcurM4IquamVrDIyNgBG <- token ]'\\", utils.ErrorToString(errsURLFields), prefix)
			}
		case "email":
			// smtp://username:password@host:port[/path]
			if s.GetURLField("host") == "" {
				errsURLFields = fmt.Errorf("%s%s  host: <required> e.g. 'smtp.example.io'\\", utils.ErrorToString(errsURLFields), prefix)
			}
		case "gotify":
			// gotify://host:port/path/token
			if s.GetURLField("host") == "" {
				errsURLFields = fmt.Errorf("%s%s  host: <required> e.g. 'gotify.example.io'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("token") == "" {
				errsURLFields = fmt.Errorf("%s%s  token: <required> e.g. 'Aod9Cb0zXCeOrnD'\\", utils.ErrorToString(errsURLFields), prefix)
			}
		case "google_chat":
			// googlechat://url
			if s.GetURLField("raw") == "" {
				errsURLFields = fmt.Errorf("%s%s  raw: <required> e.g. 'https://chat.googleapis.com/v1/spaces/FOO/messages?key=bar&token=baz'\\", utils.ErrorToString(errsURLFields), prefix)
			}
		case "ifttt":
			// ifttt://webhook_id
			if s.GetURLField("webhook_id") == "" {
				errsURLFields = fmt.Errorf("%s%s  webhook_id: <required> e.g. 'h1fyLh42h7lDI2L11T-bv'\\", utils.ErrorToString(errsURLFields), prefix)
			}
		case "join":
			// join://apiKey@join
			if s.GetURLField("api_key") == "" {
				errsURLFields = fmt.Errorf("%s%s  api_key: <required> e.g. 'f8eae56127864015b0d2f4d8db6ff53f'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("devices") == "" {
				errsURLFields = fmt.Errorf("%s%s  devices: <required> e.g. '550ddc132c2b4fd28b8b89f735860db1,7294feb73974e5c99d7479ab7b73ba39,d2d775a2f453237d733aa2b7ea2c3ecd'\\", utils.ErrorToString(errsURLFields), prefix)
			}
		case "mattermost":
			// mattermost://[username@]host[:port][/path]/token[/channel]
			if s.GetURLField("host") == "" {
				errsURLFields = fmt.Errorf("%s%s  host: <required> e.g. 'mattermost.example.io'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("token") == "" {
				errsURLFields = fmt.Errorf("%s%s  token: <required> e.g. 'Aod9Cb0zXCeOrnD'\\", utils.ErrorToString(errsURLFields), prefix)
			}
		case "matrix":
			// matrix://user:password@host
			if s.GetURLField("host") == "" {
				errsURLFields = fmt.Errorf("%s%s  host: <required> e.g. 'matrix.example.io'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("password") == "" {
				errsURLFields = fmt.Errorf("%s%s  password: <required> e.g. '<password>' OR '<access_token>'\\", utils.ErrorToString(errsURLFields), prefix)
			}
		case "ops_genie":
			// opsgenie://host[:port][/path]/apiKey
			if s.GetURLField("api_key") == "" {
				errsURLFields = fmt.Errorf("%s%s  api_key: <required> e.g. 'xxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx'\\", utils.ErrorToString(errsURLFields), prefix)
			}
		case "pushbullet":
			// pushbullet://token/targets
			if s.GetURLField("token") == "" {
				errsURLFields = fmt.Errorf("%s%s  token: <required> e.g. 'o.5NfxzU9yH4xBZlEXZArRtyUB4S4Ua8Hd'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("targets") == "" {
				errsURLFields = fmt.Errorf("%s%s  targets: <required> e.g. 'fpwfXzDCYsTxw4VfAAoHiR,5eAzVLKp5VRUMJeYehwbzv,XR7VKoK5b2MYWDpstD3Hfq'\\", utils.ErrorToString(errsURLFields), prefix)
			}
		case "pushover":
			// pushover://token@user
			if s.GetURLField("token") == "" {
				errsURLFields = fmt.Errorf("%s%s  token: <required> e.g. 'aayohdg8gqjj3ssszuqwwmuipt5gcd'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("user") == "" {
				errsURLFields = fmt.Errorf("%s%s  user: <required> e.g. '2QypyiVSnURsw72cpnXCuVAQMJpKKY'\\", utils.ErrorToString(errsURLFields), prefix)
			}
		case "rocket_chat":
			// rocket_chat://[username@]host:port[/port]/tokenA/tokenB/channel
			if s.GetURLField("host") == "" {
				errsURLFields = fmt.Errorf("%s%s  host: <required> e.g. 'rocket-chat.example.io'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("port") == "" {
				errsURLFields = fmt.Errorf("%s%s  port: <required> e.g. '443'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("token_a") == "" {
				errsURLFields = fmt.Errorf("%s%s  token_a: <required> e.g. '8eGdRzc9r4YYNyvge'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("token_b") == "" {
				errsURLFields = fmt.Errorf("%s%s  token_b: <required> e.g. '2XYQcX9NBwJBKfQnphpebPcnXZcPEi32Nt4NKJfrnbhsbRfX'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("channel") == "" {
				errsURLFields = fmt.Errorf("%s%s  channel: <required> e.g. 'argusChannel' or '@user'\\", utils.ErrorToString(errsURLFields), prefix)
			}
		case "slack":
			// slack://token:token@channel
			slacktype := s.GetURLField("slacktype")
			if slacktype == "" {
				errsURLFields = fmt.Errorf("%s%s  slacktype: <required> e.g. 'bot' or 'webhook'\\", utils.ErrorToString(errsURLFields), prefix)
			} else if slacktype != "bot" && slacktype != "webhook" {
				errsURLFields = fmt.Errorf("%s%s  slacktype: <invalid> %q (should be either 'bot' or 'webhook')\\", utils.ErrorToString(errsURLFields), prefix, slacktype)
			}
			if s.GetURLField("token") == "" {
				errsURLFields = fmt.Errorf("%s%s  token: <required> e.g. '123456789012-1234567890123-4mt0t4l1YL3g1T5L4cK70k3N'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("channel") == "" {
				errsURLFields = fmt.Errorf("%s%s  channel: <required> e.g. 'C001CH4NN3L' or 'webhook'\\", utils.ErrorToString(errsURLFields), prefix)
			}
		case "teams":
			// teams://[group@][tenant][/altid][/groupowner]
			if s.GetURLField("organization") == "" {
				errsURLFields = fmt.Errorf("%s%s  organization: <required> e.g. 'https://<ORGANIZATION>.webhook.office.com/webhookb2/<group>@<tenant>/IncomingWebhook/<altId>/<groupOwner>'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("group") == "" {
				errsURLFields = fmt.Errorf("%s%s  group: <required> e.g. 'https://<organization>.webhook.office.com/webhookb2/<GROUP>@<tenant>/IncomingWebhook/<altId>/<groupOwner>'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("tenant") == "" {
				errsURLFields = fmt.Errorf("%s%s  tenant: <required> e.g. 'https://<organization>.webhook.office.com/webhookb2/<group>@<TENANT>/IncomingWebhook/<altId>/<groupOwner>'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("alt_id") == "" {
				errsURLFields = fmt.Errorf("%s%s  alt_id: <required> e.g. 'https://<organization>.webhook.office.com/webhookb2/<group>@<tenant>/IncomingWebhook/<ALT-ID>/<groupOwner>'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("group_owner") == "" {
				errsURLFields = fmt.Errorf("%s%s  group_owner: <required> e.g. 'https://<organization>.webhook.office.com/webhookb2/<group>@<tenant>/IncomingWebhook/<altId>/<GROUP-OWNER>'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetParam("host") == "" {
				errsParams = fmt.Errorf("%s%s  host: <required> e.g. 'example.webhook.office.com'\\", utils.ErrorToString(errsParams), prefix)
			}
		case "telegram":
			// telegram://token@telegram
			if s.GetURLField("token") == "" {
				errsURLFields = fmt.Errorf("%s%s  token: <required> e.g. '110201543:AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsaw'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("chats") == "" {
				errsURLFields = fmt.Errorf("%s%s  token: <required> e.g. 'channelName' or 'chat_id'\\", utils.ErrorToString(errsURLFields), prefix)
			}
		case "zulip_chat":
			// zulip://botMail:botKey@host:port
			if s.GetURLField("host") == "" {
				errsURLFields = fmt.Errorf("%s%s  host: <required> e.g. 'example.zulipchat.com'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("port") == "" {
				errsURLFields = fmt.Errorf("%s%s  port: <required> e.g. '443'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("bot_mail") == "" {
				errsURLFields = fmt.Errorf("%s%s  bot_mail: <required> e.g. 'my-bot@zulipchat.com'\\", utils.ErrorToString(errsURLFields), prefix)
			}
			if s.GetURLField("bot_key") == "" {
				errsURLFields = fmt.Errorf("%s%s  bot_key: <required> e.g. 'correcthorsebatterystable'\\", utils.ErrorToString(errsURLFields), prefix)
			}
		case "shoutrrr":
			// Raw
			if s.GetURLField("raw") == "" {
				errsURLFields = fmt.Errorf("%s%s  raw: <required> e.g. 'service://foo:bar@something'\\", utils.ErrorToString(errsURLFields), prefix)
			}
		default:
			// Invalid/Unknown type
			if s.Type != "" {
				errs = fmt.Errorf("%s%stype: <invalid> e.g. 'slack', see the docs for possible types\\", utils.ErrorToString(errs), prefix)
			}
		}

		sender, _ := shoutrrr_lib.CreateSender()
		_, err := sender.Locate(s.GetURL())
		if err != nil && s.Type != "" {
			errsLocate = fmt.Errorf("%s%s^ %q\\", utils.ErrorToString(errs), prefix, err.Error())
		}
	}

	// Trivial corrections
	// ZulipChat - Replace to match example - https://containrrr.dev/shoutrrr/v0.5/services/zulip/
	bot_mail := s.GetSelfURLField("bot_mail")
	if bot_mail != "" {
		s.SetURLField("bot_mail", strings.ReplaceAll(bot_mail, "@", "%40"))
	}
	// Slack - # -> %23
	for key := range *s.Params {
		// https://containrrr.dev/shoutrrr/v0.5/services/slack/
		// The format for the Color prop follows the slack docs but # needs to be escaped as %23 when passed in a URL.
		// So #ff8000 would be %23ff8000 etc.
		if strings.Contains(key, "color") && s.GetSelfParam(key) != "" {
			s.SetParam(key, strings.Replace(s.GetSelfParam(key), "#", "%23", 1))
		}
	}

	if errsOptions != nil {
		errs = fmt.Errorf("%s%soptions:\\%w", utils.ErrorToString(errs), prefix, errsOptions)
	}
	if errsURLFields != nil {
		errs = fmt.Errorf("%s%surl_fields:\\%w", utils.ErrorToString(errs), prefix, errsURLFields)
	}
	if errsParams != nil {
		errs = fmt.Errorf("%s%sparams:\\%w", utils.ErrorToString(errs), prefix, errsParams)
	}
	if errsLocate != nil {
		errs = fmt.Errorf("%s%w", utils.ErrorToString(errs), errsLocate)
	}
	return
}

// Print the Slice.
func (n *Slice) Print(prefix string) bool {
	if n == nil || len(*n) == 0 {
		return false
	}

	fmt.Printf("%snotify:\n", prefix)
	for shoutrrrID, shoutrrr := range *n {
		fmt.Printf("%s  %s:\n", prefix, shoutrrrID)
		shoutrrr.Print(prefix + "    ")
	}
	return true
}

// Print the Shourrr Struct.
func (s *Shoutrrr) Print(prefix string) {
	utils.PrintlnIfNotDefault(s.Type, fmt.Sprintf("%stype: %s", prefix, s.Type))

	if s.Options != nil && len(*s.Options) > 0 {
		fmt.Printf("%soptions:\n", prefix)
		for key := range *s.Options {
			utils.PrintlnIfNotDefault(s.GetSelfOption(key), fmt.Sprintf("%s  %s: %s", prefix, key, s.GetSelfOption(key)))
		}
	}

	if s.URLFields != nil && len(*s.URLFields) > 0 {
		fmt.Printf("%surl_fields:\n", prefix)
		for key := range *s.URLFields {
			utils.PrintlnIfNotDefault(s.GetSelfURLField(key), fmt.Sprintf("%s  %s: %s", prefix, key, s.GetSelfURLField(key)))
		}
	}

	if s.Params != nil && len(*s.Params) != 0 {
		fmt.Printf("%sparams:\n", prefix)
		for key := range *s.Params {
			utils.PrintlnIfNotDefault(s.GetSelfParam(key), fmt.Sprintf("%s  %s: %s", prefix, key, s.GetSelfParam(key)))
		}
	}
}
