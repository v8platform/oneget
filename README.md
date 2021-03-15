# oneget
Консольная утилита для загрузки пакетов с releases.1c.ru


[![Release](https://img.shields.io/github/release/v8platform/oneget.svg?style=for-the-badge)](https://github.com/v8platform/oneget/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](/LICENSE.md)
[![Build status](https://img.shields.io/github/workflow/status/v8platform/oneget/goreleaser?style=for-the-badge)](https://github.com/v8platform/oneget/actions?workflow=goreleaser)
[![Codecov branch](https://img.shields.io/codecov/c/github/v8platform/oneget/master.svg?style=for-the-badge)](https://codecov.io/gh/v8platform/oneget)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](http://godoc.org/github.com/v8platform/oneget)
[![SayThanks.io](https://img.shields.io/badge/SayThanks.io-%E2%98%BC-1EAEDB.svg?style=for-the-badge)](https://saythanks.io/to/khorevaa)
[![Powered By: GoReleaser](https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=for-the-badge)](https://github.com/goreleaser)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)

### Команда `get`

Использование:

```shell
export ONEC_USERNAME=user
export $ONEC_PASSWORD=password
oneget get --path ./tmp/dist/ --nick platform83 --version 8.3.18.1334 --filter="deb64_.*.tar.gz$"

# or
oneget --user user --pwd password get --path ./tmp/dist/ --nick platform83 --version 8.3.18.1334 --filter="deb64_.*.tar.gz$"

```


## Запуск в докере

```shell
docker run -v $(pwd):/tmp/dist demoncat/oneget \
    --user $ONEC_USERNAME \
    --pwd $ONEC_PASSWORD \
    --path /tmp/dist/ \
    --nicks platform83 \
    --version-filter 8.3.16.1876 \
    --distrib-filter 'deb64.tar.gz$'
```

## Настройка логов

### Через файл настройки
Создать рядом с приложением файл `logos.yaml` с содержимым

```yaml
appenders:
  console:
    - name: CONSOLE
      target: stdout
      encoder:
        console:

  rolling_file:
    - name: FILE
      file_name: ./logs/oneget.log
      max_size: 100
      encoder:
        json:
loggers:
  root:
    level: info
    appender_refs:
      - CONSOLE
  logger:
    - name: "github.com/v8platform/oneget"
      appender_refs:
        - CONSOLE
        - FILE
      level: debug     

```

### Через переменные окружения
```shell
export LOGOS_CONFIG="appenders.rolling_file.0.name=FILE;
appenders.rolling_file.0.file_name=./logs/oneget.log;
appenders.rolling_file.0.max_size=100;
appenders.rolling_file.0.encoder.json;
loggers.logger.0.level=debug;
loggers.logger.0.name=github.com/v8platform/oneget;
loggers.logger.0.appender_refs.0=CONSOLE;
loggers.logger.0.appender_refs.1=FILE;"
```
#TODO

Идея и часть кода взята из [этого](https://github.com/korableg/Downloader1C) проекта 