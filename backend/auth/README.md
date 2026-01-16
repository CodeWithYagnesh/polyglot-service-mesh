# Auth API

REST API for user authentication and profile management in Micro-Bingo.

---

## Endpoints

### POST `/user/register`

**Request Body**
```json
{
  "email": "string",
  "username": "string",
  "password": "string"
}
```

**Response**
```json
{
  "userId": "string",
  "message": "User registered successfully."
}
```

---

### POST `/user/login`

**Request Body**
```json
{
  "email": "string",
  "password": "string"
}
```

**Response**
```json
{
  "userId": "string"
}
```

---

### GET `/user/get?userId=string`

**Response**
```json
{
  "userId": "string",
  "email": "string",
  "username": "string"
}
```

---

### PUT `/user/update`

**Request Body**
```json
{
  "userId": "string",
  "email": "string",     // optional
  "username": "string"   // optional
}
```

**Response**
```json
{
  "success": true,
  "message": "User updated successfully."
}
```

---

## Notes

- All endpoints use JSON for requests and responses.
- Make sure your database is running and the `users` table exists:

```sql
CREATE TABLE users (
  id VARCHAR(32) PRIMARY KEY,
  username VARCHAR(64) NOT NULL,
  email VARCHAR(128) NOT NULL UNIQUE,
  password VARCHAR(128) NOT NULL
);
```
