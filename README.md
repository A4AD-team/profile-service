# profile-service

Управление профилями пользователей форума.

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)

## Функции

- Просмотр / редактирование профиля
- Аватар (ссылка)
- Статистика: посты, комментарии, репутация
- Публичная страница профиля

## Технологии

- Go 1.23+
- PostgreSQL + pgx / sqlc
- Fiber / Gin
- Redis (кэш профиля)

## Структура

```
profile-service/
├── cmd/
│   └── server/
├── internal/
│   ├── handler/
│   ├── repository/
│   └── service/
├── migrations/
└── docker-compose.yml
```


## Запуск

```bash
docker compose up -d postgres-profile redis

go run ./cmd/server
```
