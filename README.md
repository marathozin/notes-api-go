# Notes API

REST API для управления заметками с тегами и JWT-авторизацией.

Документация API: https://notes-api-go-dw61.onrender.com/swagger/index.html

**Стек:** 
- Go 1.26
- net/http
- pgx v5
- bcrypt
- JWT (HS256)

## Структура проекта

```
notes-api/
├── cmd/api/main.go                     # точка входа, сборка зависимостей
├── docs/                               # документация
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/
│   ├── config/config.go                # конфиг из env-переменных
│   ├── handler/
│   │   ├── auth_test.go                # тесты для auth handlers
│   │   ├── auth.go                     # register, login, refresh, me
│   │   ├── notes_test.go               # тесты для notes handlers
│   │   ├── notes.go                    # CRUD заметок
│   │   └── swagger_types.go            # типы для Swagger-схем ответов
│   ├── middleware/
│   │   ├── auth.go                     # проверка JWT Bearer
│   │   └── logging.go                  # логирование и recover panic
│   ├── model/
│   │   ├── note.go                     # Note model
│   │   └── user.go                     # User model
│   ├── service/
│   │   └── token.go                    # выдача и проверка JWT
│   └── store/
│       ├── memory.go                   # интерфейсы UserStore, NoteStore
│       └── postgres/
│           ├── db.go                   # создание pgxpool
│           ├── user.go                 # UserStore (postgres)
│           └── note.go                 # NoteStore (postgres)
├── testutil/
│   ├── helpers.go                      # вспомогательные функции, чтобы убрать повторяющийся бойлерплейт из тестов
│   └── mocks.go                        # mock-реализации всех store-интерфейсов, работающие в памяти
├── migrations
│   ├── 001_create_users_table.down.sql # откат таблицы Users
│   ├── 001_create_users_table.up.sql   # создание таблицы Users
│   ├── 002_create_notes_table.down.sql # откат таблицы Notes
│   └── 002_create_notes_table.up.sql   # создание таблицы Users
├── pkg/response/response.go            # вспомогательные функции JSON/Error
└── .env.example
```

---

## Быстрый старт

### 1. Клонируйте и установите зависимости

```bash
git clone https://github.com/marathozin/notes-api-go.git
cd notes-api-go
go mod tidy
```

### 2. Настройте окружение
```bash
cp .env.example .env
# отредактируйте .env
# сгенерируйте JWT_SECRET: openssl rand -hex 64
# docker compose по умолчанию использует DATABASE_URL из .env.example
```

### 4. Запуск
```bash
docker compose up --build -d
```

## API

### Авторизация

| Метод | Путь              | Описание                                             | 
|-------|-------------------|------------------------------------------------------|
| POST  | /auth/register    | Регистрация                                          |
| POST  | /auth/login       | Вход, возвращает токены                              |
| POST  | /auth/refresh     | Обновить access по refresh токену                    |
| GET   | /auth/me          | Данные текущего пользователя (требуется авторизация) |

### Заметки (все требуют авторизации)

| Метод  | Путь          | Описание                        |
|--------|---------------|---------------------------------|
| GET    | /notes        | Список заметок текущего юзера   |
| POST   | /notes        | Создать заметку                 |
| GET    | /notes/{id}   | Получить заметку                |
| PUT    | /notes/{id}   | Обновить заметку                |
| DELETE | /notes/{id}   | Удалить заметку                 |

---

## Примеры запросов

```bash
# Регистрация
curl -X POST localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","username":"user","password":"secret123"}'

# Вход
curl -X POST localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"secret123"}'

# Создать заметку
curl -X POST localhost:8080/notes \
  -H 'Authorization: Bearer <access_token>' \
  -H 'Content-Type: application/json' \
  -d '{"title":"Заметка 1","content":"Текст заметки"}'

# Список заметок
curl localhost:8080/notes \
  -H 'Authorization: Bearer <access_token>'

# Обновить refresh токен
curl -X POST localhost:8080/auth/refresh \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token":"<refresh_token>"}'
```

## Планируемые улучшения:
1. Пагинация
2. Поиск по заметкам
3. Rate limiting
