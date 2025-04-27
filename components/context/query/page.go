package query

// Page returns the page.
func (q *query) Page(defaultValue ...uint) uint {
	if v := q.Get("page").UInt(); v != 0 {
		return v
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return 1
}
