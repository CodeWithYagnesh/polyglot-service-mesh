# Category API

REST API for managing categories in Micro-Bingo.

---

## Endpoints

### POST `/category/add`

**Request Body**
```json
{
  "userId": "string",
  "name": "string",
  "description": "string"
}
```

**Response**
```json
{
  "categoryId": "string"
}
```

---

### GET `/category/list?userId=string`

**Response**
```json
[
  {
    "categoryId": "string",
    "name": "string",
    "description": "string"
  }
  // ...
]
```

---

### GET `/category/:categoryId?userId=string`

**Response**
```json
{
  "categoryId": "string",
  "name": "string",
  "description": "string"
}
```

---

### PUT `/category/:categoryId`

**Request Body**
```json
{
  "userId": "string",
  "name": "string",         // optional
  "description": "string"   // optional
}
```

**Response**
```json
{
  "success": true
}
```

---

### DELETE `/category/:categoryId?userId=string`

**Response**
```json
{
  "success": true
}
```

---

## Notes

- All endpoints require `userId` for authentication/authorization.
- The `description` field is optional.
- Make sure your database is running and the `categories` table exists:

```sql
CREATE TABLE categories (
  category_id VARCHAR(32) PRIMARY KEY,
  user_id VARCHAR(64) NOT NULL,
  name VARCHAR(64) NOT NULL,
  description VARCHAR(255)
);
```
