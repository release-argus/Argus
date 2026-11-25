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
	"slices"
	"strconv"
	"strings"
	"time"

	goshoutrrr "github.com/nicholas-fedor/shoutrrr"
	"github.com/nicholas-fedor/shoutrrr/pkg/format"
	goshoutrrrTypes "github.com/nicholas-fedor/shoutrrr/pkg/types"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
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
				&decode.ErrKeyField{
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
			&decode.ErrKeyField{
				Key: "options",
				Err: err,
			},
		)
	}
	if err := s.checkValuesURLFields(); err != nil {
		errs = append(
			errs,
			&decode.ErrKeyField{
				Key: "url_fields",
				Err: err,
			},
		)
	}
	if err := s.checkValuesParams(); err != nil {
		errs = append(
			errs,
			&decode.ErrKeyField{
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

// CheckURL returns the error Shoutrrr gives when locating the service for the receiver's URL.
func (s *Shoutrrr) CheckURL() error {
	// Legacy Teams configs (Office 365 Connectors) skip Locate, as Shoutrrr rejects their URL.
	if s.GetType() == "teams" && s.GetURLField("group") != "" {
		return nil
	}

	//#nosec G104 -- Disregard as we are not giving any rawURLs.
	sender, _ := goshoutrrr.CreateSenderWithOptions(goshoutrrrTypes.SenderOptions{})
	_, err := sender.Locate(s.BuildURL())
	return err //nolint:wrapcheck
}

// CheckURLs returns the [Shoutrrr.CheckURL] errors of each Shoutrrr.
func (s *Shoutrrrs) CheckURLs() error {
	if s == nil {
		return nil
	}

	var errs []error
	for _, key := range util.SortedKeys(*s) {
		if err := (*s)[key].CheckURL(); err != nil {
			errs = append(
				errs,
				&decode.ErrKeyField{
					Key: key,
					Err: err,
				},
			)
		}
	}
	return errors.Join(errs...)
}

// LogWarnings logs each Shoutrrr of serviceID that will fail to send, and the unknown params each will drop.
func (s *Shoutrrrs) LogWarnings(serviceID string) {
	if s == nil {
		return
	}

	for _, key := range util.SortedKeys(*s) {
		notify := (*s)[key]
		logFrom := logx.LogFrom{Primary: serviceID, Secondary: key}
		if err := notify.CheckURL(); err != nil {
			logx.Warn(fmt.Sprintf("failed to locate notify, so it will fail to send: %v", err), logFrom, true)
			continue
		}

		//#nosec G104 -- Disregard as we are not giving any rawURLs.
		sender, _ := goshoutrrr.CreateSenderWithOptions(goshoutrrrTypes.SenderOptions{})
		url := notify.BuildURL()
		if unknown := unknownParams(sender, url, *notify.BuildParams(serviceinfo.ServiceInfo{})); len(unknown) > 0 {
			logx.Warn(fmt.Sprintf("dropping unknown params %q", unknown), logFrom, true)
		}
	}
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
			&decode.ErrKeyField{
				Key: "options",
				Err: err,
			},
		)
	}
	if err := b.checkValuesParams(itemType); err != nil {
		errs = append(
			errs,
			&decode.ErrKeyField{
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
				&decode.ErrKeyField{
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
	if !slices.Contains(SupportedTypes, typeName) {
		errs = append(
			errs,
			polymorphic.ErrInvalidValue{
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
	case "bark":
		// Deprecated: scheme -> disabletls.
		if scheme := b.GetParam("scheme"); scheme != "" {
			if b.GetParam("disabletls") == "" {
				logx.Deprecated("Replacing 'notify.bark.params.scheme' with 'notify.bark.params.disabletls'")
				disableTLS := "no"
				if strings.EqualFold(scheme, "http") {
					disableTLS = "yes"
				}
				b.setParam("disabletls", disableTLS)
			}
			b.deleteParam("scheme")
			changed = true
		}
	case "generic":
		// Deprecated: requestmethod -> method.
		if v := b.GetParam("requestmethod"); v != "" {
			if b.GetParam("method") == "" {
				logx.Deprecated("Renaming 'notify.generic.params.requestmethod' to 'notify.generic.params.method'")
				b.setParam("method", v)
			}
			b.deleteParam("requestmethod")
			changed = true
		}
		// Deprecated: custom_headers -> headers.
		if customHeaders := b.getURLField("custom_headers"); customHeaders != "" {
			if headers := b.getURLField("headers"); headers == "" {
				logx.Deprecated("Renaming 'notify.generic.url_fields.custom_headers' to 'notify.generic.url_fields.headers'")
				b.setURLField("headers", customHeaders)
			}
			b.deleteURLField("custom_headers")
			changed = true
		}
	case "discord":
		// Deprecated: threadid -> thread_id.
		if v := b.GetParam("threadid"); v != "" {
			if b.GetParam("thread_id") == "" {
				logx.Deprecated("Renaming 'notify.discord.params.threadid' to 'notify.discord.params.thread_id'")
				b.setParam("thread_id", v)
			}
			b.deleteParam("threadid")
			changed = true
		}
	case "gotify":
		// Deprecated: insecureskipverify -> skiptlsverify.
		if v := b.GetParam("insecureskipverify"); v != "" {
			if b.GetParam("skiptlsverify") == "" {
				logx.Deprecated("Renaming 'notify.gotify.params.insecureskipverify' to 'notify.gotify.params.skiptlsverify'")
				b.setParam("skiptlsverify", v)
			}
			b.deleteParam("insecureskipverify")
			changed = true
		}
		// Deprecated: Gotify sets the date itself.
		if b.GetParam("date") != "" {
			logx.Deprecated("Removing 'notify.gotify.params.date', as Gotify sets the date of each message")
			b.deleteParam("date")
			changed = true
		}
	case "ifttt":
		// Deprecated: usemessageasvalue -> messagevalue.
		if v := b.GetParam("usemessageasvalue"); v != "" {
			if b.GetParam("messagevalue") == "" {
				logx.Deprecated("Renaming 'notify.ifttt.params.usemessageasvalue' to 'notify.ifttt.params.messagevalue'")
				b.setParam("messagevalue", v)
			}
			b.deleteParam("usemessageasvalue")
			changed = true
		}
		// Deprecated: usetitleasvalue -> titlevalue.
		if v := b.GetParam("usetitleasvalue"); v != "" {
			if b.GetParam("titlevalue") == "" {
				logx.Deprecated("Renaming 'notify.ifttt.params.usetitleasvalue' to 'notify.ifttt.params.titlevalue'")
				b.setParam("titlevalue", v)
			}
			b.deleteParam("usetitleasvalue")
			changed = true
		}
	case "matrix":
		// Deprecated: username -> user.
		if v := b.getURLField("username"); v != "" {
			if b.getURLField("user") == "" {
				logx.Deprecated("Renaming 'notify.matrix.url_fields.username' to 'notify.matrix.url_fields.user'")
				b.setURLField("user", v)
			}
			b.deleteURLField("username")
			changed = true
		}
		// Remove #'s in channel aliases.
		if rooms := b.GetParam("rooms"); strings.Contains(rooms, "#") {
			b.setParam(
				"rooms",
				strings.ReplaceAll(rooms, "#", ""),
			)
			changed = true
		}
	case "mqtt":
		// QoS, map Shoutrrr's numeric aliases to the names.
		if qos := b.GetParam("qos"); qos != "" {
			if i, err := strconv.Atoi(qos); err == nil && i >= 0 && i < len(mqttParamQoS) {
				b.setParam("qos", mqttParamQoS[i])
				changed = true
			}
		}
	case "mattermost":
		// Channel, strip leading /
		if channel, ok := strings.CutPrefix(b.getURLField("channel"), "/"); ok {
			b.setURLField("channel", channel)
			changed = true
		}
	case "opsgenie":
		// Deprecated: 1-5 -> P1-P5.
		if priority := b.GetParam("priority"); len(priority) == 1 && priority >= "1" && priority <= "5" {
			logx.Deprecated(fmt.Sprintf(
				"Replacing 'notify.opsgenie.params.priority' %q with %q",
				priority, "P"+priority,
			))
			b.setParam("priority", "P"+priority)
			changed = true
		}
	case "ntfy":
		// Deprecated: disabletls -> skiptlsverify.
		if disableTLS := b.GetParam("disabletls"); disableTLS != "" {
			if b.GetParam("skiptlsverify") == "" {
				logx.Deprecated("Renaming 'notify.ntfy.params.disabletls' to 'notify.ntfy.params.skiptlsverify'")
				b.setParam("skiptlsverify", disableTLS)
			}
			b.deleteParam("disabletls")
			changed = true
		}
		// Deprecated: disabletlsverification -> skiptlsverify.
		if disableTLSVerification := b.GetParam("disabletlsverification"); disableTLSVerification != "" {
			if b.GetParam("skiptlsverify") == "" {
				logx.Deprecated("Renaming 'notify.ntfy.params.disabletlsverification' to 'notify.ntfy.params.skiptlsverify'")
				b.setParam("skiptlsverify", disableTLSVerification)
			}
			b.deleteParam("disabletlsverification")
			changed = true
		}
		// Deprecated: scheme -> disabletls_.
		if scheme := b.GetParam("scheme"); scheme != "" {
			if b.GetParam("disabletls_") == "" {
				logx.Deprecated(
					"Replacing 'notify.ntfy.params.scheme' with 'notify.ntfy.params.disabletls_'" +
						" - the next major release will rename it to 'notify.ntfy.params.disabletls'",
				)
				disableTLS := "no"
				if strings.EqualFold(scheme, "http") {
					disableTLS = "yes"
				}
				b.setParam("disabletls_", disableTLS)
			}
			b.deleteParam("scheme")
			changed = true
		}
	case "signal":
		// Deprecated: apikey -> token.
		if v := b.GetParam("apikey"); v != "" {
			if b.GetParam("token") == "" {
				logx.Deprecated("Renaming 'notify.signal.params.apikey' to 'notify.signal.params.token'")
				b.setParam("token", v)
			}
			b.deleteParam("apikey")
			changed = true
		}
	case "slack":
		// Deprecated: threadts -> thread_ts.
		if v := b.GetParam("threadts"); v != "" {
			if b.GetParam("thread_ts") == "" {
				logx.Deprecated("Renaming 'notify.slack.params.threadts' to 'notify.slack.params.thread_ts'")
				b.setParam("thread_ts", v)
			}
			b.deleteParam("threadts")
			changed = true
		}
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
			b.deleteParam("skiptlsverification")
			changed = true
		}
		// Timeout, treat integers as seconds by default.
		if timeout := b.GetParam("timeout"); timeout != "" {
			if _, err := strconv.Atoi(timeout); err == nil {
				b.setParam("timeout", timeout+"s")
				changed = true
			}
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
	if !slices.Contains(SupportedTypes, sType) {
		sTypeWithoutID := util.FirstNonDefault(s.Type, s.Main.Type)
		if sTypeWithoutID == "" {
			return &decode.ErrField{
				Key:         "type",
				Description: "e.g. 'slack', see the docs for possible types - https://release-argus.io/docs/config/notify",
			}
		}
	}

	// Check the Type doesn't differ in the Main.
	if s.Main.Type != "" && sType != s.Main.Type {
		return &decode.ErrField{
			Key:   "type",
			Value: sType,
			Description: fmt.Sprintf(
				"must be the same as the root notify.%s.type (%s)",
				s.ID, s.Main.Type,
			),
		}
	}

	// Invalid/Unknown type.
	if !slices.Contains(SupportedTypes, sType) {
		return polymorphic.ErrInvalidValue{
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
				&decode.ErrField{
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
					&decode.ErrField{
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
						&decode.ErrField{
							Key:         "max_tries",
							Value:       maxTriesStr,
							Description: "must be positive",
						},
					)
					// Positive.
				} else {
					errs = append(
						errs,
						&decode.ErrField{
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
					&decode.ErrField{
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
			&decode.ErrField{
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
				&decode.ErrField{
					Key: "devicekey",
				},
			)
		}
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key: "host",
				},
			)
		}
	case "discord":
		// discord://token@webhookid
		if s.GetURLField("token") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "token",
					Description: "e.g. 'https://discord.com/api/webhooks/[ 975870285909737583 <- webhookid ]/[ QEdyk-Qi5AiMXoZdxQFpWNcwEfmz5oOm_1Rni9DnjQAUap4zWcurM4IquamVrDIyNgBG <- TOKEN ]'",
				},
			)
		}
		if s.GetURLField("webhookid") == "" {
			errs = append(
				errs,
				&decode.ErrField{
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
				&decode.ErrField{
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
				&decode.ErrField{
					Key:         "host",
					Description: "e.g. 'gotify.example.com'",
				},
			)
		}
		if s.GetURLField("token") == "" {
			errs = append(
				errs,
				&decode.ErrField{
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
				&decode.ErrField{
					Key:         "raw",
					Description: "e.g. 'https://chat.googleapis.com/v1/spaces/FOO/messages?key=bar&token=baz'",
				},
			)
		}
	case "homeassistant":
		// homeassistant://token@host[:port][/path]
		if s.GetURLField("token") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "token",
					Description: "long-lived access token",
				},
			)
		}
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "host",
					Description: "e.g. 'homeassistant.example.com'",
				},
			)
		}
	case "ifttt":
		// ifttt://webhookid
		if s.GetURLField("webhookid") == "" {
			errs = append(
				errs,
				&decode.ErrField{
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
				&decode.ErrField{
					Key:         "apikey",
					Description: "e.g. 'f8eae56127864015b0d2f4d8db6ff53f'",
				},
			)
		}
	case "lark":
		// lark://host/token
		if s.GetURLField("token") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "token",
					Description: "e.g. 'https://open.larksuite.com/open-apis/bot/v2/hook/[ TOKEN ]'",
				},
			)
		}
	case "matrix":
		// matrix://user:password@host[:port]/[?rooms=!roomID1,roomAlias2][&disabletls=yes]
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "host",
					Description: "e.g. 'matrix.example.com'",
				},
			)
		}
		if s.GetURLField("password") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "password",
					Description: "e.g. 'pass123' (with user) OR 'access_token' (no user)",
				},
			)
		}
	case "mattermost":
		// mattermost://[username@]host[:port]/token[/channel]
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "host",
					Description: "e.g. 'mattermost.example.com'",
				},
			)
		}
		if s.GetURLField("token") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "token",
					Description: "e.g. 'Aod9Cb0zXCeOrnD'",
				},
			)
		}
	case "mqtt":
		// mqtt://[username[:password]@]host[:port]/topic
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "host",
					Description: "e.g. 'mqtt.example.com'",
				},
			)
		}
		if s.GetURLField("topic") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "topic",
					Description: "e.g. 'home/argus/releases'",
				},
			)
		}
	case "ntfy":
		// ntfy://[username]:[password]@[host][:port]/topic
		if s.GetURLField("topic") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key: "topic",
				},
			)
		}
	case "notifiarr":
		// notifiarr://apikey[?name=X&channel=X&thumbnail=X&image=X&color=X]
		if s.GetURLField("apikey") == "" {
			errs = append(
				errs,
				&decode.ErrField{
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
				&decode.ErrField{
					Key:         "apikey",
					Description: "e.g. 'xxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx'",
				},
			)
		}
	case "pagerduty":
		// pagerduty://[host[:port]]/integrationkey
		if s.GetURLField("integrationkey") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "integrationkey",
					Description: "32 hex characters, from the Events API v2 integration",
				},
			)
		}
	case "pushbullet":
		// pushbullet://token/targets
		if s.GetURLField("token") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "token",
					Description: "e.g. 'o.5NfxzU9yH4xBZlEXZArRtyUB4S4Ua8Hd'",
				},
			)
		}
		if s.GetURLField("targets") == "" {
			errs = append(
				errs,
				&decode.ErrField{
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
				&decode.ErrField{
					Key:         "token",
					Description: "e.g. 'aayohdg8gqjj3ssszuqwwmuipt5gcd'",
				},
			)
		}
		if s.GetURLField("user") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "user",
					Description: "e.g. '2QypyiVSnURsw72cpnXCuVAQMJpKKY'",
				},
			)
		}
	case "rocketchat":
		// rocketchat://[username@]host[:port]/tokenA/tokenB[/channel]
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "host",
					Description: "e.g. 'rocket-chat.example.com'",
				},
			)
		}
		if s.GetURLField("tokena") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "tokena",
					Description: "e.g. '8eGdRzc9r4YYNyvge'",
				},
			)
		}
		if s.GetURLField("tokenb") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "tokenb",
					Description: "e.g. '2XYQcX9NBwJBKfQnphpebPcnXZcPEi32Nt4NKJfrnbhsbRfX'",
				},
			)
		}
	case "signal":
		// signal://[user:password@]host[:port]/source/recipient1[/recipient2...]
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "host",
					Description: "e.g. 'signal-api.example.com'",
				},
			)
		}
		if s.GetURLField("source") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "source",
					Description: "e.g. '+1234567890'",
				},
			)
		}
		if s.GetURLField("recipients") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "recipients",
					Description: "e.g. '+1234567890,group.abc'",
				},
			)
		}
	case "signalgrid":
		// signalgrid://clientkey@channel
		if s.GetURLField("clientkey") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key: "clientkey",
				},
			)
		}
		if s.GetURLField("channel") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key: "channel",
				},
			)
		}
	case "slack":
		// slack://token:token@channel
		if s.GetURLField("token") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "token",
					Description: "e.g. '123456789012-1234567890123-4mt0t4l1YL3g1T5L4cK70k3N'",
				},
			)
		}
		if s.GetURLField("channel") == "" {
			errs = append(
				errs,
				&decode.ErrField{
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
				&decode.ErrField{
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
				&decode.ErrField{
					Key:         "token",
					Description: "e.g. '110201543:AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsaw'",
				},
			)
		}
	case "twilio":
		// twilio://accountsid:authtoken@fromnumber/tonumber1[/tonumber2...]
		if s.GetURLField("accountsid") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "accountsid",
					Description: "e.g. 'ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx'",
				},
			)
		}
		if s.GetURLField("authtoken") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key: "authtoken",
				},
			)
		}
		if s.GetURLField("fromnumber") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "fromnumber",
					Description: "e.g. '+1234567890'",
				},
			)
		}
		if s.GetURLField("tonumbers") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "tonumbers",
					Description: "e.g. '+1234567890,+1987654321'",
				},
			)
		}
	case "wecom":
		// wecom://key
		if s.GetURLField("key") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "key",
					Description: "the key=... value of the group bot webhook URL",
				},
			)
		}
	case "xmpp":
		// xmpp://user:password@host[:port]/
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "host",
					Description: "e.g. 'xmpp.example.com'",
				},
			)
		}
		if s.GetURLField("user") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "user",
					Description: "e.g. 'argus' or 'argus@example.com'",
				},
			)
		}
		if s.GetURLField("password") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key: "password",
				},
			)
		}
	case "zulip":
		// zulip://botMail:botKey@host:port
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "host",
					Description: "e.g. 'zulipchat.example.com'",
				},
			)
		}
		if s.GetURLField("botmail") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "botmail",
					Description: "e.g. 'my-bot@zulipchat.com'",
				},
			)
		}
		if s.GetURLField("botkey") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "botkey",
					Description: "e.g. 'correcthorsebatterystable'",
				},
			)
		}
	case "generic":
		// generic://[username[:password]@]host[:port][/path]
		if s.GetURLField("host") == "" {
			errs = append(
				errs,
				&decode.ErrField{
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
						&decode.ErrField{
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
	case "homeassistant":
		// homeassistant://token@host[:port][/path]/?service=domain.action
		if service := s.GetParam("service"); service != "" &&
			!util.RegexCheck(homeAssistantParamService, service) {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "service",
					Value:       service,
					Description: "e.g. 'notify.mobile_app_phone'",
				},
			)
		}
	case "ifttt":
		// ifttt://webhookid/?events=event1[,event2,...]&value1=value1&value2=value2&value3=value3
		if s.GetParam("events") == "" {
			errs = append(
				errs,
				&decode.ErrField{
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
				&decode.ErrField{
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
				&decode.ErrField{
					Key:         "fromaddress",
					Description: "e.g. 'service@gmail.com'",
				},
			)
		}
		if s.GetParam("toaddresses") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "toaddresses",
					Description: "e.g. 'name@gmail.com'",
				},
			)
		}
	case "teams":
		// teams://?host=fullPowerAutomateURL
		// don't fail on old, no-longer functional instances.
		if s.GetParam("host") == "" && s.GetURLField("group") == "" {
			errs = append(
				errs,
				&decode.ErrField{
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
				&decode.ErrField{
					Key:         "chats",
					Description: "e.g. '@channelName' or 'chatID'",
				},
			)
		}
	case "xmpp":
		// xmpp://user:password@host[:port]/?to=X[&rooms=X]
		if s.GetParam("to") == "" && s.GetParam("rooms") == "" {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "to",
					Description: "'to' and/or 'rooms' required, e.g. 'user@example.com'",
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

	// Durations, must be positive.
	for _, key := range durationParams[itemType] {
		if value := b.GetParam(key); value != "" {
			if d, err := time.ParseDuration(value); err != nil || d <= 0 {
				errs = append(
					errs,
					&decode.ErrField{
						Key:         key,
						Value:       value,
						Description: "use a positive 'AhBmCs' duration, e.g. '10s'",
					},
				)
			}
		}
	}

	for _, key := range boolParams[itemType] {
		value := b.GetParam(key)
		if value == "" {
			continue
		}
		if _, ok := format.ParseBool(util.EvalEnvVars(value), false); !ok {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         key,
					Value:       value,
					Description: "use 'yes' or 'no'",
				},
			)
		}
	}

	// Params.*
	for key, value := range b.Params {
		if !util.CheckTemplate(value) {
			errs = append(
				errs,
				&decode.ErrField{
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
		if err := b.validateParamSelect("sound", barkParamSound); err != nil {
			errs = append(errs, err)
		}
	case "generic":
		if err := b.validateParamSelect("method", genericParamMethod); err != nil {
			errs = append(errs, err)
		}
	case "mqtt":
		if err := b.validateParamSelect("qos", mqttParamQoS); err != nil {
			errs = append(errs, err)
		}
	case "ntfy":
		if err := b.validateParamSelect("priority", ntfyParamPriority); err != nil {
			errs = append(errs, err)
		}
	case "pagerduty":
		if err := b.validateParamSelect("action", pagerDutyParamAction); err != nil {
			errs = append(errs, err)
		}
		if err := b.validateParamSelect("severity", pagerDutyParamSeverity); err != nil {
			errs = append(errs, err)
		}
	case "signal":
		if err := b.validateParamSelect("textmode", signalParamTextMode); err != nil {
			errs = append(errs, err)
		}
	case "signalgrid":
		if err := b.validateParamSelect("type", signalgridParamType); err != nil {
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

	return polymorphic.ErrInvalidValue{
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
