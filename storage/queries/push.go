package queries

const (
	//GetPushReceiversData - return push receivers data
	GetPushReceiversData = `SELECT u.name, d.device_id, d.message_system FROM (SELECT id, name, %s FROM users WHERE id in (SELECT user_id2 as id FROM friends WHERE user_id1 = ?) AND %s > (SELECT %s FROM users WHERE id=?) AND %s < ?) u INNER JOIN device_ids d ON u.id=d.user_id`
)
