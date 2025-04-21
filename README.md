## PVZ Backend Service

Небольшой демо-сервис на Go для работы пвз. Можно создавать ПВЗ, открывать приёмки товаров, добавлять товары в приёмку и создавать новых     сотрудников.

Проект создан как тестовое задание для стажировки, по сути в учебных целях, поэтому код простой и понятный (но это не точно). Использовал генерацию HTTP API через OpenAPI, gRPC, PostgreSQL, Gin, JWT, Prometheus.

## Что под капотом:
    HTTP REST API на Gin (генерируется из OpenAPI 3)
    Немного gRPC для демонстрации альтернативы REST
    Простенький репозиторий с PostgreSQL
    JWT-авторизация (с заглушкой /dummyLogin для тестов)
    Метрики Prometheus 
    Docker Compose для удобного запуска 
    Тесты: юнит-тесты с покрытием больше 90% + интеграционный тест

## Запуск и проверка
    1. docker compose up --build -d

    2. Проверка работы сервиса
    Получите тестовый JWT-токен для роли модератора:

    curl -X POST http://localhost:8080/dummyLogin \
    -H 'Content-Type: application/json' \
    -d '{"role":"moderator"}'

    Скопируйте полученный токен и создайте новый ПВЗ (например, в Казани):
    curl -X POST http://localhost:8080/pvz \
    -H 'Authorization: Bearer ВАШ_ТОКЕН' \
    -H 'Content-Type: application/json' \
    -d '{"city":"Казань"}'

    3. Проверка метрик

    # 1) Количество созданных ПВЗ
    curl -s http://localhost:9000/metrics \
    | grep '^pvz_created_total ' \
    | awk '{print $2}'
    
    # 2) Количество созданных приёмок
    curl -s http://localhost:9000/metrics \
    | grep '^reception_created_total ' \
    | awk '{print $2}'
    
    # 3) Количество добавленных товаров
    curl -s http://localhost:9000/metrics \
    | grep '^products_added_total ' \
    | awk '{print $2}'
    
    # 4) Общее число HTTP‑запросов (сумма по всем методам/пути/статусам)
    curl -s http://localhost:9000/metrics \
    | grep '^http_requests_total' \
    | awk '{sum += $2} END {print sum}'

    # 5) Среднее время ответа: (sum/count)
    curl -s http://localhost:9000/metrics \
    | awk '/^http_request_duration_seconds_sum/ {s=$2} /^http_request_duration_seconds_count/ {c=$2} END{printf "%.3f\n", s/c}'

    4. Проверка gRPC
    grpcurl -plaintext localhost:3000 list
