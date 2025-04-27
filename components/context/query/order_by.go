package query

import (
	"fmt"
	"strings"
)

// OrderBy is the struct that wraps the basic fields.
func (q *query) OrderBy() *OrderBy {
	var orderBy OrderBy

	orderByRaw := q.Get("orderBy").String()
	if orderByRaw == "" {
		orderByRaw = q.Get("order-by").String()
	}

	if orderByRaw != "" {
		orderByRaws := strings.Split(orderByRaw, ",")
		for _, one := range orderByRaws {
			one = strings.TrimSpace(one)
			if one == "" {
				continue
			}

			orderByRaws := strings.Split(one, ":")
			if len(orderByRaws) != 2 {
				continue
			}

			key := orderByRaws[0]
			order := strings.ToLower(orderByRaws[1])
			isDESC := false
			if order == "desc" {
				isDESC = true
			} else if order == "DESC" {
				isDESC = true
			}

			orderBy.Set(key, isDESC)
		}
	}

	return &orderBy
}

////

// OrderByOne is a single order by.
type OrderByOne struct {
	Key    string
	IsDESC bool

	//
	clause string
}

// String returns the string of the order by.
func (w *OrderByOne) String() string {
	return w.Clause()
}

// Clause returns the clause of the order by.
func (w *OrderByOne) Clause() string {
	if w.clause == "" {
		orderMod := "ASC"
		if w.IsDESC {
			orderMod = "DESC"
		}

		w.clause = fmt.Sprintf("%s %s", w.Key, orderMod)
	}

	return w.clause
}

// OrderBy is a list of order bys.
type OrderBy []OrderByOne

// Set sets a order by.
func (w *OrderBy) Set(key string, IsDESC bool) {
	*w = append(*w, OrderByOne{
		Key:    key,
		IsDESC: IsDESC,
	})
}

// Get gets a order by.
func (w *OrderBy) Get(key string) (bool, bool) {
	for _, v := range *w {
		if v.Key == key {
			return v.IsDESC, true
		}
	}

	return false, false
}

// Add adds a order by.
func (w *OrderBy) Add(key string, IsDESC bool) {
	*w = append(*w, OrderByOne{
		Key:    key,
		IsDESC: IsDESC,
	})
}

// AddClause adds a order by clause.
func (w *OrderBy) AddClause(clause string) {
	*w = append(*w, OrderByOne{
		clause: clause,
	})
}

// Del deletes a order by.
func (w *OrderBy) Del(key string) {
	for i, v := range *w {
		if v.Key == key {
			*w = append((*w)[:i], (*w)[i+1:]...)
			break
		}
	}
}

// Debug prints the order bys.
func (w *OrderBy) Debug() {
	for _, v := range *w {
		var desc string
		if v.IsDESC {
			desc = "DESC"
		} else {
			desc = "ASC"
		}

		fmt.Printf("[order_by] %s %s\n", v.Key, desc)
	}
}

// Length returns the length of the order bys.
func (w *OrderBy) Length() int {
	return len(*w)
}

// GetClause gets the order by clause.
func (w *OrderBy) GetClause(key string) string {
	for _, v := range *w {
		if v.Key == key {
			return v.clause
		}
	}

	return ""
}

// Build builds the order bys.
func (w *OrderBy) Build() string {
	orders := []string{}
	for _, order := range *w {
		orders = append(orders, order.Clause())
	}

	return strings.Join(orders, ",")
}

// Reset resets the order bys.
func (w *OrderBy) Reset() {
	*w = []OrderByOne{}
}
