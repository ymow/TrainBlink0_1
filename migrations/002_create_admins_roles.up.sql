-- ============================================================
-- Admins Table (Admin Users)
-- ============================================================
CREATE TABLE admins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Authentication
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,

    -- Profile
    name VARCHAR(100) NOT NULL,

    -- Role & Permissions
    role VARCHAR(50) NOT NULL DEFAULT 'moderator',
    permissions JSONB DEFAULT '[]',

    -- Status
    is_active BOOLEAN DEFAULT true,

    -- Security
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip INET,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES admins(id) ON DELETE SET NULL,

    CONSTRAINT check_role CHECK (role IN ('super_admin', 'admin', 'moderator'))
);

CREATE INDEX idx_admins_email ON admins(email);
CREATE INDEX idx_admins_role ON admins(role);
CREATE INDEX idx_admins_is_active ON admins(is_active) WHERE is_active = true;

CREATE TRIGGER update_admins_updated_at BEFORE UPDATE ON admins
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- Roles & Permissions
-- ============================================================
CREATE TABLE roles (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    permissions JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default roles
INSERT INTO roles (id, name, description, permissions) VALUES
('super_admin', 'Super Administrator', 'Full system access', '["*"]'::jsonb),
('admin', 'Administrator', 'Manage users and content',
 '["user.*", "content.*", "report.*", "station.*", "analytics.view"]'::jsonb),
('moderator', 'Moderator', 'Review reports and moderate content',
 '["user.view", "user.ban", "content.view", "content.delete", "report.*"]'::jsonb);

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add foreign key for users.banned_by after admins table exists
ALTER TABLE users ADD CONSTRAINT fk_users_banned_by
    FOREIGN KEY (banned_by) REFERENCES admins(id) ON DELETE SET NULL;
