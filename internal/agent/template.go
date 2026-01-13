package agent

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"text/template"
)

type Filler interface {
	Fill(w io.Writer, opts ...fillerOpt) error
}

var _ Filler = templateFiller{}

//go:embed user.gotmpl
var templateDoc string

type templateFiller struct {
	GitLog         string
	Scope          string
	Diff           string
	IdealFuture    []idealSection
	IdealSeparator string

	hasSeparator bool
}

type shortenCommitMessageFiller struct {
	GitCommit string
}

type idealSection struct {
	Name  string
	Notes []string
}

type fillerOpt func(templateFiller) templateFiller

func (tf templateFiller) Fill(w io.Writer, opts ...fillerOpt) error {
	for _, opt := range opts {
		tf = opt(tf)
	}
	t, err := template.New("user").Parse(templateDoc)
	if err != nil {
		return fmt.Errorf("parse embedded template: %w", err)
	}
	var out bytes.Buffer
	if err := t.Execute(io.MultiWriter(&out, w), tf); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	if tf.hasSeparator {
		if len(bytes.Split(out.Bytes(), []byte(tf.IdealSeparator))) != 2 {
			return fmt.Errorf("inappropriate ideal separator %q", tf.IdealSeparator)
		}
	}
	return nil
}

const DefaultIdealSeparator = `
-------------------------------------------------
-------- NEXT STEPS AND WORK IN PROGRESS --------
-------------------------------------------------
`
