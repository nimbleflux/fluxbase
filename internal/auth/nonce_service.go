package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type NonceService struct {
	nonceRepo *NonceRepository
	userRepo  *UserRepository
}

func NewNonceService(nonceRepo *NonceRepository, userRepo *UserRepository) *NonceService {
	return &NonceService{
		nonceRepo: nonceRepo,
		userRepo:  userRepo,
	}
}

func (n *NonceService) Reauthenticate(ctx context.Context, userID string) (string, error) {
	_, err := n.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	nonce := uuid.New().String()

	if err := n.nonceRepo.Set(ctx, nonce, userID, 5*time.Minute); err != nil {
		return "", fmt.Errorf("failed to store nonce: %w", err)
	}

	return nonce, nil
}

func (n *NonceService) VerifyNonce(ctx context.Context, nonce, userID string) bool {
	valid, err := n.nonceRepo.Validate(ctx, nonce, userID)
	if err != nil {
		return false
	}
	return valid
}

func (n *NonceService) CleanupExpiredNonces(ctx context.Context) (int64, error) {
	return n.nonceRepo.Cleanup(ctx)
}
