package main

import (
	"strings"
	"time"

	check "gopkg.in/check.v1"

	"github.com/bbklab/adbot/types"
)

var (
	userName    = "admin"
	userPass    = types.Password("admin")
	userDesc    = "the only privileged user for integration test cases"
	defaultUser = &types.User{
		Name:     userName,
		Password: userPass,
		Desc:     userDesc,
	}
)

func (s *ApiSuite) TestUserProfile(c *check.C) {
	startAt := time.Now()

	user, err := s.client.UserProfile()
	c.Assert(err, check.IsNil)
	c.Assert(user.Name, check.Equals, userName)
	c.Assert(string(user.Password), check.Equals, "******")
	c.Assert(user.CreatedAt.IsZero(), check.Equals, false)
	c.Assert(user.LastLoginAt.IsZero(), check.Equals, false)

	costPrintln("TestUserProfile() passed", startAt)
}

func (s *ApiSuite) TestUserList(c *check.C) {
	startAt := time.Now()

	users, err := s.client.ListUsers()
	c.Assert(err, check.IsNil)
	c.Assert(len(users), check.Equals, 1)
	c.Assert(users[0].Name, check.Equals, userName)
	c.Assert(string(users[0].Password), check.Equals, "******")

	costPrintln("TestUserList() passed", startAt)
}

func (s *ApiSuite) TestUserAny(c *check.C) {
	startAt := time.Now()

	any, err := s.client.AnyUsers()
	c.Assert(err, check.IsNil)
	c.Assert(any, check.Equals, true)

	costPrintln("TestUserAny() passed", startAt)
}

func (s *ApiSuite) TestUserCreate(c *check.C) {
	startAt := time.Now()

	_, err := s.client.CreateUser(defaultUser)
	c.Assert(err, check.NotNil)
	c.Assert(err, check.ErrorMatches, "403 - .*there already has one super admin user.*")

	costPrintln("TestUserCreate() passed", startAt)
}

func (s *ApiSuite) TestUserChgPassword(c *check.C) {
	startAt := time.Now()

	// invalid change password
	datas := map[*types.ReqChangePassword]string{
		&types.ReqChangePassword{userPass, "x"}:                                     "400 - .*password invalid.*",
		&types.ReqChangePassword{userPass, types.Password(strings.Repeat("x", 65))}: "400 - .*password invalid.*",
		&types.ReqChangePassword{userPass, userPass}:                                "400 - .*new password should not be the same as original.*",
		&types.ReqChangePassword{"xxxx", "1234"}:                                    "403 - .*authentication failed.*",
	}
	for req, errmsg := range datas {
		err := s.client.ChangeUserPassword(req)
		c.Assert(err, check.NotNil)
		c.Assert(err, check.ErrorMatches, errmsg)
	}

	// change password
	err := s.client.ChangeUserPassword(&types.ReqChangePassword{userPass, "xxxx"})
	c.Assert(err, check.IsNil)
	// relogin required after change password
	token, err := s.client.Login(&types.ReqLogin{userName, types.Password("xxxx")})
	c.Assert(err, check.IsNil)
	s.client.SetHeader("Admin-Access-Token", token)

	// change back
	err = s.client.ChangeUserPassword(&types.ReqChangePassword{"xxxx", userPass})
	c.Assert(err, check.IsNil)
	// relogin required after change password
	token, err = s.client.Login(&types.ReqLogin{userName, types.Password(userPass)})
	c.Assert(err, check.IsNil)
	s.client.SetHeader("Admin-Access-Token", token)

	costPrintln("TestUserChgPassword() passed", startAt)
}
