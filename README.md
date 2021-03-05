# oneget
Консольная утилита для загрузки пакетов с releases.1c.ru


[![Release](https://img.shields.io/github/release/v8platform/oneget.svg?style=for-the-badge)](https://github.com/v8platform/oneget/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](/LICENSE.md)
[![Build status](https://img.shields.io/github/workflow/status/v8platform/oneget/goreleaser?style=for-the-badge)](https://github.com/v8platform/oneget/actions?workflow=releaser)
[![Codecov branch](https://img.shields.io/codecov/c/github/v8platform/oneget/master.svg?style=for-the-badge)](https://codecov.io/gh/v8platform/oneget)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](http://godoc.org/github.com/v8platform/oneget)
[![SayThanks.io](https://img.shields.io/badge/SayThanks.io-%E2%98%BC-1EAEDB.svg?style=for-the-badge)](https://saythanks.io/to/khorevaa)
[![Powered By: GoReleaser](https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=for-the-badge)](https://github.com/goreleaser)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)


## Запуск в докере

```shell
docker run -v $(pwd):/tmp/dist demoncat/oneget \
    --user $ONEC_USERNAME \
    --pwd $ONEC_PASSWORD \
    --path /tmp/dist/
    --nicks platform83 \
    --version-filter 8.3.16.1876 \
    --distrib-filter 'deb64.tar.gz$'
```

#TODO

Идея и часть кода взята из [этого](https://github.com/korableg/Downloader1C) проекта 