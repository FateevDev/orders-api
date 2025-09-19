package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/FateevDev/orders-api/model"
	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	Client *redis.Client
}

func orderIdKey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}

func (r *RedisRepository) Insert(ctx context.Context, order model.Order) error {
	marshal, err := json.Marshal(order)

	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	key := orderIdKey(order.OrderID)
	res := r.Client.SetNX(ctx, key, marshal, 0)

	if err := res.Err(); err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	return nil
}

func ErrNotFound(id uint64) error {
	return fmt.Errorf("order %d not found", id)
}

func (r *RedisRepository) FindById(ctx context.Context, id uint64) (model.Order, error) {
	value, err := r.Client.Get(ctx, orderIdKey(id)).Result()

	if errors.Is(err, redis.Nil) {
		return model.Order{}, ErrNotFound(id)
	} else if err != nil {
		return model.Order{}, fmt.Errorf("failed to get order: %w", err)
	}

	var order model.Order
	err = json.Unmarshal([]byte(value), &order)

	if err != nil {
		return model.Order{}, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return order, nil
}

func (r *RedisRepository) List(ctx context.Context) ([]model.Order, error) {
	value, err := r.Client.GetRange(ctx, "order", 0, 100).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	orderList := make([]model.Order, 0)
	err = json.Unmarshal([]byte(value), &orderList)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal orders: %w", err)
	}

	return orderList, nil
}

func (r *RedisRepository) Delete(ctx context.Context, id uint64) error {
	err := r.Client.Del(ctx, orderIdKey(id)).Err()

	if errors.Is(err, redis.Nil) {
		return ErrNotFound(id)
	} else if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	return nil
}

func (r *RedisRepository) Update(ctx context.Context, id uint64, order model.Order) error {
	set := r.Client.Set(ctx, orderIdKey(id), order, 0)

	if err := set.Err(); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}
