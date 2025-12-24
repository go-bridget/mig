package introspect

import (
	"sort"
	"strings"

	"github.com/go-bridget/mig/model"
)

// sortIndexes sorts indexes consistently: primary key first, then by column names they index.
func sortIndexes(indexes []*model.Index) {
	sort.Slice(indexes, func(i, j int) bool {
		// Primary key indexes come first
		if indexes[i].Primary != indexes[j].Primary {
			return indexes[i].Primary
		}
		// Then sort by concatenated column names for consistent ordering
		iCols := strings.Join(indexes[i].Columns, ",")
		jCols := strings.Join(indexes[j].Columns, ",")
		return iCols < jCols
	})
}
