package api

// AuthLoginRequest defines the request to /api/auth/login
type AuthLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthLoginReply defines the reply from /api/auth/login
type AuthLoginReply struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// AuthLogoutReply defines the reply from /api/auth/logout
type AuthLogoutReply struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
