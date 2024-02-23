---
layout: page
title: proxmox_virtual_environment_user
parent: Resources
subcategory: Virtual Environment
---

# Resource: proxmox_virtual_environment_user

Manages a user.

## Example Usage

```terraform
resource "proxmox_virtual_environment_user" "operations_automation" {
  comment  = "Managed by Terraform"
  password = "a-strong-password"
  user_id  = "operations-automation@pve"
}

resource "proxmox_virtual_environment_role" "operations_monitoring" {
  role_id = "operations-monitoring"

  privileges = [
    "VM.Monitor",
  ]
}
```

## Argument Reference

- `comment` - (Optional) The user comment.
- `email` - (Optional) The user's email address.
- `enabled` - (Optional) Whether the user account is enabled.
- `expiration_date` - (Optional) The user account's expiration date (RFC 3339).
- `first_name` - (Optional) The user's first name.
- `groups` - (Optional) The user's groups.
- `keys` - (Optional) The user's keys.
- `last_name` - (Optional) The user's last name.
- `password` - (Optional) The user's password. Required for PVE or PAM realms.
- `user_id` - (Required) The user identifier.

## Attribute Reference

There are no additional attributes available for this resource.

## Import

Instances can be imported using the `user_id`, e.g.,

```bash
terraform import proxmox_virtual_environment_user.operations_automation operations-automation@pve
```
