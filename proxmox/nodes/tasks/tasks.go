/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package tasks

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/bpg/terraform-provider-proxmox/proxmox/api"
)

// GetTaskStatus retrieves the status of a task.
func (c *Client) GetTaskStatus(ctx context.Context, upid string) (*GetTaskStatusResponseData, error) {
	resBody := &GetTaskStatusResponseBody{}

	err := c.DoRequest(
		ctx,
		http.MethodGet,
		c.ExpandPath(fmt.Sprintf("%s/status", url.PathEscape(upid))),
		nil,
		resBody,
	)
	if err != nil {
		return nil, fmt.Errorf("error retrievinf task status: %w", err)
	}

	if resBody.Data == nil {
		return nil, api.ErrNoDataObjectInResponse
	}

	return resBody.Data, nil
}

// WaitForTask waits for a specific task to complete.
func (c *Client) WaitForTask(ctx context.Context, upid string, timeout, delay time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	b := backoff.WithContext(backoff.NewConstantBackOff(delay), ctx)

	err := backoff.Retry(func() error {
		status, err := c.GetTaskStatus(ctx, upid)
		if err != nil {
			return backoff.Permanent(err)
		}

		if status.Status != "running" {
			if status.ExitCode != "OK" {
				return backoff.Permanent(fmt.Errorf(
					"task \"%s\" failed to complete with exit code: %s",
					upid,
					status.ExitCode,
				))
			}

			return nil
		}

		return errors.New("not ready")
	}, b)
	if err != nil {
		return fmt.Errorf(
			"error waiting for task \"%s\" to complete: %w",
			upid,
			err,
		)
	}

	return nil
}
