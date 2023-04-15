package downloader

import (
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_filterReleaseFiles(t *testing.T) {
	type args struct {
		list              []ReleaseFileInfo
		filters           []FileFilter
		additionalFilters []FileFilter
	}
	tests := []struct {
		name             string
		args             args
		wantFilteredList int
	}{
		{
			"Platform83Project with addin_8_3_18_1363",
			args{
				list: []ReleaseFileInfo{
					{
						"Тонкий клиент 1С:Предприятия для DEB-based Linux-систем",
						"/version_file?nick=Platform83&ver=8.3.18.1363&path=Platform%5C8_3_18_1363%5Cthin.client_8_3_18_1363.deb32.tar.gz",
					},
					{
						"Тонкий клиент 1С:Предприятия для RPM-based Linux-систем",
						"/version_file?nick=Platform83&ver=8.3.18.1363&path=Platform%5C8_3_18_1363%5Cthin.client_8_3_18_1363.rpm32.tar.gz",
					}, {
						"Технология внешних компонент",
						"/version_file?nick=Platform83&ver=8.3.18.1363&path=Platform%5C8_3_18_1363%5Caddin_8_3_18_1363.zip",
					},
				},
				filters: []FileFilter{
					regexp.MustCompile(".*deb32.tar.gz"),
					regexp.MustCompile("Технология внешних компонент"),
					NewFileFilterMust(Platform83Project, "thin.deb"),
				},
			},
			2,
		},
		{
			"DevelopmentTools10 with Bellsoft JDK",
			args{
				list: []ReleaseFileInfo{
					{
						"Дистрибутив 1C:EDT для ОС Windows 64 бит",
						"/version_file?nick=DevelopmentTools10&ver=2020.6.2&path=DevelopmentTools%5C2020_6_2%5C1c_edt_distr_2020.6.2_8_windows_x86_64.zip",
					},
					{
						"Дистрибутив для оффлайн установки 1C:EDT для ОС Linux 64 бит",
						"/version_file?nick=DevelopmentTools10&ver=2020.6.2&path=DevelopmentTools%5C2020_6_2%5C1c_edt_distr_offline_2020.6.2_8_linux_x86_64.tar.gz",
					}, {
						"Bellsoft JDK Full (64-bit) для Windows",
						"/version_file?nick=DevelopmentTools10&ver=2020.6.2&path=DevelopmentTools%5C2020_6%5Cbellsoft_jdk11.0.9_12_windows_amd64_full.msi",
					}, {
						"Дистрибутив для оффлайн установки 1C:EDT для ОС Windows 64 бит",
						"/version_file?nick=DevelopmentTools10&ver=2020.6.2&path=DevelopmentTools%5C2020_6_2%5C1c_edt_distr_offline_2020.6.2_8_windows_x86_64.zip",
					},
				},
				filters: []FileFilter{
					NewFileFilterMust(EDTProject, "deb"),
				},
			},
			1,
		},
		{
			"DevelopmentTools10 with Bellsoft JDK",
			args{
				list: []ReleaseFileInfo{
					{
						"Дистрибутив для оффлайн установки 1C:EDT для ОС Linux 64 бит",
						"/version_file?nick=DevelopmentTools10&ver=2020.6.2&path=DevelopmentTools%5C2020_6_2%5C1c_edt_distr_offline_2020.6.2_8_linux_x86_64.tar.gz",
					},
					{
						"Дистрибутив 1C:EDT для ОС Linux 64 бит",
						"/version_file?nick=DevelopmentTools10&ver=2020.6.2&path=DevelopmentTools%5C2020_6_2%5C1c_edt_distr_2020.6.2_8_linux_x86_64.tar.gz",
					}, {
						"Bellsoft JDK Full (64-bit) для DEB-based Linux-систем",
						"/version_file?nick=DevelopmentTools10&ver=2020.6.2&path=DevelopmentTools%5C2020_6%5Cbellsoft_jdk11.0.9_12_linux_amd64_full.deb",
					}, {
						"Bellsoft JDK Full (64-bit) для RPM-based Linux-систем",
						"/version_file?nick=DevelopmentTools10&ver=2020.6.2&path=DevelopmentTools%5C2020_6%5Cbellsoft_jdk11.0.9_12_linux_amd64_full.rpm",
					},
				},
				filters: []FileFilter{
					NewFileFilterMust(EDTProject, "deb"),
				},
			},
			2,
		},
		{
			"DevelopmentTools10 only Bellsoft JDK",
			args{
				list: []ReleaseFileInfo{
					{
						"Дистрибутив для оффлайн установки 1C:EDT для ОС Linux 64 бит",
						"/version_file?nick=DevelopmentTools10&ver=2020.6.2&path=DevelopmentTools%5C2020_6_2%5C1c_edt_distr_offline_2020.6.2_8_linux_x86_64.tar.gz",
					},
					{
						"Дистрибутив 1C:EDT для ОС Linux 64 бит",
						"/version_file?nick=DevelopmentTools10&ver=2020.6.2&path=DevelopmentTools%5C2020_6_2%5C1c_edt_distr_2020.6.2_8_linux_x86_64.tar.gz",
					}, {
						"Bellsoft JDK Full (64-bit) для DEB-based Linux-систем",
						"/version_file?nick=DevelopmentTools10&ver=2020.6.2&path=DevelopmentTools%5C2020_6%5Cbellsoft_jdk11.0.9_12_linux_amd64_full.deb",
					}, {
						"Bellsoft JDK Full (64-bit) для RPM-based Linux-систем",
						"/version_file?nick=DevelopmentTools10&ver=2020.6.2&path=DevelopmentTools%5C2020_6%5Cbellsoft_jdk11.0.9_12_linux_amd64_full.rpm",
					},
				},
				filters: []FileFilter{
					NewFileFilterMust(EDTProject, "deb.jdk.x64"),
				},
			},
			1,
		},
		{
			"PosqtreSQL with 14.1-2.1C",
			args{
				list: []ReleaseFileInfo{
					{
						"Дистрибутив СУБД PostgreSQL для Windows (64-bit) одним архивом",
						"/version_file?nick=AddCompPostgre&ver=14.1-2.1C&path=AddCompPostgre%5c14_1_2_1C%5cpostgresql_14.1_2.1C_x64.zip",
					}, {
						"Дистрибутив СУБД PostgreSQL для Linux x86 (64-bit) одним архивом (RPM)",
						"/version_file?nick=AddCompPostgre&ver=14.1-2.1C&path=AddCompPostgre%5c14_1_2_1C%5cpostgresql_14.1_2.1C_x86_64_rpm.tar.bz2",
					}, {
						"Дистрибутив СУБД PostgreSQL для Linux x86 (64-bit) (дополнительные модули) одним архивом (RPM)",
						"/version_file?nick=AddCompPostgre&ver=14.1-2.1C&path=AddCompPostgre%5c14_1_2_1C%5cpostgresql_14.1_2.1C_x86_64_addon_rpm.tar.bz2",
					}, {
						"Дистрибутив СУБД PostgreSQL для Linux x86 (64-bit) одним архивом (DEB)",
						"/version_file?nick=AddCompPostgre&ver=14.1-2.1C&path=AddCompPostgre%5c14_1_2_1C%5cpostgresql_14.1_2.1C_amd64_deb.tar.bz2",
					}, {
						"Дистрибутив СУБД PostgreSQL для Linux x86 (64-bit) (дополнительные модули) одним архивом (DEB)",
						"/version_file?nick=AddCompPostgre&ver=14.1-2.1C&path=AddCompPostgre%5c14_1_2_1C%5cpostgresql_14.1_2.1C_amd64_addon_deb.tar.bz2",
					},
				},
				filters: []FileFilter{
					NewFileFilterMust(PostgreSQLProject, "deb.x64"),
				},
			},
			2,
		},
		{
			"Platform83Project with --filter",
			args{
				list: []ReleaseFileInfo{
					{
						"Тонкий клиент 1С:Предприятия для Linux",
						"/version_file?nick=Platform83&ver=8.3.20.2180&path=Platform%5C8_3_20_2180%5Cthin.client32_8_3_20_2180.tar.gz",
					},
					{
						"Технологическая платформа 1С:Предприятия (64-bit) для Windows",
						"https://releases.1c.ru/https://releases.1c.ru/version_file?nick=Platform83&ver=8.3.20.2180&path=Platform%5C8_3_20_2180%5Cwindows64full_8_3_20_2180.rar",
					},
					{
						"Технологическая платформа 1С:Предприятия (64-bit) для Windows + Тонкий клиент для Windows, Linux и MacOS для автоматического обновления клиентов через веб-сервер",
						"https://releases.1c.ru/version_file?nick=Platform83&ver=8.3.21.1302&path=Platform%5С8_3_21_1302%5Сwindows64full_with_all_clients_8_3_21_1302.rar",
					},
				},
				filters: []FileFilter{
					NewFileFilterMust(Platform83Project, "win.full.x64"),
				},
				additionalFilters: []FileFilter{
					regexp.MustCompile("windows64full_8"),
				},
			},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotFilteredList := filterReleaseFiles(tt.args.list, tt.args.filters, tt.args.additionalFilters); !reflect.DeepEqual(len(gotFilteredList), tt.wantFilteredList) {
				t.Errorf("filterReleaseFiles() = %v, want %v", len(gotFilteredList), tt.wantFilteredList)
			}
		})
	}
}

func TestOnegetDownloader_Platform_getFilesExist(t *testing.T) {
	platform := "Platform83"
	// Список существующих платформ
	versions := []string{
		"8.3.21.1302",
		"8.3.21.1302",
		"8.3.22.1672",
		"8.3.21.1607",
	}
	for _, ver := range versions {
		assert.True(t, releaseExist(t, platform, ver))
	}

}

// На текущий момент тест не актуален, так как релизы все доступны
func TestOnegetDownloader_Platform_getFilesNotExist(t *testing.T) {
	platform := "Platform83"
	// Список существующих платформ
	versions := []string{
		"8.3.22.1672",
	}
	for _, ver := range versions {
		assert.True(t, releaseExist(t, platform, ver))
	}
}

func getAuth(t *testing.T) (string, string) {
	login := os.Getenv("User")
	pwd := os.Getenv("Password")

	if login == "" || pwd == "" {
		t.Skipf("Параметры авторизации на https://releases.1c.ru не заданы. тест пропускаем")
	}
	return login, pwd
}

func releaseExist(t *testing.T, platform string, ver string) bool {
	login, pwd := getAuth(t)
	dl := NewDownloader(
		login,
		pwd,
	)
	dl.client, _ = NewClient(loginURL, releasesURL, login, pwd)

	versionFilter, _ := NewVersionFilter("", ver)
	config := DownloadConfig{
		Project:           platform,
		Version:           versionFilter,
		Filters:           nil,
		AdditionalFilters: nil,
	}
	releases, _ := dl.getProjectReleases(config)
	return len(releases) != 0

}
