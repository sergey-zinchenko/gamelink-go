package queries

const (

	//CreateTableUsers - query to create table users
	CreateTableUsers = `
CREATE TABLE IF NOT EXISTS users (
  id         INT(11)   NOT NULL     AUTO_INCREMENT,
  vk_id      VARCHAR(45)            GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.vk_id'))) VIRTUAL,
  fb_id      VARCHAR(45)            GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.fb_id'))) VIRTUAL,
  name       VARCHAR(45)            GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.name'))) VIRTUAL,
  nickname   VARCHAR(45)            GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.nickname'))),
  sex        ENUM ('F', 'M', 'X')   GENERATED ALWAYS AS (json_unquote(
      ifnull(json_extract(data, '$.sex'), 'X'))) VIRTUAL,
  lb1        INT(11)                GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.lb1'))) VIRTUAL,
  lb2        INT(11)                GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.lb2'))) VIRTUAL,
  lb3        INT(11)                GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.lb3'))),
  bdate      TIMESTAMP              GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.bdate'))),
  email      VARCHAR(45)            GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.email'))),
  lbmeta     JSON                   GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.lbmeta'))),
  data       JSON      NOT NULL,
  created_at TIMESTAMP NOT NULL     DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL         DEFAULT NULL,
  country    VARCHAR(45)            GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.country'))),
  PRIMARY KEY (id),
  UNIQUE INDEX vk_id_UNIQUE (vk_id ASC),
  UNIQUE INDEX fb_id_UNIQUE (fb_id ASC),
  INDEX lb1 (lb1 ASC),
  INDEX lb2 (lb2 ASC),
  INDEX lb3 (lb3 ASC),
  INDEX bdate (bdate ASC),
  INDEX email (email ASC),
  INDEX nickname (nickname ASC),
  INDEX name (name ASC)
)
  ENGINE = InnoDB
  DEFAULT CHARACTER SET = utf8;`

	//CreateTableFriends - query to create table friends
	CreateTableFriends = `
CREATE TABLE IF NOT EXISTS friends (
  user_id1 INT(11) NOT NULL,
  user_id2 INT(11) NOT NULL,
  PRIMARY KEY (user_id1, user_id2),
  INDEX fk_users_has_users_users2_idx (user_id2 ASC),
  INDEX fk_users_has_users_users1_idx (user_id1 ASC),
  CONSTRAINT fk_users_has_users_users1
  FOREIGN KEY (user_id1)
  REFERENCES users (id)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT fk_users_has_users_users2
  FOREIGN KEY (user_id2)
  REFERENCES users (id)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION
)
  ENGINE = InnoDB
  DEFAULT CHARACTER SET = utf8;`

	//CreateTableSaves - query to create table saves
	CreateTableSaves = `CREATE TABLE IF NOT EXISTS saves (
 id INT(11) NOT NULL AUTO_INCREMENT,
 name VARCHAR(45) GENERATED ALWAYS AS (json_unquote(json_extract(data,'$.name'))) VIRTUAL,
 data JSON NOT NULL,
 created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
 updated_at TIMESTAMP NULL DEFAULT NULL,
 user_id INT(11) NOT NULL,
 PRIMARY KEY (id),
 INDEX fk_saves_users_idx (user_id ASC),
 CONSTRAINT fk_saves_users
   FOREIGN KEY (user_id)
   REFERENCES users (id)
   ON DELETE NO ACTION
   ON UPDATE NO ACTION)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8;`

	//CreateTableTournaments - query to create tournaments table
	CreateTableTournaments = `
CREATE TABLE IF NOT EXISTS tournaments (
  id INT NOT NULL AUTO_INCREMENT,	
  expired_time INT NOT NULL,
  UNIQUE INDEX id (expired_time ASC),
  PRIMARY KEY (id))
ENGINE = InnoDB;`

	//CreateTableRooms - query to create rooms table
	CreateTableRooms = `
CREATE TABLE IF NOT EXISTS rooms (
  id INT NOT NULL AUTO_INCREMENT,
  expired_time INT NOT NULL,
  created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX tournament_expired_time (expired_time ASC),
  INDEX id (id ASC),
  CONSTRAINT fk_rooms_1
    FOREIGN KEY (expired_time)
    REFERENCES tournaments (expired_time)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION)
ENGINE = InnoDB;`

	//CreateTableRoomsUsers - query to create table rooms_users
	CreateTableRoomsUsers = `
CREATE TABLE IF NOT EXISTS rooms_users (
  id INT NOT NULL AUTO_INCREMENT,
  room_id INT NOT NULL,
  expired_time INT NOT NULL,
  user_id INT(11) NOT NULL,
  score INT UNSIGNED NULL,
  created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX fk_rooms_users_rooms1_idx (room_id ASC),
  INDEX fk_rooms_users_users1_idx (user_id ASC),
  INDEX fk_rooms_users_tournament_idx (expired_time ASC),
  CONSTRAINT fk_rooms_users_rooms1
    FOREIGN KEY (room_id)
    REFERENCES rooms (id)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION, 
  CONSTRAINT fk_rooms_users_tournament_idx
    FOREIGN KEY (expired_time)
    REFERENCES tournaments (expired_time)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT fk_rooms_users_users1
    FOREIGN KEY (user_id)
    REFERENCES users (id)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION)
ENGINE = InnoDB;`

	//CreateLbView - query to create view for leaderboard
	CreateLbView = `CREATE OR REPLACE
    ALGORITHM = UNDEFINED
    SQL SECURITY DEFINER
VIEW leader_board%[1]d AS
    SELECT 
        u.id AS id,
        IFNULL(u.nickname, u.name) AS nickname,
        u.country AS country,
        u.lbmeta AS meta,
        IFNULL(u.lb%[1]d, 0) AS score
    FROM
        users u
    ORDER BY u.lb%[1]d DESC`
	//CreateUsersTournamentsTable - query to create table users in tournament
	CreateUsersTournamentsTable = `CREATE TABLE IF NOT EXISTS gamelink.users_tournaments (
  tournament_id INT NOT NULL,
  user_id INT(11) NOT NULL,
  PRIMARY KEY (user_id, tournament_id),
  INDEX fk_users_tournaments_tournaments1_idx (tournament_id ASC),
  INDEX fk_users_tournaments_users1_idx (user_id ASC),
  CONSTRAINT fk_users_tournaments_tournaments1
    FOREIGN KEY (tournament_id)
    REFERENCES gamelink.tournaments (id)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT fk_users_tournaments_users1
    FOREIGN KEY (user_id)
    REFERENCES gamelink.users (id)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION)
ENGINE = InnoDB;`
)