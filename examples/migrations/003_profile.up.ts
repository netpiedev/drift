export async function up(db: { exec: (sql: string) => Promise<void> }) {
  await db.exec(`
    CREATE TABLE IF NOT EXISTS profiles (
      id BIGSERIAL PRIMARY KEY,
      user_id BIGINT NOT NULL REFERENCES users(id),
      display_name TEXT NOT NULL
    );
  `);
}

export async function down(db: { exec: (sql: string) => Promise<void> }) {
  await db.exec("DROP TABLE IF EXISTS profiles;");
}
