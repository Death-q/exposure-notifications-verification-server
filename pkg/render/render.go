// Copyright 2020 Google LLC
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

// Package render defines rendering functionality.
package render

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/exposure-notifications-server/pkg/logging"

	"go.uber.org/zap"
)

// allowedResponseCodes are the list of allowed response codes. This is
// primarily here to catch if someone, in the future, accidentially includes a
// bad status code.
var allowedResponseCodes = map[int]struct{}{
	200: {},
	400: {},
	401: {},
	404: {},
	405: {},
	412: {},
	413: {},
	429: {},
	500: {},
}

// Renderer is responsible for rendering various content and templates like HTML
// and JSON responses.
type Renderer struct {
	// debug indicates templates should be reloaded on each invocation and real
	// error responses should be rendered. Do not enable in production.
	debug bool

	// logger is the log writer.
	logger *zap.SugaredLogger

	// rendererPool is a pool of *bytes.Buffer, used as a rendering buffer to
	// prevent partial responses being sent to clients.
	rendererPool *sync.Pool

	// templates is the actually collection of templates. templatesLoader is a
	// function for (re)loading templates. templatesLock is a mutex to prevent
	// concurrent modification of the templates field.
	templates     *template.Template
	templatesRoot string
}

// New creates a new renderer with the given details.
func New(ctx context.Context, root string, debug bool) (*Renderer, error) {
	logger := logging.FromContext(ctx)

	r := &Renderer{
		debug:  debug,
		logger: logger,
		rendererPool: &sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, 1024))
			},
		},
		templatesRoot: root,
	}

	// Load initial templates
	if err := r.loadTemplates(); err != nil {
		return nil, err
	}

	return r, nil
}

// loadTemplates loads or reloads all templates.
func (r *Renderer) loadTemplates() error {
	if r.templatesRoot == "" {
		return nil
	}

	tmpl := template.New("").
		Option("missingkey=zero").
		Funcs(templateFuncs())
	if err := loadTemplates(tmpl, r.templatesRoot); err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	r.templates = tmpl
	return nil
}

func loadTemplates(tmpl *template.Template, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".html") {
			return nil
		}

		if _, err := tmpl.ParseFiles(path); err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		return nil
	})
}

// safeHTML un-escapes known safe html.
func safeHTML(s string) template.HTML {
	return template.HTML(s)
}

func templateFuncs() template.FuncMap {
	return map[string]interface{}{
		"joinStrings":    strings.Join,
		"trimSpace":      strings.TrimSpace,
		"stringContains": strings.Contains,
		"toLower":        strings.ToLower,
		"toUpper":        strings.ToUpper,
		"safeHTML":       safeHTML,
	}
}

// AllowedResponseCode returns true if the code is a permitted response code,
// false otherwise.
func (r *Renderer) AllowedResponseCode(code int) bool {
	_, ok := allowedResponseCodes[code]
	return ok
}
