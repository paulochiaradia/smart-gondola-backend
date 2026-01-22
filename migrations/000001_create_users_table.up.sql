-- Habilita geração de UUID se ainda não estiver habilitada
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    store_id UUID, -- Pode ser NULL (Tenant Owner ou Admin)
    
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(50),
    avatar_url TEXT,
    
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL, -- admin, tenant, manager, operator
    status VARCHAR(50) NOT NULL, -- active, pending, suspended
    
    -- Metadados de Convite
    invited_by UUID,
    
    -- Preferências
    timezone VARCHAR(50) DEFAULT 'UTC',
    language VARCHAR(10) DEFAULT 'pt-BR',
    
    -- Segurança e Compliance
    email_verified_at TIMESTAMP,
    terms_accepted_at TIMESTAMP,
    password_changed_at TIMESTAMP,
    
    -- 2FA (Simplificado em colunas ou JSONB, vamos usar JSONB para flexibilidade futura)
    two_factor_settings JSONB DEFAULT '{"enabled": false}',
    
    -- Auditoria
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMP,
    last_login_at TIMESTAMP,
    last_login_ip VARCHAR(45),
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Índices para performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_org_id ON users(organization_id);
CREATE INDEX idx_users_status ON users(status);