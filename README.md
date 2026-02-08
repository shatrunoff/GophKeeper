# GopherPass

Безопасный менеджер паролей и секретов с серверной синхронизацией.

## Схема БД

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    login TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE secrets (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    encrypted_payload BYTEA NOT NULL,
    meta JSONB,
    version BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## API

### Регистрация
```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"login": "user", "password": "pass123"}'
```

### Логин
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"login": "user", "password": "pass123"}'
# Response: {"token": "eyJ..."}
```

### Создание секрета
```bash
curl -X POST http://localhost:8080/api/secrets \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"type": "credentials", "payload": {"user": "admin", "pass": "secret"}, "version": 0}'
```

### Получение секретов
```bash
curl http://localhost:8080/api/secrets \
  -H "Authorization: Bearer <token>"
```

### Удаление секрета
```bash
curl -X DELETE http://localhost:8080/api/secrets/<id> \
  -H "Authorization: Bearer <token>"
```

## Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `DB_DSN` | PostgreSQL connection string | `postgres://localhost:5432/gopherpass` |
| `JWT_SECRET` | Секрет для подписи JWT | `secret` |
| `DATA_ENCRYPTION_KEY` | Ключ AES-256 (32 байта) | `12345678901234567890123456789012` |
| `SERVER_PORT` | Порт HTTP сервера | `8080` |


## Типы секретов

- `credentials` — логин/пароль
- `text` — произвольный текст
- `binary` — бинарные данные
- `card` — банковская карта