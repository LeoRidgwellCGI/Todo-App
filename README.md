# Todo-App 📝

A simple CLI & API To-Do List Manager built in Go.  
Created by **Leo Ridgwell**.

---

## Overview

Todo-App lets you manage to-do items directly from your terminal.  
Each task includes:
- A **description**
- A **status** (`not started`, `started`, or `completed`)
- A **created_at** timestamp  

All data is stored as JSON under the automatically created `./out/` directory.

---

## Requirements

- Go **1.25.1** or newer

---

## Installation (All Modes)

First, clone the Repo and CD into the directory:
```bash
git clone https://github.com/leoridgwellcgi/todo-app.git
cd todo-app
```

---

## Running / Building (CLI Mode)

To run directly in CLI mode:
```bash
go run ./cmd/cli
```

To build in CLI mode:
```bash
go build -o bin/todo ./cmd/cli
```

---

## Running / Building (API Mode)

To directly in API mode:
```bash
go run ./cmd/api
```

To build in API mode:
```bash
go build -o bin/todo ./cmd/api
```

---

### Running / Building from Root
Running / Building from Root (aka main.go) is no longer possible.  
It will now display a warning and instructions on how to run / build in CLI or API mode.

---

## Testing (All Modes)

Run:
```bash
go test ./
```

---

## Usage (CLI Mode)

### Commands
| Command                          | Description                                                       |
| -------------------------------- | ----------------------------------------------------------------- |
| `-list`                          | List all to-do items                                              |
| `-add "<description>"`           | Add a new item                                                    |
| `-status <state>`                | Set status when adding (`not started`, `started`, or `completed`) |
| `-update <id> -newdesc "<desc>"` | Update a task description                                         |
| `-delete <id>`                   | Delete a task by ID                                               |
| `-out <path>`                    | Use a custom file path (stored under `./out/`)                    |

### Global flags
| Flag             | Description                              |
| ---------------- | ---------------------------------------- |
| `-logtext`       | Use readable text logs instead of JSON   |
| `-traceid <id>`  | Provide a custom trace ID                |
| `--traceid=<id>` | Alternate syntax for specifying trace ID |

---

## Usage (API Mode)

### Routes
| Routes                         | Description                                                                               |
| ------------------------------ | ----------------------------------------------------------------------------------------- |
| `get`                          | Get information about all tasks or a single task (See examples below)                     |
| `add`                          | POST a header and description and add a new task (See examples below)                     |
| `update`                       | POST a header and description and update an existing task (See examples below)            |
| `delete`                       | POST a header and description and delete an existing task (See examples below)            |

### Static Pages
| Pages                          | Description                                                                               |
| ------------------------------ | ----------------------------------------------------------------------------------------- |
| `list`                         | List all tasks on a static page (See examples below)                                      |
| `about`                        | View information about the application on a static page (See examples below)              |

---

## Examples (CLI Mode)

List tasks:
```bash
go run ./cmd/cli -list
```

Add a new task:
```bash
go run ./cmd/cli -add "Write documentation"
```

Set the status when adding a new task:
```bash
go run ./cmd/cli -add "Write docs" -status "started"
```

Update a task:
```bash
go run ./cmd/cli -update 1 -newdesc "Write README file"
```

Delete a task:
```bash
go run ./cmd/cli -delete 1
```

Use a custom out path:
```bash
go run ./cmd/cli -add "Change out path" -out "test/todos2.json"
```

Use text logs and a custom trace ID:
```bash
go run ./cmd/cli -logtext --traceid=my-trace-id -add "Try text logs"
```

---

## Examples (API Mode)

Get all tasks:
```bash
curl http://localhost:8080/get
```

Get a single task:
```bash
curl http://localhost:8080/get?id=1
```

Add a new task:
```bash
curl -X POST http://localhost:8080/add \
  -H 'Content-Type: application/json' \
  -d '{"description":"Buy milk"}'
```

Set the status when adding a new task:
```bash
curl -X POST http://localhost:8080/add \
  -H 'Content-Type: application/json' \
  -d '{"description":"Buy milk","status":"started"}'
```

Update a task:
```bash
curl -X POST http://localhost:8080/update \
  -H 'Content-Type: application/json' \
  -d '{"id":1, "description":"Buy milk and eggs"}'
```

Update a task and status:
```bash
curl -X POST http://localhost:8080/update \
  -H 'Content-Type: application/json' \
  -d '{"id":1, "description":"Buy milk and eggs", "status":"started"}'
```

Delete a task:
```bash
curl -X POST http://localhost:8080/delete \
  -H 'Content-Type: application/json' \
  -d '{"id":1}'
```

List all tasks (static page):
```bash
curl http://localhost:8080/list
```

Open about page (static page):
```bash
curl http://localhost:8080/about
```

---

## Author

Leo Ridgwell - Junior Software Engineer @ CGI

