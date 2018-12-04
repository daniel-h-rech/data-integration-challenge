package data

import (
	"reflect"
	"strings"
	"testing"
)

func TestReadCSVStream(t *testing.T) {

	tests := []struct {
		input          string
		expectedOutput [][]string
	}{
		{
			"name;addressZip",
			[][]string{{}}},
		{
			"name;addressZip\ntola sales group;78229",
			[][]string{{"tola sales group", "78229"}}},
		{
			"name;addressZip\ntola sales group;78229\nfoundation corrections inc;94002",
			[][]string{{"tola sales group", "78229"}, {"foundation corrections inc", "94002"}},
		},
	}

	for _, test := range tests {

		j := 0

		err := readCSVStream(strings.NewReader(test.input), func(record []string) error {

			if !reflect.DeepEqual(test.expectedOutput[j], record) {
				t.Error()
			}
			j++
			return nil
		})

		if err != nil {
			t.Error(err)
		}
	}
}
