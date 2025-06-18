# Enhanced Gin Chat App

A lightweight, self-contained real-time chat application written in Go.
It combines **Gin** for HTTP routing, **Gorilla WebSocket** for live messaging, **Server-Sent Events (SSE)** for history streaming, and an embedded **SQLite** database for persistence. The single-page front-end (HTML + vanilla JS) provides a modern, responsive UI out of the box.


[![image.png](https://i.postimg.cc/yxXVjpXQ/image.png)](https://postimg.cc/xN8DTRZM)

---

## ✨ Key Features

| Feature                     | Description                                                                           |
| --------------------------- | ------------------------------------------------------------------------------------- |
| Real-time messaging         | Incoming chat messages are pushed instantly over WebSockets.                          |
| Typing-style history replay | Past messages stream word-by-word via SSE to mimic live typing.                       |
| Presence & status           | Users are marked **online** / **away** automatically (heart-beats + last-seen check). |
| Username-only auth          | No passwords—just pick a unique handle and start chatting.                            |
| Clean responsive UI         | Pure HTML/CSS/JS (no framework) with mobile-friendly layout.                          |
| Zero external services      | Uses an on-disk `chat.db`; runs anywhere Go runs.                                     |

---

## 🏗️ Tech Stack

* **Go 1.22+**

  * [Gin](https://github.com/gin-gonic/gin) – HTTP router & middleware&#x20;
  * [Gorilla WebSocket](https://github.com/gorilla/websocket) – WS transport
  * **SQLite** via `modernc.org/sqlite` – zero-config embedded DB&#x20;
* **Vanilla JS** single-page front-end (no build step)&#x20;
* **HTML templates** served from `templates/` (Gin)&#x20;

---

## 🚀 Quick Start

```bash
# 1. Clone & enter the project
git clone https://github.com/<your-org>/<repo>.git
cd <repo>

# 2. Download Go dependencies
go mod tidy

# 3. Run the server (defaults to :8080)
go run .

# 4. Open the app
# Visit http://localhost:8080 in your browser
```

> The first launch auto-creates `chat.db` and the required tables.&#x20;

---

## 📂 Project Structure

```
.
├── assets/            # Static CSS / images (served at /assets)
├── templates/
│   └── index.html     # Single-page UI (loaded at GET /)
├── handlers.go        # HTTP + WS + SSE handlers
├── models.go          # Data models (Message, User, ChatApp)
├── utils.go           # Helpers: DB setup, broadcasting, heartbeats
├── main.go            # Entry point & route wiring
└── go.mod
```

---

## 🌐 HTTP & WebSocket Endpoints

| Method | Path           | Purpose                                                         |
| ------ | -------------- | --------------------------------------------------------------- |
| `GET`  | `/`            | Serve `index.html` template                                     |
| `POST` | `/create-user` | Create a new username (JSON `{username}`)                       |
| `POST` | `/join`        | Mark user online & pick recipient (JSON `{username,recipient}`) |
| `GET`  | `/users`       | List users + status (`?current=<me>`)                           |
| `GET`  | `/events`      | **SSE** stream of past messages for a user                      |
| `GET`  | `/ws`          | **WebSocket** endpoint for live chat                            |

---

## 🗄️ Database Schema (SQLite)

```sql
CREATE TABLE messages (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  sender    TEXT NOT NULL,
  receiver  TEXT NOT NULL,
  content   TEXT NOT NULL,
  created_at DATETIME NOT NULL
);

CREATE TABLE users (
  username  TEXT PRIMARY KEY,
  status    TEXT NOT NULL DEFAULT 'away',
  last_seen DATETIME NOT NULL,
  joined_at DATETIME NOT NULL
);
```



---

## 🙋‍♀️ Contributing

1. **Fork** the project & create your branch:
   `git checkout -b feature/my-new-feature`
2. **Commit** your changes:
   `git commit -am 'Add some feature'`
3. **Push** to the branch:
   `git push origin feature/my-new-feature`
4. **Open a Pull Request**

Please run `go vet ./...` and `go test ./...` (if you add tests) before submitting.

---

## 📜 License

Distributed under the MIT License. See `LICENSE` for more information.

---

## 📅 Roadmap / Ideas

* 💬 Group chats & rooms
* 🔒 Upgrade to token-based auth
* 🖼️ File & image sharing over WebSocket
* 📱 PWA / mobile push notifications

Feel free to open issues or PRs for any of the above!
