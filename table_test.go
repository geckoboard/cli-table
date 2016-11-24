package table

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func TestTableWriterErrors(t *testing.T) {
	var err error
	table := New(3)

	err = table.SetHeader(-1, "foo", AlignCenter)
	if err == nil {
		t.Error("expected SeHeader with col arg -1 to fail")
	}
	err = table.SetHeader(3, "foo", AlignCenter)
	if err == nil {
		t.Error("expected SeHeader with col arg 3 to fail")
	}
	err = table.SetHeader(2, "foo", AlignCenter)
	if err != nil {
		t.Errorf("expected SetHeader with col arg 2 to succeed; got %v", err)
	}

	specs := [][]string{
		[]string{""},             // less columns than required
		[]string{"", "", "", ""}, // more columns than required
	}
	for specIndex, spec := range specs {
		err = table.Append(spec)
		if err == nil {
			t.Errorf("[spec %d] expected call to Append with %d column(s) (table has 3 columns) to fail", specIndex, len(spec))
		}
	}

	err = table.AddHeaderGroup(4, "g0", AlignLeft)
	if err == nil {
		t.Error("expected call to AddHeaderGroup with a colspan exceeding the available columns to fail")
	}
}

func TestTableWriter(t *testing.T) {
	testHeaders := []string{"1", "2", "3"}
	testRows := [][]string{
		[]string{"1234567890", "1234567890", "123456789"},
		[]string{"123", "123", "123"},
	}

	specs := []struct {
		alignments []Alignment
		padding    int
		charFilter CharacterFilter
		rows       [][]string
		expOutput  string
	}{
		{
			alignments: []Alignment{AlignLeft, AlignCenter, AlignRight},
			padding:    0,
			rows:       testRows,
			expOutput: `+----------+----------+---------+
|1         |    2     |        3|
+----------+----------+---------+
|1234567890|1234567890|123456789|
|123       |   123    |      123|
+----------+----------+---------+
`,
		},
		{
			alignments: []Alignment{AlignLeft, AlignCenter, AlignRight},
			padding:    -100, // should be forced to 0
			rows:       testRows,
			expOutput: `+----------+----------+---------+
|1         |    2     |        3|
+----------+----------+---------+
|1234567890|1234567890|123456789|
|123       |   123    |      123|
+----------+----------+---------+
`,
		},
		{
			alignments: []Alignment{AlignLeft, AlignCenter, AlignRight},
			padding:    1,
			rows:       append(testRows, []string{"\033[33m1234567890\033[0m", "", ""}),
			// Test Ansi support; Ansi characters should not affect column width measurements
			expOutput: `+------------+------------+-----------+
| 1          |     2      |         3 |
+------------+------------+-----------+
| 1234567890 | 1234567890 | 123456789 |
| 123        |    123     |       123 |
| ` + "\033[33m1234567890\033[0m" + ` |            |           |
+------------+------------+-----------+
`,
		},
		{
			alignments: []Alignment{AlignLeft, AlignCenter, AlignRight},
			padding:    1,
			rows:       append(testRows, []string{"\033[33m1234567890\033[0m", "", ""}),
			charFilter: StripAnsi,
			expOutput: `+------------+------------+-----------+
| 1          |     2      |         3 |
+------------+------------+-----------+
| 1234567890 | 1234567890 | 123456789 |
| 123        |    123     |       123 |
| 1234567890 |            |           |
+------------+------------+-----------+
`,
		},
		{
			alignments: []Alignment{AlignLeft, AlignCenter, AlignRight},
			padding:    1,
			rows:       [][]string{},
			expOutput: `+---+---+---+
| 1 | 2 | 3 |
+---+---+---+
`,
		},
	}

	var buf bytes.Buffer
	for specIndex, spec := range specs {
		table := New(len(testHeaders))
		for hIndex, header := range testHeaders {
			table.SetHeader(hIndex, header, spec.alignments[hIndex])
		}
		if spec.padding != 0 {
			table.SetPadding(spec.padding)
		}
		table.Append(spec.rows...)

		buf.Reset()
		table.Write(&buf, spec.charFilter)
		tableOutput := buf.String()

		if tableOutput != spec.expOutput {
			t.Errorf(
				"[spec %d]\n expected output to be: \n%s\n got:\n%s\n",
				specIndex,
				indent(spec.expOutput, 2),
				indent(tableOutput, 2),
			)
		}
	}
}

func TestHeaderGroups(t *testing.T) {
	specs := []struct {
		headers    []string
		groups     []headerGroup
		alignments []Alignment
		expOutput  string
	}{
		{
			headers: []string{"1", "2", "3"},
			groups: []headerGroup{
				headerGroup{header: "group 0", colSpan: 2},
				headerGroup{header: "1", colSpan: 1},
			},
			alignments: []Alignment{AlignLeft, AlignCenter, AlignRight},
			expOutput: `+-------+-+
|group 0|1|
+-------+-+
|1|  2  |3|
+-+-----+-+
`,
		},
		{
			headers: []string{"long cell header", "2", "3"},
			groups: []headerGroup{
				headerGroup{header: "group 0", colSpan: 2},
				headerGroup{header: "1", colSpan: 1},
			},
			alignments: []Alignment{AlignLeft, AlignCenter, AlignRight},
			expOutput: `+------------------+-+
|     group 0      |1|
+------------------+-+
|long cell header|2|3|
+----------------+-+-+
`,
		},
	}

	var buf bytes.Buffer
	for specIndex, spec := range specs {
		table := New(len(spec.headers))
		for hIndex, header := range spec.headers {
			table.SetHeader(hIndex, header, spec.alignments[hIndex])
		}
		for _, hg := range spec.groups {
			table.AddHeaderGroup(hg.colSpan, hg.header, AlignCenter)
		}

		buf.Reset()
		table.Write(&buf, StripAnsi)
		tableOutput := buf.String()

		if tableOutput != spec.expOutput {
			t.Errorf(
				"[spec %d]\n expected output to be: \n%s\n got:\n%s\n",
				specIndex,
				indent(spec.expOutput, 2),
				indent(tableOutput, 2),
			)
		}
	}
}

func TestUTF8Handling(t *testing.T) {
	expOutput := `+-------+-------+
| 1     | 2     |
+-------+-------+
| ▲ 123 | 123 ▾ |
| > 123 | 123 < |
+-------+-------+
`

	var buf bytes.Buffer
	table := New(2)
	table.SetHeader(0, "1", AlignLeft)
	table.SetHeader(1, "2", AlignLeft)
	table.SetPadding(1)
	table.Append(
		[]string{"▲ 123", "123 ▾"},
		[]string{"> 123", "123 <"},
	)

	buf.Reset()
	table.Write(&buf, PreserveAnsi)
	tableOutput := buf.String()

	if tableOutput != expOutput {
		t.Errorf(
			"expected output to be: \n%s\n got:\n%s\n",
			indent(expOutput, 2),
			indent(tableOutput, 2),
		)
	}
}

func indent(s string, pad int) string {
	return regexp.MustCompile("(?m)^").ReplaceAllString(s, strings.Repeat(" ", pad))
}
