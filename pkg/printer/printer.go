package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
)

type Printer struct {
	Format      string
	Pretty      bool
	SkipHeaders bool
	Itemise     bool
	Dest        io.Writer
}

func NewPrinter(format string, pretty bool, out io.Writer) *Printer {
	p := Printer{
		Format: format,
		Pretty: pretty,
		Dest:   out,
	}
	return &p
}

func (p *Printer) Print(data []map[string]string, headers []string) {
	switch p.Format {
	case "json":
		if p.Itemise {
			for _, v := range data {
				p.PrintJSON(v)
			}
		} else {
			p.PrintJSON(data)
		}
	default:
		p.PrintText(data, headers)
	}
}

func (p *Printer) PrintText(data []map[string]string, headers []string) error {
	var flags uint
	if p.Pretty {
		flags = tabwriter.Debug
	}
	w := tabwriter.NewWriter(p.Dest, 6, 4, 1, ' ', flags)
	if len(headers) == 0 || p.Format == "wide" {
		for k, _ := range data[0] {
			headers = append(headers, k)
		}
	}
	if !p.SkipHeaders {
		headersStr := ""
		for _, k := range headers {
			headersStr += k + "\t"
		}
		if _, err := fmt.Fprintln(w, headersStr); err != nil {
			return err
		}
	}
	for _, m := range data {
		valuesStr := ""
		for _, h := range headers {
			v := m[h]
			valuesStr += v + "\t"
		}
		if _, err := fmt.Fprintln(w, valuesStr); err != nil {
			return err
		}
	}
	w.Flush()
	return nil
}

func (p *Printer) PrintJSON(data any) error {
	var b []byte
	var err error
	if p.Pretty {
		b, err = json.MarshalIndent(data, "", "  ")
	} else {
		b, err = json.Marshal(data)
	}
	if err != nil {
		return err
	}
	fmt.Fprintln(p.Dest, string(b))
	return nil
}
