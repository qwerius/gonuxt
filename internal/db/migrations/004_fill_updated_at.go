package migrations

var Migration004FillUpdatedAt = Migration{
    Version: 4,
    Name:    "fill_updated_at_for_existing_users",
    Up: `
        UPDATE users
        SET updated_at = created_at
        WHERE updated_at IS NULL;
    `,
    Down: `
        -- Opsional: biasanya tidak perlu rollback
        -- Misal bisa set ke NULL
        UPDATE users
        SET updated_at = NULL
        WHERE updated_at = created_at;
    `,
}
