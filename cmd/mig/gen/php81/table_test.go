package php81

import (
	"testing"

	"github.com/apex/log"

	"github.com/go-bridget/mig/cmd/mig/internal"
)

func TestTableRendering(t *testing.T) {
	t.Parallel()

	tab := internal.Table{
		Name:    "support_issue",
		Comment: "Support issues",
		Columns: []*internal.Column{
			{Name: "id", Type: "int", Comment: "ID"},
			{Name: "title", Type: "varchar", Comment: "Title"},
		},
	}
	output, err := RenderTable(&tab, "")
	if err != nil {
		t.Error(err)
	}

	if len(output) == 0 {
		t.Error("empty output from render")
	}

	log.Infof("%s", string(output))
}
