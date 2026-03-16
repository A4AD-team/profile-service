package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/A4AD-team/profile-service/internal/service"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	exchangePost    = "post"
	exchangeComment = "comment"

	keyPostCreated    = "post.created"
	keyPostDeleted    = "post.deleted"
	keyPostLiked      = "post.liked"
	keyCommentCreated = "comment.created"
	keyCommentDeleted = "comment.deleted"
	keyCommentLiked   = "comment.liked"

	queuePostCreated    = "profile.post.created"
	queuePostDeleted    = "profile.post.deleted"
	queuePostLiked      = "profile.post.liked"
	queueCommentCreated = "profile.comment.created"
	queueCommentDeleted = "profile.comment.deleted"
	queueCommentLiked   = "profile.comment.liked"
)

type binding struct {
	queue    string
	exchange string
	key      string
}

var bindings = []binding{
	// Profile is now created via HTTP from auth-service synchronously
	{queuePostCreated, exchangePost, keyPostCreated},
	{queuePostDeleted, exchangePost, keyPostDeleted},
	{queuePostLiked, exchangePost, keyPostLiked},
	{queueCommentCreated, exchangeComment, keyCommentCreated},
	{queueCommentDeleted, exchangeComment, keyCommentDeleted},
	{queueCommentLiked, exchangeComment, keyCommentLiked},
}

type Consumer struct {
	amqpURL string
	conn    *amqp.Connection
	svc     *service.ProfileService
	logger  *zap.Logger
}

func NewConsumer(amqpURL string, svc *service.ProfileService, logger *zap.Logger) *Consumer {
	return &Consumer{amqpURL: amqpURL, svc: svc, logger: logger}
}

func (c *Consumer) Connect() error {
	conn, err := amqp.Dial(c.amqpURL)
	if err != nil {
		return fmt.Errorf("rabbitmq dial: %w", err)
	}
	c.conn = conn
	return nil
}

func (c *Consumer) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	for _, ex := range []string{exchangePost, exchangeComment} {
		if err := ch.ExchangeDeclare(ex, "topic", true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare exchange %s: %w", ex, err)
		}
	}

	for _, b := range bindings {
		if _, err := ch.QueueDeclare(b.queue, true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare queue %s: %w", b.queue, err)
		}
		if err := ch.QueueBind(b.queue, b.key, b.exchange, false, nil); err != nil {
			return fmt.Errorf("bind queue %s: %w", b.queue, err)
		}
	}

	for _, b := range bindings {
		msgs, err := ch.Consume(b.queue, "", false, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("consume queue %s: %w", b.queue, err)
		}
		go c.consume(ctx, b.key, msgs)
	}

	<-ctx.Done()
	return nil
}

func (c *Consumer) consume(ctx context.Context, key string, msgs <-chan amqp.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			if err := c.dispatch(ctx, key, msg.Body); err != nil {
				c.logger.Error("failed to process message",
					zap.String("routing_key", key),
					zap.Error(err),
				)
				msg.Nack(false, false)
				continue
			}
			msg.Ack(false)
		}
	}
}

func (c *Consumer) dispatch(ctx context.Context, key string, body []byte) error {
	switch key {
	case keyPostCreated:
		return c.handlePostCreated(ctx, body)
	case keyPostDeleted:
		return c.handlePostDeleted(ctx, body)
	case keyPostLiked:
		return c.handlePostLiked(ctx, body)
	case keyCommentCreated:
		return c.handleCommentCreated(ctx, body)
	case keyCommentDeleted:
		return c.handleCommentDeleted(ctx, body)
	case keyCommentLiked:
		return c.handleCommentLiked(ctx, body)
	default:
		c.logger.Warn("unknown routing key", zap.String("key", key))
		return nil
	}
}

func (c *Consumer) handlePostCreated(ctx context.Context, body []byte) error {
	var e PostCreatedEvent
	if err := json.Unmarshal(body, &e); err != nil {
		return fmt.Errorf("unmarshal PostCreatedEvent: %w", err)
	}
	authorID, err := uuid.Parse(e.AuthorID)
	if err != nil {
		return fmt.Errorf("invalid authorId: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.svc.HandlePostCreated(ctx, authorID)
}

func (c *Consumer) handlePostDeleted(ctx context.Context, body []byte) error {
	var e PostDeletedEvent
	if err := json.Unmarshal(body, &e); err != nil {
		return fmt.Errorf("unmarshal PostDeletedEvent: %w", err)
	}
	authorID, err := uuid.Parse(e.AuthorID)
	if err != nil {
		return fmt.Errorf("invalid authorId: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.svc.HandlePostDeleted(ctx, authorID)
}

func (c *Consumer) handlePostLiked(ctx context.Context, body []byte) error {
	var e PostLikedEvent
	if err := json.Unmarshal(body, &e); err != nil {
		return fmt.Errorf("unmarshal PostLikedEvent: %w", err)
	}
	authorID, err := uuid.Parse(e.AuthorID)
	if err != nil {
		return fmt.Errorf("invalid authorId: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.svc.HandlePostLiked(ctx, authorID, e.Delta)
}

func (c *Consumer) handleCommentCreated(ctx context.Context, body []byte) error {
	var e CommentCreatedEvent
	if err := json.Unmarshal(body, &e); err != nil {
		return fmt.Errorf("unmarshal CommentCreatedEvent: %w", err)
	}
	authorID, err := uuid.Parse(e.AuthorID)
	if err != nil {
		return fmt.Errorf("invalid authorId: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.svc.HandleCommentCreated(ctx, authorID)
}

func (c *Consumer) handleCommentDeleted(ctx context.Context, body []byte) error {
	var e CommentDeletedEvent
	if err := json.Unmarshal(body, &e); err != nil {
		return fmt.Errorf("unmarshal CommentDeletedEvent: %w", err)
	}
	authorID, err := uuid.Parse(e.AuthorID)
	if err != nil {
		return fmt.Errorf("invalid authorId: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.svc.HandleCommentDeleted(ctx, authorID)
}

func (c *Consumer) handleCommentLiked(ctx context.Context, body []byte) error {
	var e CommentLikedEvent
	if err := json.Unmarshal(body, &e); err != nil {
		return fmt.Errorf("unmarshal CommentLikedEvent: %w", err)
	}
	authorID, err := uuid.Parse(e.AuthorID)
	if err != nil {
		return fmt.Errorf("invalid authorId: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.svc.HandleCommentLiked(ctx, authorID, e.Delta)
}
