package fwprovider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	customtypes "github.com/bpg/terraform-provider-proxmox/fwprovider/types"
)

type userResourceModel struct {
	ID types.String `tfsdk:"user_id"`

	Comment        types.String            `tfsdk:"comment"`
	Email          types.String            `tfsdk:"email"`
	Enabled        types.Bool              `tfsdk:"enabled"`
	ExpirationDate customtypes.RFC3339Time `tfsdk:"expiration_date"`
	FirstName      types.String            `tfsdk:"first_name"`
	Groups         types.Set               `tfsdk:"groups"`
	Keys           types.String            `tfsdk:"keys"`
	LastName       types.String            `tfsdk:"last_name"`
	Password       types.String            `tfsdk:"password"`

	ACL []userACLResourceModel `tfsdk:"acl"`
}

type userACLResourceModel struct {
	Path      types.String `tfsdk:"path"`
	Propagate types.Bool   `tfsdk:"propagate"`
	RoleID    types.String `tfsdk:"role_id"`
}
