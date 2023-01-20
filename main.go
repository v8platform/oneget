package main

import (
	"fmt"

	"github.com/khorevaa/logos"
	"github.com/urfave/cli/v2"
	"github.com/v8platform/oneget/cmd"

	"os"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

var log = logos.New("github.com/v8platform/oneget").Sugar()

func setFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "user",
			Aliases:  []string{"u"},
			EnvVars:  []string{"ONEGET_USER", "ONEC_USERNAME"},
			Usage:    "Пользователь портала 1С",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "pwd",
			Aliases:  []string{"p"},
			EnvVars:  []string{"ONEGET_PASSWORD", "ONEC_PASSWORD"},
			Usage:    "Пароль пользователя портала 1С",
			Required: true,
		},
		&cli.BoolFlag{
			Name:    "debug",
			EnvVars: []string{"ONEGET_DEBUG"},
			Usage:   "Режим отладки приложения",
		},
		&cli.StringFlag{
			Name:        "logs",
			DefaultText: "oneget.log",
			Value:       "oneget.log",
			Usage:       "Файл лога загрузки",
		},
		&cli.BoolFlag{
			Name:    "enableHttp",
			Aliases: []string{"s"},
			Value:   false,
			EnvVars: []string{"ONEGET_ENABLE_HTTP_SERVER"},
			Usage:   "Запустить http сервер для доступа к скачаным файлам",
		},
		&cli.StringFlag{
			Name:    "serverPort",
			Aliases: []string{"sp"},
			Value:   "8080",
			EnvVars: []string{"ONEGET_HTTP_SERVER_PORT"},
		},
	}
}

func main() {
	app := &cli.App{
		Name:    "oneget",
		Usage:   "Приложение для загрузки релизов сайта релизов 1С",
		Version: buildVersion(),
		Flags:   setFlags(),
	}

	for _, command := range cmd.Commands {
		app.Commands = append(app.Commands, command.Cmd())
	}

	err := app.Run(os.Args)
	defer log.Sync()
	if err != nil {
		log.Fatal(err.Error())
	}
}

func buildVersion() string {
	var result = version
	if commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, commit)
	}
	if date != "" {
		result = fmt.Sprintf("%s\nbuilt at: %s", result, date)
	}
	if builtBy != "" {
		result = fmt.Sprintf("%s\nbuilt by: %s", result, builtBy)
	}
	return result
}
