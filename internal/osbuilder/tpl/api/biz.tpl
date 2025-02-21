// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://code.byted.org/ies/qagents. The professional
// version of this repository is https://code.byted.org/ies/qagents.

package job

// Package job implements the business logic for job management. It provides
// interfaces and methods for creating, updating, deleting, retrieving,
// and listing jobs.

//go:generate mockgen -destination mock_job.go -package job code.byted.org/ies/qagents/internal/qagents/biz/v1/job JobBiz

import (
	"context"

	"code.byted.org/ies/qastack/pkg/store/where"
	"github.com/jinzhu/copier"

	"code.byted.org/ies/qagents/internal/apiserver/model"
	"code.byted.org/ies/qagents/internal/apiserver/pkg/conversion"
	"code.byted.org/ies/qagents/internal/apiserver/store"
	jobv1 "code.byted.org/ies/qagents/pkg/api/apiserver/v1/job"
)

// JobBiz defines the interface for job business logic.
type JobBiz interface {
	// Create creates a new job based on the provided request.
	Create(ctx context.Context, rq *jobv1.CreateJobRequest) (*jobv1.CreateJobResponse, error)

	// Update updates an existing job based on the provided request.
	Update(ctx context.Context, rq *jobv1.UpdateJobRequest) (*jobv1.UpdateJobResponse, error)

	// Delete deletes an existing job based on the provided request.
	Delete(ctx context.Context, rq *jobv1.DeleteJobRequest) (*jobv1.DeleteJobResponse, error)

	// Get retrieves details of a single job based on the provided request.
	Get(ctx context.Context, rq *jobv1.GetJobRequest) (*jobv1.GetJobResponse, error)

	// List retrieves a paginated list of jobs based on the provided query request.
	List(ctx context.Context, rq *jobv1.ListJobRequest) (*jobv1.ListJobResponse, error)

	// JobExpansion is reserved for extending the interface with additional methods.
	JobExpansion
}

// JobExpansion is an empty interface that can be used for extending the JobBiz interface.
type JobExpansion interface{}

// jobBiz implements the JobBiz interface and serves as the specific instance
// for handling job-related business operations with a store dependency.
type jobBiz struct {
	store store.IStore
}

// Ensure jobBiz implements the JobBiz interface.
var _ JobBiz = (*jobBiz)(nil)

// New creates a new instance of jobBiz with the provided store dependency.
func New(store store.IStore) *jobBiz {
	return &jobBiz{store: store}
}

// Create creates a new job based on the provided request.
func (b *jobBiz) Create(ctx context.Context, rq *jobv1.CreateJobRequest) (*jobv1.CreateJobResponse, error) {
	var jobM model.JobM
	_ = copier.Copy(&jobM, rq) // Copies request data to the job model.

	// TODO: Implement additional business logic here.

	if err := b.store.Jobs().Create(ctx, &jobM); err != nil {
		return nil, err
	}

	return &jobv1.CreateJobResponse{Id: &jobM.ID}, nil
}

// Update updates an existing job based on the provided request.
func (b *jobBiz) Update(ctx context.Context, rq *jobv1.UpdateJobRequest) (*jobv1.UpdateJobResponse, error) {
	whr := where.T(ctx).F("id", rq.Job.Id) // Filters by job ID.
	jobM, err := b.store.Jobs().Get(ctx, whr)
	if err != nil {
		return nil, err
	}

	// TODO: Implement additional business logic here.

	if err := b.store.Jobs().Update(ctx, jobM); err != nil {
		return nil, err
	}

	return &jobv1.UpdateJobResponse{}, nil
}

// Delete deletes an existing job based on the provided request.
func (b *jobBiz) Delete(ctx context.Context, rq *jobv1.DeleteJobRequest) (*jobv1.DeleteJobResponse, error) {
	// TODO: This might need to be modified
	whr := where.T(ctx).F("id", rq.JobIDs) // Filters by job IDs.
	if err := b.store.Jobs().Delete(ctx, whr); err != nil {
		return nil, err
	}

	return &jobv1.DeleteJobResponse{}, nil
}

// Get retrieves details of a single job based on the provided request.
func (b *jobBiz) Get(ctx context.Context, rq *jobv1.GetJobRequest) (*jobv1.GetJobResponse, error) {
	// TODO: This might need to be modified
	whr := where.T(ctx).F("id", rq.Id) // Filters by job ID.
	jobM, err := b.store.Jobs().Get(ctx, whr)
	if err != nil {
		return nil, err
	}

	return &jobv1.GetJobResponse{Job: conversion.JobModelToJobV1(jobM)}, nil
}

// List retrieves a paginated list of jobs based on the provided query request.
func (b *jobBiz) List(ctx context.Context, rq *jobv1.ListJobRequest) (*jobv1.ListJobResponse, error) {
	whr := where.T(ctx).P(int(*rq.Page), int(*rq.PageSize)) // Filters based on pagination.
	count, jobList, err := b.store.Jobs().List(ctx, whr)
	if err != nil {
		return nil, err
	}

	// Convert stored job models into Thrift API job responses.
	jobs := make([]*jobv1.Job, 0, len(jobList))
	for _, job := range jobList {
		convertedJob := conversion.JobModelToJobV1(job)
		jobs = append(jobs, convertedJob)
	}

	return &jobv1.ListJobResponse{Total: &count, Jobs: jobs}, nil
}
