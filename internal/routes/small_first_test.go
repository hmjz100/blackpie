package routes

import (
	"slices"
	"testing"
)

func TestSmallFirstOrdersKnownWindowsBeforeUnknownAndSpent(t *testing.T) {
	seats := []Seat{
		{Key: "large", Window: 1000000},
		{Key: "unknown"},
		{Key: "small", Window: 4000},
		{Key: "spent", Window: 1000, Used: 100},
	}
	var got []string
	for _, i := range Apply(SmallFirst, "group", seats) {
		got = append(got, seats[i].Key)
	}
	if want := []string{"small", "large", "unknown", "spent"}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
