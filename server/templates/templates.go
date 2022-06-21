package templates

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

var templates *template.Template

func LoadTemplates(pattern string) error {
	templatesTemp, err := template.ParseGlob(pattern)
	if err != nil {
		return err
	}
	templates = templatesTemp
	log.Debug(fmt.Sprintf("Successfully loaded templates using pattern: %s", pattern))
	return nil
}

func ExecuteTemplate(responseWriter http.ResponseWriter, templateName string, data interface{}) {
	if err := templates.ExecuteTemplate(responseWriter, templateName, data); err != nil {
		log.Error(fmt.Sprintf("Could not render template '%s': %s", templateName, err.Error()))
		return
	}
}
