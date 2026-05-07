export async function up() {
  return "";
}

export async function down(db: { exec: (sql: string) => Promise<void> }) {
  await db.exec("DROP TABLE IF EXISTS profiles;");
}
