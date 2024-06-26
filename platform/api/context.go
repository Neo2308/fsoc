// Copyright 2022 Cisco Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"golang.org/x/term"

	"github.com/cisco-open/fsoc/config"
)

type callContext struct {
	goContext context.Context
	cfg       *config.Context
	spinner   *spinner.Spinner
}

var statusChar = map[bool]string{
	false: color.RedString("\u00d7"),   // cross mark
	true:  color.GreenString("\u2713"), // checkmark
}

func newCallContext(goContext context.Context, quiet bool) *callContext {
	// get current config context
	cfg := config.GetCurrentContext()
	if cfg == nil {
		log.Fatal(`Missing context; use "fsoc config create" to configure your context`)
		panic("unreachable") // keep golintci happy (until it recognizes apex/log fatals)
	}
	log.WithFields(log.Fields{"context": cfg.Name, "url": cfg.URL, "tenant": cfg.Tenant}).Info("Using context")

	// use background if no context was provided
	if goContext == nil {
		goContext = context.Background()
	}

	// create spinner if needed
	var spinnerObj *spinner.Spinner
	if !quiet && term.IsTerminal(int(os.Stderr.Fd())) {
		spinnerObj = spinner.New(spinner.CharSets[21], 50*time.Millisecond, spinner.WithWriterFile(os.Stderr))
	}

	// prepare call context
	callCtx := callContext{
		goContext,
		cfg,
		spinnerObj,
	}

	return &callCtx
}

func (c *callContext) startSpinner(msg string) {
	if c.spinner != nil {
		if msg != "" {
			c.spinner.Suffix = " " + msg + " in progress"
			//TODO: consider making leaving the message/status optional; or just drop it
			//c.spinner.FinalMSG = statusChar[false] + " " + msg + "\n" // jic
		} else {
			c.spinner.Suffix = ""
			c.spinner.FinalMSG = ""
		}
		_ = c.spinner.Color("cyan")
		c.spinner.Start()
	}
}

func (c *callContext) stopSpinner(ok bool) {
	c.stopSpinnerHide()
	if c.spinner != nil {
		_, msg, parsed := strings.Cut(c.spinner.FinalMSG, " ") // first blank after mark
		if parsed {
			c.spinner.FinalMSG = statusChar[ok] + " " + msg
		}
		c.spinner.Stop()
	}
}

func (c *callContext) stopSpinnerHide() {
	if c.spinner != nil {
		c.spinner.FinalMSG = ""
		c.spinner.Stop()
	}
}
