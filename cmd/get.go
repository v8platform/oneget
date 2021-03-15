package cmd

import (
	"fmt"
	"github.com/khorevaa/logos"
	"go.uber.org/multierr"
	"strings"
	"sync"
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
	Filter    cli.StringSlice

	releases []string
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

	releases := getMapFromStrings(c.releases, "@", "latest")
	filters := getMapFromStrings(c.Filter.Value(), "=", "")

	var downloads []downloadConfig
	for project, version := range releases {
		downloads = append(downloads, downloadConfig{
			project,
			version,
			filters[project],
		})
	}

	wg := sync.WaitGroup{}

	var err error

	for _, download := range downloads {
		wg.Add(1)
		go func(info downloadConfig) {

			dl := dloader.NewDownloader(dloader.OneConfig{
				Login:    c.User,
				Password: c.Password,
				BasePath: c.BaseDir,
				Project:  info.Project,
				Version:  info.Version,
				Filter:   info.Filter,
			})

			_, errGet := dl.Get()
			if errGet != nil {
				err = multierr.Append(err, errGet)
			}
			wg.Done()
		}(download)

	}
	wg.Wait()

	if err != nil {
		return err
	}

	log.Infof("Downloaded <%d> releases", len(downloads))

	return nil
}

func (c *getCmd) Cmd() *cli.Command {

	cmd := &cli.Command{
		Name:      "get",
		Usage:     "Получение релиза сайта релизов 1С",
		ArgsUsage: "RELEASE...",
		CustomHelpTemplate: cli.CommandHelpTemplate + `ARGUMENTS:
   RELEASE - описание релиза в формате platform83@8.3.18.1334

`,
		Flags:  c.Flags(),
		Action: c.run,
		Before: func(ctx *cli.Context) error {

			if !ctx.Args().Present() {
				err := cli.ShowSubcommandHelp(ctx)
				if err != nil {
					return err
				}
				return fmt.Errorf("WRONG USAGE: Requires a RELEASE argument")
			}

			c.releases = ctx.Args().Slice()

			return nil
		},
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
			//Required: true,
		},
		&cli.StringFlag{
			Destination: &c.Version,
			EnvVars:     []string{"ONEGET_NICKS_VERSION"},
			Name:        "version",
			Usage:       "Фильтр версий по номеру",
			//Required:    true,
		},
		&cli.StringFlag{
			Destination: &c.Version,
			EnvVars:     []string{"ONEGET_NICKS_VERSION"},
			Name:        "platform-filter",
			Usage: `Фильтр по типу ОС для платформы (platform83)]
							macOS - дистрибутив для OS X			
							windows - дистрибутив для Windows			
							deb - дистрибутив для DEB-based Linux-систем			
							deb - дистрибутив для RPM-based Linux-систем			
`,
			//Required:    true,
		},
		&cli.TimestampFlag{
			DefaultText: time.Now().Format("2006-01-02"),
			Layout:      "2006-01-02",
			EnvVars:     []string{"ONEGET_START_DATE"},
			Name:        "start-date",
			Usage:       "Фильтр версий по номеру",
		},
		&cli.StringSliceFlag{
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

func getMapFromStrings(arr []string, sep string, defValue string) map[string]string {

	result := make(map[string]string)

	for _, str := range arr {

		values := strings.SplitN(str, sep, 2)

		key := values[0]
		value := defValue

		if len(values) == 2 {
			value = values[1]
		}

		result[key] = value

	}

	return result
}

type downloadConfig struct {
	Project string
	Version string
	Filter  string
}
