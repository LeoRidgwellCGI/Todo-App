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

## Installation

Clone and build:
```bash
git clone https://github.com/leoridgwellcgi/todo-app.git
cd todo-app
go build -o todo-app
```
or run directly:
```bash
go run .
```

---

## Testing

Run:
```bash
go test ./
```

---

## CLI Usage

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

## API Usage

### Routes
| Routes                         | Description                                                         |
| ------------------------------ | ------------------------------------------------------------------- |
| `coming soon`                  | Coming soon                                                         |

---

## CLI Examples

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

## API Examples

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

