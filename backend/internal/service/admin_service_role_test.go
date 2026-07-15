//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

const (
	roleTestSuperAdminID    int64 = 9001
	roleTestSuperAdminEmail       = "super-admin@example.com"
)

func configureRoleTestSuperAdmin(t *testing.T, repo *userRepoStub) int64 {
	t.Helper()
	t.Setenv("ADMIN_EMAIL", roleTestSuperAdminEmail)
	if repo.usersByID == nil {
		repo.usersByID = make(map[int64]*User)
	}
	repo.usersByID[roleTestSuperAdminID] = &User{
		ID:    roleTestSuperAdminID,
		Email: roleTestSuperAdminEmail,
		Role:  RoleAdmin,
	}
	return roleTestSuperAdminID
}

func TestAdminService_CreateUser_WithAdminRole(t *testing.T) {
	repo := &userRepoStub{nextID: 30}
	actorID := configureRoleTestSuperAdmin(t, repo)
	svc := &adminServiceImpl{userRepo: repo}

	user, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:        "admin@test.com",
		Password:     "strong-pass",
		Role:         RoleAdmin,
		ActorAdminID: actorID,
	})
	require.NoError(t, err)
	require.Equal(t, RoleAdmin, user.Role)
}

func TestAdminService_CreateUser_WithProviderRole(t *testing.T) {
	repo := &userRepoStub{nextID: 33}
	actorID := configureRoleTestSuperAdmin(t, repo)
	svc := &adminServiceImpl{userRepo: repo}

	user, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:        "provider@test.com",
		Password:     "strong-pass",
		Role:         RoleProvider,
		ActorAdminID: actorID,
	})
	require.NoError(t, err)
	require.Equal(t, RoleProvider, user.Role)
	require.True(t, user.IsProvider())
}

func TestAdminService_CreateUser_WithAdminProviderRole(t *testing.T) {
	repo := &userRepoStub{nextID: 35}
	actorID := configureRoleTestSuperAdmin(t, repo)
	svc := &adminServiceImpl{userRepo: repo}

	user, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:        "admin-provider@test.com",
		Password:     "strong-pass",
		Role:         RoleAdminProvider,
		ActorAdminID: actorID,
	})
	require.NoError(t, err)
	require.Equal(t, RoleAdminProvider, user.Role)
	require.True(t, user.IsAdmin())
	require.True(t, user.IsProvider())
}

func TestAdminService_CreateUser_DefaultsToUserRole(t *testing.T) {
	repo := &userRepoStub{nextID: 31}
	svc := &adminServiceImpl{userRepo: repo}

	user, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "plain@test.com",
		Password: "strong-pass",
	})
	require.NoError(t, err)
	require.Equal(t, RoleUser, user.Role)
}

func TestAdminService_CreateUser_InvalidRoleRejected(t *testing.T) {
	repo := &userRepoStub{nextID: 32}
	svc := &adminServiceImpl{userRepo: repo}

	_, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "bad@test.com",
		Password: "strong-pass",
		Role:     "superuser",
	})
	require.Error(t, err)
	require.Empty(t, repo.created, "非法角色不应写入用户")
}

func TestAdminService_CreateUser_NonSuperAdminCannotCreateAdmin(t *testing.T) {
	t.Setenv("ADMIN_EMAIL", roleTestSuperAdminEmail)
	repo := &userRepoStub{
		nextID: 34,
		user:   &User{ID: 2, Email: "manager@test.com", Role: RoleAdmin},
	}
	svc := &adminServiceImpl{userRepo: repo}

	_, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:        "admin@test.com",
		Password:     "strong-pass",
		Role:         RoleAdmin,
		ActorAdminID: 2,
	})
	require.Error(t, err)
	require.Empty(t, repo.created)
}

func TestAdminService_UpdateUser_PromoteToAdmin(t *testing.T) {
	base := &userRepoStub{user: &User{ID: 42, Email: "u@example.com", Role: RoleUser}}
	actorID := configureRoleTestSuperAdmin(t, base)
	repo := &rpmUserRepoStub{userRepoStub: base}
	invalidator := &authCacheInvalidatorStub{}
	svc := &adminServiceImpl{
		userRepo:             repo,
		redeemCodeRepo:       &redeemRepoStub{},
		authCacheInvalidator: invalidator,
	}

	updated, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{Role: RoleAdmin, ActorAdminID: actorID})
	require.NoError(t, err)
	require.Equal(t, RoleAdmin, updated.Role)
	require.Equal(t, []int64{42}, invalidator.userIDs, "角色变更应失效认证缓存")
}

func TestAdminService_UpdateUser_RoleOmittedKeepsExisting(t *testing.T) {
	base := &userRepoStub{user: &User{ID: 42, Email: "u@example.com", Role: RoleAdmin}}
	repo := &rpmUserRepoStub{userRepoStub: base}
	svc := &adminServiceImpl{userRepo: repo, redeemCodeRepo: &redeemRepoStub{}}

	newName := "renamed"
	updated, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{Username: &newName})
	require.NoError(t, err)
	require.Equal(t, RoleAdmin, updated.Role, "未提供 role 时不应改变现有角色")
}

func TestAdminService_UpdateUser_InvalidRoleRejected(t *testing.T) {
	base := &userRepoStub{user: &User{ID: 42, Email: "u@example.com", Role: RoleUser}}
	repo := &rpmUserRepoStub{userRepoStub: base}
	svc := &adminServiceImpl{userRepo: repo, redeemCodeRepo: &redeemRepoStub{}}

	_, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{Role: "root"})
	require.Error(t, err)
	require.Nil(t, repo.lastUpdated, "非法角色不应触发持久化")
}

func TestAdminService_UpdateUser_AssignProviderRole(t *testing.T) {
	base := &userRepoStub{user: &User{ID: 42, Email: "u@example.com", Role: RoleUser}}
	actorID := configureRoleTestSuperAdmin(t, base)
	repo := &rpmUserRepoStub{userRepoStub: base}
	invalidator := &authCacheInvalidatorStub{}
	svc := &adminServiceImpl{
		userRepo:             repo,
		redeemCodeRepo:       &redeemRepoStub{},
		authCacheInvalidator: invalidator,
	}

	updated, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{Role: RoleProvider, ActorAdminID: actorID})
	require.NoError(t, err)
	require.Equal(t, RoleProvider, updated.Role)
	require.True(t, updated.IsProvider())
	require.Equal(t, []int64{42}, invalidator.userIDs)
}

func TestAdminService_UpdateUser_NonSuperAdminCannotChangeRole(t *testing.T) {
	t.Setenv("ADMIN_EMAIL", roleTestSuperAdminEmail)
	base := &userRepoStub{
		user: &User{ID: 42, Email: "u@example.com", Role: RoleUser},
		usersByID: map[int64]*User{
			42: &User{ID: 42, Email: "u@example.com", Role: RoleUser},
			2:  &User{ID: 2, Email: "manager@test.com", Role: RoleAdmin},
		},
	}
	repo := &rpmUserRepoStub{userRepoStub: base}
	svc := &adminServiceImpl{userRepo: repo, redeemCodeRepo: &redeemRepoStub{}}

	_, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{Role: RoleAdmin, ActorAdminID: 2})
	require.Error(t, err)
	require.Nil(t, repo.lastUpdated)
}

func TestAdminService_UpdateUser_NonSuperAdminCannotEditSuperAdmin(t *testing.T) {
	t.Setenv("ADMIN_EMAIL", roleTestSuperAdminEmail)
	base := &userRepoStub{
		user: &User{ID: roleTestSuperAdminID, Email: roleTestSuperAdminEmail, Role: RoleAdmin},
		usersByID: map[int64]*User{
			roleTestSuperAdminID: &User{ID: roleTestSuperAdminID, Email: roleTestSuperAdminEmail, Role: RoleAdmin},
			2:                    &User{ID: 2, Email: "manager@test.com", Role: RoleAdmin},
		},
	}
	repo := &rpmUserRepoStub{userRepoStub: base}
	svc := &adminServiceImpl{userRepo: repo, redeemCodeRepo: &redeemRepoStub{}}

	newName := "edited"
	_, err := svc.UpdateUser(context.Background(), roleTestSuperAdminID, &UpdateUserInput{Username: &newName, ActorAdminID: 2})
	require.Error(t, err)
	require.Nil(t, repo.lastUpdated)
}

func TestAdminService_UpdateUser_CannotChangeSuperAdminEmail(t *testing.T) {
	t.Setenv("ADMIN_EMAIL", roleTestSuperAdminEmail)
	base := &userRepoStub{user: &User{ID: roleTestSuperAdminID, Email: roleTestSuperAdminEmail, Role: RoleAdmin}}
	repo := &rpmUserRepoStub{userRepoStub: base}
	svc := &adminServiceImpl{userRepo: repo, redeemCodeRepo: &redeemRepoStub{}}

	_, err := svc.UpdateUser(context.Background(), roleTestSuperAdminID, &UpdateUserInput{
		Email:        "other@example.com",
		ActorAdminID: roleTestSuperAdminID,
	})
	require.Error(t, err)
	require.Nil(t, repo.lastUpdated)
}

// roleGuardUserRepoStub supplies a controllable admin count for last-admin guard tests.
type roleGuardUserRepoStub struct {
	*rpmUserRepoStub
	adminTotal int64
	listCalls  int
}

func (s *roleGuardUserRepoStub) ListWithFilters(_ context.Context, _ pagination.PaginationParams, _ UserListFilters) ([]User, *pagination.PaginationResult, error) {
	s.listCalls++
	return nil, &pagination.PaginationResult{Total: s.adminTotal}, nil
}

func TestAdminService_UpdateUser_DemoteLastAdminRejected(t *testing.T) {
	base := &userRepoStub{user: &User{ID: 42, Email: "a@example.com", Role: RoleAdmin}}
	actorID := configureRoleTestSuperAdmin(t, base)
	repo := &roleGuardUserRepoStub{rpmUserRepoStub: &rpmUserRepoStub{userRepoStub: base}, adminTotal: 1}
	svc := &adminServiceImpl{userRepo: repo, redeemCodeRepo: &redeemRepoStub{}}

	_, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{Role: RoleUser, ActorAdminID: actorID})
	require.Error(t, err)
	require.Contains(t, err.Error(), "last admin")
	require.Nil(t, repo.lastUpdated, "最后一个管理员不应被降级持久化")
	require.Equal(t, 1, repo.listCalls, "降级路径应触发管理员计数")
}

func TestAdminService_UpdateUser_DemoteLastAdminToProviderRejected(t *testing.T) {
	base := &userRepoStub{user: &User{ID: 42, Email: "a@example.com", Role: RoleAdmin}}
	actorID := configureRoleTestSuperAdmin(t, base)
	repo := &roleGuardUserRepoStub{rpmUserRepoStub: &rpmUserRepoStub{userRepoStub: base}, adminTotal: 1}
	svc := &adminServiceImpl{userRepo: repo, redeemCodeRepo: &redeemRepoStub{}}

	_, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{Role: RoleProvider, ActorAdminID: actorID})
	require.Error(t, err)
	require.Contains(t, err.Error(), "last admin")
	require.Nil(t, repo.lastUpdated)
	require.Equal(t, 1, repo.listCalls)
}

func TestAdminService_UpdateUser_DemoteLastCombinedAdminToProviderRejected(t *testing.T) {
	base := &userRepoStub{user: &User{ID: 42, Email: "a@example.com", Role: RoleAdminProvider}}
	actorID := configureRoleTestSuperAdmin(t, base)
	repo := &roleGuardUserRepoStub{rpmUserRepoStub: &rpmUserRepoStub{userRepoStub: base}, adminTotal: 1}
	svc := &adminServiceImpl{userRepo: repo, redeemCodeRepo: &redeemRepoStub{}}

	_, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{Role: RoleProvider, ActorAdminID: actorID})
	require.Error(t, err)
	require.Contains(t, err.Error(), "last admin")
	require.Nil(t, repo.lastUpdated)
	require.Equal(t, 1, repo.listCalls)
}

func TestAdminService_UpdateUser_AddProviderCapabilityKeepsAdmin(t *testing.T) {
	base := &userRepoStub{user: &User{ID: 42, Email: "a@example.com", Role: RoleAdmin}}
	actorID := configureRoleTestSuperAdmin(t, base)
	repo := &roleGuardUserRepoStub{rpmUserRepoStub: &rpmUserRepoStub{userRepoStub: base}, adminTotal: 1}
	svc := &adminServiceImpl{userRepo: repo, redeemCodeRepo: &redeemRepoStub{}}

	updated, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{Role: RoleAdminProvider, ActorAdminID: actorID})
	require.NoError(t, err)
	require.True(t, updated.IsAdmin())
	require.True(t, updated.IsProvider())
	require.Equal(t, 0, repo.listCalls)
}

func TestAdminService_UpdateUser_DemoteAdminAllowedWhenOthersExist(t *testing.T) {
	base := &userRepoStub{user: &User{ID: 42, Email: "a@example.com", Role: RoleAdmin}}
	actorID := configureRoleTestSuperAdmin(t, base)
	repo := &roleGuardUserRepoStub{rpmUserRepoStub: &rpmUserRepoStub{userRepoStub: base}, adminTotal: 2}
	invalidator := &authCacheInvalidatorStub{}
	svc := &adminServiceImpl{
		userRepo:             repo,
		redeemCodeRepo:       &redeemRepoStub{},
		authCacheInvalidator: invalidator,
	}

	updated, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{Role: RoleUser, ActorAdminID: actorID})
	require.NoError(t, err)
	require.Equal(t, RoleUser, updated.Role)
	require.NotNil(t, repo.lastUpdated)
	require.Equal(t, RoleUser, repo.lastUpdated.Role, "存在其他管理员时允许降级")
}

func TestAdminService_UpdateUser_PromoteDoesNotCountAdmins(t *testing.T) {
	base := &userRepoStub{user: &User{ID: 42, Email: "u@example.com", Role: RoleUser}}
	actorID := configureRoleTestSuperAdmin(t, base)
	repo := &roleGuardUserRepoStub{rpmUserRepoStub: &rpmUserRepoStub{userRepoStub: base}, adminTotal: 1}
	svc := &adminServiceImpl{
		userRepo:             repo,
		redeemCodeRepo:       &redeemRepoStub{},
		authCacheInvalidator: &authCacheInvalidatorStub{},
	}

	updated, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{Role: RoleAdmin, ActorAdminID: actorID})
	require.NoError(t, err)
	require.Equal(t, RoleAdmin, updated.Role)
	require.Equal(t, 0, repo.listCalls, "升级路径不应触发管理员计数")
}
