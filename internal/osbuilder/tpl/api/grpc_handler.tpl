// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://code.byted.org/ies/qagents. The professional
// version of this repository is https://code.byted.org/ies/qagents.

package handler

import (
	"context"

	jobv1 "code.byted.org/ies/qagents/pkg/api/apiserver/v1/job"
)

// CreateJob handles the creation of a new job.
func (h *QAgentsHandler) CreateJob(ctx context.Context, rq *jobv1.CreateJobRequest) (*jobv1.CreateJobResponse, error) {
	return h.biz.JobsV1().Create(ctx, rq)
}

// UpdateJob handles updating an existing job's details.
func (h *QAgentsHandler) UpdateJob(ctx context.Context, rq *jobv1.UpdateJobRequest) (*jobv1.UpdateJobResponse, error) {
	return h.biz.JobsV1().Update(ctx, rq)
}

// DeleteJob handles the deletion of one or more jobss.
func (h *QAgentsHandler) DeleteJob(ctx context.Context, rq *jobv1.DeleteJobRequest) (*jobv1.DeleteJobResponse, error) {
	return h.biz.JobsV1().Delete(ctx, rq)
}

// GetJob retrieves information about a specific job.
func (h *QAgentsHandler) GetJob(ctx context.Context, rq *jobv1.GetJobRequest) (*jobv1.GetJobResponse, error) {
	return h.biz.JobsV1().Get(ctx, rq)
}

// ListJob retrieves a list of jobs based on query parameters.
func (h *QAgentsHandler) ListJob(ctx context.Context, rq *jobv1.ListJobRequest) (*jobv1.ListJobResponse, error) {
	return h.biz.JobsV1().List(ctx, rq)
}
