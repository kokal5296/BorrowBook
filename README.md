# BorrowBook

## Prerequisites

- Go 1.22.2 or later
- PostgreSQL 13 or later

## Installation

1. Clone the repository:

```sh
git clone https://github.com/kokal5296/BorrowBook.git
cd BorrowBook
```

2. Install dependencies:

```sh
go mod tidy
```

## Configuration

Modify a `.env` file in the root directory of the project and add the following environment variables:

```sh
POSTGRESQL_URI="postgres://<username>:<password>@localhost:<port>/"
POSTGRESQL_DB_NAME="<database_name>"
PORT=":3000"
```

Replace `<username>`, `<password>`, `<port>`, and `<database_name>` with your PostgreSQL credentials and database details.

## Running the Application

Start the server using `main.go`:

```sh
go run main.go
```

## Making Requests

### Create User

**Endpoint:** `POST /user`

**Example JSON Payload:**

```json
{
  "first_name": "Tine",
  "last_name": "Kokalj"
}
```

### Get User

**Endpoint:** `GET /user/:id`

**Example JSON Payload:**

```json
{}
```

### Get All Users

**Endpoint:** `GET /users`

**Example JSON Payload:**

```json
{}
```

### Update User

**Endpoint:** `PUT /user/:id`

**Example JSON Payload:**

```json
{
  "first_name": "Janez",
  "last_name": "Novak"
}
```

### Delete User

**Endpoint:** `DELETE /user/:id`

**Example JSON Payload:**

```json
{}
```

### Create Book

**Endpoint:** `POST /book`

**Example JSON Payload:**

```json
{
  "title": "The Lord Of The Rings: Fellowship of the Ring",
  "quantity": 5
}
```

### Get Book

**Endpoint:** `GET /book/:id`

**Example JSON Payload:**

```json
{}
```

### Get All Books

**Endpoint:** `GET /books`

**Example JSON Payload:**

```json
{}
```

### Update Book: Changing only quantity

**Endpoint:** `PUT /book/:id`

**Example JSON Payload:**

```json
{
  "title": "The Lord Of The Rings: Fellowship of the Ring",
  "quantity": 10
}
```

### Update Book 

**Endpoint:** `PUT /book/:id`

**Example JSON Payload:**

```json
{
  "title": "The Lord Of The Rings: The two Towers",
  "quantity": 4
}
```

### Delete Book

**Endpoint:** `DELETE /book/:id`

**Example JSON Payload:**

```json
{}
```

### Get Available Books

**Endpoint:** `GET /book_borrow`

**Example JSON Payload:**

```json
{}
```

### Get All Borrowed Books

**Endpoint:** `GET /book_borrowed`

**Example JSON Payload:**

```json
{}
```

### Borrow Book

**Endpoint:** `POST /book_borrow`

**Example JSON Payload:**

```json
{
  "book_id": 1,
  "user_id": 1
}
```

### Return Book

**Endpoint:** `PUT /book_borrow`

**Example JSON Payload:**

```json
{
  "book_id": 1,
  "user_id": 1
}
```
