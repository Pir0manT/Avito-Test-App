# Используем официальный образ Go
FROM golang:1.23.1-alpine

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы проекта
COPY . .

# Устанавливаем зависимости
RUN go mod download

# Собираем проект
RUN go build -o main main.go

# Экспонируем порт
EXPOSE 8080

# Запускаем проект
CMD ["./main"]