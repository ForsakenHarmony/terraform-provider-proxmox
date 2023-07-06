package types

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ basetypes.StringValuable = RFC3339Time{}

// RFC3339TimeValue creates an RFC3339Time with a known value.
func RFC3339TimeValue(value time.Time) RFC3339Time {
	return RFC3339Time{
		StringValue: types.StringValue(value.Format(time.RFC3339)),
		value:       value,
	}
}

// RFC3339TimePointerValue creates an RFC3339Time with a null value if nil or a known value.
func RFC3339TimePointerValue(value *time.Time) RFC3339Time {
	if value == nil {
		return RFC3339TimeNull()
	}

	return RFC3339Time{
		StringValue: types.StringValue(value.Format(time.RFC3339)),
		value:       *value,
	}
}

// RFC3339TimeNull creates an RFC3339Time with a null value.
func RFC3339TimeNull() RFC3339Time {
	return RFC3339Time{
		StringValue: types.StringNull(),
		value:       time.Time{},
	}
}

// RFC3339TimeUnknown creates an RFC3339Time with an unknown value.
func RFC3339TimeUnknown() RFC3339Time {
	return RFC3339Time{
		StringValue: types.StringUnknown(),
		value:       time.Time{},
	}
}

// RFC3339Time is a string value that represents an RFC3339Time timestamp.
type RFC3339Time struct {
	basetypes.StringValue

	value time.Time
}

// Type returns a StringType.
func (r RFC3339Time) Type(_ context.Context) attr.Type {
	return RFC3339TimeType{}
}

// Equal returns true if `other` is an RFC3339Time and has the same value as `r`.
func (r RFC3339Time) Equal(other attr.Value) bool {
	o, ok := other.(RFC3339Time)

	if !ok {
		return false
	}

	return r.StringValue.Equal(o.StringValue)
}

// ValueTime returns the known time value. If String is null or unknown, returns
// the default value.
func (r RFC3339Time) ValueTime() time.Time {
	return r.value
}

// ValueTimePointer returns a pointer to the known time value, nil for a null
// value, or a pointer to the default time for an unknown value.
func (r RFC3339Time) ValueTimePointer() *time.Time {
	if r.IsNull() {
		return nil
	}

	return &r.value
}
