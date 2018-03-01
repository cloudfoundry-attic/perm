package ceflog

import (
	"bytes"
	"strings"
	"testing"
)

const (
	testVendor  = "acme"
	testProduct = "flux capacitor"
	testVersion = "42"
)

func TestCEFEvent(t *testing.T) {
	var tests = []struct {
		name, signature string
		severity        int
		extension       []string

		want string
	}{
		{
			name:      "Scary event",
			signature: "scary.event",
			want:      "CEF:0|acme|flux capacitor|42|scary.event|Scary event|0|",
		},
		{
			name:      "Tricky event",
			signature: "tricky|event",
			want:      "CEF:0|acme|flux capacitor|42|tricky\\|event|Tricky event|0|",
		},
		{
			name:      "Sneaky|event",
			signature: "sneaky.event",
			want:      "CEF:0|acme|flux capacitor|42|sneaky.event|Sneaky\\|event|0|",
		},
		{
			name:      "\\Nasty|ev=ent",
			signature: "nasty.event",
			extension: []string{"src", "ugly=s|ou\\rce"},
			want:      "CEF:0|acme|flux capacitor|42|nasty.event|\\\\Nasty\\|ev=ent|0|src=ugly\\=s|ou\\\\rce",
		},
		{
			name:      "Informational event",
			signature: "informational.event",
			extension: []string{"src", "127.0.0.1"},
			want:      "CEF:0|acme|flux capacitor|42|informational.event|Informational event|0|src=127.0.0.1",
		},
		{
			name:      "Really informational event",
			signature: "really.informational.event",
			extension: []string{"src", "127.0.0.1", "dst", "127.0.0.2"},
			want:      "CEF:0|acme|flux capacitor|42|really.informational.event|Really informational event|0|src=127.0.0.1 dst=127.0.0.2",
		},
		{
			name:      "Severe event",
			signature: "severe.event",
			severity:  10,
			want:      "CEF:0|acme|flux capacitor|42|severe.event|Severe event|10|",
		},
		{
			name:      "Severity overflow",
			signature: "severity.overflow",
			severity:  100,
			want:      "CEF:0|acme|flux capacitor|42|severity.overflow|Severity overflow|10|",
		},
		{
			name:      "Severity underflow",
			signature: "severity.underflow",
			severity:  -1,
			want:      "CEF:0|acme|flux capacitor|42|severity.underflow|Severity underflow|0|",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var w bytes.Buffer
			l := New(&w, testVendor, testProduct, testVersion)

			l.LogEvent(tc.signature, tc.name, Sev(tc.severity), Ext(tc.extension...))

			actual := strings.TrimSpace(w.String())
			if actual != tc.want {
				t.Errorf("log line mismatch; want: %v, got: %v", tc.want, actual)
			}
		})
	}
}

func TestOddExtPair(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected Ext() to panic on an odd number of keys but it did not")
		}
	}()

	Ext("happy-key", "happy-value", "lonely-key")
}
