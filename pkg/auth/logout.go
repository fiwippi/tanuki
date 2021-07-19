package auth

// LogoutReply defines the reply from /auth/logout
type LogoutReply struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
