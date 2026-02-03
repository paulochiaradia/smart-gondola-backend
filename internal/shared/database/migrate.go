package database

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Driver para Postgres
	_ "github.com/golang-migrate/migrate/v4/source/file"       // Driver para ler de arquivos
)

// RunMigrations executa as migrações pendentes
func RunMigrations(dbURL string) error {
	// 1. Cria a instância do Migrator apontando para a pasta local "migrations"
	m, err := migrate.New(
		"file://migrations",
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("erro ao criar instância de migração: %w", err)
	}

	// 2. Executa o Up (aplica tudo que falta)
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Nenhuma migração nova para aplicar.")
			return nil // Não é um erro, apenas não tinha nada novo
		}
		return fmt.Errorf("erro ao aplicar migrações: %w", err)
	}

	log.Println("Migrações aplicadas com sucesso!")
	return nil
}
