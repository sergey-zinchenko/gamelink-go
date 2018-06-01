package queries

const (
	//AllUsersLeaderboard1Query - mysql query to get 1st leader board against all users
	AllUsersLeaderboard1Query = `
SELECT JSON_OBJECT(
    "id", k.id,
    "position", k.pos,
    "name", k.name,
    "score", k.score,
    "top100", CAST(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
                                                            '"id":', b.id, ',',
                                                            '"name":', JSON_QUOTE(b.name), ',',
                                                            '"score":', b.lb2,
                                                            '}')), ']') AS JSON))
FROM (SELECT
        s.id,
        count(*) + 1 as pos,
        s.name,
        s.score
      from leader_board1 l, (select
                               id,
                               lb2 as score,
                               name
                             from leader_board1 o
                             where o.id = ?) s
      where l.lb2 IS NOT NULL AND l.lb2 > s.score) k,
  (SELECT
     l.id,
     l.name,
     l.lb2
   FROM leader_board1 l
   LIMIT 100) b
GROUP BY k.id`
	//AllUsersLeaderboard2Query - mysql query to get second leader board against all users
	AllUsersLeaderboard2Query = `
SELECT JSON_OBJECT(
    "id", k.id,
    "position", k.pos,
    "name", k.name,
    "score", k.score,
    "top100", CAST(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
                                                            '"id":', b.id, ',',
                                                            '"name":', JSON_QUOTE(b.name), ',',
                                                            '"score":', b.lb2,
                                                            '}')), ']') AS JSON))
FROM (SELECT
        s.id,
        count(*) + 1 as pos,
        s.name,
        s.score
      from leader_board1 l, (select
                               id,
                               lb2 as score,
                               name
                             from leader_board1 o
                             where o.id = ?) s
      where l.lb2 IS NOT NULL AND l.lb2 > s.score) k,
  (SELECT
     l.id,
     l.name,
     l.lb2
   FROM leader_board1 l
   LIMIT 100) b
GROUP BY k.id`
	//FriendsLeaderboard1Query - mysql query to get first leader board against friends
	FriendsLeaderboard1Query = `
SELECT JSON_OBJECT(
    "id", k.id,
    "position", k.pos,
    "name", k.name,
    "score", k.score,
    "friends",
    CAST(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
                                                  '"id":', p.id, ',',
                                                  '"name":', JSON_QUOTE(p.name), ',',
                                                  '"score":', p.lb2,
                                                  '}')), ']') AS JSON))
FROM (SELECT
        v.id,
        v.name,
        v.lb2
      FROM (SELECT
              u.id,
              u.name,
              u.lb2
            FROM friends f, users u
            WHERE f.user_id2 = ? AND f.user_id1 = u.id
            UNION SELECT
                    u.id,
                    u.name,
                    u.lb2
                  FROM friends f, users u
                  WHERE f.user_id1 = ? AND f.user_id2 = u.id) v
      ORDER BY v.lb2) p,
  (SELECT
     m.id,
     count(*) + 1 as pos,
     m.name,
     m.score
   FROM (SELECT
           l.id,
           l.name,
           l.lb2 as score
         FROM leader_board1 l
         WHERE l.id = ?) m,
     (SELECT u.lb2 as score
      FROM friends f, users u
      WHERE f.user_id2 = ? AND f.user_id1 = u.id
      UNION
      SELECT u.lb2 as score
      FROM friends f, users u
      WHERE f.user_id1 = ? AND f.user_id2 = u.id) s
   where m.score IS NOT NULL AND s.score > m.score) k
GROUP BY k.id`
	//FriendsLeaderboard2Query - mysql query to get second leader board against friends
	FriendsLeaderboard2Query = `
SELECT JSON_OBJECT(
    "id", k.id,
    "position", k.pos,
    "name", k.name,
    "score", k.score,
    "friends",
    CAST(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
                                                  '"id":', p.id, ',',
                                                  '"name":', JSON_QUOTE(p.name), ',',
                                                  '"score":', p.lb2,
                                                  '}')), ']') AS JSON))
FROM (SELECT
        v.id,
        v.name,
        v.lb2
      FROM (SELECT
              u.id,
              u.name,
              u.lb2
            FROM friends f, users u
            WHERE f.user_id2 = ? AND f.user_id1 = u.id
            UNION SELECT
                    u.id,
                    u.name,
                    u.lb2
                  FROM friends f, users u
                  WHERE f.user_id1 = ? AND f.user_id2 = u.id) v
      ORDER BY v.lb2) p,
  (SELECT
     m.id,
     count(*) + 1 as pos,
     m.name,
     m.score
   FROM (SELECT
           l.id,
           l.name,
           l.lb2 as score
         FROM leader_board1 l
         WHERE l.id = ?) m,
     (SELECT u.lb2 as score
      FROM friends f, users u
      WHERE f.user_id2 = ? AND f.user_id1 = u.id
      UNION
      SELECT u.lb2 as score
      FROM friends f, users u
      WHERE f.user_id1 = ? AND f.user_id2 = u.id) s
   where m.score IS NOT NULL AND s.score > m.score) k
GROUP BY k.id`
)
