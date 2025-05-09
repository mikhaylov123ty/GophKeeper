# GophKeeper
## Локальный запуск Сервера
Все команды рекомендуется запускать из корневой директории проекта

### БД
Установить и настроить Postgres для вашей платформы
```
https://www.postgresql.org/download/
```
### Сервер
1. Сгенерировать пару сертификат\ключ
```
go run ./cmd/certManager/cert.go
```
По умолчанию ключи создадутся в корне проекта

2. Собрать сервер

```
go build -ldflags "-X main.buildVersion=v1 -X "main.buildDate=$(date +"%Y/%m/%d")"" -o ./cmd/server/GophServer ./cmd/server/main.go
```
Скомпилированный файл будет лежать в папке проекта /cmd/server

3. Запуск сервера
###### C файлом конфигурации
Конфигурация находится в папке проекта /cmd/server вместе с исполняемым файлом
```
cmd/server/config.json
```
Все необходимы поля уже заполнены, необходимо только скорректировать:
* DSN для БД, если строка подключения отличается по умолчанию
* Migrations Dir - в корне проекта по умолчанию, изменить, если планируется перенести в другое место
* JWT key - задать на свое усмотрение
* Crypto keys - в корне проекта по умолчанию, изменить, если планируется перенести в другое место

Пример строки запуска:
```
./cmd/server/GophServer -config cmd/server/config.json
```

###### С флагами запуска:
```
-host -имя хоста сервера
-grpc-port - порт gRPC сервера
-l - уровень логирования
-d - DSN БД
-m - директория с миграциями
-private-key - путь к private.key
-certificate - путь к public.crt
-jwt-key - ключ подписи JWT
-config - путь к файлу конфигурации
-a - альтернатива флагам -host + -grpc-port, принимает целиком адрес,
прим. localhost:443
```
Пример строки запуска:
```
./cmd/server/server -grpc-port 4443 -l info -d "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" -m "file://./migrations" -private-key "./private.key" -certificate "./public.key" -jwt-key 123
```

## Запуск клиента

Выбрать клиент для своей платформы из папки:
```
cmd/client/
```

###### С файлом конфигурации
Конфигурация находится в папке проекта /cmd/server вместе с исполняемым файлов
```
cmd/server/config.json
```

Пример строки запуска:
```
./cmd/client/yourClient -config ./cmd/client/config.json
```
Все необходимы поля уже заполнены, необходимо только скорректировать:
* Public cert - в корне проекта по умолчанию, изменить, если планируется перенести в другое место
* Files Output Folder - указать папку для сохранения скачаных файлов из приложения. Необходимо создать предварительно

###### С флагами конфгурации
Флаги запуска:
```
-host -имя хоста сервера
-grpc-port - порт gRPC сервера
-certificate - путь к public.crt
-files-output -путь к папке для сохранения скачаных файлов из приложения
-a - альтернатива флагам -host + -grpc-port, принимает целиком адрес,
прим. localhost:443
```
Пример для Mac OS:
```
./cmd/client/yourClient -grpc-port 4443 -certificate ./public.crt -files-output "/Users/your user name/Downloads/Output"
```
Пример для Windows:
```
./cmd/client/yourClient -grpc-port 4443 -certificate ./public.crt -files-output "C:/Users/your user name/Downloads/Output"
```

После запуска следовать инструкциям внизу экрана

#### В случае перемещения исполняемого файла или файла конфигурации - необходимо скорректировать пути к ключам, файлам миграции соответственно

## Запуск сервера в Docker контейнере
1. Установить Docker.

После установки проверить командой,
```
docker -v
```
В результате должна появиться версия:
```
Docker version 28.1.1, build 4eba377
```

2. Из директории проекта собрать образ сервера
```
docker build --no-cache -t gopher_keeper_server:v1 .
```
3. В корне проекта создать .env файл и добавить переменные окружения
Пример:
```
JWT_KEY: 1234
PRIVATE_KEY: ./private.key
CERTIFICATE: ./public.crt
DATABASE_DSN: postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable
MIGRATIONS_DIR: file://./migrations
ADDRESS: 0.0.0.0:4443
LOG_LEVEL: info

POSTGRES_DB: postgres
POSTGRES_USER: postgres
POSTGRES_PASSWORD: postgres
```
4. Поменять маппинг локальной директории в 51й строчке файла ```docker-compose.yaml``` 
на свою локальную директорию

Пример:
```
device:  /Users/your user name/Downloads/shared/
```

5. Запустить контейнер с сервисами БД, созданием серификатов, сервером
```
docker compose up --build --no-recreate
```
Успешный запуск сопровождается сообщением о старте сервера после миграции и создания пары ключей

Пример:
```
level=INFO msg="starting server" address=0.0.0.0:4443
```

6. Перейти в локальную директорию из П.4. Забрать сгенерированный ```public.crt``` для клиента 
и положить его в корень каталога проекта.

7. Запустить клиент по инструкции.

### ВАЖНО!
Шифрование данных производится сертификатом, при его утере или замене зашифрованные другим сертификатом данные не смогут расшифроваться.


### Запуск клиента на другом ПК\вне проекта
1. Скопировать клиент для своей платформы из папки:
```
cmd/client/
```
2. Скопировать файл конфигурации клиента из папки (не требуется, если планируется запуск через флаги):
```
cmd/client/
```
3. Скопировать файл сертификата в зависимости от варианта запуска сервера.

Локальная сборка - корень проекта

Docker контейнер - общая папка на локальном диске

4. Сложить все в одну папку и скорректировать имя хоста сервера и output папку для сохранения скачаных файлов.

Формат пути для Windows:
```
C:/Users/your user name/Downloads/GophKeeper/Output
```
Формат пути для MacOS и Linux

```
/Users/your user name/Downloads/GophKeeper/Output
```
В самом приложении соблюдать форматирование не нужно.