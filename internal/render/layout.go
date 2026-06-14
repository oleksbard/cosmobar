package render

// Fit assigns item indices (in input order) into at most maxRows visual rows,
// where each row's total width (items + separators of width sepW) fits within
// cols. When everything cannot fit in maxRows rows, the lowest-priority items
// are dropped one at a time until it does. Returns rows as slices of original
// indices; dropped indices appear in no row.
func Fit(widths, prios []int, sepW, cols, maxRows int) [][]int {
	idx := make([]int, len(widths))
	for i := range idx {
		idx[i] = i
	}
	for {
		rows := wrap(idx, widths, sepW, cols)
		if len(rows) <= maxRows || len(idx) <= 1 {
			return rows
		}
		idx = removeValue(idx, lowestPrio(idx, prios))
	}
}

func wrap(idx, widths []int, sepW, cols int) [][]int {
	var rows [][]int
	var cur []int
	curW := 0
	for _, i := range idx {
		add := widths[i]
		if len(cur) > 0 {
			add += sepW
		}
		if len(cur) > 0 && curW+add > cols {
			rows = append(rows, cur)
			cur = nil
			curW = 0
			add = widths[i]
		}
		cur = append(cur, i)
		curW += add
	}
	if len(cur) > 0 {
		rows = append(rows, cur)
	}
	return rows
}

// lowestPrio returns the value in idx with the smallest priority (ties: last).
func lowestPrio(idx, prios []int) int {
	victim := idx[0]
	best := prios[idx[0]]
	for _, i := range idx {
		if prios[i] <= best {
			best = prios[i]
			victim = i
		}
	}
	return victim
}

func removeValue(s []int, v int) []int {
	out := s[:0:0]
	for _, x := range s {
		if x != v {
			out = append(out, x)
		}
	}
	return out
}
