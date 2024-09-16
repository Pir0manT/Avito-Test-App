# My Go Application

## Описание
Это пример Go приложения, которое запускается в Docker контейнере.

## Требования
Для правильной работы приложения необходимы следующие переменные окружения:
- `SERVER_ADDRESS`: Адрес сервера (например, `0.0.0.0:8080`)
- `POSTGRES_CONN`: Строка подключения к базе данных PostgreSQL (например, `postgres://{username}:{password}@{host}:{5432}/{dbname}`)
- `GIN_MODE`: Режим работы Gin (например, `release`)

## Сборка и запуск приложения в Docker

### Шаг 1: Клонирование репозитория
```sh
git clone https://github.com/Pir0manT/Avito-Test-App.git
cd Avito-Test-App
```

### Шаг 2: Сборка Docker образа
```sh
docker build -t my-go-app .
```

### Шаг 3: Запуск приложения
```sh
docker run -d -p 8080:8080 \
  -e SERVER_ADDRESS=0.0.0.0:8080 \
  -e POSTGRES_CONN=postgres://{username}:{password}@{host}:{5432}/{dbname} \
  -e GIN_MODE=release \
  my-go-app
```

## Сборка приложения при помощи make, необходимо наличие установленного GoLang версии 1.23.1

### Шаг 1: Клонирование репозитория
```sh
git clone https://github.com/Pir0manT/Avito-Test-App.git
cd Avito-Test-App
```

### Шаг 2: Сборка 
```sh
make
```
#### После окончания сборки в папке bin будут доступны исполняемые файлы.