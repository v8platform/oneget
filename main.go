package main

import (
	"log"
	"os"

	dloader "TheDemonCat/oneget/downloader"
	"github.com/urfave/cli/v2"
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
			Name:     "nicks",
			Usage:    `Имена приложений, разделенные запятой (например \"platform83, EnterpriseERP20\"), 
					подсмотреть можно в адресе, ссылки имею вид например https://releases.1c.ru/project/EnterpriseERP20`,
		},
		&cli.StringFlag{
			Name:     "version-filter",
			Usage:    "Фильтр версий по номеру (регулярное выражение)",
		},
			&cli.StringFlag{
			Name:     "path",
			DefaultText: "./downloads",
			Usage:    "Путь к каталогу выгрузки",
		},
		&cli.StringFlag{
			Name:     		"logs",
			DefaultText: 	"oneget.logs",
			Value: 			"oneget.logs",
			Usage:    		"Файл лога загрузки",
		},
	}
}

func main() {
	app := &cli.App{
		Name : "oneget",
		Usage: "Приложение для загрузки релизов сайта релизов 1С",
		Flags: setFlags(),
		Action: func(c *cli.Context) error {
			downloaderConfig := dloader.Downloader{
				Login:    		c.String("user"),
				Password: 		c.String("pwd"),
				BasePath: 		c.String("path"),
				StartDate : 	StartDate(c.String("startDate")),
				Nicks  :    	Nicks(c.String("nicks")),
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
