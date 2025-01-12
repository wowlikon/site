# Personal Website

Персональный веб-сайт с системой аутентификации, защитой от ботов и интеграцией с GitHub API.

## Технологии

### Бэкенд
- Go 1.20 и выше
- Gin Web Framework
- JWT для аутентификации
- TOML для конфигурации
- bcrypt для хеширования паролей

### Фронтенд
- HTML5
- CSS3
- JavaScript

## Основные функции
- 🔐 Система аутентификации (регистрация/вход)
- 🛡️ Защита от ботов и rate limiting
- 📊 Интеграция с GitHub API
- 📱 Адаптивный дизайн
- 🔄 Автоматическое обновление списков блокировки
- 📜 Поддержка HTTPS

## Установка и запуск

### Предварительные требования
- Go 1.x или выше
- SSL сертификаты (для HTTPS)

### Настройка, установка и запуск

1. Клонируйте репозиторий
```bash
git clone https://github.com/wowlikon/site.git
```

2. Перейдите в директорию проекта
```bash
cd site
```

3. Установите сервис systemd
```bash
sudo cp ./site.service /etc/systemd/system/site.service
```

4. Установите зависимости
```bash
go mod download
```

5. Добавьте файлы ssl для https
```
./ssl/youredomain.com.crt
./ssl/youredomain.com.key
```

6. Настройте config.toml
```bash
cp config.toml.example config.toml
nano config.toml # or use vim
```

7. Настройте `certificates.json` и `repositories.json`
```bash
cp ./data/certificates.json.example ./data/certificates.json
nano ./data/certificates.json # or use vim

cp ./data/repositories.json.example ./data/repositories.json
nano ./data/repositories.json # or use vim
```

8. Скомпилируйте проект
```bash
go build .
```

9. Добавьте сервис в автозагрузку и запустите
```bash
sudo systemctl enable site.service --now
```

## API Endpoints
| Endpoint | Метод | Описание |
|----------|-------|-----------|
| /account/register | POST | Регистрация нового пользователя |
| /account/login | POST | Аутентификация пользователя |
| /account/profile | GET | Получение профиля пользователя (требует JWT) |
| /api/repos/:username/:repo | GET | Получение информации о GitHub репозитории |
| /api/stats | GET | Получение системной статистики |

## Структура проекта
```
site/
├── data/
│   ├── blocked_paths.txt
│   ├── blocked_ua.txt
│   ├── repositories.json
│   └── certificates.json
├── ssl/
│   ├── sitedomain.crt (for HTTPS)
│   └── sitedomain.key (for HTTPS)
├── static/
│   ├── fonts/
│   ├── images/
│   ├── pages/
│   ├── scripts/
│   └── styles/
├── ...
├── main.go
├── config.toml
├── site.service (for systemd)
├── status.sh (script using systemctl)
├── update.sh (script update code from repo, compile and restart service)
└── README.md
```

## Безопасность
- Rate limiting для защиты от DDoS атак
- Блокировка подозрительных User-Agent
- Фильтрация запросов по путям
- Хеширование паролей с помощью bcrypt
- JWT для авторизации

## TODO
- использование базы данных
- улучшение безопасности авторизации и регистрации
  - подтверждение email
  - TOTP 2FA
  - восстановление пароля
- добавление API токена
- админ-панель
- новые функции API
- облачное хранилище и нестройка выделения памяти на пользователя
- загрузка файлов в хранилище
  - HTTP/HTTPS
  - bit-torrent
  - другие источники
- telegram/discord бот

## Автор
[wowlikon](https://github.com/wowlikon)

## Лицензия
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
