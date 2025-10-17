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

---

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
| Routes                         | Description                                                         |
| ------------------------------ | ------------------------------------------------------------------- |
| `coming soon`                  | Coming soon                                                         |

---

## Examples (CLI Mode)

Add a new task:
```bash
go run . -add "Write documentation"
```

Update a task:
```bash
go run . -update 1 -newdesc "Write README file"
```

Delete a task:
```bash
go run . -delete 1
```

List tasks:
```bash
go run . -list
```

Use text logs and a custom trace ID:
```bash
go run . -logtext --traceid=my-trace-id -add "Try text logs"
```

---

## Examples (API Mode)

Add a new task:
```bash
coming soon
```

Update a task:
```bash
coming soon
```

Delete a task:
```bash
coming soon
```

List tasks:
```bash
coming soon
```

---

## Author

Leo Ridgwell - Junior Software Engineer @ CGI

