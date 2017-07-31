package mail

import (
	"html/template"
	"testing"
)

func TestExecuteTemplate(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic.")
		}
	}()
	//TODO(amwolff) Write better tests
	Mail.executeTemplate(
		Mail{},
		template.Must(template.ParseFiles("../cmd/default-templates/subject.html")),
		nil,
	)
}
