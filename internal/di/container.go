package di

import (
	"database/sql"
	"fmt"

	orgUseCase "github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/usecase"
	orgRepo "github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/infrastructure/repository"
	orgHandler "github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/interface/http/handler"

	userUseCase "github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/application/usecase"
	userRepo "github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/infrastructure/repository"
	userHandler "github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/interface/http/handler"

	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/config"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/database"
)

type Container struct {
	UserHandler *userHandler.UserHandler
	OrgHandler  *orgHandler.OrganizationHandler
	DB          *sql.DB //ex: health check simples)
}

// NewContainer inicializa tudo e retorna:
// 1. O Container preenchido
// 2. Uma função de limpeza (cleanup) para fechar conexões
// 3. Erro (se houver)
func NewContainer(cfg *config.Config) (*Container, func(), error) {
	// 1. O Container assume a responsabilidade de abrir o banco
	db, err := database.NewPostgres(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao iniciar banco no container: %w", err)
	}

	// 2. Define a função de limpeza (Closure)
	cleanup := func() {
		if db != nil {
			db.Close() // Fecha o banco quando chamado
		}
	}

	// --- Módulo Users ---
	uRepo := userRepo.NewUserRepository(db)
	uUseCase := userUseCase.NewUserUseCase(uRepo)
	uHandler := userHandler.NewUserHandler(uUseCase)

	// --- Módulo Organizations ---
	oRepo := orgRepo.NewOrganizationRepository(db)
	oUseCase := orgUseCase.NewOrganizationUseCase(oRepo)
	oHandler := orgHandler.NewOrganizationHandler(oUseCase)

	return &Container{
		UserHandler: uHandler,
		OrgHandler:  oHandler,
		DB:          db,
	}, cleanup, nil
}
