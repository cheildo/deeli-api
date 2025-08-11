# Deeli API - Article Saver and Recommendation Service

Deeli API is a backend service built with Go that allows users to save articles, fetches their metadata, and provides personalized recommendations based on user ratings. It features a robust architecture with a background worker for asynchronous tasks and a clean, testable codebase.

## Features

-   **User Authentication**: Secure user registration and login using JWT (JSON Web Tokens).
-   **Article Curation**: Save articles via URL. The service automatically fetches the article's `title`, `description`, and `image` in the background.
-   **Metadata Fetching with Retries**: A background worker asynchronously scrapes article metadata. If a scrape fails, it automatically retries up to 3 times with a 5-minute interval.
-   **Article Rating**: Users can rate their saved articles on a scale of 1-5.
-   **Personalized Recommendations**: A `GET /recommendations` endpoint provides article suggestions based on a collaborative filtering algorithm that analyzes the ratings of similar users.
-   **Paginated Lists**: The user's article list (`GET /articles`) is paginated for efficient data retrieval.
-   **Fully Tested**: Includes an integration test suite that runs against a separate, containerized test database.

## System Design and Architecture

This project is built using a modern, layered architecture in Go to ensure a clean separation of concerns, maintainability, and testability.

-   **Layered Architecture**: The core logic is split into distinct layers:
    -   **Handler Layer**: Responsible for handling HTTP requests and responses (using the Gin framework). It validates input and orchestrates calls to the business logic layer.
    -   **Repository Layer**: Responsible for all database interactions. This is the only layer that directly communicates with the PostgreSQL database (via the GORM ORM). It abstracts away the database querying logic.
    -   **Service Layer (`recommendation` package)**: Encapsulates complex business logic that may involve multiple data sources or operations, such as the recommendation algorithm.
-   **Asynchronous Worker**: For long-running tasks like scraping website metadata, the API immediately responds to the user and delegates the task to a background worker. This is implemented using a simple Goroutine and a Ticker for periodic retries, preventing API timeouts and improving user experience.
-   **Configuration Management**: The application uses a `.env` file for managing environment variables. This allows for different configurations for development, testing, and production without changing the code.
-   **Database and Migrations**: The project uses PostgreSQL as its database, managed via Docker Compose for easy setup. GORM's `AutoMigrate` feature is used for simplicity, though for a production environment, a versioned migration tool like `golang-migrate` would be recommended.
-   **Testing Strategy**: The project includes a suite of integration tests. The tests run against a real, separate test database (also managed by Docker Compose) to ensure all layers of the application work together correctly.

## Tech Stack

-   **Language**: Go
-   **Framework**: Gin
-   **Database**: PostgreSQL
-   **ORM**: GORM
-   **Authentication**: JWT (JSON Web Tokens)
-   **Containerization**: Docker & Docker Compose
-   **Testing**: Go's native testing package + `testify`

## API Endpoints

A full list of endpoints and their usage will be available via Swagger documentation once the service is running.

-   `POST /signup` - Register a new user.
-   `POST /login` - Log in and receive a JWT.
-   `GET /me` - Get the current user's information.
-   `POST /articles` - Save a new article by URL.
-   `GET /articles` - Get a paginated list of the user's saved articles.
-   `DELETE /articles/:id` - Delete a saved article.
-   `POST /articles/:id/rate` - Add or update a rating for an article.
-   `GET /articles/:id/rate` - Get the user's rating for an article.
-   `DELETE /articles/:id/rate` - Remove a rating.
-   `GET /recommendations` - Get personalized article recommendations.

---

## Setup and Installation

### Prerequisites

-   [Go](https://golang.org/doc/install) (version 1.18 or higher)
-   [Docker](https://www.docker.com/products/docker-desktop) and Docker Compose

### Running the Application

1.  **Clone the Repository**
    ```sh
    git clone https://github.com/cheildo/deeli-api.git
    cd deeli-api
    ```

2.  **Start the Databases**
    This command will start two PostgreSQL containers: one for development (`article_db`) and one for running automated tests (`article_db_test`).
    ```sh
    docker compose up -d
    ```

3.  **Configure Environment Variables**
    Copy the example environment file and edit it if necessary (the defaults should work with the Docker setup).
    ```sh
    cp .env.example .env
    ```

4.  **Install Go Dependencies**
    ```sh
    go mod tidy
    ```

5.  **Run the API Server**
    ```sh
    go run cmd/api/main.go
    ```
    The server will start on `http://localhost:8080`. You should see log messages indicating a successful database connection and the worker starting.

### Running the Tests

To run the entire suite of integration tests, make sure the Docker containers are running and then execute the following command from the project root:

```sh
# This command will automatically use the .env.test configuration
go test ./...