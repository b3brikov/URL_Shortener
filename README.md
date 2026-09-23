# URL Shortener

Сервис сокращения URL на Go.

Проект реализует REST API для создания коротких ссылок и получения исходного URL. Для хранения данных используется PostgreSQL, Redis используется для кеширования и накопления статистики переходов.

## Стек

* Go
* PostgreSQL
* Redis
* JWT
* Docker / Docker Compose
* net/http
* gRPC
* `slog`

## Архитектура

```text
Client
  │
  ▼
HTTP API
  │
  ▼
Handler
  │
  ▼
Service
  │
  ├──► Redis
  │
  └──► Repository
          │
          ▼
      PostgreSQL
```

Для работы со статистикой переходов используется background worker:

```text
Request
  │
  ▼
Redis
(click counter)
  │
  ▼
Background Worker
  │
  ▼
PostgreSQL
```

Также сервис предоставляет gRPC API для взаимодействия с другими сервисами.

## Возможности

* создание коротких URL;
* получение оригинального URL по короткому коду;
* JWT-аутентификация;
* PostgreSQL для постоянного хранения;
* Redis-кеширование;
* накопление статистики переходов через background worker;
* graceful shutdown;
* HTTP API;
* gRPC API;

## Запуск

### Docker Compose

```bash
docker compose up --build
```

После запуска сервисы будут доступны согласно настройкам из `.env`.

## Также есть возможность использовать make

```bash
make run
```

### Локальный запуск

Необходимы:

* Go
* PostgreSQL
* Redis

Создайте `.env` на основе используемой конфигурации проекта и запустите:

```bash
go run ./cmd
```

или соответствующую точку входа проекта.

## gRPC

gRPC API используется Agent-проектом для взаимодействия с Shortener.

Основные операции:

```text
CreateNewCode
GetOriginalUrl
```

Proto-контракты находятся в:

```text
internal/core/transport/gRPC/proto
```

## Graceful Shutdown

При получении `SIGINT` или `SIGTERM` приложение корректно завершает:

* HTTP server;
* gRPC server;
* background worker.

Для shutdown используется общий timeout context.

## Статус

Проект используется как backend-проект для практики:

* Go backend;
* PostgreSQL и Redis;
* gRPC;
* background workers;
* graceful shutdown;
* Docker;
* архитектуры handler → service → repository;
* взаимодействия микросервисов;
