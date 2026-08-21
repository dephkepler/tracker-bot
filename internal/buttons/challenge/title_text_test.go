package challenge

import (
	"strings"
	"testing"
	"tracker-bot/internal/models"
)

func TestAllocateSegments_SumsExactlyToWidth(t *testing.T) {
	cases := []struct {
		name   string
		width  int
		values []int
	}{
		{"even thirds", 14, []int{1, 1, 1}},
		{"one dominant", 14, []int{97, 2, 1}},
		{"one category is everything", 14, []int{0, 0, 5}},
		{"two categories tie", 10, []int{5, 5, 0}},
		{"tiny total, wide strip", 14, []int{1, 0, 0}},
		{"odd remainders", 7, []int{1, 1, 1}},
	}
	for _, c := range cases {
		total := 0
		for _, v := range c.values {
			total += v
		}
		out := allocateSegments(c.width, total, c.values...)
		sum := 0
		for _, v := range out {
			sum += v
		}
		if sum != c.width {
			t.Errorf("%s: sum(%v) = %d, want %d", c.name, out, sum, c.width)
		}
		for i, v := range out {
			if v < 0 {
				t.Errorf("%s: out[%d] = %d, want >= 0", c.name, i, v)
			}
		}
	}
}

func TestAllocateSegments_ZeroValueGetsNoCells(t *testing.T) {
	out := allocateSegments(14, 5, 0, 5, 0)
	if out[0] != 0 || out[2] != 0 {
		t.Errorf("out = %v, want the two zero categories to stay at 0", out)
	}
	if out[1] != 14 {
		t.Errorf("out[1] = %d, want 14 (the only nonzero category takes the whole width)", out[1])
	}
}

func TestSegmentStrip_EmptyRendersAllPending(t *testing.T) {
	got := segmentStrip(0, 0, 0)
	want := strings.Repeat("⚪", segmentWidth)
	if got != want {
		t.Errorf("segmentStrip(0,0,0) = %q, want %q", got, want)
	}
}

func TestSegmentStrip_RenderedWidthMatchesSegmentWidth(t *testing.T) {
	got := segmentStrip(7, 2, 1)
	if n := len([]rune(got)); n != segmentWidth {
		t.Errorf("rune length = %d, want %d", n, segmentWidth)
	}
}

func TestTrendStrip_MapsStatusToFixedLevels(t *testing.T) {
	trend := []models.ChallengeDayStatus{
		models.ChallengeDayDone, models.ChallengeDaySkipped, models.ChallengeDayPending,
	}
	got := trendStrip(trend)
	want := trendCharDone + trendCharSkipped + trendCharPending
	if got != want {
		t.Errorf("trendStrip(...) = %q, want %q", got, want)
	}
}

func TestTrendStrip_Empty(t *testing.T) {
	if got := trendStrip(nil); got != "" {
		t.Errorf("trendStrip(nil) = %q, want empty", got)
	}
}
