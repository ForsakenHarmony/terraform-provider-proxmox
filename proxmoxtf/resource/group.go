/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package resource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bpg/terraform-provider-proxmox/proxmox/access"
	"github.com/bpg/terraform-provider-proxmox/proxmox/api"
	"github.com/bpg/terraform-provider-proxmox/proxmoxtf"
)

const (
	dvResourceVirtualEnvironmentGroupComment = ""

	mkResourceVirtualEnvironmentGroupComment = "comment"
	mkResourceVirtualEnvironmentGroupID      = "group_id"
	mkResourceVirtualEnvironmentGroupMembers = "members"
)

// Group returns a resource that manages a group in the Proxmox VE access control list.
func Group() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			mkResourceVirtualEnvironmentGroupComment: {
				Type:        schema.TypeString,
				Description: "The group comment",
				Optional:    true,
				Default:     dvResourceVirtualEnvironmentGroupComment,
			},
			mkResourceVirtualEnvironmentGroupID: {
				Type:        schema.TypeString,
				Description: "The group id",
				Required:    true,
				ForceNew:    true,
			},
			mkResourceVirtualEnvironmentGroupMembers: {
				Type:        schema.TypeSet,
				Description: "The group members",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
		CreateContext: groupCreate,
		ReadContext:   groupRead,
		UpdateContext: groupUpdate,
		DeleteContext: groupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func groupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(proxmoxtf.ProviderConfiguration)

	client, err := config.GetClient()
	if err != nil {
		return diag.FromErr(err)
	}

	comment := d.Get(mkResourceVirtualEnvironmentGroupComment).(string)
	groupID := d.Get(mkResourceVirtualEnvironmentGroupID).(string)

	body := &access.GroupCreateRequestBody{
		Comment: &comment,
		ID:      groupID,
	}

	err = client.Access().CreateGroup(ctx, body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(groupID)

	return groupRead(ctx, d, m)
}

func groupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	config := m.(proxmoxtf.ProviderConfiguration)

	client, err := config.GetClient()
	if err != nil {
		return diag.FromErr(err)
	}

	groupID := d.Id()

	group, err := client.Access().GetGroup(ctx, groupID)
	if err != nil {
		if errors.Is(err, api.ErrResourceDoesNotExist) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if group.Comment != nil {
		err = d.Set(mkResourceVirtualEnvironmentGroupComment, group.Comment)
	} else {
		err = d.Set(mkResourceVirtualEnvironmentGroupComment, "")
	}

	diags = append(diags, diag.FromErr(err)...)

	err = d.Set(mkResourceVirtualEnvironmentGroupMembers, group.Members)
	diags = append(diags, diag.FromErr(err)...)

	return diags
}

func groupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(proxmoxtf.ProviderConfiguration)

	client, err := config.GetClient()
	if err != nil {
		return diag.FromErr(err)
	}

	comment := d.Get(mkResourceVirtualEnvironmentGroupComment).(string)
	groupID := d.Id()

	body := &access.GroupUpdateRequestBody{
		Comment: &comment,
	}

	err = client.Access().UpdateGroup(ctx, groupID, body)
	if err != nil {
		return diag.FromErr(err)
	}

	return groupRead(ctx, d, m)
}

func groupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(proxmoxtf.ProviderConfiguration)

	client, err := config.GetClient()
	if err != nil {
		return diag.FromErr(err)
	}

	groupID := d.Id()

	err = client.Access().DeleteGroup(ctx, groupID)

	if err != nil && !errors.Is(err, api.ErrResourceDoesNotExist) {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
