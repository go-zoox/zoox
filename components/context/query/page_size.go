package query

// ConstantsQueryPageSizeKeys is the keys that are used to identify the page size.
var ConstantsQueryPageSizeKeys = []string{
	"page_size",
	"pageSize",
}

// PageSize returns the page size.
// If the page size is not set, it returns 10.
func (q *query) PageSize(defaultValue ...uint) uint {
	for _, key := range ConstantsQueryPageSizeKeys {
		if v := q.Get(key).UInt(); v != 0 {
			return v
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return 10
}
