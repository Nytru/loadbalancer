package datebase

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Repo struct {
	cnnString string
}

type ClientLimitDb struct {
	ID               string
	Capacity         int
	RefillIntervalMs int64
}

func NewRepo(cnnString string) *Repo {
	return &Repo{
		cnnString: cnnString,
	}
}

func (r *Repo) Get(ctx context.Context, id string) (ClientLimitDb, error) {
	c, err := pgx.Connect(ctx, r.cnnString)
	if err != nil {
		return ClientLimitDb{}, err
	}
	defer c.Close(ctx)

	var cl ClientLimitDb
	err = c.
	QueryRow(ctx, "SELECT id, capacity, refill_interval_milliseconds FROM client_limits WHERE id = $1", id).
	Scan(&cl.ID, &cl.Capacity, &cl.RefillIntervalMs)

	if err != nil {
		if err == pgx.ErrNoRows {
			return ClientLimitDb{}, nil
		}
		return ClientLimitDb{}, err
	}

	return cl, nil
}

func (r *Repo) GetAll(ctx context.Context) ([]ClientLimitDb, error) {
	c, err := pgx.Connect(ctx, r.cnnString)

	if err != nil {
		return nil, err
	}
	defer c.Close(ctx)

	rows, err := c.Query(ctx, "SELECT id, capacity, refill_interval_milliseconds FROM client_limits")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []ClientLimitDb
	for rows.Next() {
		var cl ClientLimitDb
		if err := rows.Scan(&cl.ID, &cl.Capacity, &cl.RefillIntervalMs); err != nil {
			return nil, err
		}
		clients = append(clients, cl)
	}

	return clients, nil
}

func (r *Repo) Add(ctx context.Context, cl ClientLimitDb) error {
	c, err := pgx.Connect(ctx, r.cnnString)
	if err != nil {
		return err
	}
	defer c.Close(ctx)

	_, err = c.Exec(ctx, "INSERT INTO client_limits (id, capacity, refill_interval_milliseconds) VALUES ($1, $2, $3)",
		cl.ID, cl.Capacity, cl.RefillIntervalMs)

	return err
}

func (r *Repo) Remove(ctx context.Context, id string) error {
	c, err := pgx.Connect(ctx, r.cnnString)
	if err != nil {
		return err
	}
	defer c.Close(ctx)

	_, err = c.Exec(ctx, "DELETE FROM client_limits WHERE id = $1", id)

	return err
}

func (r *Repo) Update(ctx context.Context, cl ClientLimitDb) error {
	c, err := pgx.Connect(ctx, r.cnnString)
	if err != nil {
		return err
	}
	defer c.Close(ctx)

	_, err = c.Exec(ctx, "UPDATE client_limits SET capacity = $1, refill_interval_milliseconds = $2 WHERE id = $3",
		cl.Capacity, cl.RefillIntervalMs, cl.ID)

	return err
}
