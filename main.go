package main

import (
	"log"
	"os"
	"strings"
	"fmt"

	dloader "TheDemonCat/oneget/downloader"
	"github.com/urfave/cli/v2"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

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
		&cli.StringFlag{
			Name: "nicks",
			Usage: `Имена приложений, разделенные запятой (например \"platform83, EnterpriseERP20\"), 
					подсмотреть можно в адресе, ссылки имею вид например https://releases.1c.ru/project/EnterpriseERP20`,
		},
		&cli.StringFlag{
			Name:  "version-filter",
			Usage: "Фильтр версий по номеру (регулярное выражение)",
		},
		&cli.StringFlag{
			Name:  "distrib-filter",
			Usage: "Дополнительный фильтр пакетов (регулярное выражение)",
		},
		&cli.StringFlag{
			Name:        "path",
			DefaultText: "./downloads",
			Usage:       "Путь к каталогу выгрузки",
		},
		&cli.StringFlag{
			Name:        "logs",
			DefaultText: "oneget.logs",
			Value:       "oneget.logs",
			Usage:       "Файл лога загрузки",
		},
	}
}

func main() {
	app := &cli.App{
		Name:    "oneget",
		Usage:   "Приложение для загрузки релизов сайта релизов 1С",
		Version: buildVersion(),
		Flags:   setFlags(),
		Action: func(c *cli.Context) error {
			downloaderConfig := dloader.Downloader{
				Login:         c.String("user"),
				Password:      c.String("pwd"),
				BasePath:      c.String("path"),
				StartDate:     StartDate(c.String("startDate")),
				Nicks:         Nicks(strings.ToLower(c.String("nicks"))),
				VersionFilter: c.String("version-filter"),
				DistribFilter: c.String("distrib-filter"),
			}

			downloader := dloader.New(&downloaderConfig)
			fileLogs, err := LogFile(c.String("logs"))
			if err != nil {
				handleError(err, "Ошибка записи файла логирования")
			}
			downloader.SetLogOutput(fileLogs)
			_, err = downloader.Get()
			if err != nil {
				log.Fatal(err)
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
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
