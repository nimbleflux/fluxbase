package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/email"
	"github.com/rs/zerolog/log"
)

type InvitationHandler struct {
	invitationService *auth.InvitationService
	dashboardAuth     *auth.DashboardAuthService
	emailService      email.Service
	baseURL           string
}

func NewInvitationHandler(
	invitationService *auth.InvitationService,
	dashboardAuth *auth.DashboardAuthService,
	emailService email.Service,
	baseURL string,
) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
		dashboardAuth:     dashboardAuth,
		emailService:      emailService,
		baseURL:           baseURL,
	}
}

func invitationErrorDetails(err error) string {
	switch {
	case errors.Is(err, auth.ErrInvitationExpired):
		return "Invitation has expired"
	case errors.Is(err, auth.ErrInvitationAlreadyAccepted):
		return "Invitation has already been accepted"
	case errors.Is(err, auth.ErrInvitationNotFound):
		return "Invitation not found"
	default:
		return "Invalid token"
	}
}

func invitationErrorStatus(err error) int {
	switch {
	case errors.Is(err, auth.ErrInvitationExpired):
		return 410
	case errors.Is(err, auth.ErrInvitationAlreadyAccepted):
		return 409
	case errors.Is(err, auth.ErrInvitationNotFound):
		return 404
	default:
		return 400
	}
}

type CreateInvitationRequest struct {
	Email          string `json:"email" validate:"required,email"`
	Role           string `json:"role" validate:"required,oneof=instance_admin tenant_admin"`
	ExpiryDuration int64  `json:"expiry_duration,omitempty"`
}

type CreateInvitationResponse struct {
	Invitation  *auth.InvitationToken `json:"invitation"`
	InviteLink  string                `json:"invite_link"`
	EmailSent   bool                  `json:"email_sent"`
	EmailStatus string                `json:"email_status,omitempty"`
}

type ValidateInvitationResponse struct {
	Valid      bool                  `json:"valid"`
	Invitation *auth.InvitationToken `json:"invitation,omitempty"`
	Error      string                `json:"error,omitempty"`
}

type AcceptInvitationRequest struct {
	Password string `json:"password" validate:"required,min=12"`
	Name     string `json:"name" validate:"required,min=2"`
}

type AcceptInvitationResponse struct {
	User         *auth.DashboardUser `json:"user"`
	AccessToken  string              `json:"access_token"`
	RefreshToken string              `json:"refresh_token"`
	ExpiresIn    int64               `json:"expires_in"`
}

func (h *InvitationHandler) CreateInvitation(c fiber.Ctx) error {
	ctx := context.Background()

	inviterID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	inviterUUID, err := uuid.Parse(inviterID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req CreateInvitationRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := auth.ValidateDashboardRole(req.Role); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	inviterRole, ok := c.Locals("user_role").(string)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User role not found",
		})
	}

	if req.Role == "tenant_admin" && inviterRole != "instance_admin" {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Only instance_admin can invite tenant_admin users",
		})
	}

	expiryDuration := 7 * 24 * time.Hour
	if req.ExpiryDuration > 0 {
		expiryDuration = time.Duration(req.ExpiryDuration) * time.Second
	}

	invitation, err := h.invitationService.CreateInvitation(ctx, req.Email, req.Role, &inviterUUID, expiryDuration)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create invitation: %v", err),
		})
	}

	inviteLink := fmt.Sprintf("%s/invite/%s", h.baseURL, invitation.Token)

	emailSent := false
	emailStatus := ""

	if h.emailService != nil {
		inviterName := "An administrator"
		if err := h.emailService.SendInvitationEmail(ctx, req.Email, inviterName, inviteLink); err != nil {
			log.Warn().Err(err).Str("email", req.Email).Msg("Failed to send invitation email")
			emailStatus = fmt.Sprintf("Failed to send email: %v. Share the invite link manually.", err)
		} else {
			emailSent = true
			emailStatus = "Invitation email sent successfully"
		}
	} else {
		emailStatus = "Email service not configured. Share the invite link manually."
	}

	return c.Status(http.StatusCreated).JSON(CreateInvitationResponse{
		Invitation:  invitation,
		InviteLink:  inviteLink,
		EmailSent:   emailSent,
		EmailStatus: emailStatus,
	})
}

func (h *InvitationHandler) ValidateInvitation(c fiber.Ctx) error {
	ctx := context.Background()

	token := c.Params("token")
	if token == "" {
		return c.Status(http.StatusBadRequest).JSON(ValidateInvitationResponse{
			Valid: false,
			Error: "Token is required",
		})
	}

	invitation, err := h.invitationService.ValidateToken(ctx, token)
	if err != nil {
		return c.JSON(ValidateInvitationResponse{
			Valid: false,
			Error: invitationErrorDetails(err),
		})
	}

	invitation.Token = ""

	return c.JSON(ValidateInvitationResponse{
		Valid:      true,
		Invitation: invitation,
	})
}

func (h *InvitationHandler) AcceptInvitation(c fiber.Ctx) error {
	ctx := context.Background()

	token := c.Params("token")
	if token == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Token is required",
		})
	}

	var req AcceptInvitationRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := auth.ValidateDashboardPassword(req.Password); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	invitation, err := h.invitationService.ValidateToken(ctx, token)
	if err != nil {
		return c.Status(invitationErrorStatus(err)).JSON(fiber.Map{
			"error": invitationErrorDetails(err),
		})
	}

	_, err = h.dashboardAuth.GetDB().Exec(ctx, "SET LOCAL app.invitation_token = $1", token)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to set session context",
		})
	}

	user, err := h.dashboardAuth.CreateUser(ctx, invitation.Email, req.Password, req.Name)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create user: %v", err),
		})
	}

	_, err = h.dashboardAuth.GetDB().Exec(ctx, `
		UPDATE platform.users
		SET role = $1, email_verified = true
		WHERE id = $2
	`, invitation.Role, user.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to set user role and verify email",
		})
	}

	if err := h.invitationService.AcceptInvitation(ctx, token); err != nil {
		log.Warn().
			Err(err).
			Str("token", token).
			Str("email", invitation.Email).
			Msg("Failed to mark invitation as accepted")
	}

	loggedInUser, loginResp, err := h.dashboardAuth.Login(ctx, invitation.Email, req.Password, nil, c.Get("User-Agent"))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "User created but failed to generate access token",
		})
	}

	return c.Status(http.StatusCreated).JSON(AcceptInvitationResponse{
		User:         loggedInUser,
		AccessToken:  loginResp.AccessToken,
		RefreshToken: loginResp.RefreshToken,
		ExpiresIn:    loginResp.ExpiresIn,
	})
}

func (h *InvitationHandler) ListInvitations(c fiber.Ctx) error {
	ctx := context.Background()

	includeAccepted := c.Query("include_accepted", "false") == "true"
	includeExpired := c.Query("include_expired", "false") == "true"

	invitations, err := h.invitationService.ListInvitations(ctx, includeAccepted, includeExpired)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list invitations",
		})
	}

	for i := range invitations {
		invitations[i].Token = ""
	}

	return c.JSON(fiber.Map{
		"invitations": invitations,
	})
}

func (h *InvitationHandler) RevokeInvitation(c fiber.Ctx) error {
	ctx := context.Background()

	token := c.Params("token")
	if token == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Token is required",
		})
	}

	if err := h.invitationService.RevokeInvitation(ctx, token); err != nil {
		if errors.Is(err, auth.ErrInvitationNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Invitation not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to revoke invitation",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Invitation revoked successfully",
	})
}
