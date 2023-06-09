package template_parser

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"html/template"
	"io/ioutil"
)

type TemplateParams struct {
	TemplateName string
	Vars         interface{}
}

const (
	globalTemplateDir = "templates/html"
)

func TemplateParser(params TemplateParams) ([]byte, error) {
	var body bytes.Buffer
	filePath := fmt.Sprintf("%s/%s", globalTemplateDir, params.TemplateName)
	t, err := template.ParseFiles(filePath)
	if err != nil {
		return []byte{}, errors.Errorf("error while parsing html template %s: %s", params.TemplateName, err.Error())
	}
	err = t.Execute(&body, params.Vars)
	if err != nil {
		return []byte{}, errors.Errorf("error while executing html template  %s: %s", params.TemplateName, err.Error())
	}

	return body.Bytes(), nil
}

func GetTemplates(fileNames []string) ([]string, error) {
	templates := make([]string, len(fileNames))
	for ind, fileName := range fileNames {
		data, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", globalTemplateDir, fileName))
		if err != nil {
			return []string{}, errors.Errorf("error while reading file %s: %s", fileName, err.Error())
		}
		templates[ind] = string(data)
	}
	return templates, nil
}
