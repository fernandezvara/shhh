package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/fernandezvara/shhh/internal/ops"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/urfave/cli/v2"
)

func main() {

	app := &cli.App{
		Name:  "shhh",
		Usage: "shhh! don't say others ...",
		Commands: []*cli.Command{
			{
				Name:   "create",
				Usage:  "creates a new secrets file",
				Action: createFunc,
				Flags:  commonFlags,
			},
			{
				Name:    "set",
				Aliases: []string{"s"},
				Usage:   "sets a secret entry",
				Action:  setFunc,
				Flags:   append(commonFlags, groupFlag, keyFlag, valueFlag, fileFlag),
			},
			{
				Name:    "get",
				Aliases: []string{"g"},
				Usage:   "gets a secret entry",
				Action:  getFunc,
				Flags:   append(commonFlags, groupFlag, keyFlag, fileFlag),
			},
			{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "deletes a secret entry",
				Action:  deleteFunc,
				Flags:   append(commonFlags, groupFlag, keyFlag, forceFlag),
			},
			{
				Name:   "deletegroup",
				Usage:  "deletes a group of secrets",
				Action: deleteGroupFunc,
				Flags:  append(commonFlags, groupFlag, keyFlag, forceFlag),
			},
			{
				Name:   "version",
				Usage:  "shows executable version",
				Action: versionFunc,
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

var commonFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "db",
		Value:   "./shhh.db",
		Usage:   "filename where the secrets will be stored",
		EnvVars: []string{"SHHH_DB"},
	},
	&cli.StringFlag{
		Name:    "passwd",
		Aliases: []string{"p"},
		Usage:   "password to en/decrypt information (ensure this is not logged)",
		EnvVars: []string{"SHHH_PASSWD"},
	},
}

var groupFlag = &cli.StringFlag{
	Name:    "group",
	Aliases: []string{"g"},
	Usage:   "group ",
	EnvVars: []string{"SHHH_GROUP"},
}

var keyFlag = &cli.StringFlag{
	Name:    "key",
	Aliases: []string{"k"},
	Usage:   "key to store the information",
	EnvVars: []string{"SHHH_KEY"},
}

var valueFlag = &cli.StringFlag{
	Name:    "value",
	Aliases: []string{"v"},
	Usage:   "value to set",
	EnvVars: []string{"SHHH_VALUE"},
}

var fileFlag = &cli.StringFlag{
	Name:    "file",
	Aliases: []string{"f"},
	Usage:   "file to write/read",
	EnvVars: []string{"SHHH_FILE"},
}

var forceFlag = &cli.BoolFlag{
	Name:    "force",
	Usage:   "force changes (don't ask for confirmation)",
	EnvVars: []string{"SHHH_FORCE"},
}

func createFunc(c *cli.Context) error {

	var (
		client *ops.Ops
		passwd string
		err    error
	)

	client, err = ops.Open(c.String("db"), true)
	er(err)

	err = client.SetSalt()
	er(err)

	passwd = mustAsk(c, "passwd", "Password", "", true, nil)

	err = client.SetKey(passwd)
	er(err)

	return client.Close()

}

func setFunc(c *cli.Context) error {

	var (
		client                *ops.Ops
		id, key, file, passwd string
		value                 []byte
		err                   error
	)

	client, err = ops.Open(c.String("db"), false)
	er(err)
	defer client.Close()

	passwd = mustAsk(c, "passwd", "Password", "", true, nil)

	err = client.GetSalt()
	er(err)

	err = client.GetKey(passwd)
	er(err)

	id = mustAsk(c, "group", "   Group", "", false, validationRequired)
	key = mustAsk(c, "key", "     Key", "", false, validationRequired)
	file = c.String("file")

	if file != "" {
		value, err = os.ReadFile(file)
		er(err)
	} else {
		value = []byte(mustAsk(c, "value", "   Value", "", true, nil))
	}

	err = client.Set(id, key, value)
	er(err)

	return nil

}

func getFunc(c *cli.Context) error {

	var (
		client                *ops.Ops
		id, key, passwd, file string
		values                map[string][]byte
		t                     table.Writer
		err                   error
	)

	client, err = ops.Open(c.String("db"), false)
	er(err)
	defer client.Close()

	passwd = mustAsk(c, "passwd", "Password", "", true, nil)

	err = client.GetSalt()
	er(err)

	err = client.GetKey(passwd)
	er(err)

	id = mustAsk(c, "group", "   Group", "", false, validationRequired)
	key = mustAsk(c, "key", "     Key (blank for all)", "", false, nil)

	values, err = client.Get(id, key)
	er(err)

	if len(values) == 0 {
		err = ops.ErrNotExist
		er(err)
	}

	if key != "" && file != "" {
		fmt.Println("WriteFile")
		err = os.WriteFile(file, values[key], 0644)
		er(err)
		return nil
	}

	t = table.NewWriter()

	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.Style().Format.Header = text.FormatTitle
	t.AppendHeader(table.Row{"Key", "Value"})

	for k, v := range values {
		t.AppendRow(table.Row{k, string(v)})
	}

	t.Render()

	return nil
}

func deleteFunc(c *cli.Context) error {

	var (
		client          *ops.Ops
		id, key, passwd string
		apply           bool
		err             error
	)

	client, err = ops.Open(c.String("db"), false)
	er(err)
	defer client.Close()

	passwd = mustAsk(c, "passwd", "Password", "", true, nil)

	err = client.GetSalt()
	er(err)

	err = client.GetKey(passwd)
	er(err)

	id = mustAsk(c, "group", "   Group", "", false, validationRequired)
	key = mustAsk(c, "key", "     Key", "", false, validationRequired)
	apply = c.Bool("force")

	if !apply {
		apply, err = promptTrueFalseBool("Are you sure?", "Yes", "No", false)
		er(err)
	}

	if apply {
		err = client.DeleteKey(id, key)
		er(err)
	}

	return nil

}

func deleteGroupFunc(c *cli.Context) error {

	var (
		client     *ops.Ops
		id, passwd string
		apply      bool
		err        error
	)

	client, err = ops.Open(c.String("db"), false)
	er(err)
	defer client.Close()

	passwd = mustAsk(c, "passwd", "Password", "", true, nil)

	err = client.GetSalt()
	er(err)

	err = client.GetKey(passwd)
	er(err)

	id = mustAsk(c, "group", "   Group", "", false, validationRequired)
	apply = c.Bool("force")

	if !apply {
		apply, err = promptTrueFalseBool("Are you sure?", "Yes", "No", false)
		er(err)
	}

	if apply {
		err = client.Delete(id)
		er(err)
	}

	return nil

}

func versionFunc(c *cli.Context) error {

	fmt.Printf(`
     Version: %s
  GIT Commit: %s
     GIT URL: %s
	
`, Version, FullCommit, GitURL)

	return nil

}
