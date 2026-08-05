package services

import "testing"

func TestPercentUsesIntegerMoneyAndRoundsToTwoDecimals(t *testing.T) {
	if got := percent(-3000, 45000); got != -6.67 {
		t.Fatalf("got %v, want -6.67", got)
	}
	if got := percent(-3000, 42000); got != -7.14 {
		t.Fatalf("got %v, want -7.14", got)
	}
	if got := percent(1, 0); got != 0 {
		t.Fatalf("zero base: got %v", got)
	}
}

func TestGroupNumber(t *testing.T) {
	if got := groupNumber(45000); got != "45 000" {
		t.Fatalf("got %q", got)
	}
}
