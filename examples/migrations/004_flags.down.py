def up(db):
    return None

def down(db):
    db.exec("ALTER TABLE users DROP COLUMN IF EXISTS flags;")
