package funnel

import (
	"bytes"
	"reflect"
	"regexp"
	"testing"
	"text/template"
)

func TestGetLineProcessor(t *testing.T) {
	cfg := &Config{
		DirName: "something",
	}

	lp := GetLineProcessor(cfg)
	if _, ok := lp.(*NoProcessor); !ok {
		t.Errorf("Incorrect line processor returned. Expected *funnel.NoProcessor, Got %s", reflect.TypeOf(lp))
	}

	cfg = &Config{
		DirName:      "something",
		PrependValue: "prepender] ",
	}
	lp = GetLineProcessor(cfg)
	if _, ok := lp.(*SimpleLineProcessor); !ok {
		t.Errorf("Incorrect line processor returned. Expected *funnel.SimpleLineProcessor, Got %s", reflect.TypeOf(lp))
	}

	cfg = &Config{
		DirName:      "something",
		PrependValue: "prepender {{.Timestamp}}",
	}

	lp = GetLineProcessor(cfg)
	if _, ok := lp.(*TemplateLineProcessor); !ok {
		t.Errorf("Incorrect line processor returned. Expected *funnel.TemplateLineProcessor, Got %s", reflect.TypeOf(lp))
	}
}

func TestNoProcessor(t *testing.T) {
	lp := &NoProcessor{}
	line := "write this line"

	var b bytes.Buffer
	err := lp.Write(&b, line)
	if err != nil {
		t.Fatal(err)
		return
	}
	if b.String() != line {
		t.Errorf("Did not match. Expected \"%s\", Got \"%s\"", line, b.String())
	}
}

func TestSimpleProcessor(t *testing.T) {
	lp := &SimpleLineProcessor{prependStr: "prepend this"}
	line := "write this line"

	var b bytes.Buffer
	err := lp.Write(&b, line)
	if err != nil {
		t.Fatal(err)
		return
	}
	if b.String() != lp.prependStr+line {
		t.Errorf("Did not match. Expected \"%s\", Got \"%s\"", lp.prependStr+line, b.String())
	}
}

func TestTemplateProcessor(t *testing.T) {
	tpl := template.Must(template.New("line").Parse("[myapp {{.UnixTimestamp}}]- "))
	lp := &TemplateLineProcessor{template: tpl}
	line := "write this line"

	var b bytes.Buffer
	err := lp.Write(&b, line)
	if err != nil {
		t.Fatal(err)
	}
	regexStr := "[myapp [0-9]{19}]- " + line
	matched, err := regexp.MatchString(regexStr, b.String())
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Errorf("Did not match. Expected \"%s\", Got \"%s\"", regexStr, b.String())
	}

	// Testing ISO8601 timestamp
	b.Reset()
	// 2018-01-03T10:53:03+0530
	tpl = template.Must(template.New("line").Parse("[myapp {{.ISO8601Timestamp}}]- "))
	lp = &TemplateLineProcessor{template: tpl}
	line = "write this line"

	err = lp.Write(&b, line)
	if err != nil {
		t.Fatal(err)
	}
	regexStr = "[myapp [0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[Z|+|-]([0-9]{4})?]- " + line
	matched, err = regexp.MatchString(regexStr, b.String())
	if err != nil {
		t.Fatal(err)
		return
	}
	if !matched {
		t.Errorf("Did not match. Expected \"%s\", Got \"%s\"", regexStr, b.String())
	}

}

func BenchmarkNoProcessor(b *testing.B) {
	lp := &NoProcessor{}

	var buff bytes.Buffer
	for n := 0; n < b.N; n++ {
		lp.Write(&buff, string(randStringBytes(50)))
	}
}

func BenchmarkSimpleProcessor(b *testing.B) {
	lp := &SimpleLineProcessor{prependStr: "prepend this"}

	var buff bytes.Buffer
	for n := 0; n < b.N; n++ {
		lp.Write(&buff, string(randStringBytes(50)))
	}
}

func BenchmarkTemplateProcessor(b *testing.B) {
	tpl := template.Must(template.New("line").Parse("[myapp {{.UnixTimestamp}}]- "))
	lp := &TemplateLineProcessor{template: tpl}

	var buff bytes.Buffer
	for n := 0; n < b.N; n++ {
		lp.Write(&buff, string(randStringBytes(50)))
	}
}
