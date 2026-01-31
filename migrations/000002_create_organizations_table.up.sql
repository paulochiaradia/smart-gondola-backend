-- TABELA ORGANIZATIONS (Atualizada)
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    document VARCHAR(50), 
    slug VARCHAR(100) NOT NULL UNIQUE,
    plan VARCHAR(50) NOT NULL DEFAULT 'free',
    sector VARCHAR(50) NOT NULL DEFAULT 'other',
    
    -- Coluna Mágica para limites e configs futuras
    settings JSONB NOT NULL DEFAULT '{}', 
    
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_organizations_sector ON organizations(sector);

CREATE TABLE IF NOT EXISTS stores (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50), 
    
    -- Endereço e Timezone
    address_street VARCHAR(255),
    address_number VARCHAR(20),
    address_complement VARCHAR(100),
    address_district VARCHAR(100),
    address_city VARCHAR(100),
    address_state VARCHAR(2),
    address_zip_code VARCHAR(20),
    
    timezone VARCHAR(50) DEFAULT 'America/Sao_Paulo',
    
    is_active BOOLEAN DEFAULT TRUE,
    deleted_at TIMESTAMP,
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_stores_organization 
        FOREIGN KEY (organization_id) 
        REFERENCES organizations(id) ON DELETE CASCADE,

    -- [SUGESTÃO A] Constraint de Unicidade Composta
    -- Tradução: "Nesta organização, este código só pode aparecer uma vez"
    CONSTRAINT uq_stores_org_code UNIQUE (organization_id, code)
);
-- Index para buscar todas as lojas de uma Org rapidamente
CREATE INDEX idx_stores_org_id ON stores(organization_id);

-- TABELA USERS (Já existe, mas precisamos garantir as FKs)
-- Lembre-se: User tem organization_id (Obrigatório) e store_id (Opcional)
ALTER TABLE users 
    ADD CONSTRAINT fk_users_organization 
    FOREIGN KEY (organization_id) REFERENCES organizations(id);

ALTER TABLE users 
    ADD CONSTRAINT fk_users_store 
    FOREIGN KEY (store_id) REFERENCES stores(id);