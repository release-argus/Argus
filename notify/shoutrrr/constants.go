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

var (
	barkNtfyParamScheme       = []string{"http", "https"}
	barkParamSound            = []string{"alarm", "anticipate", "bell", "birdsong", "bloom", "calypso", "chime", "choo", "descent", "electronic", "fanfare", "glass", "gotosleep", "healthnotification", "horn", "ladder", "mailsent", "minuet", "multiwaynotification", "newmail", "newsflash", "noir", "paymentsuccess", "shake", "sherwoodforest", "silence", "spell", "suspense", "telegraph", "tiptoes", "typewriters", "update"}
	genericParamRequestmethod = []string{"CONNECT", "DELETE", "GET", "HEAD", "OPTIONS", "POST", "PUT", "TRACE"}
	ntfyParamPriority         = []string{"min", "low", "default", "high", "max"}
	smtpParamAuth             = []string{"None", "Unknown", "Plain", "CramMD5", "OAuth2"}
	smtpParamEncryption       = []string{"Auto", "ExplicitTLS", "ImplicitTLS", "None"}
	telegramParamParsemode    = []string{"None", "HTML", "Markdown"}
)
