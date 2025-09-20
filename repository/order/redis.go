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

func ErrOrderWithIdNotFound(id uint64) error {
	return fmt.Errorf("order %d not found", id)
}
func ErrOrdersNotFound() error {
	return errors.New("orders not found")
}

const SetKey = "orders"

func (r *RedisRepository) Insert(ctx context.Context, order model.Order) error {
	marshal, err := json.Marshal(order)

	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	key := orderIdKey(order.OrderID)

	txn := r.Client.TxPipeline()
	defer txn.Discard()

	setNxCmd := txn.SetNX(ctx, key, marshal, 0)
	txn.SAdd(ctx, SetKey, key)

	_, err = txn.Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	if setNxCmd.Err() != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	return nil
}

func (r *RedisRepository) FindById(ctx context.Context, id uint64) (model.Order, error) {
	value, err := r.Client.Get(ctx, orderIdKey(id)).Result()

	if errors.Is(err, redis.Nil) {
		return model.Order{}, ErrOrderWithIdNotFound(id)
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

type FindAllPage struct {
	Size   uint64
	Offest uint64
}

type FindAllResult struct {
	Orders []model.Order
	Cursor uint64
}

func (r *RedisRepository) FindAll(ctx context.Context, p FindAllPage) (FindAllResult, error) {
	keys, cursor, err := r.Client.SScan(ctx, SetKey, p.Offest, "*", int64(p.Size)).Result()

	if errors.Is(err, redis.Nil) {
		return FindAllResult{}, ErrOrdersNotFound()
	} else if err != nil {
		return FindAllResult{}, fmt.Errorf("failed to get orders: %w", err)
	}

	if len(keys) == 0 {
		return FindAllResult{}, nil
	}

	result, err := r.Client.MGet(ctx, keys...).Result()

	if err != nil {
		return FindAllResult{}, fmt.Errorf("failed to get orders: %w", err)
	}

	orderList := make([]model.Order, len(result))

	for i, v := range result {
		v := v.(string)
		var order model.Order
		err = json.Unmarshal([]byte(v), &order)

		if err != nil {
			return FindAllResult{}, fmt.Errorf("failed to unmarshal order: %w", err)
		}

		orderList[i] = order
	}

	return FindAllResult{orderList, cursor}, nil
}

func (r *RedisRepository) Delete(ctx context.Context, id uint64) error {
	key := orderIdKey(id)

	txn := r.Client.TxPipeline()
	defer txn.Discard()

	delCmd := txn.Del(ctx, key)
	txn.SRem(ctx, SetKey, key)

	_, err := txn.Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	result, err := delCmd.Result()

	if result == 0 {
		return ErrOrderWithIdNotFound(id)
	} else if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	return nil
}

func (r *RedisRepository) Update(ctx context.Context, id uint64, order model.Order) error {
	marshal, err := json.Marshal(order)

	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	err = r.Client.SetXX(ctx, orderIdKey(id), marshal, 0).Err()

	if errors.Is(err, redis.Nil) {
		return ErrOrderWithIdNotFound(id)
	} else if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}
