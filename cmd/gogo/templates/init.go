package templates

import (
	"regexp"
	"strings"
	"text/template"
)

var (
	box *template.Template
)

func Box() *template.Template {
	return box
}

func init() {
	// register templates
	box = template.New("gogo").Funcs(template.FuncMap{
		"lowercase": strings.ToLower,
	})

	// commons
	template.Must(box.New("gitignore").Parse(gitIgnoreTemplate))
	template.Must(box.New("env.sh").Parse(envTemplate))
	template.Must(box.New("go.mod").Parse(modTemplate))
	template.Must(box.New("readme").Parse(
		format(readmeTemplate),
	))
	template.Must(box.New("makefile").Parse(
		format(makefileTemplate),
	))
	template.Must(box.New("main.go").Parse(
		format(mainTemplate),
	))
	template.Must(box.New("errors.go").Parse(
		format(errorsTemplate),
	))
	// controllers
	template.Must(box.New("application.go").Parse(
		format(applicationTemplate),
	))
	template.Must(box.New("application_testing.go").Parse(
		format(applicationTestingTemplate),
	))
	template.Must(box.New("application_config.go").Parse(
		format(applicationConfigTemplate),
	))
	template.Must(box.New("application_config_test.go").Parse(
		format(applicationConfigTestTemplate),
	))
	template.Must(box.New("application_config.json").Parse(
		format(applicationConfigJSONTemplate),
	))
	template.Must(box.New("getting_start.go").Parse(
		format(gettingStartTemplate),
	))
	template.Must(box.New("getting_start_test.go").Parse(
		format(gettingStartTestTemplate),
	))
	// filters
	template.Must(box.New("filter_testing.go").Parse(
		format(filterTestingTemplate),
	))
	template.Must(box.New("filter_recovery.go").Parse(
		format(filterRecoveryTemplate),
	))
	template.Must(box.New("filter_recovery_test.go").Parse(
		format(filterRecoveryTestTemplate),
	))
	// models
	template.Must(box.New("model.go").Parse(
		format(modelTemplate),
	))
	template.Must(box.New("model_test.go").Parse(
		format(modelTestTemplate),
	))
	// templates
	template.Must(box.New("template_controller").Parse(
		format(componentControllerTemplate),
	))
	template.Must(box.New("template_controller_test").Parse(
		format(componentControllerTestTemplate),
	))
	template.Must(box.New("template_filter").Parse(
		format(componentFilterTemplate),
	))
	template.Must(box.New("template_filter_test").Parse(
		format(componentFilterTestTemplate),
	))
	template.Must(box.New("template_model").Parse(
		format(componentModelTemplate),
	))
	template.Must(box.New("template_model_test").Parse(
		format(componentModelTestTemplate),
	))
}

var (
	langOpenTag  = regexp.MustCompile(`<(\w+)>`)
	langCloseTag = regexp.MustCompile(`</(\w+)>`)
)

func format(tpl string) string {
	tpl = langOpenTag.ReplaceAllStringFunc(tpl, func(tag string) string {
		return "```" + strings.Trim(tag, "<>")
	})
	tpl = langCloseTag.ReplaceAllStringFunc(tpl, func(tag string) string {
		return "```"
	})

	return tpl
}
