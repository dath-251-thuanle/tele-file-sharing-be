--- migrations/004_select_file_userid.sql

SELECT
    f.id,
    f.object_key,
    f.filename,
    f.size,
    f.mime,
    f.status,
    f.created_at,
    f.updated_at
FROM files f
JOIN users u ON u.id = f.owner_user_id
WHERE u.telegram_user_id = $1
ORDER BY f.created_at DESC;