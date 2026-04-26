package auth

import (
	"context"
	"fmt"
)

// DefaultOTPSender implements OTPSender using email service
type DefaultOTPSender struct {
	emailService EmailService
	fromAddress  string
	appName      string
}

// NewDefaultOTPSender creates a new OTP sender
func NewDefaultOTPSender(emailService EmailService, fromAddress, appName string) *DefaultOTPSender {
	if appName == "" {
		appName = "Fluxbase"
	}
	if fromAddress == "" {
		fromAddress = "noreply@fluxbase.app"
	}
	return &DefaultOTPSender{
		emailService: emailService,
		fromAddress:  fromAddress,
		appName:      appName,
	}
}

// SendEmailOTP sends an OTP code via email
func (s *DefaultOTPSender) SendEmailOTP(ctx context.Context, to, code, purpose string) error {
	// Check if email service is available
	if s.emailService == nil {
		// Log but don't fail - OTP is still stored and can be verified
		// This allows OTP to work in test environments without email configuration
		return nil
	}

	subject := s.getEmailSubject(purpose)
	body := s.getEmailBody(code, purpose)

	if err := s.emailService.Send(ctx, to, subject, body); err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	return nil
}

// SendSMSOTP sends an OTP code via SMS
func (s *DefaultOTPSender) SendSMSOTP(ctx context.Context, to, code, purpose string) error {
	// SMS sending not yet implemented - would require Twilio/similar integration
	return fmt.Errorf("SMS OTP sending is not yet implemented")
}

// getEmailSubject returns the email subject based on purpose
func (s *DefaultOTPSender) getEmailSubject(purpose string) string {
	switch purpose {
	case "signin":
		return fmt.Sprintf("Your %s Sign In Code", s.appName)
	case "signup":
		return fmt.Sprintf("Verify your %s account", s.appName)
	case "recovery":
		return fmt.Sprintf("Your %s Account Recovery Code", s.appName)
	case "email_change":
		return fmt.Sprintf("Verify your new %s email", s.appName)
	case "phone_change":
		return fmt.Sprintf("Verify your new %s phone", s.appName)
	default:
		return fmt.Sprintf("Your %s Verification Code", s.appName)
	}
}

// getEmailBody returns the email body based on purpose
func (s *DefaultOTPSender) getEmailBody(code, purpose string) string {
	var action string
	switch purpose {
	case "signin":
		action = "sign in to your account"
	case "signup":
		action = "complete your account registration"
	case "recovery":
		action = "recover your account"
	case "email_change":
		action = "verify your new email address"
	case "phone_change":
		action = "verify your new phone number"
	default:
		action = "complete verification"
	}

	return fmt.Sprintf(`Hello,

Your verification code is: %s

Use this code to %s. This code will expire in 15 minutes.

If you didn't request this code, please ignore this email.

Best regards,
The %s Team`, code, action, s.appName)
}

// NoOpOTPSender is a no-op OTP sender for testing
type NoOpOTPSender struct{}

// SendEmailOTP does nothing
func (s *NoOpOTPSender) SendEmailOTP(ctx context.Context, to, code, purpose string) error {
	return nil
}

// SendSMSOTP does nothing
func (s *NoOpOTPSender) SendSMSOTP(ctx context.Context, to, code, purpose string) error {
	return nil
}
