package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"text/template"

	"strings"

	"github.com/coreos/matchbox/matchbox/server"
)

const (
	contentType     = "Content-Type"
	jsonContentType = "application/json"
)

// renderJSON encodes structs to JSON, writes the response to the
// ResponseWriter, and logs encoding errors.
func (s *Server) renderJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		s.logger.Errorf("error JSON encoding: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.writeJSON(w, js)
}

// writeJSON writes the given bytes with a JSON Content-Type.
func (s *Server) writeJSON(w http.ResponseWriter, data []byte) {
	w.Header().Set(contentType, jsonContentType)
	_, err := w.Write(data)
	if err != nil {
		s.logger.Errorf("error writing to response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Server) renderTemplate(w io.Writer, data interface{}, contents ...string) (err error) {
	return s.renderTemplateWithFuncMap(w, template.FuncMap{}, data, contents...)
}

func (s *Server) renderTemplateWithFuncMap(
	w io.Writer, funcs template.FuncMap, data interface{}, contents ...string,
) (err error) {
	tmpl := template.New("").Funcs(funcs).Option("missingkey=error")
	for _, content := range contents {
		tmpl, err = tmpl.Parse(content)
		if err != nil {
			s.logger.Errorf("error parsing template: %v", err)
			return err
		}
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		s.logger.Errorf("error rendering template: %v", err)
		return err
	}
	return nil
}

func (s *Server) templateFuncMap(ctx context.Context, core server.Server) template.FuncMap {
	return template.FuncMap{
		"indent": func(spaces int, content string) string {
			lines := strings.Split(content, "\n")

			var output string
			for _, line := range lines {
				output += "\n" + strings.Repeat(" ", spaces) + line
			}

			return output
		},
		"include": func(name string, data interface{}) (string, error) {
			contents, err := core.IgnitionGet(ctx, name)
			if err != nil {
				return "", fmt.Errorf("No include template named: %s", name)
			}

			var buf bytes.Buffer
			funcs := s.templateFuncMap(ctx, core)
			err = s.renderTemplateWithFuncMap(&buf, funcs, data, contents)
			return buf.String(), err
		},
	}
}
