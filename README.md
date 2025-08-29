# To-Do List API

RESTful API для приложения менеджера задач, написанный на Go, PostgreSql, Docker Compose, JWT.

### Требования для запуска:
-   Docker
-   Docker Compose

### Запуск:

1.  **Клонирование репозитория:**
    ```bash
    git clone https://github.com/makson2134/go-to-do-list-rest
    cd go-to-do-list-rest
    ```

2.  **Получение JWT-ключа:**
    Самый простой способ получить ключ - сгенерировать его через openssl
    ```bash
    openssl rand -base64 32
    ```

3.  **Создание файла `.env`:**
    Создайте файл `.env` в корне проекта. Пример .env файла:
    ```env
    SERVER_PORT=8080
    SERVER_TIMEOUT=10s
    SERVER_IDLETIMEOUT=60s
    
    DB_HOST=postgres
    DB_PORT=5432
    DB_USER=todolist_user
    DB_PASSWORD=todolist_password
    DB_NAME=todolist_db
    DB_SSLMODE=disable
    DB_MAXIDLECONNS=5
    DB_MAXOPENCONNS=15
    DB_CONNMAXLIFETIME=10m
    
    DB_URL=postgres://todolist_user:todolist_password@postgres:5432/todolist_db?sslmode=disable
    
    #Ключ JWT
    JWT_SECRET_KEY=
    JWT_TTL=12h
    ```

4.  **Запуск в Docker Compose:**
    ```bash
    docker compose up -d
    ```
    Сервис будет доступен по порту, указанному в поле `SERVER_PORT`.

## Инструменты

-   **Язык:** Golang `1.24.4`
-   **Контейнеризация:** Docker, Docker Compose
-   **База данных:** PostgreSQL `17-alpine`
-   **Миграции:** `migrate/migrate:v4.17.1`

## Go Библиотеки

В проекте используется следующие библиотеки:

-   `github.com/go-chi/chi/v5 v5.2.2`  - роутер
-   `github.com/go-chi/render v1.0.3` - для работы с json 
-   `github.com/go-playground/validator/v10 v10.27.0` - для валидации вводимых полей
-   `github.com/golang-jwt/jwt/v5 v5.3.0` - для работы с JWT
-   `github.com/ilyakaznacheev/cleanenv v1.5.0` - для конфига
-   `github.com/joho/godotenv v1.5.1` - для получения переменных окружения
-   `github.com/lib/pq v1.10.9` - драйвер для PostgreSql
-   `github.com/stretchr/testify v1.11.1` - для тестов
-   `golang.org/x/crypto v0.41.0` - для bcrypt


## To-Do:

-    Вынести параметры логгера (уровень, формат) в `.env`.
-    Добавить TLS.
-    Дописать тесты.
-    Реализовать документацию API с помощью Swagger/OpenAPI.
-    Настроить CI/CD пайплайн.
-    Подключить метрики Grafana + Prometheus.
