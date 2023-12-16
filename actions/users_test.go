package actions

import (
	"net/http"

	"event_planner/models"
)

func (as *ActionSuite) Test_Users_New() {
	res := as.HTML("/users/new").Get()
	as.Equal(http.StatusOK, res.Code)
}

func (as *ActionSuite) CreateTestUser(u *models.User, shouldFail bool) {
	count, err := as.DB.Count("users")
	as.NoError(err)

	res := as.HTML("/users").Post(u)
	if shouldFail {
		as.Equal(http.StatusOK, res.Code)

		count2, err2 := as.DB.Count("users")
		as.NoError(err2)
		as.Equal(count, count2)
	} else {
		as.Equal(http.StatusFound, res.Code)

		count2, err2 := as.DB.Count("users")
		as.NoError(err2)
		as.Equal(count+1, count2)
	}
}

func (as *ActionSuite) Test_Users_Create() {
	u := &models.User{
		Email:                "mark@example.com",
		Password:             "password",
		PasswordConfirmation: "password",
	}
	as.CreateTestUser(u, false)
}

func (as *ActionSuite) Test_Users_Recover() {
	password := "password"
	u := &models.User{
		Email:                "mark@example.com",
		Password:             password,
		PasswordConfirmation: password,
	}
	as.CreateTestUser(u, false)

	res := as.HTML("/recovery").Get()
	as.Equal(http.StatusOK, res.Code)

	req := &RecoveryRequest{
		Email: u.Email,
	}
	res = as.HTML("/requestRecovery").Post(req)
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/recover", res.Location())

	res = as.HTML("/recover").Get()
	as.Equal(http.StatusOK, res.Code)

	u2 := &models.User{}
	as.DB.First(u2)
	as.Equal(u.Email, u2.Email)

	upreq := &RecoveryUpdate{
		Email:                u.Email,
		Code:                 u2.RecoveryCode.String,
		Password:             password + "2",
		PasswordConfirmation: password + "2",
	}
	res = as.HTML("/completeRecover").Post(upreq)
	as.Equal(http.StatusFound, res.Code)

	u2.Password = password + "2"
	res = as.HTML("/signin").Post(u2)
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/", res.Location())
}

func (as *ActionSuite) Test_Users_Recover_Fail() {
	password := "password"
	u := &models.User{
		Email:                "mark@example.com",
		Password:             password,
		PasswordConfirmation: password,
	}
	as.CreateTestUser(u, false)

	req := &RecoveryRequest{Email: u.Email}
	res := as.HTML("/requestRecovery").Post(req)
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/recover", res.Location())

	res = as.HTML("/recover").Get()
	as.Equal(http.StatusOK, res.Code)

	u2 := &models.User{}
	as.DB.First(u2)
	as.Equal(u.Email, u2.Email)

	upreq := &RecoveryUpdate{
		Email:                u.Email,
		Code:                 u2.RecoveryCode.String + "nope",
		Password:             password + "2",
		PasswordConfirmation: password + "2",
	}
	res = as.HTML("/completeRecover").Post(upreq)
	as.Equal(403, res.Code)

	u2.Password = password
	res = as.HTML("/signin").Post(u2)
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/", res.Location())
}
