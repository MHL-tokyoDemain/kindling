package firestore

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kindling/kindling/pkg/types"
)

func TestProjectIDAccessor(t *testing.T) {
	c := &Client{projectID: "my-project-123"}
	if c.ProjectID() != "my-project-123" {
		t.Errorf("expected my-project-123, got %q", c.ProjectID())
	}
}

func TestMapFirestoreErrorPermissionDenied(t *testing.T) {
	err := status.Error(codes.PermissionDenied, "no access")
	mapped := mapFirestoreError(err)

	var we *WriteError
	if !errors.As(mapped, &we) {
		t.Fatalf("expected *WriteError, got %T", mapped)
	}
	if we.Code != types.ErrAuthSessionInvalid {
		t.Errorf("expected code %q, got %q", types.ErrAuthSessionInvalid, we.Code)
	}
	if we.Message != "permission denied" {
		t.Errorf("expected 'permission denied', got %q", we.Message)
	}
}

func TestMapFirestoreErrorResourceExhausted(t *testing.T) {
	err := status.Error(codes.ResourceExhausted, "too many")
	mapped := mapFirestoreError(err)

	var we *WriteError
	if !errors.As(mapped, &we) {
		t.Fatalf("expected *WriteError, got %T", mapped)
	}
	if we.Code != types.ErrFirestoreQuota {
		t.Errorf("expected code %q, got %q", types.ErrFirestoreQuota, we.Code)
	}
}

func TestMapFirestoreErrorDefaultGRPCCode(t *testing.T) {
	err := status.Error(codes.Unavailable, "backend unavailable")
	mapped := mapFirestoreError(err)

	var we *WriteError
	if !errors.As(mapped, &we) {
		t.Fatalf("expected *WriteError, got %T", mapped)
	}
	if we.Code != types.ErrInternal {
		t.Errorf("expected code %q, got %q", types.ErrInternal, we.Code)
	}
	if we.Message != "backend unavailable" {
		t.Errorf("expected 'backend unavailable', got %q", we.Message)
	}
}

func TestMapFirestoreErrorNonStatusError(t *testing.T) {
	err := errors.New("plain non-grpc error")
	mapped := mapFirestoreError(err)

	var we *WriteError
	if !errors.As(mapped, &we) {
		t.Fatalf("expected *WriteError, got %T", mapped)
	}
	if we.Code != types.ErrInternal {
		t.Errorf("expected code %q, got %q", types.ErrInternal, we.Code)
	}
	if we.Message != "unexpected Firestore error" {
		t.Errorf("expected 'unexpected Firestore error', got %q", we.Message)
	}
	if !errors.Is(mapped, err) {
		t.Error("expected mapped error to wrap the original error")
	}
}
