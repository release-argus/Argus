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

const (
	// homeAssistantParamService matches a Home Assistant action, as 'action' or 'domain.action'.
	homeAssistantParamService = `^[a-zA-Z0-9_]+(\.[a-zA-Z0-9_]+)?$`

	// smtpAuthUnknown is the smtp 'auth' value meaning "no method given".
	smtpAuthUnknown = "Unknown"
)

// durationParams are the params of each type that hold a duration.
var durationParams = map[string][]string{
	"pushover": {"retry", "expire"},
	"smtp":     {"timeout"},
}

var (
	barkParamSound         = []string{"alarm", "anticipate", "bell", "birdsong", "bloom", "calypso", "chime", "choo", "descent", "electronic", "fanfare", "glass", "gotosleep", "healthnotification", "horn", "ladder", "mailsent", "minuet", "multiwaynotification", "newmail", "newsflash", "noir", "paymentsuccess", "shake", "sherwoodforest", "silence", "spell", "suspense", "telegraph", "tiptoes", "typewriters", "update"}
	genericParamMethod     = []string{"CONNECT", "DELETE", "GET", "HEAD", "OPTIONS", "POST", "PUT", "TRACE"}
	mqttParamQoS           = []string{"AtMostOnce", "AtLeastOnce", "ExactlyOnce"}
	ntfyParamPriority      = []string{"min", "low", "default", "high", "max"}
	pagerDutyParamAction   = []string{"trigger", "acknowledge", "resolve"}
	pagerDutyParamSeverity = []string{"critical", "error", "warning", "info"}
	signalParamTextMode    = []string{"None", "Normal", "Styled"}
	signalgridParamType    = []string{"SUCCESS", "INFO", "WARN", "CRIT"}
	smtpParamAuth          = []string{"None", smtpAuthUnknown, "Plain", "Login", "CRAMMD5", "OAuth2"}
	smtpParamEncryption    = []string{"Auto", "ExplicitTLS", "ImplicitTLS", "None"}
	telegramParamParsemode = []string{"None", "HTML", "Markdown", "MarkdownV2"}
	zulipParamType         = []string{"channel", "direct"}
)
