# Orders API

This is a simple RESTful API for managing orders, built with Go. It uses Redis as its data store.

## Features

*   Create, Read, Update, and Delete (CRUD) operations for orders.
*   RESTful API endpoints.
*   Uses Redis for data persistence.
*   Dockerized Redis instance for easy setup.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

*   [Go](https://golang.org/doc/install)
*   [Docker](https://docs.docker.com/get-docker/)
*   [Make](https://www.gnu.org/software/make/)

### Installing

1.  **Clone the repository:**

    ```sh
    git clone https://github.com/FateevDev/orders-api.git
    cd orders-api
    ```

2.  **Start the Redis container:**

    This command will start a Redis container in the background.

    ```sh
    make up
    ```

3.  **Run the application:**

    This command will start the API server.

    ```sh
    go run main.go
    ```

The API will be running at `http://localhost:3000`.

## API Endpoints

The following endpoints are available:

| Method   | Endpoint            | Description          |
| -------- | ------------------- | -------------------- |
| `GET`    | `/api/orders`       | Get all orders       |
| `POST`   | `/api/orders`       | Create a new order   |
| `GET`    | `/api/orders/{id}`  | Get an order by ID   |
| `PUT`    | `/api/orders/{id}`  | Update an order by ID|
| `DELETE` | `/api/orders/{id}`  | Delete an order by ID|

### Example Usage

You can use a tool like `curl` to interact with the API.

**Create an order:**

```sh
curl -X POST -H "Content-Type: application/json" -d '{"line_items":[{"item":"Laptop","quantity":1,"price":1200}]}' http://localhost:3000/api/orders
```

**Get all orders:**

```sh
curl http://localhost:3000/api/orders
```

## Makefile Commands

*   `make up`: Start the Redis container in detached mode.
*   `make down`: Stop and remove the Redis container.
*   `make redis-sh`: Access the shell within the Redis container.

# TODOs

* write tests to endpoints
* create openapi specs for api
