package variables

import (
	"fmt"
	"text/template"
)

func (c command) GetTemplateFunction() template.FuncMap {
	return template.FuncMap{
		"customVariable": func(userId string, name string) string {
			list := loadList(userId)

			if value, ok := list[name]; ok {
				return value
			}

			return fmt.Sprintf("_unknown variable: %s_", name)
		},
	}
}
