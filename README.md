# crawler
Сервис берущий на себе ответственность за сбор данных с внешних агрегаторов событий

### Контракт сообщений RedPanda
``` json
{
  "external_source": "timepad",
  "external_source_id": "1234567",
  "external_url": "https://timepad.ru/event/1234567",
  "title": "Митап по Go-разработке в Москве",
  "description": "Обсуждаем внутренности дженериков, каналы и планировщик GMP...",
  "starts_at": "2026-11-10T18:00:00Z",
  "ends_at": "2026-11-10T21:00:00Z",
  "max_participants": 100,
  "join_policy": "open",
  "location": {
    "name": "Точка Кипения - Арбат",
    "address": "Москва, Новый Арбат, 36",
    "latitude": 55.7539,
    "longitude": 37.5844
  },
  "tags": [
    "golang",
    "backend",
    "meetup"
  ]
}
```