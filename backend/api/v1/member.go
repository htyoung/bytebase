package v1

import (
	"context"
	"slices"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bytebase/bytebase/backend/common"
	api "github.com/bytebase/bytebase/backend/legacyapi"
	"github.com/bytebase/bytebase/backend/store"
)

func isOwnerOrDBA(user *store.UserMessage) bool {
	return slices.Contains(user.Roles, api.WorkspaceAdmin) || slices.Contains(user.Roles, api.WorkspaceDBA)
}

// isProjectOwnerOrDeveloper returns whether a principal is a project owner or developer in the project.
// TODO(p0ny): remove this function after iam migration.
func isProjectOwnerOrDeveloper(principalID int, projectPolicy *store.IAMPolicyMessage) bool {
	for _, binding := range projectPolicy.Bindings {
		if binding.Role != api.ProjectOwner && binding.Role != api.ProjectDeveloper {
			continue
		}
		for _, member := range binding.Members {
			if member.ID == principalID || member.Email == api.AllUsers {
				return true
			}
		}
	}
	return false
}

// TODO(p0ny): remove this function after iam migration.
func getUserBelongingProjects(ctx context.Context, s *store.Store, userUID int) (map[string]bool, error) {
	projects, err := s.ListProjectV2(ctx, &store.FindProjectMessage{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list projects")
	}

	projectIDs := map[string]bool{}
	for _, project := range projects {
		policy, err := s.GetProjectPolicy(ctx, &store.GetProjectPolicyMessage{ProjectID: &project.ResourceID})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get project %q iam policy", project.ResourceID)
		}
		if isProjectMember(userUID, policy) {
			projectIDs[project.ResourceID] = true
		}
	}
	return projectIDs, nil
}

// TODO(p0ny): remove this function after iam migration.
func isUserAtLeastProjectViewer(ctx context.Context, s *store.Store, requestProjectID string) (bool, error) {
	principalID, ok := ctx.Value(common.PrincipalIDContextKey).(int)
	if !ok {
		return false, status.Errorf(codes.Internal, "principal ID not found")
	}
	user, err := s.GetUserByID(ctx, principalID)
	if err != nil {
		return false, errors.Wrapf(err, "failed to get user %d", principalID)
	}

	if isOwnerOrDBA(user) {
		return true, nil
	}

	policy, err := s.GetProjectPolicy(ctx, &store.GetProjectPolicyMessage{ProjectID: &requestProjectID})
	if err != nil {
		return false, errors.Wrapf(err, "failed to get project iam policy")
	}

	if isProjectOwnerDeveloperOrViewer(principalID, policy) {
		return true, nil
	}
	return false, nil
}

// isProjectOwnerDeveloperOrViewer returns whether a principal is a project owner or developer in the project.
// TODO(p0ny): remove this function after iam migration.
func isProjectOwnerDeveloperOrViewer(principalID int, projectPolicy *store.IAMPolicyMessage) bool {
	for _, binding := range projectPolicy.Bindings {
		if binding.Role != api.ProjectOwner && binding.Role != api.ProjectDeveloper && binding.Role != api.ProjectViewer {
			continue
		}
		for _, member := range binding.Members {
			if member.ID == principalID || member.Email == api.AllUsers {
				return true
			}
		}
	}
	return false
}

// TODO(p0ny): remove this function after iam migration.
func isUserAtLeastProjectDeveloper(ctx context.Context, s *store.Store, requestProjectID string) (bool, error) {
	principalID, ok := ctx.Value(common.PrincipalIDContextKey).(int)
	if !ok {
		return false, status.Errorf(codes.Internal, "principal ID not found")
	}
	user, err := s.GetUserByID(ctx, principalID)
	if err != nil {
		return false, errors.Wrapf(err, "failed to get user %d", principalID)
	}

	if isOwnerOrDBA(user) {
		return true, nil
	}

	policy, err := s.GetProjectPolicy(ctx, &store.GetProjectPolicyMessage{ProjectID: &requestProjectID})
	if err != nil {
		return false, errors.Wrapf(err, "failed to get project iam policy")
	}

	if isProjectOwnerOrDeveloper(principalID, policy) {
		return true, nil
	}
	return false, nil
}

// TODO(p0ny): remove this function after iam migration.
func isUserAtLeastProjectMember(ctx context.Context, s *store.Store, requestProjectID string) (bool, error) {
	principalID, ok := ctx.Value(common.PrincipalIDContextKey).(int)
	if !ok {
		return false, status.Errorf(codes.Internal, "principal ID not found")
	}
	user, err := s.GetUserByID(ctx, principalID)
	if err != nil {
		return false, errors.Wrapf(err, "failed to get user %d", principalID)
	}

	if isOwnerOrDBA(user) {
		return true, nil
	}

	policy, err := s.GetProjectPolicy(ctx, &store.GetProjectPolicyMessage{ProjectID: &requestProjectID})
	if err != nil {
		return false, errors.Wrapf(err, "failed to get project iam policy")
	}

	if isProjectMember(principalID, policy) {
		return true, nil
	}
	return false, nil
}
