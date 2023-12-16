package actions

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/pkg/errors"

	"event_planner/models"
)

// UsersNew renders the users form.
func UsersNew(c buffalo.Context) error {
	if uid := c.Session().Get("current_user_id"); uid != nil {
		c.Flash().Add("info", "You are logged in")
		return c.Redirect(302, "/")
	}

	u := models.User{}
	c.Set("user", u)
	return c.Render(http.StatusOK, r.HTML("users/new.plush.html"))
}

// UserRecovery returns the html form for requesting a
// recovery code
func PasswordResetForm(c buffalo.Context) error {
	req := RecoveryRequest{}
	c.Set("req", req)
	return c.Render(http.StatusOK, r.HTML("users/recovery.html"))
}

// UserRecover returns the html form for using the recovery code
func AccountRecoveryForm(c buffalo.Context) error {
	req := RecoveryUpdate{}
	c.Set("req", req)
	return c.Render(http.StatusOK, r.HTML("users/recover.html"))
}

// UsersCreate registers a new user with the application.
func UsersCreate(c buffalo.Context) error {
	u := &models.User{}
	if err := c.Bind(u); err != nil {
		return errors.WithStack(err)
	}

	tx := c.Value("tx").(*pop.Connection)
	verrs, err := u.Create(tx)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("user", u)
		c.Set("errors", verrs)
		return c.Render(http.StatusOK, r.HTML("users/new.plush.html"))
	}

	c.Session().Set("current_user_id", u.ID)
	c.Flash().Add("success", "Welcome to event-planner")

	return c.Redirect(http.StatusFound, "/")
}

// SetCurrentUser attempts to find a user based on the current_user_id
// in the session. If one is found it is set on the context.
func SetCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if uid := c.Session().Get("current_user_id"); uid != nil {
			u := &models.User{}
			tx := c.Value("tx").(*pop.Connection)
			err := tx.Find(u, uid)
			if err != nil {
				c.Session().Delete("current_user_id")
				c.Session().Set("redirectURL", c.Request().URL.String())
				c.Flash().Add("danger", "You must be authorized with a correct user to see that page")
				return c.Redirect(http.StatusFound, "loginPath()")
			}
			c.Set("current_user", u)
		}
		return next(c)
	}
}

// Authorize require a user be logged in before accessing a route
func Authorize(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if uid := c.Session().Get("current_user_id"); uid == nil {
			c.Session().Set("redirectURL", c.Request().URL.String())

			err := c.Session().Save()
			if err != nil {
				return errors.WithStack(err)
			}

			c.Flash().Add("danger", "You must be authorized to see that page")
			return c.Redirect(http.StatusFound, "loginPath()")
		}
		return next(c)
	}
}

// Sender interface for the recovery sender
type Sender interface {
	Send(map[string]interface{}) error
}

// MockSender mock sender, should be customized!
type MockSender struct{}

// Send mock sender, should be customized!
func (s MockSender) Send(data map[string]interface{}) error {
	fmt.Printf("Sending data:%+v....", data)
	return nil
}

// SetupRecoverySender sets the sender on the context.
func SetupRecoverySender(s Sender) func(next buffalo.Handler) buffalo.Handler {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			c.Set("recovery_sender", s)
			return next(c)
		}
	}
}

func verifyCode() string {
	i := rand.Intn(1000000)
	code := strconv.Itoa(i)

	for len(code) < 6 {
		code = "0" + code
	}

	return code
}

// RecoveryRequest request struct for a recovery code
type RecoveryRequest struct {
	Email string `json:"email"`
}

// UserRequestRecovery handles requesting recovery for an account
// when user provides matching email
func PasswordReset(c buffalo.Context) error {
	req := &RecoveryRequest{}

	if err := c.Bind(req); err != nil {
		return errors.WithStack(err)
	}

	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	u := &models.User{}
	if err := tx.Where(`Lower(email) = ?`, strings.ToLower(req.Email)).First(u); err != nil {
		// return c.Error(404, errors.New("Could not recover"))
		// TODO we cannot 404 on the email to obfuscate whether the email is found or not.
		c.Flash().Add("danger", "Invalid data")
		return c.Redirect(http.StatusMovedPermanently, "accountRecoveryPath()")
	}
	u.RecoveryCode = nulls.NewString(verifyCode())
	u.RecoveryExp = nulls.NewTime(time.Now().UTC().Add(time.Minute * 10))

	verrs, err := u.Update(tx)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("errors", verrs)
		c.Set("req", req)
		return c.Render(http.StatusOK, r.HTML("users/recovery.html"))
	}
	go func(u *models.User) {
		sender, ok := c.Value("recovery_sender").(Sender)
		if !ok {
			fmt.Printf("ERROR SENDING RECOVERY: no recovery sender found\n")
			return
		}

		if err := sender.Send(map[string]interface{}{
			"code":           u.RecoveryCode.String,
			"sender_email":   "system@example.com",
			"receiver_email": u.Email,
		}); err != nil {
			fmt.Printf("ERROR SENDING RECOVERY: %v\n", err)
		}
	}(u)

	// TODO what is this error?
	if ct, ok := c.Value("contentType").(string); !ok || strings.Contains(ct, "html") || strings.Contains(ct, "form") {
		c.Flash().Add("danger", "Something bad happened")
		return c.Redirect(http.StatusFound, "accountRecoveryForm()")
	}
	return c.Render(http.StatusOK, r.Auto(c, "Sent"))
}

// RecoveryUpdate request struct for using recovery code.
type RecoveryUpdate struct {
	Email                string `json:"email"`
	Code                 string `json:"code"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirm"`
}

// UserRequestRecover is used to allow users to reset their passwords
// using the recovery code we've sent them
func AccountRecovery(c buffalo.Context) error {
	req := &models.RecoveryUpdate{}
	if err := c.Bind(req); err != nil {
		return errors.WithStack(err)
	}

	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	u := &models.User{}
	verrs := req.Validate()
	if verrs.HasAny() {
		c.Set("errors", verrs)
		c.Set("req", req)
		return c.Render(http.StatusOK, r.HTML("users/recover.html"))
	}

	if err := tx.Where("Lower(email) = ?", strings.ToLower(req.Email)).First(u); err != nil {
		c.Flash().Add("danger", "Recovery code is not valid")
		return c.Redirect(http.StatusMovedPermanently, "accountRecoveryPath()")
	}

	if u.RecoveryExp.Time.After(time.Now().UTC()) && u.RecoveryCode.String == req.Code {
		u.Password = req.Password
		u.PasswordConfirmation = req.PasswordConfirmation

		u.RecoveryCode = nulls.NewString("")
		verrs, err := u.Update(tx)
		if err != nil {
			return errors.WithStack(err)
		}
		if verrs.HasAny() {
			c.Set("errors", verrs)
			c.Set("req", req)
			return c.Render(http.StatusOK, r.HTML("users/recover.html"))
		}
	} else {
		c.Flash().Add("warning", "Recovery code is not valid")
		return c.Redirect(http.StatusMovedPermanently, "accountRecoveryPath()")
	}

	if ct, ok := c.Value("contentType").(string); !ok || strings.Contains(ct, "html") || strings.Contains(ct, "form") {
		c.Flash().Add("success", "Your password has been reset")
		return c.Redirect(http.StatusFound, "rootPath()")
	}
	return c.Render(http.StatusOK, r.Auto(c, u))
}
