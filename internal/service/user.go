package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/cache/rds"
	"github.com/erewhile/iam/internal/consts"
	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/dto/resp"
	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/logger"
	"github.com/erewhile/iam/internal/model"
	"github.com/erewhile/iam/internal/repository"
	"github.com/erewhile/iam/internal/token"
	"github.com/erewhile/iam/pkg/hash"
	"github.com/erewhile/iam/pkg/password"
	"github.com/erewhile/iam/pkg/utils"
	"github.com/google/uuid"
)

type UserService struct {
	repo         repository.UserRepository
	token        repository.TokenRepository
	roleRepo     repository.RoleRepository
	transactor   *repository.Transactor
	tokenCache   rds.TokenCache
	sessionCache rds.IAMSessionCache
	loginAttempt rds.LoginAttemptCache
}

func NewUserService(
	repo repository.UserRepository,
	token repository.TokenRepository,
	roleRepo repository.RoleRepository,
	transactor *repository.Transactor,
	tokenCache rds.TokenCache,
	sessionCache rds.IAMSessionCache,
	loginAttempt rds.LoginAttemptCache,
) *UserService {
	return &UserService{
		repo:         repo,
		token:        token,
		roleRepo:     roleRepo,
		transactor:   transactor,
		tokenCache:   tokenCache,
		sessionCache: sessionCache,
		loginAttempt: loginAttempt,
	}
}

func (s *UserService) Login(ctx context.Context, body req.UserLogin) (*token.TokenPair, string, error) {
	sec := config.Get().LoginSecurity
	ip := body.RequestMeta.IP

	locked, ttl, err := s.loginAttempt.IsLocked(ctx, body.Username)
	if err != nil {
		logger.Error("check login lock failed", err)
		return nil, "", errors.New("login failed, please try again later")
	}
	if locked {
		accountLocked := errors.New("account is temporarily locked due to too many failed login attempts")
		return nil, "", fmt.Errorf("%w: try again in %d seconds", accountLocked, int(ttl.Seconds()))
	}

	userInfo, err := s.repo.GetByUsername(ctx, body.Username)
	if err != nil {
		if db.IsNotFound(err) {
			s.recordFailure(ctx, body.Username, ip, sec)
			return nil, "", errors.New("user not found")
		}
		logger.Error("login failed", err)
		return nil, "", errors.New("login failed")
	}

	if userInfo.Status != model.UserStatusActive {
		return nil, "", errors.New("account is not active")
	}

	ok, err := password.Validate(body.Password, string(userInfo.PasswordHash))
	if err != nil {
		logger.Error("password check failed", err)
		return nil, "", errors.New("password check failed, please try again later")
	}
	if !ok {
		s.recordFailure(ctx, body.Username, ip, sec)
		return nil, "", errors.New("wrong password")
	}

	_ = s.loginAttempt.ResetFailure(ctx, body.Username)
	_ = s.loginAttempt.ResetFailureByIP(ctx, ip)

	roleCodes, err := s.getUserRoleCodes(ctx, userInfo.ID)
	if err != nil {
		logger.Error("get user roles failed", err)
		return nil, "", errors.New("login failed")
	}

	sessionID := uuid.New()
	userPayload := token.UserPayload{
		UserID:        userInfo.ID,
		UserUUID:      userInfo.UUID,
		ApplicationID: nil,
		Roles:         roleCodes,
	}

	tokenPair, err := s.issueTokenPair(ctx, userPayload, sessionID, body.RequestMeta)
	if err != nil {
		return nil, "", err
	}

	sid, err := s.StartSession(ctx, userInfo.ID, userInfo.UUID)
	if err != nil {
		logger.Error("start iam session failed", err)
		return nil, "", errors.New("login failed")
	}

	return tokenPair, sid, nil
}

func (s *UserService) recordFailure(ctx context.Context, username, ip string, sec config.LoginSecurity) {
	count, err := s.loginAttempt.IncrFailure(ctx, username, sec.AttemptWindow)
	if err != nil {
		logger.Error("incr login failure failed", err)
		return
	}
	if count >= int64(sec.MaxAttempts) {
		if err := s.loginAttempt.Lock(ctx, username, sec.LockoutDuration); err != nil {
			logger.Error("lock account failed", err)
		}
		_ = s.loginAttempt.ResetFailure(ctx, username)
	}

	if ip != "" {
		_, err := s.loginAttempt.IncrFailureByIP(ctx, ip, sec.AttemptWindow)
		if err != nil {
			logger.Error("incr login failure by ip failed", err)
		}
	}
}

func (s *UserService) LoginWithOAuthCode(ctx context.Context, payload *rds.OAuthCodePayload, applicationID int, meta req.RequestMeta) (*token.TokenPair, error) {
	roleCodes, err := s.getUserRoleCodes(ctx, payload.UserID)
	if err != nil {
		logger.Error("get user roles failed", err)
		return nil, errors.New("issue token failed")
	}

	userPayload := token.UserPayload{
		UserID:        payload.UserID,
		UserUUID:      payload.UserUUID,
		ApplicationID: &applicationID,
		Roles:         roleCodes,
	}
	return s.issueTokenPair(ctx, userPayload, payload.SessionID, meta)
}

func (s *UserService) Profile(ctx context.Context, userID int) {}

func (s *UserService) Refresh(ctx context.Context, body req.UserRefresh) (*token.TokenPair, error) {
	claims, payload, err := token.Validate(
		body.Token,
		body.RequestMeta,
		[]byte(config.Get().Token.Aad),
		token.TokenTypeRefresh,
	)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	online, err := s.tokenCache.ExistsRefresh(ctx, claims.SessionID)
	if err != nil || !online {
		return nil, errors.New("refresh token expired")
	}

	tokenHashed := hash.HashBlake2b256([]byte(body.Token))
	if !s.isTokenValid(ctx, tokenHashed, model.TokenTypeRefresh) {
		s.invalidateToken(ctx, claims.SessionID)
		return nil, errors.New("refresh token revoked or not found")
	}

	userInfo, err := s.repo.GetByID(ctx, payload.UserID)
	if err != nil {
		if db.IsNotFound(err) {
			s.invalidateToken(ctx, claims.SessionID)
			return nil, errors.New("user not found")
		}
		logger.Error("refresh failed", err)
		return nil, errors.New("refresh failed")
	}

	if userInfo.Status != model.UserStatusActive {
		s.invalidateToken(ctx, claims.SessionID)
		return nil, errors.New("account is not active")
	}

	roleCodes, err := s.getUserRoleCodes(ctx, payload.UserID)
	if err != nil {
		logger.Error("get user roles failed", err)
		return nil, errors.New("refresh failed")
	}

	if err := s.token.RevokeBySession(ctx, claims.SessionID); err != nil {
		logger.Error("revoke failed", err)
		return nil, errors.New("revoke failed")
	}

	s.invalidateToken(ctx, claims.SessionID)

	newSessionID := uuid.New()
	userPayload := token.UserPayload{
		UserID:        payload.UserID,
		UserUUID:      payload.UserUUID,
		ApplicationID: payload.ApplicationID,
		Roles:         roleCodes,
	}
	return s.issueTokenPair(ctx, userPayload, newSessionID, body.RequestMeta)
}

func (s *UserService) Logout(ctx context.Context, sessionID uuid.UUID, iamSID string) error {
	s.invalidateToken(ctx, sessionID)

	if err := s.token.RevokeBySession(ctx, sessionID); err != nil {
		logger.Error("logout failed", err)
		return errors.New("logout failed")
	}

	if iamSID != "" {
		if err := s.ClearSession(ctx, iamSID); err != nil {
			logger.Error("clear iam session failed", err)
		}
	}
	return nil
}

func (s *UserService) issueTokenPair(
	ctx context.Context,
	userPayload token.UserPayload,
	sessionID uuid.UUID,
	meta req.RequestMeta,
) (*token.TokenPair, error) {
	tokenPair, err := token.Generate(
		userPayload,
		sessionID,
		meta,
		[]byte(config.Get().Token.Aad),
	)
	if err != nil {
		logger.Error("generate token failed", err)
		return nil, errors.New("generate token failed")
	}

	accessJti := uuid.New()
	refreshJti := uuid.New()
	now := utils.Now()
	tokenCfg := config.Get().Token

	err = s.transactor.WithTx(ctx, func(ctx context.Context, txClient *db.Client) error {
		txTokenRepo := repository.NewTokenRepository(txClient)

		if err := txTokenRepo.Create(ctx, req.TokenCreate{
			UserID:    userPayload.UserID,
			Jti:       accessJti,
			SessionID: sessionID,
			Type:      model.TokenTypeAccess,
			TokenHash: hash.HashBlake2b256([]byte(tokenPair.AccessToken)),
			ExpiresAt: now.Add(tokenCfg.AccessTokenTTL),
			IP:        meta.IP,
			UserAgent: meta.UserAgent,
		}); err != nil {
			return err
		}

		return txTokenRepo.Create(ctx, req.TokenCreate{
			UserID:    userPayload.UserID,
			Jti:       refreshJti,
			SessionID: sessionID,
			Type:      model.TokenTypeRefresh,
			TokenHash: hash.HashBlake2b256([]byte(tokenPair.RefreshToken)),
			ExpiresAt: now.Add(tokenCfg.RefreshTokenTTL),
			IP:        meta.IP,
			UserAgent: meta.UserAgent,
		})
	})
	if err != nil {
		logger.Error("save token failed", err)
		return nil, errors.New("save token failed")
	}

	_ = s.tokenCache.SetAccess(ctx, sessionID, config.Get().Token.AccessTokenTTL)
	_ = s.tokenCache.SetRefresh(ctx, sessionID, config.Get().Token.RefreshTokenTTL)

	return tokenPair, nil
}

func (s *UserService) isTokenValid(ctx context.Context, hashed []byte, tt model.TokenType) bool {
	_, err := s.token.GetIfValid(ctx, hashed, tt)
	return err == nil
}

func (s *UserService) invalidateToken(ctx context.Context, sessionID uuid.UUID) {
	if err := s.tokenCache.DelAccess(ctx, sessionID); err != nil {
		logger.Error("del access cache failed", err)
	}
	if err := s.tokenCache.DelRefresh(ctx, sessionID); err != nil {
		logger.Error("del refresh cache failed", err)
	}
}

func (s *UserService) getUserRoleCodes(ctx context.Context, userID int) ([]string, error) {
	roles, err := s.roleRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(roles))
	for _, r := range roles {
		codes = append(codes, r.Code)
	}
	return codes, nil
}

func (s *UserService) StartSession(ctx context.Context, userID int, userUUID uuid.UUID) (string, error) {
	sid, err := utils.RandomString(32)
	if err != nil {
		return "", err
	}
	if err := s.sessionCache.Set(ctx, sid, rds.IAMSessionPayload{
		UserID:   userID,
		UserUUID: userUUID,
	}, config.Get().Session.CookieTTL); err != nil {
		return "", err
	}
	return sid, nil
}

func (s *UserService) CheckSession(ctx context.Context, sid string) (userID int, userUUID uuid.UUID, ok bool) {
	if sid == "" {
		return 0, uuid.Nil, false
	}
	payload, err := s.sessionCache.Get(ctx, sid)
	if err != nil || payload == nil {
		return 0, uuid.Nil, false
	}

	totalTTL := config.Get().Session.CookieTTL
	now := utils.Now().Unix()
	remainingTime := payload.ExpiredAt - now
	halfTTL := int64(totalTTL.Seconds() / 2)

	if remainingTime < halfTTL {
		asyncCtx := context.Background()

		go func(p rds.IAMSessionPayload) {
			if err := s.sessionCache.Refresh(asyncCtx, sid, &p, totalTTL); err != nil {
				logger.Error(fmt.Sprintf("async refresh session failed for sid: %s", sid), err)
			}
		}(*payload)
	}

	return payload.UserID, payload.UserUUID, true
}

func (s *UserService) ClearSession(ctx context.Context, sid string) error {
	if sid == "" {
		return nil
	}
	return s.sessionCache.Del(ctx, sid)
}

func (s *UserService) List(ctx context.Context, params req.UserList) ([]resp.UserListItem, int, error) {
	content, count, err := s.repo.List(ctx, params)
	if err != nil {
		logger.Error("failed to retrieve the list", err)
		return nil, 0, errors.New("failed to retrieve the list")
	}

	return content, count, nil
}

func (s *UserService) Info(ctx context.Context, params req.InfoPathParams) (*resp.UserInfo, error) {
	userInfo, err := s.repo.GetByID(ctx, params.ID)

	if err != nil {
		if db.IsNotFound(err) {
			return nil, errors.New("user not found")
		}
		logger.Error("failed to get user info", err)
		return nil, errors.New("failed to get user info")
	}

	if userInfo.Status != model.UserStatusActive {
		return nil, errors.New("account is disabled")
	}

	return &resp.UserInfo{
		ID:           userInfo.ID,
		Username:     userInfo.Username,
		Email:        userInfo.Email,
		UUID:         userInfo.UUID,
		StatusDetail: userInfo.Status.String(),
	}, nil
}

func (s *UserService) Create(ctx context.Context, body req.UserCreate) error {
	if !body.Status.IsValid() {
		return errors.New("invalid user status")
	}

	exists, err := s.repo.Duplicate(ctx, body.Username, body.Email)
	if err != nil {
		logger.Error("failed to check if user exists", err)
		return errors.New("failed to check if user exists")
	}

	if exists {
		return errors.New("username or email already exists")
	}

	hashed, err := password.Hash(body.Password)
	if err != nil {
		logger.Error("failed to hash password", err)
		return errors.New("failed to hash password")
	}

	_, err = s.repo.Create(ctx, body, hashed, model.UserStandard)
	if err != nil {
		logger.Error("failed to create user", err)
		return errors.New("failed to create user")
	}
	return nil
}

func (s *UserService) Update(ctx context.Context, params req.UserUpdatePathParams, body req.UserUpdate) error {
	if !body.Status.IsValid() {
		return errors.New("invalid user status")
	}

	if body.Password != "" && len(body.Password) < 6 {
		// return errors.New("password must be greater than 6 characters")
		return errors.New("password must be at least 6 characters long")
	}

	userInfo, err := s.repo.GetByID(ctx, params.UserID)
	if err != nil {
		if db.IsNotFound(err) {
			return errors.New("user not found")
		}
		logger.Error("get user failed", err)
		return errors.New("failed to get user info")
	}

	if body.Status == model.UserStatusDisabled && userInfo.Status == model.UserStatusActive {
		if err := s.ensureNotLastAdmin(ctx, userInfo.ID); err != nil {
			return err
		}
	}

	exists, err := s.repo.Duplicate(ctx, body.Username, body.Email, params.UserID)
	if err != nil {
		logger.Error("failed to check if user exists", err)
		return errors.New("failed to check if user exists")
	}
	if exists {
		return errors.New("username or email already exists")
	}

	var hashed string
	if body.Password != "" {
		hashed, err = password.Hash(body.Password)
		if err != nil {
			logger.Error("failed to hash password", err)
			return errors.New("failed to hash password")
		}
	}

	_, err = s.repo.Update(ctx, params, body, hashed)
	if err != nil {
		logger.Error("failed to update user", err)
		return errors.New("failed to update user")
	}

	if body.Status == model.UserStatusDisabled {
		_ = s.InvalidateAllSessions(ctx, params.UserID)
	}

	return nil
}

func (s *UserService) Delete(ctx context.Context, params req.DeletePathParams) error {
	userInfo, err := s.repo.GetByID(ctx, params.ID)
	if err != nil {
		if db.IsNotFound(err) {
			return errors.New("user not found")
		}
		logger.Error("get user failed", err)
		return errors.New("failed to get user info")
	}

	if userInfo.IsSystem {
		return errors.New("the system cannot delete")
	}

	if err := s.ensureNotLastAdmin(ctx, userInfo.ID); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, params); err != nil {
		logger.Error("failed to delete user", err)
		return errors.New("failed to delete user")
	}
	return nil
}

func (s *UserService) ensureNotLastAdmin(ctx context.Context, userID int) error {
	isAdmin, err := s.roleRepo.UserHasRole(ctx, userID, consts.RoleSuperAdmin)
	if err != nil {
		return errors.New("failed to verify user role")
	}
	if !isAdmin {
		return nil
	}
	count, err := s.roleRepo.CountUsersByRoleCode(ctx, consts.RoleSuperAdmin)
	if err != nil {
		return errors.New("failed to verify admin count")
	}
	if count <= 1 {
		return errors.New("cannot delete or disable the last admin account")
	}
	return nil
}

func (s *UserService) InvalidateAllSessions(ctx context.Context, userID int) error {
	if err := s.token.RevokeAllByUser(ctx, userID); err != nil {
		return fmt.Errorf("revoke tokens failed: %w", err)
	}

	sessionIDs, err := s.token.ListActiveSessionsByUser(ctx, userID)
	if err != nil {
		logger.Error("list active sessions failed", err)
	}
	for _, sid := range sessionIDs {
		s.invalidateToken(ctx, sid)
	}

	if err := s.sessionCache.DelAllByUser(ctx, userID); err != nil {
		logger.Error("clear iam sessions failed", err)
	}

	return nil
}
