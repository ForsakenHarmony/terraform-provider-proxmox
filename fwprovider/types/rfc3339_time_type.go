package types

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ basetypes.StringTypable = RFC3339TimeType{}

// RFC3339TimeType is a type that represents an RFC3339 timestamp.
type RFC3339TimeType struct {
	basetypes.StringType
}

// Equal ...
func (t RFC3339TimeType) Equal(o attr.Type) bool {
	other, ok := o.(RFC3339TimeType)
	if !ok {
		return false
	}

	return t == other
}

// String returns the type name.
func (t RFC3339TimeType) String() string {
	return "pvetypes.RFC3339TimeType"
}

// ValueFromString converts a basetypes.StringValue into a RFC3339Time.
func (t RFC3339TimeType) ValueFromString(
	_ context.Context,
	s basetypes.StringValue,
) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if s.IsUnknown() {
		return RFC3339TimeUnknown(), diags
	}

	if s.IsNull() {
		return RFC3339TimeNull(), diags
	}

	value, err := time.Parse(time.RFC3339, s.ValueString())
	if err != nil {
		diags.AddError(
			"Error converting StringValue to RFC3339Time",
			fmt.Sprintf("unexpected error: %s", err.Error()),
		)

		return nil, diags
	}

	return RFC3339Time{
		StringValue: s,
		value:       value,
	}, diags
}

// ValueFromTerraform creates an RFC3339Time from a generic Terraform value.
func (t RFC3339TimeType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("unexpected error converting Terraform value to StringValue: %w", err)
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to RFC3339Time: %w", DiagsError(diags))
	}

	return stringValuable, nil
}

// ValueType returns the Value type.
func (t RFC3339TimeType) ValueType(_ context.Context) attr.Value {
	return RFC3339Time{}
}

// Validate validates that the terraform value is a string containing an RFC3339Time timestamp.
func (t RFC3339TimeType) Validate(_ context.Context, value tftypes.Value, p path.Path) diag.Diagnostics {
	if value.IsNull() || !value.IsKnown() {
		return nil
	}

	var (
		diags       diag.Diagnostics
		valueString string
	)

	if err := value.As(&valueString); err != nil {
		diags.AddAttributeError(
			p,
			"RFC3339Time Validation Error: Invalid Terraform Value",
			"An unexpected error occurred while attempting to convert a Terraform value to a string. "+
				"This generally is an issue with the provider schema implementation. "+
				"Please contact the provider developers.\n\n"+
				"Path: "+p.String()+"\n"+
				"Error: "+err.Error(),
		)

		return diags
	}

	if _, err := time.Parse(time.RFC3339, valueString); err != nil {
		diags.AddAttributeError(
			p,
			"RFC3339Time Validation Error: Invalid RFC 3339 String Value",
			"An unexpected error occurred while converting a string value that was expected to be RFC 3339 format. "+
				"The RFC 3339 string format is YYYY-MM-DDTHH:MM:SSZ,"+
				" such as 2006-01-02T15:04:05Z or 2006-01-02T15:04:05+07:00.\n\n"+
				"Path: "+p.String()+"\n"+
				"Given Value: "+valueString+"\n"+
				"Error: "+err.Error(),
		)

		return diags
	}

	return diags
}
