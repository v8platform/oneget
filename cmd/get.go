package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/khorevaa/logos"
	server "github.com/v8platform/oneget/http-server"
	"github.com/v8platform/oneget/unpacker"
	"go.uber.org/multierr"

	"github.com/urfave/cli/v2"
	dloader "github.com/v8platform/oneget/downloader"
)

var log = logos.New("github.com/v8platform/oneget").Sugar()

type getCmd struct {
	User         string
	Password     string
	BaseDir      string
	StartDate    time.Time
	Rename       bool
	Extract      bool
	ExtractDir   string
	Filter       cli.StringSlice
	EnableServer bool
	ServerPort   string

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
	additionalFilterStr := getAdditionalFilters(c.Filter.Value(), "=", "")

	var downloadConfigs []dloader.DownloadConfig

	for project, version := range releases {

		projectIdAlias := getProjectId(project)

		projectId := dloader.GetProjectIDByAlias(projectIdAlias)
		additionalFilters := compileFilters(additionalFilterStr[projectIdAlias]...)

		var projectFilters []dloader.FileFilter

		if fileFilter, err := getProjectFilter(project); err != nil {
			log.Errorf("error get project <%s> file filter: %s", projectIdAlias, err.Error())
			return fmt.Errorf("project <%s> %s", projectIdAlias, err.Error())
		} else if fileFilter != nil {
			projectFilters = append(projectFilters, fileFilter)
		}

		versionFilter, err := dloader.NewVersionFilter(projectId, version)
		if err != nil {
			return err
		}

		downloadConfigs = append(downloadConfigs, dloader.DownloadConfig{
			BasePath:          getAbsolutePath(c.BaseDir),
			Project:           projectId,
			Version:           versionFilter,
			Filters:           projectFilters,
			AdditionalFilters: additionalFilters,
		})
	}

	var err error
	dl := dloader.NewDownloader(
		c.User,
		c.Password,
	)

	files, errGet := dl.Get(downloadConfigs...)
	if errGet != nil {
		err = multierr.Append(err, errGet)
	}
	if err != nil {
		return err
	}
	log.Infof("Downloaded <%d> releases, files <%d>", len(downloadConfigs), len(files))

	if c.Extract {
		err := c.extractFiles(files)
		if err != nil {
			return err
		}
	}

	c.EnableServer = ctx.Bool("enableHttp")
	c.ServerPort = ctx.String("serverPort")
	if c.EnableServer {
		server.Run(getAbsolutePath(c.BaseDir), c.ServerPort)
	}
	return nil
}

func (c *getCmd) Cmd() *cli.Command {

	cmd := &cli.Command{
		Name:      "get",
		Usage:     "Получение релиза сайта релизов 1С",
		ArgsUsage: "RELEASE...",
		CustomHelpTemplate: cli.CommandHelpTemplate + `ARGUMENTS:
   RELEASE - описание релиза в формате platform83[:filter.[filter]...]@8.3.18.1334
                                           ^               ^                ^ 
                                        имя проекта   набор фильтров   версия релиза
                                                        (см. ниже)       (см. ниже)

   > Имя проекта - подсмотреть можно в адресе, ссылки имею вид например https://releases.1c.ru/project/EnterpriseERP20
   Синонимы проектов для быстрого доступа: 
     * platform - platform83
     * edt      - DevelopmentTools10
     * ring     - EnterpriseLicenseTools
     * executor - Executor
     * pg       - AddCompPostgre

   > Набор фильтров - список предопределенных фильтров для проектов:
     - По ОС:
       * win, windows  - фильтр по MS Windows
       * linux         - фильтр по Linux (для платформы выше 8.3.20)
       * mac           - фильтр по OS X
       * deb           - фильтр по DEB-based Linux-систем (для платформы ниже 8.3.20)
       * rpm           - фильтр по RPM-based Linux-систем (для платформы ниже 8.3.20)
       Например, platform:deb - будет скачаны файлы с фильтрацией по DEB-based Linux-систем
     
     - По разрядности OS:
       * x32           - фильтр по 32-bit (по умолчанию, если не указан фильтр разрядности) 
       * x64           - фильтр по 64-bit  
       Например, platform:deb.x64 - будет скачаны файлы с фильтрацией по DEB-based Linux-систем и 64-bit
       
     - Для проекта platform (platform83)
       * thin-client, thin   - фильтр для файлов тонкого клиента 1С.Предприятие
       * client              - фильтр для файлов клиента 1С.Предприятие
       * server              - фильтр для файлов сервера 1С.Предприятие
       * full                - фильтр для файлов "Технологическая платформа" (только для Windows)
       Например, platform:deb.server.x64 - будет скачаны файлы сервера с фильтрацией по DEB-based Linux-систем и 64-bit 
    
    - Для проекта edt (DevelopmentTools10)
       * jdk    - фильтр для файлов Bellsoft JDK
       * online - фильтр для файлов онлайн установщика 1С:EDT

       Например, edt:deb - будет скачаны файлы: 
            - Дистрибутив для оффлайн установки 1C:EDT для ОС Linux 64 бит
            - Bellsoft JDK Full (64-bit) для DEB-based Linux-систем
   
   > Версии релиза:
       В версии релиза может быть указан номер версии или специальные фильтры версии.
       Если версия релиза пустая то подставляется фильтр "latest" ( "edt" воспринимается как "edt@latest"
       
       Специальные фильтры версии релиза:
         * latest           - выбирает наиболее старшую версию релиза
         * latest:[regexp]  - фильтрует список версию по <regexp>, и берет наиболее старшую
         * from:<date>      - фильтрует список версий по дате, у которых дата релиза больше <date> 
            где, date - формате 02.06.21
         * from-v:<version> - фильтрует список версий, которые старше версии релиза <version> 
            где, version - формате номер версии
         * <regexp>        - фильтрует список по регулярному выражению указанному в <regexp>

       Примеры: 
          1. "platform@from:01.01.21" - будут загружена все версии релизов, выпущенные начиная с даты 2020.01.01
          2. "platform@from-v:8.3.16" - будут загружена все версии релизов, у которых версия старше чем 8.3.16
          3. "platform@latest:8.3.16" - будут загружена последняя версия релиза 8.3.16
          4. "platform@8.3.16"        - будут загружена все версии релизов 8.3.16
   
   > Пример полного использования:
      Загрузка дистрибутивов платформа 1С.Предприятие последней версии 8.3.18 и 1C:EDT версии 2020.6.2 для OS Windows
      - oneget get platform:win.x64@latest:8.3.18 edt:win@2020.6.2 
      
      Загрузка дистрибутивов платформа 1С.Предприятие последней версии 8.3.18 и 1C:EDT версии 2020.6.2 для OS X (Mac OS)
      - oneget get platform:mac.x64@latest:8.3.18 edt:mac@2020.6.2 

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

		&cli.StringSliceFlag{
			Destination: &c.Filter,
			EnvVars:     []string{"ONEGET_FILTER"},
			Aliases:     []string{"F"},
			Name:        "filter",
			Usage: `Дополнительный фильтр пакетов (регулярное выражение)
                          Задается для каждого типа релиза отдельно. 
                          Например, edt=".*JDK.*"
`,
		},
		&cli.StringFlag{
			Destination: &c.BaseDir,
			Name:        "path",
			Aliases:     []string{"out"},
			Value:       getAbsolutePath("downloads"),
			DefaultText: getAbsolutePath("downloads"),
			Usage:       "Путь к каталогу выгрузки",
		},
		&cli.BoolFlag{
			Name:        "extract",
			Destination: &c.Extract,
			Aliases:     []string{"E"},
			EnvVars:     []string{"ONEGET_EXTRACT"},
			Value:       false,
			Usage:       "Распаковывать дистрибутив (только для файлов tar.gz)",
		},
		&cli.BoolFlag{
			Name:        "rename",
			Aliases:     []string{"R"},
			Destination: &c.Rename,
			EnvVars:     []string{"ONEGET_EXTRACT_RENAME"},
			Value:       false,
			Usage: `Переименовывать дистрибутивы при распаковке. 
				Примеры: 
					1c-enterprise-8.3.18.1334-client_8.3.18-1334_amd64.deb -> client-8.3.18.1334.deb
					1c-enterprise83-server_8.3.16-1876_amd64.deb -> server_8.3.16-1876.deb`,
		},
	}
}

func (c *getCmd) extractFiles(files []string) error {

	log.Infof("Extracting <files <%d>", len(files))

	var mErr error
	for _, file := range files {
		if strings.ToLower(filepath.Ext(file)) == ".tar.gz" {
			continue
		}

		extractDir := strings.Replace(file, "_", ".", -1) + ".extract"

		if len(c.ExtractDir) > 0 {
			_, filename := filepath.Split(file)
			extractDir = filepath.Join(c.ExtractDir, filename+"_extract")
		}

		err := unpacker.Extract(file, extractDir)
		if err != nil {
			log.Errorf(err.Error())
			multierr.Append(mErr, err)
			continue
		}

		if c.Rename {
			err := renameFiles(extractDir)
			if err != nil {
				multierr.Append(mErr, err)
				continue
			}

		}
	}

	return mErr
}

func renameFiles(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Errorf("Error find files in dir <%s> to rename: %s", dir, err.Error())
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		oldName := file.Name()
		newName := unpacker.GetAliasesDistrib(oldName)
		err := os.Rename(
			filepath.Join(dir, oldName),
			filepath.Join(dir, newName))
		if err != nil {
			log.Errorf("Error rename file <%s> to <%s>: %s", oldName, newName, err.Error())
			continue
		}

	}
	return nil
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

func getAdditionalFilters(arr []string, sep string, defValue string) map[string][]string {

	result := make(map[string][]string)

	for _, str := range arr {

		values := strings.SplitN(str, sep, 2)

		key := values[0]
		value := defValue

		if len(values) == 2 {
			value = values[1]
		}

		if result[key] == nil {
			result[key] = []string{}
		}

		result[key] = append(result[key], value)

	}

	return result
}

func getProjectId(project string) string {

	values := strings.SplitN(project, ":", 2)

	key := values[0]

	return key

}

func compileFilters(filters ...string) []dloader.FileFilter {
	var result []dloader.FileFilter
	for _, filter := range filters {

		compile := regexp.MustCompile(filter)

		result = append(result, compile)
	}

	return result
}

func getProjectFilter(project string) (dloader.FileFilter, error) {

	values := strings.SplitN(project, ":", 2)

	key := values[0]

	if len(values) == 2 {
		return dloader.NewFileFilter(dloader.GetProjectIDByAlias(key), values[1])
	}
	return nil, nil
}

func getAbsolutePath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Errorf("Error getting current path %s", err.Error())
	}
	return path.Join(dir, p)
}
