package auth

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
	"github.com/vx-labs/mqtt-broker/services/auth/pb"
	"github.com/vx-labs/mqtt-broker/services/auth/providers"
)

func splitUsername(username string) (string, string) {
	idx := strings.LastIndexByte(username, '@')
	if len(username) == 1 {
		return username, username
	}
	if idx == -1 {
		return username, username
	}
	return username[:idx], username[idx+1:]
}

func (s *server) CreateToken(ctx context.Context, in *pb.CreateTokenInput) (*pb.CreateTokenOutput, error) {
	entityID, tenantID := splitUsername(in.Protocol.Username)
	entity, err := s.store.GetEntitiesByID(tenantID, entityID)
	if err != nil {
		s.logger.Info("failed to fetch tenant for user", zap.String("username", in.Protocol.Username))
		return nil, status.Error(codes.FailedPrecondition, "authentication failed")
	}
	didAuthSucceeded := providers.Get(entity.AuthProvider).Authenticate(tenantID, entityID, in.Protocol, in.Transport)
	if !didAuthSucceeded {
		s.logger.Info("user authentication failed", zap.String("username", in.Protocol.Username))
		return nil, status.Error(codes.FailedPrecondition, "authentication failed")
	}
	sessionID := uuid.New().String()
	var interval int64 = 60
	if in.Protocol.KeepaliveInterval > 0 {
		interval = in.Protocol.KeepaliveInterval
	}
	token, err := EncodeSessionToken(SigningKey(), tenantID, entityID, sessionID, interval)
	if err != nil {
		s.logger.Error("failed to encode jwt", zap.String("username", in.Protocol.Username), zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	refreshToken, err := EncodeRefreshToken(SigningKey(), tenantID, entityID, sessionID, interval)
	if err != nil {
		s.logger.Error("failed to encode refresh token", zap.String("username", in.Protocol.Username), zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.logger.Info("user authentication succeeded", zap.String("username", in.Protocol.Username), zap.String("session_id", sessionID))
	return &pb.CreateTokenOutput{
		JWT:          token,
		Tenant:       tenantID,
		EntityID:     entityID,
		SessionID:    sessionID,
		RefreshToken: refreshToken,
	}, nil
}
func (s *server) RefreshToken(ctx context.Context, in *pb.RefreshTokenInput) (*pb.RefreshTokenOutput, error) {
	token, err := DecodeToken(SigningKey(), in.RefreshToken)
	if err != nil {
		return nil, err
	}
	if !token.VerifyIssuer("mqtt-auth", true) || !token.VerifyAudience("mqtt-auth", true) {
		return nil, status.Error(codes.Internal, "invalid token")
	}
	var interval int64 = 60
	if token.RefreshInterval > 0 {
		interval = token.RefreshInterval
	}
	idToken, err := EncodeSessionToken(SigningKey(), token.SessionTenant, token.SessionEntity, token.SessionID, interval)
	if err != nil {
		s.logger.Error("failed to encode refeshed jwt", zap.String("session_id", token.SessionID), zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RefreshTokenOutput{
		JWT:       idToken,
		Tenant:    token.SessionTenant,
		EntityID:  token.SessionEntity,
		SessionID: token.SessionID,
	}, nil
}
