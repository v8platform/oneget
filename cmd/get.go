package cmd

import (
	"github.com/khorevaa/logos"
	"time"

	"github.com/urfave/cli/v2"
	dloader "github.com/v8platform/oneget/downloader"
)

var log = logos.New("github.com/v8platform/oneget").Sugar()

type getCmd struct {
	User      string
	Password  string
	BaseDir   string
	StartDate time.Time
	Type      string
	Version   string
	Filter    string
}

func (c *getCmd) run(ctx *cli.Context) error {

	c.StartDate = time.Time{}

	startDate := ctx.Timestamp("start-date")

	if startDate != nil {
		c.StartDate = *startDate
	}
	c.User = ctx.String("user")
	c.Password = ctx.String("pwd")

	if ctx.Bool("debug") {
		logos.SetLevel("github.com/v8platform/oneget", logos.DebugLevel)
	}

	downloaderConfig := dloader.Config{
		Login:     c.User,
		Password:  c.Password,
		BasePath:  c.BaseDir,
		StartDate: c.StartDate,
		Nicks: map[string]bool{
			c.Type: true,
		},
		VersionFilter: c.Version,
		DistribFilter: c.Filter,
	}

	downloader := dloader.New(downloaderConfig)

	files, err := downloader.Get()
	if err != nil {
		return err
	}
	log.Infof("Downloaded <%d> files", len(files))

	return nil
}

func (c *getCmd) Cmd() *cli.Command {

	cmd := &cli.Command{
		Name:   "get",
		Usage:  "Получение релиза сайта релизов 1С",
		Flags:  c.Flags(),
		Action: c.run,
	}

	return cmd
}

func (c *getCmd) Flags() []cli.Flag {
	return []cli.Flag{

		&cli.StringFlag{
			Destination: &c.Type,
			EnvVars:     []string{"ONEGET_NICKS"},
			Name:        "nick",
			Usage: `Имена приложений (например \"platform83 или EnterpriseERP20\"), 
					подсмотреть можно в адресе, ссылки имею вид например https://releases.1c.ru/project/EnterpriseERP20`,
			Required: true,
		},
		&cli.StringFlag{
			Destination: &c.Version,
			EnvVars:     []string{"ONEGET_NICKS_VERSION"},
			Name:        "version",
			Usage:       "Фильтр версий по номеру",
			Required:    true,
		},
		&cli.TimestampFlag{
			DefaultText: time.Now().Format("2006-01-02"),
			Layout:      "2006-01-02",
			EnvVars:     []string{"ONEGET_START_DATE"},
			Name:        "start-date",
			Usage:       "Фильтр версий по номеру",
		},
		&cli.StringFlag{
			Destination: &c.Filter,
			EnvVars:     []string{"ONEGET_NICKS_FILTER"},
			Aliases:     []string{"filter"},
			Name:        "distrib-filter",
			Usage:       "Дополнительный фильтр пакетов (регулярное выражение)",
		},
		&cli.StringFlag{
			Destination: &c.BaseDir,
			Name:        "path",
			Aliases:     []string{"out"},
			Value:       "./downloads",
			DefaultText: "./downloads",
			Usage:       "Путь к каталогу выгрузки",
		},
	}
}
