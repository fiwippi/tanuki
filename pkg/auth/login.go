package auth

// LoginRequest defines the request to /auth/login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginReply defines the reply from /auth/login
type LoginReply struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
