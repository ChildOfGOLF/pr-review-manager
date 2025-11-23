# PR Review Manager

Сервис для автоматического назначения ревьюверов на Pull Request'ы.

## Стек

- Go 1.23 (chi router)
- PostgreSQL 15
- Docker & Docker Compose

## Быстрый старт

Запуск сервиса и базы данных:

```bash
docker-compose up --build
```

## Особенности реализации и решения

В ходе разработки были приняты следующие решения для соответствия ТЗ:

1. **Выбор ревьюверов**: Используется `rand.Shuffle` для случайного выбора из активных участников команды.
2. **Переназначение**: Новый ревьювер выбирается из команды *заменяемого* участника (согласно ТЗ), а не автора PR.
3. **Идемпотентность Merge**: Повторный вызов `/merge` не возвращает ошибку, а отдает текущий статус.
4. **Массовая деактивация**: Реализована через batch-запросы в одной транзакции. Это позволяет обрабатывать большие объемы данных (60+ PR) быстрее 100мс.
   - Сначала обновляем статусы пользователей.
   - Затем одним запросом удаляем их из ревьюверов.
   - Одним запросом находим замены и вставляем новых.

   **Результаты замеров:**
   - 3 пользователя + 5 PR: 76ms
   - 3 пользователя + 10 PR: 90ms
   - 5 пользователей + 40 PR: 94ms
   - 6 пользователей + 60 PR: 180ms
5. **ID**: Используются строковые ID (как в GitHub/GitLab), а не числовые.

## API

Основные примеры запросов.

### Команды

#### Создать команду

```bash
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {"user_id": "u1", "username": "Vadim", "is_active": true},
      {"user_id": "u2", "username": "Dasha", "is_active": true}
    ]
  }'
```

#### Массовая деактивация (с переназначением PR)

```bash
curl -X POST http://localhost:8080/team/deactivate \
  -H "Content-Type: application/json" \
  -d '{"team_name": "backend"}'
```

### Пользователи

#### Сменить активность

```bash
curl -X POST http://localhost:8080/users/setIsActive \
  -H "Content-Type: application/json" \
  -d '{"user_id": "u2", "is_active": false}'
```

#### Посмотреть назначенные PR

```bash
curl "http://localhost:8080/users/getReview?user_id=u1"
```

### Pull Requests

#### Создать PR (автоматически назначит ревьюверов)

```bash
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "pull_request_name": "Feature X",
    "author_id": "u1"
  }'
```

#### Смержить

```bash
curl -X POST http://localhost:8080/pullRequest/merge \
  -H "Content-Type: application/json" \
  -d '{"pull_request_id": "pr-1001"}'
```

#### Переназначить ревьювера

```bash
curl -X POST http://localhost:8080/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "old_user_id": "u2"
  }'
```

## Разработка

**Локальный запуск (без Docker-контейнера приложения):**

```bash
# Поднять только базу
docker-compose up -d postgres

# Запустить приложение
export DB_HOST=localhost DB_PORT=5432 DB_USER=pruser DB_PASSWORD=prpass DB_NAME=pr_review_db DB_SSLMODE=disable
make run
```

**Тесты:**

```bash
# Интеграционные тесты
go test -v ./tests/integration/...

# Нагрузочные (k6)
k6 run tests/load/load_test.js
```

**Линтер:**

Используется `golangci-lint` с настройками для `gosimple`, `staticcheck`, `gosec` и др.

```bash
golangci-lint run
```
