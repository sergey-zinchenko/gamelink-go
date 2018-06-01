package queries

const (
	//GetAllSavesQuery - mysql query to get all saves of given user
	GetAllSavesQuery = `
SELECT
  CAST(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{', '"id":', s.id, ',', '"name":', JSON_QUOTE(s.name), '}')), ']')
       AS JSON)
FROM saves s
WHERE s.user_id = ?
GROUP BY s.user_id`
	//GetSaveQuery - mysql query to get specified save of given user
	GetSaveQuery = `
SELECT JSON_OBJECT('id', s.id, 'name', s.name)
FROM saves s
WHERE s.id = ? AND s.user_id = ?`
	//GetSaveDataQuery - mysql query to get save's json field data
	GetSaveDataQuery = `
SELECT data
FROM saves s
WHERE s.id = ?`
	//UpdateSaveDataQuery - mysql query to update save's data field
	UpdateSaveDataQuery = `
UPDATE saves s
SET s.data = ?
WHERE s.id = ?`

	//DeleteSaveQuery - mysql query to delete save
	DeleteSaveQuery = `
DELETE FROM saves
WHERE id = ?
`
	//CreateSaveQuery - mysql query to create save
	CreateSaveQuery = `INSERT INTO saves (data, user_id) VALUES (?, ?)`
)
