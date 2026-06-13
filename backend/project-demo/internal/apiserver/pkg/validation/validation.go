package validation

import (
	"context"
	"regexp"

	"github.com/google/wire"

	"github.com/mungdong/devkit/internal/apiserver/store"
	"github.com/mungdong/devkit/internal/pkg/errno"
	v1 "github.com/mungdong/devkit/pkg/api/apiserver/v1"
)

// Validator handles custom business validation logic.
// It holds dependencies required for deep validation, such as database access.
type Validator struct {
	// Some complex validation logic may require direct database queries.
	// This is just an example. If validation requires other dependencies
	// like clients, services, resources, etc., they can all be injected here.
	store store.IStore
}

// Use globally precompiled regular expressions to avoid creating and compiling them repeatedly.
var (
	lengthRegex = regexp.MustCompile(`^.{3,20}$`)                                        // Length between 3 and 20 characters
	validRegex  = regexp.MustCompile(`^[A-Za-z0-9_]+$`)                                  // Only letters, numbers, and underscores
	letterRegex = regexp.MustCompile(`[A-Za-z]`)                                         // At least one letter
	numberRegex = regexp.MustCompile(`\d`)                                               // At least one number
	emailRegex  = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`) // Email format
	phoneRegex  = regexp.MustCompile(`^1[3-9]\d{9}$`)                                    // Chinese phone number
)

// ProviderSet is the Wire provider set for the validation package.
var ProviderSet = wire.NewSet(New, wire.Bind(new(any), new(*Validator)))

// New creates and initializes a new Validator instance with the required dependencies.
func New(ds store.IStore) *Validator {
	return &Validator{store: ds}
}

// isValidUsername validates if a username is valid.
func isValidUsername(username string) bool {
	// Validate length
	if !lengthRegex.MatchString(username) {
		return false
	}
	// Validate character legality
	if !validRegex.MatchString(username) {
		return false
	}
	return true
}

// isValidPassword checks whether a password meets complexity requirements.
func isValidPassword(password string) error {
	switch {
	// Check if the new password is empty
	case password == "":
		return errno.ErrInvalidArgument.WithMessage("password cannot be empty")
	// Check the length requirement of the new password
	case len(password) < 6:
		return errno.ErrInvalidArgument.WithMessage("password must be at least 6 characters long")
	// Use a regular expression to check if it contains at least one letter
	case !letterRegex.MatchString(password):
		return errno.ErrInvalidArgument.WithMessage("password must contain at least one letter")
	// Use a regular expression to check if it contains at least one number
	case !numberRegex.MatchString(password):
		return errno.ErrInvalidArgument.WithMessage("password must contain at least one number")
	}
	return nil
}

// isValidEmail checks whether an email is valid.
func isValidEmail(email string) error {
	// Check if the email is empty
	if email == "" {
		return errno.ErrInvalidArgument.WithMessage("email cannot be empty")
	}

	// Validate email format using a regular expression
	if !emailRegex.MatchString(email) {
		return errno.ErrInvalidArgument.WithMessage("invalid email format")
	}

	return nil
}

// isValidPhone checks whether a phone number is valid.
func isValidPhone(phone string) error {
	// Check if the phone number is empty
	if phone == "" {
		return errno.ErrInvalidArgument.WithMessage("phone cannot be empty")
	}

	// Validate the phone number format (assumed to be a Chinese phone number, 11 digits)
	if !phoneRegex.MatchString(phone) {
		return errno.ErrInvalidArgument.WithMessage("invalid phone format")
	}

	return nil
}

// ValidateLoginRequest validates the LoginRequest parameters.
func (v *Validator) ValidateLoginRequest(ctx context.Context, rq *v1.LoginRequest) error {
	if !isValidUsername(rq.GetUsername()) {
		return errno.ErrInvalidArgument.WithMessage("username must be 3-20 characters and contain only letters, numbers, and underscores")
	}
	if rq.GetPassword() == "" {
		return errno.ErrInvalidArgument.WithMessage("password cannot be empty")
	}
	return nil
}

// ValidateChangePasswordRequest validates the ChangePasswordRequest parameters.
func (v *Validator) ValidateChangePasswordRequest(ctx context.Context, rq *v1.ChangePasswordRequest) error {
	if err := isValidPassword(rq.GetOldPassword()); err != nil {
		return err
	}
	if err := isValidPassword(rq.GetNewPassword()); err != nil {
		return err
	}
	return nil
}

// ValidateCreateUserRequest validates the CreateUserRequest parameters.
func (v *Validator) ValidateCreateUserRequest(ctx context.Context, rq *v1.CreateUserRequest) error {
	if !isValidUsername(rq.GetUsername()) {
		return errno.ErrInvalidArgument.WithMessage("username must be 3-20 characters and contain only letters, numbers, and underscores")
	}
	if err := isValidPassword(rq.GetPassword()); err != nil {
		return err
	}
	return nil
}

// ValidateUpdateUserRequest validates the UpdateUserRequest parameters.
func (v *Validator) ValidateUpdateUserRequest(ctx context.Context, rq *v1.UpdateUserRequest) error {
	return nil
}

// ValidateDeleteUserRequest validates the DeleteUserRequest parameters.
func (v *Validator) ValidateDeleteUserRequest(ctx context.Context, rq *v1.DeleteUserRequest) error {
	if rq.GetUserID() == "" {
		return errno.ErrInvalidArgument.WithMessage("userID cannot be empty")
	}
	return nil
}

// ValidateGetUserRequest validates the GetUserRequest parameters.
func (v *Validator) ValidateGetUserRequest(ctx context.Context, rq *v1.GetUserRequest) error {
	if rq.GetUserID() == "" {
		return errno.ErrInvalidArgument.WithMessage("userID cannot be empty")
	}
	return nil
}

// ValidateListUserRequest validates the ListUserRequest parameters.
func (v *Validator) ValidateListUserRequest(ctx context.Context, rq *v1.ListUserRequest) error {
	return nil
}
