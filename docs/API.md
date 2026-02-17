# SkillSync API Documentation

Base URL: `http://localhost:8080/api/v1`

## Authentication

### POST /auth/register
Register a new user.

**Body:**
```json
{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "securepassword",
  "full_name": "John Doe",
  "skills_teach": ["Go", "React"],
  "skills_learn": ["Python", "ML"]
}
```

### POST /auth/login
Authenticate and receive a JWT token.

**Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

### POST /auth/refresh
Refresh an existing JWT token. Requires `Authorization: Bearer <token>` header.

---

## Users (Protected)

### GET /users
List users. Query params: `skill`, `level`.

### GET /users/:id
Get user by ID.

### PUT /users/me
Update current user's profile.

### GET /users/me/reputation
Get current user's reputation breakdown.

---

## Matches (Protected)

### POST /matches
Create a match request.

**Body:**
```json
{
  "target_user_id": "uuid",
  "skill_offered": "Go",
  "skill_wanted": "Python"
}
```

### GET /matches
List current user's matches.

### GET /matches/:id
Get match details.

### PUT /matches/:id/status
Update match status (`accepted`, `rejected`, `completed`).

---

## Assessment (Protected)

### POST /assessment
Submit skill assessment answers for Claude AI evaluation.

**Body:**
```json
{
  "skill": "Go",
  "answers": ["answer1", "answer2", "answer3"]
}
```

---

## Ratings (Protected)

### POST /ratings
Submit a rating for a completed match.

**Body:**
```json
{
  "match_id": "uuid",
  "rated_user_id": "uuid",
  "score": 5,
  "communication": 4,
  "knowledge": 5,
  "helpfulness": 5,
  "comment": "Great session!"
}
```

### GET /leaderboard
Get the top users by reputation.

---

## Insights (Protected)

### GET /insights/pairing/:matchId
Get Claude-generated pairing insights for a match.

---

## WebSocket

### GET /ws?token=<jwt>
Connect to the WebSocket server for real-time messaging.

**Message types:**
- `join_room` — Join a chat room (match ID)
- `leave_room` — Leave a chat room
- `message` — Send a text message
