package queries

const (
	//AllUsersLeaderboardQuery - mysql query template to get leader board against all users
	AllUsersLeaderboardQuery = `
SELECT JSON_OBJECT(
    "id", k.id,
    "position", k.pos,
    "name", k.name,
    "score", k.score,
    "top100", CAST(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
                                                            '"id":', b.id, ',',
                                                            '"name":', JSON_QUOTE(b.name), ',',
                                                            '"score":', b.lb%[1]d
                                                            ,
                                                            '}')), ']') AS JSON))
FROM (SELECT
        s.id,
        count(*) + 1 as pos,
        s.name,
        s.score
      from leader_board1 l, (select
                               id,
                               lb%[1]d
                                as score,
                               name
                             from leader_board1 o
                             where o.id = ?) s
      where l.lb%[1]d
       IS NOT NULL AND l.lb%[1]d
        > s.score) k,
  (SELECT
     l.id,
     l.name,
     l.lb%[1]d

   FROM leader_board1 l
   LIMIT 100) b
GROUP BY k.id`

	//FriendsLeaderboardQuery - mysql query template to get leader board against friends
	FriendsLeaderboardQuery = `
SET @param = ?;
SELECT JSON_OBJECT(
    "id", k.id,
    "position", k.pos,
    "name", k.name,
    "score", k.score,
    "friends",
    CAST(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
                                                  '"id":', p.id, ',',
                                                  '"name":', JSON_QUOTE(p.name), ',',
                                                  '"score":', p.lb%[1]d
                                                  ,
                                                  '}')), ']') AS JSON))
FROM (SELECT
        v.id,
        v.name,
        v.lb%[1]d

      FROM (SELECT
              u.id,
              u.name,
              u.lb%[1]d

            FROM friends f, users u
            WHERE f.user_id2 = @param AND f.user_id1 = u.id
            UNION SELECT
                    u.id,
                    u.name,
                    u.lb%[1]d

                  FROM friends f, users u
                  WHERE f.user_id1 = @param AND f.user_id2 = u.id) v
      ORDER BY v.lb%[1]d
      ) p,
  (SELECT
     m.id,
     count(*) + 1 as pos,
     m.name,
     m.score
   FROM (SELECT
           l.id,
           l.name,
           l.lb%[1]d
            as score
         FROM leader_board1 l
         WHERE l.id = @param) m,
     (SELECT u.lb%[1]d
      as score
      FROM friends f, users u
      WHERE f.user_id2 = @param AND f.user_id1 = u.id
      UNION
      SELECT u.lb%[1]d
       as score
      FROM friends f, users u
      WHERE f.user_id1 = @param AND f.user_id2 = u.id) s
   where m.score IS NOT NULL AND s.score > m.score) k
GROUP BY k.id`
)
