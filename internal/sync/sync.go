package sync

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

type Options struct {
	Namespace     string
	DeleteMissing bool
	DryRun        bool
	TenantID      string
	CreatedBy     string
}

type Summary struct {
	Created   int `json:"created"`
	Updated   int `json:"updated"`
	Deleted   int `json:"deleted"`
	Unchanged int `json:"unchanged"`
	Errors    int `json:"errors"`
}

type SyncError struct {
	Name   string `json:"name"`
	Error  string `json:"error"`
	Action string `json:"action,omitempty"`
}

type Result struct {
	Message   string      `json:"message,omitempty"`
	Namespace string      `json:"namespace,omitempty"`
	Summary   Summary     `json:"summary"`
	Details   Details     `json:"details"`
	Errors    []SyncError `json:"errors,omitempty"`
	DryRun    bool        `json:"dry_run"`
}

type Details struct {
	Created   []string `json:"created"`
	Updated   []string `json:"updated"`
	Deleted   []string `json:"deleted"`
	Unchanged []string `json:"unchanged,omitempty"`
}

type ItemSpec interface {
	GetName() string
}

type Syncer[T ItemSpec] interface {
	ListExisting(ctx context.Context, opts Options) (map[string]string, error)
	IsChanged(ctx context.Context, existingID string, item T, opts Options) (bool, error)
	Preprocess(ctx context.Context, item T) error
	Create(ctx context.Context, item T, opts Options) error
	Update(ctx context.Context, item T, existingID string, opts Options) error
	Delete(ctx context.Context, name string, existingID string, opts Options) (bool, error)
	PostSync(ctx context.Context, result *Result, opts Options) error
}

func Execute[T ItemSpec](ctx context.Context, syncer Syncer[T], items []T, opts Options) (*Result, error) {
	result := &Result{
		Namespace: opts.Namespace,
		DryRun:    opts.DryRun,
		Details: Details{
			Created:   []string{},
			Updated:   []string{},
			Deleted:   []string{},
			Unchanged: []string{},
		},
		Errors: []SyncError{},
	}

	payloadNames := make(map[string]bool)
	for _, item := range items {
		payloadNames[item.GetName()] = true
	}

	existing, err := syncer.ListExisting(ctx, opts)
	if err != nil {
		return nil, err
	}

	type updateItem struct {
		item       T
		existingID string
	}

	var toCreate []T
	var toUpdate []updateItem

	for _, item := range items {
		if existingID, exists := existing[item.GetName()]; exists {
			changed, checkErr := syncer.IsChanged(ctx, existingID, item, opts)
			if checkErr != nil {
				result.Errors = append(result.Errors, SyncError{
					Name: item.GetName(), Error: checkErr.Error(), Action: "check",
				})
				result.Summary.Errors++
				continue
			}
			if !changed {
				result.Details.Unchanged = append(result.Details.Unchanged, item.GetName())
				result.Summary.Unchanged++
				continue
			}
			toUpdate = append(toUpdate, updateItem{item: item, existingID: existingID})
		} else {
			toCreate = append(toCreate, item)
		}
	}

	if opts.DryRun {
		for _, item := range toCreate {
			result.Details.Created = append(result.Details.Created, item.GetName())
			result.Summary.Created++
		}
		for _, u := range toUpdate {
			result.Details.Updated = append(result.Details.Updated, u.item.GetName())
			result.Summary.Updated++
		}
		if opts.DeleteMissing {
			for name := range existing {
				if !payloadNames[name] {
					result.Details.Deleted = append(result.Details.Deleted, name)
					result.Summary.Deleted++
				}
			}
		}
		result.Message = "Dry run - no changes made"
		return result, nil
	}

	for _, item := range toCreate {
		if err := syncer.Preprocess(ctx, item); err != nil {
			result.Errors = append(result.Errors, SyncError{Name: item.GetName(), Error: err.Error(), Action: "preprocess"})
			result.Summary.Errors++
			continue
		}
		if err := syncer.Create(ctx, item, opts); err != nil {
			result.Errors = append(result.Errors, SyncError{Name: item.GetName(), Error: err.Error(), Action: "create"})
			result.Summary.Errors++
			continue
		}
		result.Details.Created = append(result.Details.Created, item.GetName())
		result.Summary.Created++
	}

	for _, u := range toUpdate {
		if err := syncer.Preprocess(ctx, u.item); err != nil {
			result.Errors = append(result.Errors, SyncError{Name: u.item.GetName(), Error: err.Error(), Action: "preprocess"})
			result.Summary.Errors++
			continue
		}
		if err := syncer.Update(ctx, u.item, u.existingID, opts); err != nil {
			result.Errors = append(result.Errors, SyncError{Name: u.item.GetName(), Error: err.Error(), Action: "update"})
			result.Summary.Errors++
			continue
		}
		result.Details.Updated = append(result.Details.Updated, u.item.GetName())
		result.Summary.Updated++
	}

	if opts.DeleteMissing {
		for name, existingID := range existing {
			if !payloadNames[name] {
				deleted, err := syncer.Delete(ctx, name, existingID, opts)
				if err != nil {
					result.Errors = append(result.Errors, SyncError{Name: name, Error: err.Error(), Action: "delete"})
					result.Summary.Errors++
					continue
				}
				if deleted {
					result.Details.Deleted = append(result.Details.Deleted, name)
					result.Summary.Deleted++
				}
			}
		}
	}

	if err := syncer.PostSync(ctx, result, opts); err != nil {
		result.Errors = append(result.Errors, SyncError{Name: "post_sync", Error: err.Error()})
	}

	if len(result.Errors) == 0 {
		result.Message = "Sync completed"
	}
	return result, nil
}

func GetNamespace(c fiber.Ctx) string {
	ns := c.Query("namespace")
	if ns == "" {
		return "default"
	}
	return ns
}

func DefaultNamespace(ns string) string {
	if ns == "" {
		return "default"
	}
	return ns
}
