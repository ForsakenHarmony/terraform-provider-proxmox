package tests

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource(t *testing.T) {
	t.Parallel()

	resourceName := "proxmox_virtual_environment_user.test"

	accProviders := testAccMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: `resource "proxmox_virtual_environment_user" "test" {
					comment  = "Managed by Terraform"
					password = "a-strong-password"
					user_id  = "test@pve"
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "comment", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "password", "a-strong-password"),
					resource.TestCheckResourceAttr(resourceName, "user_id", "test@pve"),
				),
			},
		},
	})
}
