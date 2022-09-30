package bucket

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// checkWidthsCoherence checks if headers and data are both sets, that the
// width is consistent between them.
func (r Results) checkWidthsCoherence() bool {
	headerWidth := len(r.headers)

	// check if every width in data is the same
	var dataWidth int
	for _, d := range r.data {
		if len(r.data[0]) != len(d) {
			return false
		}
	}
	if len(r.data) > 0 {
		dataWidth = len(r.data[0])
	}

	// one is not set
	if headerWidth == 0 || dataWidth == 0 {
		return true
	}
	return headerWidth == dataWidth
}

func nrColumnsToMaxWidth(termWidth int, n int) int {
	return (termWidth - 4 - ((n - 1) * 3)) / n
}

func (r Results) formatTable(outputWidth int) string {
	if len(r.data) == 0 || len(r.headers) == 0 {
		return ""
	}

	t := table.NewWriter()
	// Unfortunately this library does not propose to enfore a width on the
	// global table and I haven't found if you can specify of a tunning for
	// every column so I'm overshooting here.
	maxWidth := nrColumnsToMaxWidth(outputWidth, len(r.headers))
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AlignHeader: text.AlignCenter, WidthMaxEnforcer: text.WrapSoft, WidthMax: maxWidth},
		{Number: 2, AlignHeader: text.AlignCenter, WidthMaxEnforcer: text.WrapSoft, WidthMax: maxWidth},
		{Number: 3, AlignHeader: text.AlignCenter, WidthMaxEnforcer: text.WrapSoft, WidthMax: maxWidth},
		{Number: 4, AlignHeader: text.AlignCenter, WidthMaxEnforcer: text.WrapSoft, WidthMax: maxWidth},
		{Number: 5, AlignHeader: text.AlignCenter, WidthMaxEnforcer: text.WrapSoft, WidthMax: maxWidth},
		{Number: 6, AlignHeader: text.AlignCenter, WidthMaxEnforcer: text.WrapSoft, WidthMax: maxWidth},
	})
	headers := make(table.Row, len(r.headers))
	for i := range r.headers {
		headers[i] = r.headers[i]
	}
	t.AppendHeader(headers)
	for _, content := range r.data {
		row := make(table.Row, len(content))
		copy(row, content)
		t.AppendRow(row)
	}

	return t.Render()
}

func (r Results) Human(opts ResultsOpts) string {
	var output strings.Builder

	if opts.ShowName == nil || *opts.ShowName {
		output.WriteString(fmt.Sprintf("### %s ###\n", strings.ToUpper(r.bucketName)))
	}
	if len(r.comments) != 0 {
		if opts.ShowComments == nil || *opts.ShowComments {
			output.WriteString("Comments:\n")
			for _, comment := range r.comments {
				output.WriteString(fmt.Sprintf("- %s\n", comment))
			}
		}
	}
	if opts.ShowData == nil || *opts.ShowData {
		// the table will be written only if the headers and the data has been
		// filled
		table := r.formatTable(opts.OutputWidth)
		if table != "" {
			output.WriteString(table)
			output.WriteString("\n")
		}
	}
	return output.String()
}

func (r Results) JSON(opts ResultsOpts) (string, error) {
	if !r.checkWidthsCoherence() {
		return "", fmt.Errorf("cannot output JSON for bucket %q, inconsistence between width of headers and data", r.bucketName)
	}

	type jsonOutput struct {
		Bucket   string                   `json:"bucket"`
		Comments []string                 `json:"comments,omitempty"`
		Results  []map[string]interface{} `json:"results,omitempty"`
		Result   map[string]interface{}   `json:"result,omitempty"`
	}

	dataMap := make([]map[string]interface{}, 0)

	if len(r.data) != 0 && len(r.headers) != 0 {
		for _, row := range r.data {
			rowMap := make(map[string]interface{})
			for headerNr, header := range r.headers {
				rowMap[header] = row[headerNr]
			}
			dataMap = append(dataMap, rowMap)
		}
	}

	var b []byte
	var err error
	// if hide name and comments, directly output an array of results
	if (opts.ShowName != nil && !*opts.ShowName) && (opts.ShowComments != nil && !*opts.ShowComments) {
		if len(dataMap) == 1 {
			// flatten it, the result is not iterable
			b, err = json.Marshal(dataMap[0])
		} else {
			b, err = json.Marshal(dataMap)
		}
	} else {
		o := jsonOutput{}
		// TODO maybe add omitempty
		if opts.ShowName == nil || *opts.ShowName {
			o.Bucket = r.bucketName
		}
		if opts.ShowComments == nil || *opts.ShowComments {
			o.Comments = r.comments
		}
		if opts.ShowData == nil || *opts.ShowData {
			if len(dataMap) == 1 {
				// flatten it, the result is not iterable
				o.Result = dataMap[0]
			} else {
				o.Results = dataMap
			}
		}

		b, err = json.Marshal(o)
	}
	if err != nil {
		panic(err)
	}

	return string(b), nil
}
