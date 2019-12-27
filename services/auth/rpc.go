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
	token, err := EncodeSessionToken(SigningKey(), tenantID, entityID, sessionID)
	if err != nil {
		s.logger.Error("failed to encode jwt", zap.String("username", in.Protocol.Username), zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.logger.Info("user authentication succeeded", zap.String("username", in.Protocol.Username), zap.String("session_id", sessionID))
	return &pb.CreateTokenOutput{
		JWT:       token,
		Tenant:    tenantID,
		EntityID:  entityID,
		SessionID: sessionID,
	}, nil
}
