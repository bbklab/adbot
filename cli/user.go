package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/bbklab/adbot/cli/helpers"
	"github.com/bbklab/adbot/pkg/color"
	"github.com/bbklab/adbot/pkg/template"
	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/types"
)

// nolint
var (
	UserSessionTableHeader = "ID\t" + color.Yellow("CURRENT") + "\tIP\tDEVICE\tLAST ACTIVE AT\t\n"
	UserSessionTableLine   = "{{.ID}}\t{{if .Current}}{{green .Current}}{{else}}{{cyan .Current}}{{end}}\t{{.Remote}}\t{{.Device}}/{{.OS}}/{{.Browser}}\t{{tformat .LastActiveAt}}\t\n"

	userTableHeader = "USER ID\tNAME\tPASSWORD\tCREATED AT\tLASTLOGIN AT\t\n"
	userTableLine   = "{{.ID}}\t{{.Name}}\t{{.Password}}\t{{tformat .CreatedAt}}\t{{tformat .LastLoginAt}}\t\n"
)

var (
	createUserFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "name",
			Usage: "uniq name of user",
			Value: "",
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "user password",
			Value: "",
		},
		cli.StringFlag{
			Name:  "desc",
			Usage: "user description text",
			Value: "",
		},
	}

	listUserFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "quiet,q",
			Usage: "only display numeric IDs",
		},
	}

	chgUserPasswordFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "old",
			Usage: "old password",
			Value: "",
		},
		cli.StringFlag{
			Name:  "new",
			Usage: "new password",
			Value: "",
		},
	}
)

// UserCommand is exported
func UserCommand() cli.Command {
	return cli.Command{
		Name:  "user",
		Usage: "user management",
		Subcommands: []cli.Command{
			userAnyCommand(),         // any
			userListCommand(),        // ls
			userProfileCommand(),     // profile
			userSessionCommand(),     // session
			userCreateCommand(),      // create
			userChgPasswordCommand(), // chgpassword
		},
	}
}

func userAnyCommand() cli.Command {
	return cli.Command{
		Name:   "any",
		Usage:  "has any users",
		Action: anyUsers,
	}
}

func userListCommand() cli.Command {
	return cli.Command{
		Name:   "ls",
		Usage:  "list all of users",
		Flags:  listUserFlags,
		Action: listUsers,
	}
}

func userProfileCommand() cli.Command {
	return cli.Command{
		Name:   "profile",
		Usage:  "show current user's profiles",
		Action: userProfile,
	}
}

func userSessionCommand() cli.Command {
	return cli.Command{
		Name:   "session",
		Usage:  "list current user's sessions",
		Action: userSessions,
	}
}

func userCreateCommand() cli.Command {
	return cli.Command{
		Name:   "create",
		Usage:  "create a user",
		Flags:  createUserFlags,
		Action: createUser,
	}
}

func userChgPasswordCommand() cli.Command {
	return cli.Command{
		Name:   "chgpassword",
		Usage:  "change user password",
		Flags:  chgUserPasswordFlags,
		Action: chgUserPassword,
	}
}

func anyUsers(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	has, err := client.AnyUsers()
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, has)
	return nil
}

func listUsers(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	users, err := client.ListUsers()
	if err != nil {
		return err
	}

	// only print ids
	if c.Bool("quiet") {
		for _, user := range users {
			fmt.Fprintln(os.Stdout, user.ID)
		}
		return nil
	}

	var (
		w         = tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', 0)
		parser, _ = template.NewParser(userTableLine)
	)

	fmt.Fprint(w, userTableHeader)
	for _, user := range users {
		parser.Execute(w, user)
	}
	w.Flush()

	return nil
}

func userProfile(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	profile, err := client.UserProfile()
	if err != nil {
		return err
	}

	return utils.PrettyJSON(nil, profile)
}

func userSessions(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	sesses, err := client.UserSessions()
	if err != nil {
		return err
	}

	var (
		w         = tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', 0)
		parser, _ = template.NewParser(UserSessionTableLine)
	)

	fmt.Fprint(w, UserSessionTableHeader)
	for _, sess := range sesses {
		parser.Execute(w, sess)
	}
	w.Flush()
	return nil
}

func createUser(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	if c.NumFlags() == 0 {
		return cli.ShowSubcommandHelp(c)
	}

	var (
		user = &types.User{
			Name:     c.String("name"),
			Password: types.Password(c.String("password")),
			Desc:     c.String("desc"),
		}
	)

	created, err := client.CreateUser(user)
	if err != nil {
		return err
	}

	os.Stdout.Write(append([]byte(created.ID), '\r', '\n'))
	return nil
}

func chgUserPassword(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	if c.NumFlags() == 0 {
		return cli.ShowSubcommandHelp(c)
	}

	var req = &types.ReqChangePassword{
		Old: types.Password(c.String("old")),
		New: types.Password(c.String("new")),
	}

	err = client.ChangeUserPassword(req)
	if err != nil {
		return err
	}

	os.Stdout.Write(append([]byte("Change Password Succeed, ReLogin Required"), '\r', '\n'))
	return nil
}
