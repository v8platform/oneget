# oneget
Консольная утилита для загрузки пакетов с releases.1c.ru


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