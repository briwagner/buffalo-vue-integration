package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const minPasswordLength int = 8

// User is a generated model from buffalo-auth, it serves as the base for username/password authentication.
type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"password_hash" db:"password_hash"`

	Password             string       `json:"-" db:"-"`
	PasswordConfirmation string       `json:"-" db:"-"`
	RecoveryCode         nulls.String `json:"-" db:"recovery_code"`
	RecoveryExp          nulls.Time   `json:"-" db:"recovery_expiration"`
}

// Create wraps up the pattern of encrypting the password and
// running validations. Useful when writing tests.
func (u *User) Create(tx *pop.Connection) (*validate.Errors, error) {
	verrs := u.validatePassword()
	if verrs.HasAny() {
		return verrs, nil
	}

	ph, err := encryptPassword(u.Password)
	if err != nil {
		return validate.NewErrors(), errors.WithStack(err)
	}
	u.PasswordHash = string(ph)
	u = u.sanitizeFields()
	return tx.ValidateAndCreate(u)
}

// Update handles the extra work possibly needed during user update,
// hashing password and making email lowercase for consistency.
func (u *User) Update(tx *pop.Connection) (*validate.Errors, error) {
	if len(u.Password) != 0 {

		verrs := u.validatePassword()
		if verrs.HasAny() {
			return verrs, nil
		}
		ph, err := encryptPassword(u.Password)
		if err != nil {
			return validate.NewErrors(), errors.WithStack(err)
		}
		u.PasswordHash = ph
	}
	u = u.sanitizeFields()
	return tx.ValidateAndUpdate(u)
}

func (u *User) sanitizeFields() *User {
	// force email to lowercase for better matching
	u.Email = strings.ToLower(u.Email)

	// wipe out password field after it's been hashed
	u.Password = ""
	u.PasswordConfirmation = ""
	return u
}

// String is not required by pop and may be deleted
func (u User) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Users is not required by pop and may be deleted
type Users []User

// String is not required by pop and may be deleted
func (u Users) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (u *User) Validate(tx *pop.Connection) (*validate.Errors, error) {
	var err error
	return validate.Validate(
		&validators.StringIsPresent{Field: u.Email, Name: "Email"},
		&validators.StringIsPresent{Field: u.PasswordHash, Name: "PasswordHash"},
		// check to see if the email address is already taken:
		&validators.FuncValidator{
			Field:   u.Email,
			Name:    "Email",
			Message: "%s is already taken",
			Fn: func() bool {
				var b bool
				q := tx.Where("email = ?", u.Email)
				if u.ID != uuid.Nil {
					q = q.Where("id != ?", u.ID)
				}
				b, err = q.Exists(u)
				if err != nil {
					return false
				}
				return !b
			},
		},
	), err
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *User) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

func (u *User) validatePassword() *validate.Errors {
	passwordLengthValidator := validators.StringLengthInRange{
		Field:   u.Password,
		Name:    "Password",
		Message: fmt.Sprintf("Password is not long enough. Minimum of %d characters required.", minPasswordLength),
		Min:     minPasswordLength,
		Max:     0,
	}
	passwordConfirmValidator := validators.StringsMatch{
		Field:   u.Password,
		Name:    "PasswordConfirmation",
		Message: "Password does not match confirmation",
		Field2:  u.PasswordConfirmation,
	}
	verrs := validate.NewErrors()
	passwordLengthValidator.IsValid(verrs)
	passwordConfirmValidator.IsValid(verrs)
	return verrs
}

func encryptPassword(password string) (string, error) {
	ph, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(ph), err
}

// RecoveryUpdate request struct for using recovery code
type RecoveryUpdate struct {
	Email                string `json:"email"`
	Code                 string `json:"code"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirm"`
}

// Validate can be run to verify a recovery request
func (ru *RecoveryUpdate) Validate() *validate.Errors {
	verrs := validate.Validate(
		&validators.StringIsPresent{Field: ru.Code, Name: "Code"},
		&validators.StringIsPresent{Field: ru.Password, Name: "Password"},
		&validators.StringIsPresent{Field: ru.PasswordConfirmation, Name: "Password confirmation"},
		&validators.StringsMatch{
			Name:    "Password",
			Field:   ru.Password,
			Field2:  ru.PasswordConfirmation,
			Message: "Password and confirmation must match",
		},
	)
	if verrs.HasAny() {
		ru.sanitizeFields()
	}
	return verrs
}

// sanitizeFields wipes the password fields set by user
func (ru *RecoveryUpdate) sanitizeFields() {
	// wipe out password fields.
	ru.Password = ""
	ru.PasswordConfirmation = ""
}
