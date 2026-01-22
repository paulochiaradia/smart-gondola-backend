package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Importa o driver do PGX para funcionar com database/sql
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/config"
)

// NewPostgres conecta ao banco e retorna a instância do pool
func NewPostgres(cfg *config.Config) (*sql.DB, error) {
	// Monta a string de conexão (DSN)
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	// Abre a conexão (mas não conecta ainda, apenas valida os argumentos)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// Configurações de Pool (Vital para performance em Cloud)
	// Limita conexões para não estourar o banco (RDS/ElephantSQL)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Tenta um Ping real com timeout curto para garantir que o banco está vivo
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("falha ao conectar no postgres: %w", err)
	}

	return db, nil
}
