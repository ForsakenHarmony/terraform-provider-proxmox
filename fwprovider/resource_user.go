package fwprovider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	customtypes "github.com/bpg/terraform-provider-proxmox/fwprovider/types"
	"github.com/bpg/terraform-provider-proxmox/proxmox/helpers/ptr"
	proxmoxtypes "github.com/bpg/terraform-provider-proxmox/proxmox/types"

	"github.com/bpg/terraform-provider-proxmox/proxmox"
	"github.com/bpg/terraform-provider-proxmox/proxmox/access"
)

type userResource struct {
	client proxmox.Client
}

var (
	_ resource.Resource                = (*userResource)(nil)
	_ resource.ResourceWithImportState = (*userResource)(nil)
)

// NewUserResource ...
func NewUserResource() resource.Resource {
	return &userResource{}
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"comment": schema.StringAttribute{
				Description: "The user comment",
				Optional:    true,
			},
			"email": schema.StringAttribute{
				Description: "The user's email address",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the user account is enabled",
				Optional:    true,
			},
			"expiration_date": schema.StringAttribute{
				Description: "The user account's expiration date",
				Optional:    true,
				Default:     stringdefault.StaticString(time.Unix(0, 0).UTC().Format(time.RFC3339)),
				CustomType:  customtypes.RFC3339TimeType{},
			},
			"first_name": schema.StringAttribute{
				Description: "The user's first name",
				Optional:    true,
			},
			"groups": schema.SetAttribute{
				Description: "The user's groups",
				Optional:    true,
				ElementType: types.StringType,
			},
			"keys": schema.StringAttribute{
				Description: "The user's keys",
				Optional:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "The user's last name",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "The user's password (required if the realm is pve)",
				Optional:    true,
				Sensitive:   true,
			},
			"user_id": schema.StringAttribute{
				Description: "The user id",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"acl": schema.ListNestedBlock{
				Description: "The cloning configuration",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							Required:    true,
							Description: "The path",
						},
						"propagate": schema.BoolAttribute{
							Optional:    true,
							Description: "Whether to propagate to child paths",
							Default:     booldefault.StaticBool(false),
						},
						"role_id": schema.StringAttribute{
							Required:    true,
							Description: "The role id",
						},
					},
				},
			},
		},
		MarkdownDescription: "manages a user in the Proxmox VE access control list",
	}
}

func (r *userResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(proxmox.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *proxmox.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state userResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var groups []string

	resp.Diagnostics.Append(state.Groups.ElementsAs(ctx, &groups, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	body := &access.UserCreateRequestBody{
		Comment:        state.Comment.ValueStringPointer(),
		Email:          state.Email.ValueStringPointer(),
		Enabled:        proxmoxtypes.CustomBoolPointer(state.Enabled.ValueBoolPointer()),
		ExpirationDate: proxmoxtypes.CustomTimestampPointer(state.ExpirationDate.ValueTimePointer()),
		FirstName:      state.FirstName.ValueStringPointer(),
		Groups:         groups,
		ID:             state.ID.ValueString(),
		Keys:           state.Keys.ValueStringPointer(),
		LastName:       state.LastName.ValueStringPointer(),
		Password:       state.Password.ValueStringPointer(),
	}

	err := r.client.Access().CreateUser(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create user", apiCallFailed+err.Error())
		return
	}

	for _, v := range state.ACL {
		aclBody := &access.ACLUpdateRequestBody{
			Delete:    proxmoxtypes.CustomBoolPointer(ptr.Ptr(false)),
			Path:      v.Path.ValueString(),
			Propagate: proxmoxtypes.CustomBoolPointer(v.Propagate.ValueBoolPointer()),
			Roles:     []string{v.RoleID.ValueString()},
			Users:     []string{state.ID.ValueString()},
		}

		err := r.client.Access().UpdateACL(ctx, aclBody)
		if err != nil {
			resp.Diagnostics.AddError("Unable to create acl", apiCallFailed+err.Error())
			return
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.Set(ctx, state)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.Access().GetUser(ctx, state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "HTTP 404") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Unable to read User", apiCallFailed+err.Error())

		return
	}

	acl, err := r.client.Access().GetACL(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read ACLs", apiCallFailed+err.Error())
		return
	}

	//nolint:prealloc
	var aclParsed []userACLResourceModel

	for _, v := range acl {
		if v.Type != "user" || v.UserOrGroupID != state.ID.ValueString() {
			continue
		}

		aclParsed = append(aclParsed, userACLResourceModel{
			Path:      types.StringValue(v.Path),
			Propagate: types.BoolPointerValue(v.Propagate.PointerBool()),
			RoleID:    types.StringValue(v.RoleID),
		})
	}

	state.ACL = aclParsed
	state.Comment = types.StringPointerValue(user.Comment)
	state.Email = types.StringPointerValue(user.Email)
	state.Enabled = types.BoolPointerValue(user.Enabled.PointerBool())
	state.ExpirationDate = customtypes.RFC3339TimePointerValue(user.ExpirationDate.PointerTime())
	state.FirstName = types.StringPointerValue(user.FirstName)

	if user.Groups != nil {
		setValue, diag := types.SetValueFrom(ctx, types.StringType, *user.Groups)
		resp.Diagnostics.Append(diag...)

		if !diag.HasError() {
			state.Groups = setValue
		}
	} else {
		state.Groups = types.SetNull(types.StringType)
	}

	state.Keys = types.StringPointerValue(user.Keys)
	state.LastName = types.StringPointerValue(user.LastName)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.Set(ctx, state)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		oldState userResourceModel
		state    userResourceModel
	)

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var groups []string

	resp.Diagnostics.Append(state.Groups.ElementsAs(ctx, &groups, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	body := &access.UserUpdateRequestBody{
		Comment:        state.Comment.ValueStringPointer(),
		Email:          state.Email.ValueStringPointer(),
		Enabled:        proxmoxtypes.CustomBoolPointer(state.Enabled.ValueBoolPointer()),
		ExpirationDate: proxmoxtypes.CustomTimestampPointer(state.ExpirationDate.ValueTimePointer()),
		FirstName:      state.FirstName.ValueStringPointer(),
		Groups:         groups,
		Keys:           state.Keys.ValueStringPointer(),
		LastName:       state.LastName.ValueStringPointer(),
	}

	err := r.client.Access().UpdateUser(ctx, state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update user", apiCallFailed+err.Error())
		return
	}

	if !state.Password.Equal(oldState.Password) {
		err := r.client.Access().ChangeUserPassword(ctx, state.ID.ValueString(), state.Password.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Unable to update user password", apiCallFailed+err.Error())
			return
		}
	}

	for _, v := range oldState.ACL {
		aclBody := &access.ACLUpdateRequestBody{
			Delete:    proxmoxtypes.CustomBool(true).Pointer(),
			Path:      v.Path.ValueString(),
			Propagate: proxmoxtypes.CustomBoolPointer(v.Propagate.ValueBoolPointer()),
			Roles:     []string{v.RoleID.ValueString()},
			Users:     []string{state.ID.ValueString()},
		}

		err := r.client.Access().UpdateACL(ctx, aclBody)
		if err != nil {
			resp.Diagnostics.AddError("Unable to delete old ACL", apiCallFailed+err.Error())
			return
		}
	}

	for _, v := range state.ACL {
		aclBody := &access.ACLUpdateRequestBody{
			Delete:    proxmoxtypes.CustomBool(false).Pointer(),
			Path:      v.Path.ValueString(),
			Propagate: proxmoxtypes.CustomBoolPointer(v.Propagate.ValueBoolPointer()),
			Roles:     []string{v.RoleID.ValueString()},
			Users:     []string{state.ID.ValueString()},
		}

		err := r.client.Access().UpdateACL(ctx, aclBody)
		if err != nil {
			resp.Diagnostics.AddError("Unable to create ACL", apiCallFailed+err.Error())
			return
		}
	}
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	for _, v := range state.ACL {
		aclBody := &access.ACLUpdateRequestBody{
			Delete:    proxmoxtypes.CustomBool(true).Pointer(),
			Path:      v.Path.ValueString(),
			Propagate: proxmoxtypes.CustomBoolPointer(v.Propagate.ValueBoolPointer()),
			Roles:     []string{v.RoleID.ValueString()},
			Users:     []string{state.ID.ValueString()},
		}

		err := r.client.Access().UpdateACL(ctx, aclBody)
		if err != nil {
			resp.Diagnostics.AddError("Unable to delete ACL", apiCallFailed+err.Error())
			return
		}
	}

	err := r.client.Access().DeleteUser(ctx, state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "HTTP 404") {
			return
		}

		resp.Diagnostics.AddError("Unable to delete User", apiCallFailed+err.Error())
	}
}

func (r *userResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("user_id"), req, resp)
}

const apiCallFailed = "API call failed: "
