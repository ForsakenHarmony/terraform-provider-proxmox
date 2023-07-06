package types

import (
	"errors"

	"github.com/hashicorp/go-multierror"
	tfdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// DiagsError creates an error from Diagnostics.
func DiagsError(diags tfdiag.Diagnostics) error {
	var errs *multierror.Error

	for _, diag := range diags {
		if diag == nil {
			continue
		}

		if diag.Severity() == tfdiag.SeverityError {
			errs = multierror.Append(errs, errors.New(diag.Detail()))
		}
	}

	//nolint:wrapcheck
	return errs.ErrorOrNil()
}
