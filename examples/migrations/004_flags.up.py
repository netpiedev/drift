def up(db):
    db.exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS flags INTEGER NOT NULL DEFAULT 0;")

def down(db):
    db.exec("ALTER TABLE users DROP COLUMN IF EXISTS flags;")
