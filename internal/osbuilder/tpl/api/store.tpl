// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://code.byted.org/ies/qagents. The professional
// version of this repository is https://code.byted.org/ies/qagents.

// nolint: dupl
package store

import (
	"context"

	genericstore "code.byted.org/ies/qastack/pkg/store"
	"code.byted.org/ies/qastack/pkg/store/logger/byted"
	"code.byted.org/ies/qastack/pkg/store/where"

	"code.byted.org/ies/qagents/internal/apiserver/model"
)

// JobStore defines the interface for managing job-related data operations.
type JobStore interface {
	// Create inserts a new Job record into the store.
	Create(ctx context.Context, obj *model.JobM) error

	// Update modifies an existing Job record in the store based on the given model.
	Update(ctx context.Context, obj *model.JobM) error

	// Delete removes Job records that satisfy the given query options.
	Delete(ctx context.Context, opts *where.Options) error

	// Get retrieves a single Job record that satisfies the given query options.
	Get(ctx context.Context, opts *where.Options) (*model.JobM, error)

	// List retrieves a list of Job records and their total count based on the given query options.
	List(ctx context.Context, opts *where.Options) (int64, []*model.JobM, error)

	// JobExpansion is a placeholder for extension methods for jobs, to be implemented by additional interfaces if needed.
	JobExpansion
}

// JobExpansion is an empty interface provided for extending the JobStore interface.
// Developers can define job-specific additional methods in this interface for future expansion.
type JobExpansion interface{}

// jobStore implements the JobStore interface and provides default implementations of the methods.
type jobStore struct {
	*genericstore.Store[model.JobM]
}

// Ensure that jobStore satisfies the JobStore interface at compile time.
var _ JobStore = (*jobStore)(nil)

// newJobStore creates a new jobStore instance with the provided datastore and logger.
func newJobStore(store *datastore) *jobStore {
	return &jobStore{
		Store: genericstore.NewStore[model.JobM](store, byted.NewLogger()),
	}
}
