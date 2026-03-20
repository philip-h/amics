# AMICS (Auto Marker for ICS)

A lightweight auto-grading platform for high school computer science courses. Students submit Python code and receive immediate feedback from pytest-based test suites.

## Features

- **Immediate Feedback** – Students see pytest results instantly with highlighted key information
- **Secure Execution** – Student code runs in isolated Docker containers
- **Unlimited Attempts** – Encourages mastery-based learning through iteration
- **Simple Stack** – Built with Go, PostgreSQL, and vanilla JavaScript

## Architecture

AMICS uses a database-as-queue pattern to process student submissions asynchronously:

1. Student submits code via web interface
2. Submission saved to PostgreSQL with `grading` status
3. Background worker polls database for grading submissions
4. Worker spawns isolated Docker container to run pytest
5. Results parsed and saved, student sees formatted output

## Tech Stack

- **Backend**: Go (standard library + PostgreSQL driver)
- **Database**: PostgreSQL
- **Frontend**: HTML templates, CSS, vanilla JavaScript
- **Testing**: Docker + Python 3.11 + pytest

## Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Docker

## Quick Start
```bash
# Clone repository
git clone https://github.com/philip-h/amics.git
cd amics

# Install dependencies
go mod tidy

# Start server (includes background worker)
go run cmd/amics/main.go
```

Visit `http://localhost:8000` to access the platform.

## Status

**Beta** – Currently in use for ICS courses. Not intended for public deployment.

---

**Notes:**
- This is a personal project for educational use
