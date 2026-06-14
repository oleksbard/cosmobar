package render

import (
	"reflect"
	"testing"
)

func TestFitSingleRow(t *testing.T) {
	// widths sum + seps small enough for one row
	widths := []int{8, 10, 4}
	prios := []int{70, 90, 80}
	got := Fit(widths, prios, 3, 120, 2)
	want := [][]int{{0, 1, 2}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Fit single row = %v", got)
	}
}

func TestFitWrapsTwoRows(t *testing.T) {
	// total 8+3+10+3+4+3+12+3+5+3+5 = 59 > 40 → two rows
	widths := []int{8, 10, 4, 12, 5, 5}
	prios := []int{70, 90, 80, 100, 60, 50}
	got := Fit(widths, prios, 3, 40, 2)
	want := [][]int{{0, 1, 2}, {3, 4, 5}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Fit wrap = %v, want %v", got, want)
	}
}

func TestFitDropsLowestPriorityWhenOverflowing(t *testing.T) {
	// With maxRows=1 and a tight width, lowest-prio items get dropped until it fits.
	widths := []int{10, 10, 10}
	prios := []int{30, 90, 60}          // index 0 is lowest prio
	got := Fit(widths, prios, 3, 23, 1) // fits two 10s + one sep = 23
	want := [][]int{{1, 2}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Fit drop = %v, want %v", got, want)
	}
}
