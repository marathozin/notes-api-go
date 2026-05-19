# Notes API

REST API для управления заметками на Go (net/http, без внешних зависимостей).

## Структура проекта

```
notes-api/
├── cmd/
│   └── api/
│       └── main.go          # точка входа, сборка зависимостей
├── internal/
│   ├── handler/
│   │   ├── notes.go         # CRUD-хэндлеры
│   │   └── routes.go        # регистрация маршрутов + middleware
│   ├── middleware/
│   │   └── logging.go       # Logging, RecoverPanic
│   ├── model/
│   │   └── note.go          # структуры Note, CreateNoteInput, UpdateNoteInput
│   └── store/
│       └── memory.go        # интерфейс NoteStore + in-memory реализация
└── pkg/
    └── response/
        └── response.go      # вспомогательные функции JSON/Error
```

**Правило разграничения пакетов:**
- `cmd/` — исполняемые приложения (wire-up зависимостей)
- `internal/` — код, не предназначенный для внешних пользователей
- `pkg/` — переиспользуемые утилиты без бизнес-логики

## Запуск

```bash
go run ./cmd/api          # порт по умолчанию 8080
PORT=3000 go run ./cmd/api
```

## API

| Метод  | Путь          | Описание              |
|--------|---------------|-----------------------|
| GET    | /notes        | Список всех заметок   |
| POST   | /notes        | Создать заметку       |
| GET    | /notes/{id}   | Получить заметку      |
| PUT    | /notes/{id}   | Обновить заметку      |
| DELETE | /notes/{id}   | Удалить заметку       |

### Примеры (curl)

```bash
# Создать
curl -X POST localhost:8080/notes \
  -H 'Content-Type: application/json' \
  -d '{"title":"Заметка 1","body":"Текст заметки"}'

# Список
curl localhost:8080/notes

# Получить по ID
curl localhost:8080/notes/1

# Обновить
curl -X PUT localhost:8080/notes/1 \
  -H 'Content-Type: application/json' \
  -d '{"title":"Обновлённая заметка","body":"Новый текст"}'

# Удалить
curl -X DELETE localhost:8080/notes/1
```