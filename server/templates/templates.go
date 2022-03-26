package templates

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/event"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

var templates *template.Template

func LoadTemplates(patterns ...string) {
	dictFunc := func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, errors.New("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, errors.New("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	}

	templates = template.New("")
	for _, pattern := range patterns {
		template.Must(templates.Funcs(template.FuncMap{"dict": dictFunc}).ParseGlob(pattern))
	}
	log.Debug(fmt.Sprintf("Templates loaded with patterns: %v", patterns))
}

func ExecuteTemplate(responseWriter http.ResponseWriter, templateName string, data interface{}) {
	if err := templates.ExecuteTemplate(responseWriter, templateName, data); err != nil {
		log.Error("Could not render template: ", err.Error())
		go event.Fatal("System Compromised", "Failed to render template: "+err.Error())
	}
}
