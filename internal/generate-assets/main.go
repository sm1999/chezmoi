package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

var (
	outputTemplate = template.Must(template.New("output").Funcs(template.FuncMap{
		"printByteSlice": printByteSlice,
	}).Parse(`// Code generated by github.com/twpayne/chezmoi/internal/generate-assets. DO NOT EDIT.
{{- if .Tags}}
// +build {{ .Tags }}
{{- end }}

package cmd

func init() {
{{- range $key, $value := .GzipedAssets }}
	gzipedAssets[{{ printf "%q" $key }}] = {{ printByteSlice $value }}
{{- end }}
}`))

	output = flag.String("o", "/dev/stdout", "output")
	tags   = flag.String("tags", "", "tags")
)

func printByteSlice(bs []byte) string {
	sb := &strings.Builder{}
	if _, err := sb.WriteString("[]byte{\n"); err != nil {
		panic(err)
	}
	const bytesPerLine = 12
	for i := 0; i < len(bs); i += bytesPerLine {
		if _, err := sb.WriteString(fmt.Sprintf("\t\t0x%02X,", bs[i])); err != nil {
			panic(err)
		}
		end := i + bytesPerLine
		if end > len(bs) {
			end = len(bs)
		}
		for _, b := range bs[i+1 : end] {
			if _, err := sb.WriteString(fmt.Sprintf(" 0x%02X,", b)); err != nil {
				panic(err)
			}
		}
		if err := sb.WriteByte('\n'); err != nil {
			panic(err)
		}
	}
	if _, err := sb.WriteString("}"); err != nil {
		panic(err)
	}
	return sb.String()
}

func run() error {
	flag.Parse()

	gzipedAssets := make(map[string][]byte)
	for _, arg := range flag.Args() {
		asset, err := ioutil.ReadFile(arg)
		if err != nil {
			return err
		}
		gzipedAsset := &bytes.Buffer{}
		w := gzip.NewWriter(gzipedAsset)
		if _, err := w.Write(asset); err != nil {
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}
		gzipedAssets[arg] = gzipedAsset.Bytes()
	}

	source := &bytes.Buffer{}
	if err := outputTemplate.Execute(source, struct {
		Tags         string
		GzipedAssets map[string][]byte
	}{
		Tags:         *tags,
		GzipedAssets: gzipedAssets,
	}); err != nil {
		return err
	}

	formattedSource, err := format.Source(source.Bytes())
	if err != nil {
		return err
	}

	return ioutil.WriteFile(*output, formattedSource, 0666)
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
