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

func LoadTemplates(pattern string) {
	templates = template.Must(template.ParseGlob(pattern))
	log.Debug(fmt.Sprintf("Templates loaded: %s", pattern))
}

func ExecuteTemplate(responseWriter http.ResponseWriter, templateName string, data interface{}) {
	templates.ExecuteTemplate(responseWriter, templateName, data)
}
