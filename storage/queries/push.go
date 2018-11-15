package queries

const (
	//GetPushReceiversData - return push receivers data
	GetPushReceiversData = `SELECT u.id, u.name, d.device_id, d.message_system FROM (SELECT id, name, ? FROM users 
							WHERE id in (SELECT user_id2 as id FROM friends WHERE user_id1 = ?) AND ? > (SELECT ? FROM users WHERE id=?) AND ? < ?) u 
                            INNER JOIN device_ids d ON u.id=d.user_id`
)
