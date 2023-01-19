package db

import "time"

type Option func(*Postgres)

func MaxPoolSize(size int) Option {
	return func(p *Postgres) {
		p.maxPoolSize = size
	}
}

func ConnAttempts(attemps int) Option {
	return func(p *Postgres) {
		p.connAttempts = attemps
	}
}

func ConnTimeOut(timeout time.Duration) Option {
	return func(p *Postgres) {
		p.connTimeout = timeout
	}
}
