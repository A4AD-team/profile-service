package messaging

// UserRegisteredEvent is published by auth-service after successful registration.
//
// TODO(auth-service): publish this event to exchange="user", routing key="user.registered"
// inside AuthService.register() after userRepository.save(user).
// Also requires adding "username" field to SignUpRequest and User entity.
//
// Expected payload:
//
//	{
//	    "userId":   "550e8400-e29b-41d4-a716-446655440000",
//	    "username": "johndoe",
//	    "email":    "john@example.com"
//	}
type UserRegisteredEvent struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// PostCreatedEvent is published by post-service when a new post is created.
// TODO(post-service): publish to exchange="post", routing key="post.created"
type PostCreatedEvent struct {
	PostID   string `json:"postId"`
	AuthorID string `json:"authorId"`
}

// PostDeletedEvent is published by post-service when a post is deleted.
// TODO(post-service): publish to exchange="post", routing key="post.deleted"
type PostDeletedEvent struct {
	PostID   string `json:"postId"`
	AuthorID string `json:"authorId"`
}

// PostLikedEvent is published by post-service on like/unlike action.
// Delta: +1 for like, -1 for unlike.
// TODO(post-service): publish to exchange="post", routing key="post.liked"
type PostLikedEvent struct {
	PostID   string `json:"postId"`
	AuthorID string `json:"authorId"`
	Delta    int    `json:"delta"`
}

// CommentCreatedEvent is published by comment-service when a comment is created.
// TODO(comment-service): publish to exchange="comment", routing key="comment.created"
type CommentCreatedEvent struct {
	CommentID string `json:"commentId"`
	AuthorID  string `json:"authorId"`
}

// CommentDeletedEvent is published by comment-service when a comment is deleted.
// TODO(comment-service): publish to exchange="comment", routing key="comment.deleted"
type CommentDeletedEvent struct {
	CommentID string `json:"commentId"`
	AuthorID  string `json:"authorId"`
}

// CommentLikedEvent is published by comment-service on like/unlike action.
// Delta: +1 for like, -1 for unlike.
// TODO(comment-service): publish to exchange="comment", routing key="comment.liked"
type CommentLikedEvent struct {
	CommentID string `json:"commentId"`
	AuthorID  string `json:"authorId"`
	Delta     int    `json:"delta"`
}
