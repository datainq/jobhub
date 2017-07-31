package notifier

import (
	"testing"
)

func TestParseTemplate(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic.")
		}
	}()
	parseTemplate("./mail_template.html", nil)
	parseTemplate("Non-existent path", nil)
}
