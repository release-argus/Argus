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
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	goshoutrrr "github.com/nicholas-fedor/shoutrrr"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/polymorphic"
)

// TestSend sends a test notification to verify the Shoutrrr configuration.
func (s *Shoutrrr) TestSend(serviceURL string) error {
	if s == nil {
		return errors.New("shoutrrr is nil")
	}

	// Ensure Options is not nil.
	s.Options = util.EnsureMap(s.Options)
	// Default delay to 0s and max_tries to 1 for the test.
	s.setOption("delay", "0s")
	s.setOption("max_tries", "1")

	testServiceInfo := s.ServiceStatus.GetServiceInfo()
	if testServiceInfo.LatestVersion == "" {
		testServiceInfo.LatestVersion = "MAJOR.MINOR.PATCH"
	}

	// Prefix 'TEST - ' if non-empty.
	title := s.Title(testServiceInfo)
	title = util.ValueUnlessZero(
		title, "TEST - "+title,
	)
	message := s.Message(testServiceInfo)
	message = "TEST" + util.ValueUnlessZero(
		message, " - "+message,
	)

	return s.Send(
		title,
		message,
		testServiceInfo,
		false,
		false,
	)
}

// CheckValues validates the fields of each Shoutrrr,
// returning errors encountered and whether any values were modified.
func (s *Shoutrrrs) CheckValues() (error, bool) {
	if s == nil {
		return nil, false
	}

	var errs []error
	changed := false
	keys := util.SortedKeys(*s)
	for _, key := range keys {
		err, keyChanged := (*s)[key].CheckValues()
		if err != nil {
			errs = append(
				errs,
				&decode.KeyFieldError{
					Key: key,
					Err: err,
				},
			)
		}
		changed = changed || keyChanged
	}

	if len(errs) == 0 {
		return nil, changed
	}
	return errors.Join(errs...), false
}

// CheckValues validates the fields of the receiver,
// returning the errors encountered and whether any values were modified.
func (s *Shoutrrr) CheckValues() (error, bool) {
	if s == nil {
		return nil, false
	}
	s.InitMaps()
	changed := s.correctSelf(s.GetType())

	var errs []error
	// Type.
	if err := s.checkValuesType(); err != nil {
		return err, false
	}
	if err := s.checkValuesOptions(); err != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "options",
				Err: err,
			},
		)
	}
	if err := s.checkValuesURLFields(); err != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "url_fields",
				Err: err,
			},
		)
	}
	if err := s.checkValuesParams(); err != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "params",
				Err: err,
			},
		)
	}

	// Exclude matrix since it logs in, so may run into a rate-limit.
	if len(errs) == 0 && s.GetType() != "matrix" {
		//#nosec G104 -- Disregard as we are not giving any rawURLs.
		sender, _ := goshoutrrr.CreateSender()
		if _, err := sender.Locate(s.BuildURL()); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil, changed
	}
	return errors.Join(errs...), false
}

// CheckValues validates the fields of the receiver,
// returning errors encountered and whether any values were modified.
func (b *Base) CheckValues(id string) (error, bool) {
	if b == nil {
		return nil, false
	}
	b.InitMaps()
	itemType := util.FirstNonDefault(b.Type, id)
	changed := b.correctSelf(itemType)

	var errs []error
	if err := b.checkValuesOptions(); err != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "options",
				Err: err,
			},
		)
	}
	if err := b.checkValuesParams(itemType); err != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "params",
				Err: err,
			},
		)
	}

	if len(errs) == 0 {
		return nil, changed
	}
	return errors.Join(errs...), false
}

// CheckValues validates the fields of each [Defaults],
// returning errors encountered and whether any values were modified.
func (s *ShoutrrrsDefaults) CheckValues() (error, bool) {
	if s == nil {
		return nil, false
	}

	var errs []error
	changed := false
	keys := util.SortedKeys(*s)
	for _, key := range keys {
		err, keyChanged := (*s)[key].CheckValues(key)
		if err != nil {
			errs = append(
				errs,
				&decode.KeyFieldError{
					Key: key,
					Err: err,
				},
			)
		}
		changed = changed || keyChanged
	}

	if len(errs) == 0 {
		return nil, changed
	}
	return errors.Join(errs...), false
}

// CheckValues validates the fields of the receiver,
// returning the errors encountered and whether any values were modified.
func (d *Defaults) CheckValues(id string) (error, bool) {
	var errs []error
	typeName := id
	if d != nil {
		typeName = util.FirstNonDefault(d.Type, id)
	}

	// Verify valid type.
	if !util.Contains(SupportedTypes, typeName) {
		errs = append(
			errs,
			polymorphic.InvalidTypeError{
				Key:     "type",
				Value:   typeName,
				Allowed: SupportedTypes,
			},
		)
	}

	changed := false
	if d != nil {
		// Run the Base checks.
		var err error
		if err, changed = d.Base.CheckValues(id); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil, changed
	}
	return errors.Join(errs...), false
}

// Print writes the ShoutrrrsDefaults to stdout with the given prefix.
func (s *ShoutrrrsDefaults) Print(prefix string) {
	if s == nil || len(*s) == 0 {
		return
	}

	str := s.String(prefix + "  ")
	fmt.Printf(
		"%snotify:\n%s",
		prefix, str,
	)
}

// correctSelf normalises config values and reports whether anything changed.
func (b *Base) correctSelf(shoutrrrType string) (changed bool) {
	// Port, strip leading :
	if port, ok := strings.CutPrefix(b.getURLField("port"), ":"); ok {
		b.setURLField("port", port)
		changed = true
	}

	// Path, strip leading /
	if path, ok := strings.CutPrefix(b.getURLField("path"), "/"); ok {
		b.setURLField("path", path)
		changed = true
	}

	// Host: Strip schema/port.
	host := b.getURLField("host")
	if h, p := parseHostPort(host); h != host {
		b.setURLField("host", h)
		if p != "" {
			b.setURLField("port", p)
		}
		changed = true
	}

	switch shoutrrrType {
	case "generic":
		// Deprecated: custom_headers -> headers.
		if customHeaders := b.getURLField("custom_headers"); customHeaders != "" {
			if headers := b.getURLField("headers"); headers == "" {
				logx.Deprecated("Renaming 'notify.generic.url_fields.custom_headers' to 'notify.generic.url_fields.headers'")
				b.setURLField("headers", customHeaders)
			}
			b.setURLField("custom_headers", "")
			changed = true
		}
	case "discord":
		// Deprecated: threadid -> thread_id.
		if v := b.GetParam("threadid"); v != "" {
			if b.GetParam("thread_id") == "" {
				logx.Deprecated("Renaming 'notify.discord.params.threadid' to 'notify.discord.params.thread_id'")
				b.setParam("thread_id", v)
			}
			b.setParam("threadid", "")
			changed = true
		}
	case "ifttt":
		// Deprecated: usemessageasvalue -> messagevalue.
		if v := b.GetParam("usemessageasvalue"); v != "" {
			if b.GetParam("messagevalue") == "" {
				logx.Deprecated("Renaming 'notify.ifttt.params.usemessageasvalue' to 'notify.ifttt.params.messagevalue'")
				b.setParam("messagevalue", v)
			}
			b.setParam("usemessageasvalue", "")
			changed = true
		}
		// Deprecated: usetitleasvalue -> titlevalue.
		if v := b.GetParam("usetitleasvalue"); v != "" {
			if b.GetParam("titlevalue") == "" {
				logx.Deprecated("Renaming 'notify.ifttt.params.usetitleasvalue' to 'notify.ifttt.params.titlevalue'")
				b.setParam("titlevalue", v)
			}
			b.setParam("usetitleasvalue", "")
			changed = true
		}
	case "matrix":
		// Remove #'s in channel aliases.
		if rooms := b.GetParam("rooms"); strings.Contains(rooms, "#") {
			b.setParam(
				"rooms",
				strings.ReplaceAll(rooms, "#", ""),
			)
			changed = true
		}
	case "mattermost":
		// Channel, strip leading /
		if channel, ok := strings.CutPrefix(b.getURLField("channel"), "/"); ok {
			b.setURLField("channel", channel)
			changed = true
		}
	case "ntfy":
		// Deprecated: disabletls -> disabletlsverification.
		if disableTLS := b.GetParam("disabletls"); disableTLS != "" {
			if disableTLSVerification := b.GetParam("disabletlsverification"); disableTLSVerification == "" {
				logx.Deprecated("Renaming 'notify.ntfy.params.disabletls' to 'notify.ntfy.params.disabletlsverification'")
				b.setParam("disabletlsverification", disableTLS)
			}
			b.setParam("disabletls", "")
			changed = true
		}
	case "slack":
		// Ensure color is stored as #RRGGBB so BuildURL's queryParam encodes it correctly.
		// Migrate %23 (old manual encoding) and bare hex to #-prefixed form.
		if color := b.GetParam("color"); color != "" {
			var newColor string
			switch {
			case strings.HasPrefix(color, "%23"):
				// Migrate old URL-encoded format; queryParam will re-encode.
				newColor = "#" + strings.TrimPrefix(color, "%23")
			case util.RegexCheck(`^[a-fA-F0-9]{6}$`, color):
				newColor = "#" + color
			}
			if newColor != "" {
				b.setParam("color", newColor)
				changed = true
			}
		}
	case "smtp":
		// Deprecated: skiptlsverification -> skiptlsverify.
		if v := b.GetParam("skiptlsverification"); v != "" {
			if b.GetParam("skiptlsverify") == "" {
				logx.Deprecated("Renaming 'notify.smtp.params.skiptlsverification' to 'notify.smtp.params.skiptlsverify'")
				b.setParam("skiptlsverify", v)
			}
			b.setParam("skiptlsverification", "")
			changed = true
		}
	case "zulip":
		// BotMail, replace the @ with a %40 - https://containrrr.dev/shoutrrr/v0.5/services/zulip/
		if botMail := b.getURLField("botmail"); strings.Contains(botMail, "@") {
			b.setURLField("botmail", strings.ReplaceAll(botMail, "@", "%40"))
			changed = true
		}
	}

	return
}

// normaliseParamSelect normalises a Param against allowed (case-insensitive),
// setting it to the correctly-cased value and returning true, or leaving it unchanged and returning false.
func (b *Base) normaliseParamSelect(key string, value string, allowed []string) bool {
	lc := strings.ToLower(value)
	for _, opt := range allowed {
		if strings.ToLower(opt) == lc {
			b.setParam(key, opt)
			return true
		}
	}
	return false
}

// checkValuesType validates the Type field against supported notification types.
func (s *Shoutrrr) checkValuesType() error {
	// Check we have a Type.
	sType := s.GetType()
	if !util.Contains(SupportedTypes, sType) {
		sTypeWithoutID := util.FirstNonDefault(s.Type, s.Main.Type)
		if sTypeWithoutID == "" {
			return &decode.FieldError{
				Key:         "type",
				Description: "e.g. 'slack', see the docs for possible types - https://release-argus.io/docs/config/notify",
			}
		}
	}

	// Check the Type doesn't differ in the Main.
	if s.Main.Type != "" && sType != s.Main.Type {
		return &decode.FieldError{
			Key:   "type",
			Value: sType,
			Description: fmt.Sprintf(
				"must be the same as the root notify.%s.type (%s)",
				s.ID, s.Main.Type,
			),
		}
	}

	// Invalid/Unknown type.
	if !util.Contains(SupportedTypes, sType) {
		return polymorphic.InvalidTypeError{
			Key:     "type",
			Value:   sType,
			Allowed: SupportedTypes,
		}
	}

	// Pass.
	return nil
}

// checkValuesOptions validates the Options field.
func (b *Base) checkValuesOptions() error {
	var errs []error
	// Options.Delay.
	if optionDelay := b.getOption("delay"); optionDelay != "" {
		// Treat integers as seconds by default.
		if _, err := strconv.Atoi(optionDelay); err == nil {
			b.Options["delay"] += "s"
		}
		if _, err := time.ParseDuration(b.Options["delay"]); err != nil {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "delay",
					Value:       optionDelay,
					Description: "use 'AhBmCs' duration format",
				},
			)
		}
	}

	// Options.MaxTries.
	if maxTriesStr := b.getOption("max_tries"); maxTriesStr != "" {
		if maxTries, err := strconv.ParseUint(maxTriesStr, 10, 64); err == nil {
			// Too large.
			if maxTries > math.MaxUint8 {
				errs = append(
					errs,
					&decode.FieldError{
						Key:         "max_tries",
						Value:       maxTriesStr,
						Description: fmt.Sprintf("must be <= %d", math.MaxUint8),
					},
				)
			}
		} else {
			// Too large?
			if util.RegexCheck(`^-?\d+$`, maxTriesStr) {
				// Negative.
				if strings.HasPrefix(maxTriesStr, "-") {
					errs = append(
						errs,
						&decode.FieldError{
							Key:         "max_tries",
							Value:       maxTriesStr,
							Description: "must be positive",
						},
					)
					// Positive.
				} else {
					errs = append(
						errs,
						&decode.FieldError{
							Key:         "max_tries",
							Value:       maxTriesStr,
							Description: fmt.Sprintf("must be <= %d", math.MaxUint8),
						},
					)
				}
			} else {
				// Not an integer.
				errs = append(
					errs,
					&decode.FieldError{
						Key:         "max_tries",
						Value:       maxTriesStr,
						Description: "must be an integer",
					},
				)
			}
		}
	}

	// Options.Message.
	optionMessage := b.getOption("message")
	if !util.CheckTemplate(optionMessage) {
		errs = append(
			errs,
			&decode.FieldError{
				Key:         "message",
				Value:       optionMessage,
				Description: "didn't pass templating",
			},
		)
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// checkValuesURLFields validates the URLFields field.
func (s *Shoutrrr) checkValuesURLFields() error {
	var errs []error

	switch s.GetType() {
	case "bark":
		// bark://:devicekey@host[:port][/path]
		if s.GetURLField("devicekey") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key: "devicekey",
				},
			)
		}
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key: "host",
				},
			)
		}
	case "discord":
		// discord://token@webhookid
		if s.GetURLField("token") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "token",
					Description: "e.g. 'https://discord.com/api/webhooks/[ 975870285909737583 <- webhookid ]/[ QEdyk-Qi5AiMXoZdxQFpWNcwEfmz5oOm_1Rni9DnjQAUap4zWcurM4IquamVrDIyNgBG <- TOKEN ]'",
				},
			)
		}
		if s.GetURLField("webhookid") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "webhookid",
					Description: "e.g. 'https://discord.com/api/webhooks/[ 975870285909737583 <- WEBHOOKID ]/[ QEdyk-Qi5AiMXoZdxQFpWNcwEfmz5oOm_1Rni9DnjQAUap4zWcurM4IquamVrDIyNgBG <- token ]'",
				},
			)
		}
	case "smtp":
		// smtp://username:password@host[:port]/?fromaddress=X&toaddresses=Y[&fromname=X]
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "host",
					Description: "e.g. 'smtp.example.com'",
				},
			)
		}
	case "gotify":
		// gotify://host[:port][/path]/token
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "host",
					Description: "e.g. 'gotify.example.com'",
				},
			)
		}
		if s.GetURLField("token") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "token",
					Description: "e.g. 'Aod9Cb0zXCeOrnD'",
				},
			)
		}
	case "googlechat":
		// googlechat://url
		if s.GetURLField("raw") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "raw",
					Description: "e.g. 'https://chat.googleapis.com/v1/spaces/FOO/messages?key=bar&token=baz'",
				},
			)
		}
	case "ifttt":
		// ifttt://webhookid
		if s.GetURLField("webhookid") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "webhookid",
					Description: "e.g. 'h1fyLh42h7lDI2L11T-bv'",
				},
			)
		}
	case "join":
		// join://apiKey@join
		if s.GetURLField("apikey") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "apikey",
					Description: "e.g. 'f8eae56127864015b0d2f4d8db6ff53f'",
				},
			)
		}
	case "matrix":
		// matrix://user:password@host[:port]/[?rooms=!roomID1,roomAlias2]][&disableTLS=yes]
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "host",
					Description: "e.g. 'matrix.example.com'",
				},
			)
		}
		if s.GetURLField("password") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "password",
					Description: "e.g. 'pass123' (with user) OR 'access_token' (no user)",
				},
			)
		}
	case "mattermost":
		// mattermost://[username@]host[:port][/path]/token[/channel]
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "host",
					Description: "e.g. 'mattermost.example.com'",
				},
			)
		}
		if s.GetURLField("token") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "token",
					Description: "e.g. 'Aod9Cb0zXCeOrnD'",
				},
			)
		}
	case "ntfy":
		// ntfy://[username]:[password]@[host][:port]/topic
		if s.GetURLField("topic") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key: "topic",
				},
			)
		}
	case "notifiarr":
		// notifiarr://apikey[?name=X&channel=X&thumbnail=X&image=X&color=X]
		if s.GetURLField("apikey") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "apikey",
					Description: "found on your Notifiarr account settings page",
				},
			)
		}
	case "opsgenie":
		// opsgenie://host[:port]/apiKey
		if s.GetURLField("apikey") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "apikey",
					Description: "e.g. 'xxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx'",
				},
			)
		}
	case "pushbullet":
		// pushbullet://token/targets
		if s.GetURLField("token") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "token",
					Description: "e.g. 'o.5NfxzU9yH4xBZlEXZArRtyUB4S4Ua8Hd'",
				},
			)
		}
		if s.GetURLField("targets") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "targets",
					Description: "e.g. 'fpwfXzDCYsTxw4VfAAoHiR,5eAzVLKp5VRUMJeYehwbzv,XR7VKoK5b2MYWDpstD3Hfq'",
				},
			)
		}
	case "pushover":
		// pushover://token@user
		if s.GetURLField("token") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "token",
					Description: "e.g. 'aayohdg8gqjj3ssszuqwwmuipt5gcd'",
				},
			)
		}
		if s.GetURLField("user") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "user",
					Description: "e.g. '2QypyiVSnURsw72cpnXCuVAQMJpKKY'",
				},
			)
		}
	case "rocketchat":
		// rocketchat://[username@]host[:port][/path]/tokenA/tokenB/channel
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "host",
					Description: "e.g. 'rocket-chat.example.com'",
				},
			)
		}
		if s.GetURLField("tokena") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "tokena",
					Description: "e.g. '8eGdRzc9r4YYNyvge'",
				},
			)
		}
		if s.GetURLField("tokenb") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "tokenb",
					Description: "e.g. '2XYQcX9NBwJBKfQnphpebPcnXZcPEi32Nt4NKJfrnbhsbRfX'",
				},
			)
		}
		if s.GetURLField("channel") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "channel",
					Description: "e.g. 'argusChannel' or '@user'",
				},
			)
		}
	case "slack":
		// slack://token:token@channel
		if s.GetURLField("token") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "token",
					Description: "e.g. '123456789012-1234567890123-4mt0t4l1YL3g1T5L4cK70k3N'",
				},
			)
		}
		if s.GetURLField("channel") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "channel",
					Description: "e.g. 'C001CH4NN3L' or 'webhook'",
				},
			)
		}
	case "shoutrrr":
		// Raw shoutrrr URL passthrough.
		if s.GetURLField("raw") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "raw",
					Description: "e.g. 'slack://TOKEN@CHANNEL'",
				},
			)
		}
	case "telegram":
		// telegram://token@telegram
		if s.GetURLField("token") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "token",
					Description: "e.g. '110201543:AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsaw'",
				},
			)
		}
	case "zulip":
		// zulip://botMail:botKey@host:port
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "host",
					Description: "e.g. 'zulipchat.example.com'",
				},
			)
		}
		if s.GetURLField("botmail") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "botmail",
					Description: "e.g. 'my-bot@zulipchat.com'",
				},
			)
		}
		if s.GetURLField("botkey") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "botkey",
					Description: "e.g. 'correcthorsebatterystable'",
				},
			)
		}
	case "generic":
		// generic://host[:port][/path]
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "host",
					Description: "e.g. 'example.com'",
				},
			)
		}
		jsonMaps := []string{"headers", "json_payload_vars", "query_vars"}
		for _, jsonMap := range jsonMaps {
			value := s.GetURLField(jsonMap)
			if value != "" {
				converted := jsonMapToString(s.GetURLField(jsonMap), "-")
				if converted == "" {
					errs = append(
						errs,
						&decode.FieldError{
							Key:         jsonMap,
							Value:       value,
							Description: "must be a JSON map",
						},
					)
				}
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

// checkValuesParams validates the Params field.
func (s *Shoutrrr) checkValuesParams() error {
	var errs []error
	itemType := s.GetType()
	if baseErrs := s.Base.checkValuesParams(itemType); baseErrs != nil {
		errs = append(errs, baseErrs)
	}

	switch itemType {
	case "ifttt":
		// ifttt://webhookid/?events=event1[,event2,...]&value1=value1&value2=value2&value3=value3
		if s.GetParam("events") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "events",
					Description: "e.g. 'event1,event2'",
				},
			)
		}
	case "join":
		// join://apiKey@join/?devices=device1[,device2, ...][&icon=icon][&title=title]
		if s.GetParam("devices") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "devices",
					Description: "e.g. '550ddc132c2b4fd28b8b89f735860db1,7294feb73974e5c99d7479ab7b73ba39,d2d775a2f453237d733aa2b7ea2c3ecd'",
				},
			)
		}
	case "smtp":
		// smtp://username:password@host[:port]/?fromaddress=X&toaddresses=Y[&fromname=X]
		if s.GetParam("fromaddress") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "fromaddress",
					Description: "e.g. 'service@gmail.com'",
				},
			)
		}
		if s.GetParam("toaddresses") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "toaddresses",
					Description: "e.g. 'name@gmail.com'",
				},
			)
		}
	case "teams":
		// teams://?host=fullPowerAutomateURL
		if s.GetParam("host") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "host",
					Description: "Full Power Automate workflow URL, e.g. 'https://prod-00.westus.logic.azure.com:443/workflows/...'",
				},
			)
		}
	case "telegram":
		// telegram://token@telegram?chats=channel-1[,chat-id-1,...]
		if s.GetParam("chats") == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "chats",
					Description: "e.g. '@channelName' or 'chatID'",
				},
			)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

// checkValuesParams validates the Params field of Base.
func (b *Base) checkValuesParams(itemType string) error {
	var errs []error

	// Normalise 'select' params.
	if e := b.checkValuesParamsSelects(itemType); e != nil {
		errs = append(errs, e)
	}

	// Params.*
	for key, value := range b.Params {
		if !util.CheckTemplate(value) {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         key,
					Value:       value,
					Description: "didn't pass templating",
				},
			)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// checkValuesParamsSelects validates Params that must match an allowed set for itemType.
func (b *Base) checkValuesParamsSelects(itemType string) error {
	var errs []error

	switch itemType {
	case "bark":
		if err := b.validateParamSelect("scheme", barkNtfyParamScheme); err != nil {
			errs = append(errs, err)
		}
		if err := b.validateParamSelect("sound", barkParamSound); err != nil {
			errs = append(errs, err)
		}
	case "generic":
		if err := b.validateParamSelect("requestmethod", genericParamRequestmethod); err != nil {
			errs = append(errs, err)
		}
	case "ntfy":
		if err := b.validateParamSelect("priority", ntfyParamPriority); err != nil {
			errs = append(errs, err)
		}
		if err := b.validateParamSelect("scheme", barkNtfyParamScheme); err != nil {
			errs = append(errs, err)
		}
	case "smtp":
		if err := b.validateParamSelect("auth", smtpParamAuth); err != nil {
			errs = append(errs, err)
		}
		if err := b.validateParamSelect("encryption", smtpParamEncryption); err != nil {
			errs = append(errs, err)
		}
	case "telegram":
		if err := b.validateParamSelect("parsemode", telegramParamParsemode); err != nil {
			errs = append(errs, err)
		}
	case "zulip":
		if err := b.validateParamSelect("type", zulipParamType); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

// validateParamSelect normalises key against allowed (case-insensitive), returning an error if not found.
func (b *Base) validateParamSelect(key string, allowed []string) error {
	value := b.GetParam(key)
	if value == "" {
		return nil
	}

	if ok := b.normaliseParamSelect(key, value, allowed); ok {
		return nil
	}

	return polymorphic.InvalidTypeError{
		Key:     key,
		Value:   value,
		Allowed: allowed,
	}
}

// parseHostPort extracts the hostname and optional port from an address.
// The address may optionally include a URL scheme.
// If the address cannot be parsed, it is returned unchanged with an empty port.
func parseHostPort(input string) (string, string) {
	address := strings.TrimSpace(input)
	// Add scheme if missing (required for net/url).
	if !strings.Contains(address, "://") {
		address = "https://" + address
	}

	u, err := url.Parse(address)
	if err != nil {
		return input, ""
	}

	return u.Hostname(), u.Port()
}
