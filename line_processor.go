package funnel

import (
	"bytes"
	"fmt"
	"io"
	"text/template"
	"time"
)

// LineProcessor interface is passed down to the consumer
// which just calls the write function
type LineProcessor interface {
	Write(io.Writer, string) error
}

// GetLineProcessor function returns the particular processor depending
// on the config.
func GetLineProcessor(cfg *Config) LineProcessor {
	// If no prepend value is needed, return no processor
	if cfg.PrependValue == "" {
		return &NoProcessor{}
	}

	t := template.Must(template.New("line").Parse(cfg.PrependValue))

	// Check if there is a template action in the string
	// If yes, return the template processor
	if len(t.Tree.Root.Nodes) > 1 {
		return &TemplateLineProcessor{template: t}
	}
	return &SimpleLineProcessor{prependStr: cfg.PrependValue}
}

// NoProcessor is used when there is no prepend value.
// It just prints the line without any other action
type NoProcessor struct {
}

func (*NoProcessor) Write(w io.Writer, line string) error {
	_, err := fmt.Fprint(w, line)
	return err
}

// SimpleLineProcessor is used when the prependValue is only a simple string
// It just concatenates the string with the line and prints it
type SimpleLineProcessor struct {
	prependStr string
}

func (lp *SimpleLineProcessor) Write(w io.Writer, line string) error {
	_, err := fmt.Fprint(w, lp.prependStr+line)
	return err
}

// TemplateLineProcessor is used when there is a template action in the prependValue
// It parses the prependValue and store the template. Then for every write call,
// it executes the template and writes it
type TemplateLineProcessor struct {
	template *template.Template
}

type templateData struct {
	RFC822Timestamp  string
	ISO8601Timestamp string
	UnixTimestamp    int64
}

func (lp *TemplateLineProcessor) Write(w io.Writer, line string) error {
	// Populating the template data struct
	t := time.Now()
	data := templateData{t.Format(time.RFC822), t.Format("2006-01-02T15:04:05Z0700"), t.UnixNano()}
	var b bytes.Buffer
	if err := lp.template.Execute(&b, data); err != nil {
		return err
	}
	if _, err := fmt.Fprint(&b, line); err != nil {
		return err
	}
	// Writing the buffer to io.Writer
	_, err := b.WriteTo(w)
	return err
}
