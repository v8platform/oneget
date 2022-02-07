package unpacker

import (
	"github.com/khorevaa/logos"
	"github.com/mholt/archiver/v3"
	"regexp"
	"strings"
)

var log = logos.New("github.com/v8platform/oneget/unpacker").Sugar()

var re = regexp.MustCompile(`^1c-enterprise[\d]*-[\d\.\d\.\d*\.\d*]*-*([a-z-]*)_([\d\.\d\.\d*\.\d*-]*)_(amd64)\.([a-z]*)$`)

func Extract(filename string, destinatin string) error {
	err := archiver.Unarchive(filename, destinatin)
	if err != nil && strings.Contains(err.Error(), "file already exist") {
		log.Warnf("the target files already exist %s", filename)
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}

func GetAliasesDistrib(fileName string) string {
	resultFileName := re.ReplaceAllString(fileName, `$1-$2.$4`)
	return resultFileName
}
