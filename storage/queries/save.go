package queries

const (
	//GetAllSavesQuery - mysql query to get all saves of given user
	GetAllSavesQuery = `
SELECT CAST(CONCAT('[',GROUP_CONCAT(DISTINCT CONCAT('{',
							'"id":', 	s.id, ',', 
                            '"name":', 	IFNULL(JSON_QUOTE(s.name), ""), ',', 
                            '"state":', s.state, ',', 
                            '"date":',	s.updated_at,
                            '}')), ']') AS JSON) as saves
    FROM (SELECT s.id, s.name, s.state, UNIX_TIMESTAMP(s.updated_at) as updated_at 
    from saves s, users u WHERE u.id=? AND u.deleted != 1  AND s.user_id=u.id) s`
	//GetSaveQuery - mysql query to get specified save of given user
	GetSaveQuery = `
SELECT CAST(CONCAT(
    '{"id":'  , 	s.id, 
	IFNULL(CONCAT(',"name":' , 	JSON_QUOTE(s.name)),""),
    IFNULL(CONCAT(',"state":', 	s.state),""), ',',
    '"date":', s.updated_at,
    '}') AS JSON) as save 
    FROM (SELECT s.id, s.name, s.state, UNIX_TIMESTAMP(s.updated_at) as updated_at 
    from saves s, users u WHERE u.id=? AND u.deleted != 1  AND s.user_id=u.id  AND s.id = ?) s`
	//GetSaveDataQuery - mysql query to get save's json field data
	GetSaveDataQuery = `
SELECT data
FROM saves s
WHERE s.id = ? AND s.user_id = (SELECT id from users WHERE id=? AND deleted != 1)`
	//UpdateSaveDataQuery - mysql query to update save's data field
	UpdateSaveDataQuery = `
UPDATE saves s
SET s.data = ?
WHERE s.id = ? AND s.user_id = (SELECT id from users WHERE id=? AND deleted != 1)`

	//DeleteSaveQuery - mysql query to delete save
	DeleteSaveQuery = `
DELETE FROM saves 
WHERE id = ? AND user_id = (SELECT id from users WHERE id=? AND deleted != 1)
`
	//CreateSaveQuery - mysql query to create save
	CreateSaveQuery = `INSERT INTO saves (data, user_id) SELECT ?, id FROM users WHERE id=? AND deleted !=1`

	//DeleteAllSaves - mysql query to delete all saves
	DeleteAllSaves = `DELETE FROM saves where user_id = ? AND s.user_id = (SELECT id from users WHERE id=? AND deleted != 1)`
)
