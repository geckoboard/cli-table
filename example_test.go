package table_test

import (
	"os"

	"github.com/geckoboard/cli-table"
)

func Example() {
	t := table.New(3)

	// Add some padding
	t.SetPadding(1)

	// Set headers
	t.SetHeader(0, "left", table.AlignLeft)
	t.SetHeader(1, "center", table.AlignCenter)
	t.SetHeader(2, "right", table.AlignRight)

	// Optionally define header groups
	t.AddHeaderGroup(2, "left group", table.AlignCenter)
	t.AddHeaderGroup(1, "right group", table.AlignRight)

	// Append single row
	t.Append([]string{"1", "2", "3"})

	// Or append a bunch of rows
	t.Append(
		[]string{"1", "2", "3"},
		[]string{"four", "five", "six"},
	)

	// Render table
	t.Write(os.Stdout, table.PreserveAnsi)

	// output:
	// +---------------+---------------+
	// |  left group   |   right group |
	// +---------------+---------------+
	// | left | center |         right |
	// +------+--------+---------------+
	// | 1    |   2    |             3 |
	// | 1    |   2    |             3 |
	// | four |  five  |           six |
	// +------+--------+---------------+
}
