package downloader

import (
	"fmt"
	"github.com/xelaj/go-dry"
	"regexp"
	"strings"
)

const (
	Platform83Project = "Platform83"
	EDTProject        = "DevelopmentTools10"
)
const (
	x64re     = "(?smU)(?:64-bit|64 бит).*"
	rpmre     = "(?smU)(?:RPM|ОС Linux|для Linux$|tar.bz2).*"
	debre     = "(?smU)(?:DEB|ОС Linux|для Linux$|tar.bz2).*"
	windowsre = "(?smU)(?:Windows|ОС Windows|zip).*"
	osxre     = "(?smU)(?:OS X|macOS|MacOS|ОС macOS).*"

	clientre      = "(?smU)Клиент"
	serverre      = "(?smU)(?:Cервер|Сервер)"
	thinre        = "(?smU)Тонкий клиент"
	offlinere     = "(?smU)оффлайн установки"
	fullwindowsre = "(?smU)Технологическая платформа"
)

/*

	Специальные фильтры для платформы и ряда других приложение

	platform:thin.mac@latest
	platform:full.win.x64@latest
	platform:win.x64@latest
	platform:server.deb.x64@latest


*/

var (
	ProjectAliases = map[string]string{
		"platform": Platform83Project,
		"edt":      "DevelopmentTools10",
		"ring":     "EnterpriseLicenseTools",
		"executor": "Executor",
		"pg":       "AddCompPostgre",
	}

	shortFilters = map[string]string{
		"mac":         osxre,
		"windows":     windowsre,
		"win":         windowsre,
		"deb":         debre,
		"rpm":         rpmre,
		"x64":         x64re,
		"client":      clientre,
		"server":      serverre,
		"thin-client": thinre,
		"thin":        thinre,
		"offline":     offlinere,
		"full":        fullwindowsre,
	}

	macSkipFilters = []string{
		"x64",
		"server",
	}

	x64Regexp = regexp.MustCompile(x64re)
)

type PlatformMatchFilter struct {
	filters      []*regexp.Regexp
	x64bitMatch  bool
	x64bitRegexp *regexp.Regexp
}

func GetProjectIDByAlias(alias string) string {

	if val, ok := ProjectAliases[alias]; ok {
		return val
	}

	return alias

}

func NewFilter(project string, filter string) (Filter, error) {

	switch project {
	case Platform83Project:
		return newPlatformMatchFilter(filter)
	case EDTProject:
		return newEdtFilter(filter)
	default:
		return nil, fmt.Errorf("unknown filter builder for project <%s>", project)
	}
}

func NewFilterMust(project string, filter string) Filter {

	newFilter, err := NewFilter(project, filter)
	if err != nil {
		log.Fatal(err.Error())
		return nil
	}

	return newFilter
}

func newPlatformMatchFilter(filter string) (*PlatformMatchFilter, error) {

	m := &PlatformMatchFilter{
		x64bitMatch:  false,
		x64bitRegexp: x64Regexp,
	}

	filters := strings.Split(filter, ".")
	filters = removeFromFilter(filters, "x32")

	if ok := dry.StringInSlice("x64", filters); ok {
		filters = removeFromFilter(filters, "x64")
		m.x64bitMatch = true
	}

	// Для Windows если стоит только фильтр по нему и другого нет, то установим для платформы скачиваем полного дистрибутива
	if ok := dry.StringInSlice("win", filters) || dry.StringInSlice("windows", filters); ok && len(filters) == 1 {
		filters = append(filters, "full")
	}

	if ok := dry.StringInSlice("mac", filters); ok {
		filters = removeFromFilter(filters, "server")
		m.x64bitMatch = false
	}

	err := m.build(filters)

	return m, err
}

func (m *PlatformMatchFilter) MatchString(source string) bool {

	if m.x64bitRegexp.MatchString(source) != m.x64bitMatch {
		return false
	}

	for _, filter := range m.filters {

		if ok := filter.MatchString(source); !ok {
			return false
		}

	}

	return true
}

func (m *PlatformMatchFilter) build(filters []string) error {

	for _, filter := range filters {

		val, ok := shortFilters[filter]

		if !ok {
			return fmt.Errorf("unknown <%s> filter", filter)
		}

		m.filters = append(m.filters, regexp.MustCompile(val))

	}

	return nil
}

func removeFromFilter(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

type EdtMatchFilter struct {
	filters        []*regexp.Regexp
	matchOffline   bool
	matchOfflineRe *regexp.Regexp
	distrRe        *regexp.Regexp
}

func newEdtFilter(filter string) (*EdtMatchFilter, error) {
	m := &EdtMatchFilter{
		matchOffline:   false,
		matchOfflineRe: regexp.MustCompile(".*1C:EDT*"),
		distrRe:        regexp.MustCompile(".*оффлайн установки*"),
	}

	filters := strings.Split(filter, ".")
	filters = removeFromFilter(filters, "x32")
	filters = removeFromFilter(filters, "x64")
	if ok := dry.StringInSlice("offline", filters); ok {
		m.matchOffline = true
		filters = removeFromFilter(filters, "offline")
	}

	if ok := dry.StringInSlice("jdk", filters); ok {
		m.filters = append(m.filters, regexp.MustCompile(".*JDK.*"))
		filters = removeFromFilter(filters, "jdk")
	}

	err := m.build(filters)
	if err != nil {
		return nil, err
	}

	return m, err
}

func (m *EdtMatchFilter) build(filters []string) error {

	for _, filter := range filters {

		val, ok := shortFilters[filter]

		if !ok {
			return fmt.Errorf("unknown <%s> filter", filter)
		}

		m.filters = append(m.filters, regexp.MustCompile(val))

	}

	return nil
}

func (m *EdtMatchFilter) MatchString(source string) bool {

	for _, filter := range m.filters {

		if ok := filter.MatchString(source); !ok {
			return false
		}

	}

	if m.distrRe.MatchString(source) {
		offlineDistr := m.matchOfflineRe.MatchString(source)

		if !m.matchOffline && offlineDistr {
			return false
		}
	}

	return true
}
