package types

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestRFC3339TimeType_ValueFromTerraform(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val         tftypes.Value
		expected    attr.Value
		expectError bool
	}{
		"null value": {
			val:      tftypes.NewValue(tftypes.String, nil),
			expected: RFC3339TimeNull(),
		},
		"unknown value": {
			val:      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expected: RFC3339TimeUnknown(),
		},
		"empty string": {
			val:         tftypes.NewValue(tftypes.String, ""),
			expectError: true,
		},
		"valid string": {
			val:      tftypes.NewValue(tftypes.String, "2006-01-02T15:04:05Z"),
			expected: RFC3339TimeValue(time.Date(2006, 1, 2, 15, 4, 5, 7, time.UTC)),
		},
		"invalid string": {
			val:         tftypes.NewValue(tftypes.String, "not ok"),
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.TODO()
			val, err := RFC3339TimeType{}.ValueFromTerraform(ctx, test.val)

			if err == nil && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if err != nil && !test.expectError {
				t.Fatalf("got unexpected error: %s", err)
			}

			if diff := cmp.Diff(test.expected, val); diff != "" {
				t.Errorf("unexpected diff (-wanted, +got): %s", diff)
			}
		})
	}
}

func TestRFC3339TimeType_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val         tftypes.Value
		expectError bool
	}{
		"not a string": {
			val:         tftypes.NewValue(tftypes.Bool, true),
			expectError: true,
		},
		"unknown string": {
			val: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		},
		"null string": {
			val: tftypes.NewValue(tftypes.String, nil),
		},
		"empty string": {
			val:         tftypes.NewValue(tftypes.String, ""),
			expectError: true,
		},
		"valid string": {
			val: tftypes.NewValue(tftypes.String, "2006-01-02T15:04:05Z"),
		},
		"invalid string": {
			val:         tftypes.NewValue(tftypes.String, "not ok"),
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.TODO()

			attributePath := path.Root("test")
			diags := RFC3339TimeType{}.Validate(ctx, test.val, attributePath)

			if !diags.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if diags.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", DiagsError(diags))
			}
		})
	}
}
