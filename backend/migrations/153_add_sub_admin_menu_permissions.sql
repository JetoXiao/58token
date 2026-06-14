-- Add menu-level read-only permissions for delegated administrators.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS admin_menu_permissions TEXT NOT NULL DEFAULT '[]';
