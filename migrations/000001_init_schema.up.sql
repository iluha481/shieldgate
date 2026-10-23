CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS services (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    path_prefix VARCHAR(100) NOT NULL,
    target_url VARCHAR(255) NOT NULL,
    rate_limit_rps INT NOT NULL DEFAULT 10,
    burst_capacity INT NOT NULL DEFAULT 20,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

 CREATE TABLE IF NOT EXISTS api_keys (
    id SERIAL PRIMARY KEY,
    service_id INT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,                     
    key_hash VARCHAR(64) UNIQUE NOT NULL,           
    key_prefix VARCHAR(12) NOT NULL,                
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

CREATE INDEX IF NOT EXISTS idx_services_path_prefix ON services(path_prefix);

CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);

CREATE INDEX IF NOT EXISTS idx_services_user_id ON services(user_id);

