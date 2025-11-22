# Результаты нагрузочного тестирования

Провёл нагрузочное тестирование с помощью k6 для проверки производительности под продолжительной нагрузкой.

## Параметры теста

- **Инструмент**: k6 v0.55.0  
- **Окружение**: Docker Compose, PostgreSQL 15-alpine
- **Профиль нагрузки**: ступенчатый рост с долгим плато
  - 1мин: рост до 5 виртуальных пользователей
  - 5мин: удержание 5 VUs (основная фаза)
  - 1мин: рост до 10 VUs (пиковая нагрузка)
  - 1мин: снижение до 0 VUs
- **Длительность**: 8 минут
- **Тестовые данные**: 5 команд, 50 пользователей

## Сценарии

Смесь операций с весовыми коэффициентами:
- 25% - Создание PR (POST /pullRequest/create)
- 20% - Получение команды (GET /team/get)
- 15% - Изменение статуса (POST /users/setIsActive)
- 15% - Получение ревью (GET /users/getReview)
- 10% - Слияние PR (POST /pullRequest/merge)
- 15% - Статистика (GET /stats)

## Результаты

### Основные метрики

| Метрика | Результат | Требование |
|---------|-----------|------------|
| RPS | 25.64 | ≥ 5 |
| Медиана времени | 11.8 ms | - |
| p(90) времени | 68.95 ms | - |
| p(95) времени | 103.49 ms | - |
| Success rate | 100% | ≥ 99.9% |
| Error rate | 0.46% | < 1% |

### Подробные результаты

```
    THRESHOLDS
    errors.........................: 0.46%
    http_req_duration..............: p(95)=103.49ms
    http_req_failed................: 0.20%

    HTTP METRICS
    http_req_duration..: avg=26.82ms min=2.17ms med=11.8ms max=669.77ms
                        p(90)=68.95ms p(95)=103.49ms
    http_reqs..........: 12313 (25.64/s)
    data_received......: 992 MB (2.1 MB/s)
    data_sent..........: 2.0 MB (4.1 kB/s)

    CHECKS (99.74% success)
      reviews retrieved: 100%
      stats retrieved: 100%
      team retrieved: 100%
      PR created: 100%
      user updated: 100%
      response time < 300ms: 100% (12288/12288)
      team created: 100%

    ITERATIONS
    iterations.........: 13701 (28.53/s)
    iteration_duration.: avg=175ms med=173ms p(95)=315ms
    vus................: min=1 max=10
```

## Запуск

```bash
docker-compose up -d
k6 run tests/load/load_test.js
```
