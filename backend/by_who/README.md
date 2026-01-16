# ByWho API

REST API for managing "by_who" (who paid/received) entities in Micro-Bingo.

---

## Endpoints

### POST `/bywho/add`

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
  "byWhoId": "string"
}
```

---

### GET `/bywho/list?userId=string`

**Response**
```json
[
  {
    "byWhoId": "string",
    "name": "string",
    "description": "string"
  }
  // ...
]
```

---

### GET `/bywho/:byWhoId?userId=string`

**Response**
```json
{
  "byWhoId": "string",
  "name": "string",
  "description": "string"
}
```

---

### PUT `/bywho/:byWhoId?userId=string`

**Request Body**
```json
{
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

### DELETE `/bywho/:byWhoId?userId=string`

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
- Make sure your database is running and the `by_who` table exists:

```sql
CREATE TABLE by_who (
  by_who_id VARCHAR(32) PRIMARY KEY,
  user_id VARCHAR(64) NOT NULL,
  name VARCHAR(64) NOT NULL,
  description VARCHAR(255)
);
```
