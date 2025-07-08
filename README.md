Made by Hubert Grzesiak

# Web Crawler & Page Analyzer

This is a full-stack web application built for a test task. The application accepts a website URL, crawls it, and displays key information about the page, such as HTML version, page title, heading counts, and link analysis.

## Tech Stack

- **Frontend:**

  - React (with Vite)
  - TypeScript
  - TanStack Table (for data grids)
  - TanStack Query (for server state management)
  - Tailwind CSS
  - shadcn/ui (for UI components)
  - Vitest & Testing Library (for testing)

- **Backend:**

  - Go (Golang)
  - Gin (Web Framework)
  - MySQL (Database)

- **DevOps:**
  - Docker & Docker Compose

---

## Prerequisites

Before you begin, ensure you have the following installed on your system:

- [Docker](https://www.docker.com/products/docker-desktop/)
- [Docker Compose](https://docs.docker.com/compose/install/) (usually included with Docker Desktop)

---

## How to Run

The entire application stack (Frontend, Backend, Database) is orchestrated with Docker Compose, ensuring a reproducible and easy-to-manage environment.

### 1. Clone the Repository

First, clone the project repository to your local machine.

```bash
git clone git@github.com:hubert-grzesiak/challange-sykell.git
cd challange-sykell


2. Build and Run with Docker Compose
This is the recommended method to run the entire application.
In the root directory of the project (where the docker-compose.yml file is located), run the following command:
docker-compose up --build


--build: This flag forces Docker to rebuild the images for the frontend and backend, ensuring all the latest code changes are included.
After the build process is complete, the services will be available at:
Frontend Application: http://localhost:5173
Backend API: http://localhost:8080
How to Use the Application
Open your web browser and navigate to http://localhost:5173.
Enter an API Key: The application's backend is protected by a simple authorization mechanism. You must enter any non-empty string into the "Enter API Key" field in the top-right corner to enable the application's functionality. For example: test1234.
Submit a URL: Enter a full website URL (e.g., https://example.com) into the main input field and click "Analyze".
View Results: The application will start polling the backend for results, which will appear in the table as they become available.
How to Run Tests
The frontend project is configured with Vitest for unit and integration testing.
To run the tests, execute the following command in the frontend directory:
# Go to frontend folder
cd frontend

# Run tests
bun test

Project Structure
The project is organized into two main parts within a monorepo structure:
/
|-- /backend         # Go application (API and Crawler)
|   |-- main.go      # Main server logic
|   |-- schema.sql   # Database schema
|   |-- Dockerfile
|
|-- /frontend        # React application (UI)
|   |-- /src
|   |   |-- /components
|   |   |-- App.tsx    # Main application component
|   |-- vite.config.ts # Vite configuration
|   |-- Dockerfile
|
|-- docker-compose.yml # Orchestrates all services
|-- README.md          # This file


```
