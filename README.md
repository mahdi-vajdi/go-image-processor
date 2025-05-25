# Go Image Processor

`go-image-processor` is a backend service designed for asynchronous image processing. It allows users to upload images,
which are then queued and processed according to predefined tasks (e.g., resizing, format conversion). The service
manages the lifecycle of these processing tasks and provides status updates.

## Features

* **Asynchronous Image Processing:** Tasks are handled in the background, allowing for non-blocking API responses.
* **Task Management:** Create, track, and manage image processing tasks.
* **Storage Integration:** Likely integrates with cloud storage (e.g., AWS S3) for original and processed images.
* **RESTful API:** Provides endpoints for uploading images and checking task status.

## Technologies Used

* **Go (version 1.24.1 or higher)**
* **Gorilla Mux:** For HTTP routing.
* **SQLx:** For database interactions (likely PostgreSQL, given `lib/pq`).
* **AWS SDK for Go v2:** For interacting with AWS services (e.g., S3).
* **Disintegration Imaging:** For image manipulation.
* **Godotenv:** For managing environment variables.

## Project Structure

```
├── cmd/api/             # Main application entry point
├── internal/            # Internal application logic (handlers, services, repository, models)
│   ├── config           # The configs for the project
│   ├── handler/         # HTTP request handlers
│   ├── middleware       # Middlewares for handlers
│   ├── model/           # Data structures
│   ├── processing/      # The main processing service
│   ├── repository/      # Data access layer
│   ├── router/          # The API routes
│   └── storage/         # The storage implementation with s3 and local
├── migrations/          # Database migration files
├── go.mod               # Go module definition
├── .env.example         # Example environment variables
└── README.md            # This file
```

## Setup & Running

1. **Prerequisites:**
    * Go 1.24.1 or later
    * A running PostgreSQL instance (or configure for your database)
    * AWS S3 bucket and credentials and set the `STORAGE_TYPE` environment variable to `s3` (if S3 storage is used)
      otherwise set the `STORAGE_TYPE` environment variable to
      `local` for local disk storage (development only)

2. **Clone the repository:**
   ```bash
   git clone https://github.com/mahdi-vajdi/go-image-processor.git
   cd go-image-processor
   ```

3. **Set up environment variables:**
   Copy `.env.example` to `.env` and fill in your configuration details (database connection, AWS credentials, etc.).
   ```bash
   cp .env.example .env
   # Edit .env with your values
   ```

4. **Install dependencies:**
   ```bash
   go mod tidy
   ```

5. **Run database migrations:**
   Use a migration tool or run them manually


6. **Run the application:**
   ```bash
   go run cmd/api/main.go
   ```
   The API server should now be running (typically on a port like `8080` or `3000`, check the `HTTP_PORT` in the .env).

## API Endpoints

* `POST /upload`: Upload an image for processing.
* `GET /status/{task_id}`: Get the status of an image processing task.
* `GET /image/{image_key}`: Retrieve a processed image.

## Potential Improvements & Next Steps

This project has several areas for potential enhancement, including but not limited to:

* **Robust Error Handling:** Implement retry logic for transient errors.
* **Dead Letter Queue (DLQ):** For tasks that repeatedly fail.
* **Metrics and Monitoring:** Integrate with systems like Prometheus.
* **Configurable Transformations:** Allow users to specify image processing parameters.
* **Wider Format Support:** Handle more input/output image formats.
* **Task Prioritization:** Implement priority queues for tasks.
* **Dynamic Worker Scaling:** Adjust the number of workers based on load.
* **Enhanced Input Validation:** Stricter checks for uploads.
* **Authentication/Authorization:** Secure API endpoints.

## License

This project is licensed under the [MIT License](https://github.com/mahdi-vajdi/go-image-processor/blob/master/LICENSE).