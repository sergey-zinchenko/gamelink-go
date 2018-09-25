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
  sex        VARCHAR(1)   			GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.sex'))) VIRTUAL,
  lb1        INT(11)                GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.lb1'))) VIRTUAL,
  lb2        INT(11)                GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.lb2'))) VIRTUAL,
  lb3        INT(11)                GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.lb3'))),
  bdate      VARCHAR(45)            GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.bdate'))),
  email      VARCHAR(45)            GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.email'))),
  meta     JSON                   GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.meta'))),
  data       JSON      NOT NULL,
  created_at TIMESTAMP NOT NULL     DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL         DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  country    VARCHAR(45)            GENERATED ALWAYS AS (json_unquote(json_extract(data, '$.country'))),
  deleted 	 TINYINT(1) NOT NULL DEFAULT 0, 
  PRIMARY KEY (id),
  UNIQUE INDEX vk_id_UNIQUE (vk_id ASC),
  UNIQUE INDEX fb_id_UNIQUE (fb_id ASC),
  INDEX lb1 (lb1 ASC),
  INDEX lb2 (lb2 ASC),
  INDEX lb3 (lb3 ASC),
  INDEX bdate (bdate ASC),
  INDEX email (email ASC),
  INDEX nickname (nickname ASC),
  INDEX name (name ASC),
  INDEX deleted (deleted ASC)
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
 state JSON GENERATED ALWAYS AS (json_unquote(json_extract(data,'$.state'))) VIRTUAL,
 data JSON NOT NULL,
 created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
 updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
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
  id INT(11) NOT NULL AUTO_INCREMENT,
  tournament_expired_time INT(11) NOT NULL,
  registration_expired_time INT(11) NOT NULL,
  users_in_room INT(11) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE INDEX id (tournament_expired_time ASC),
  INDEX tournament_exp_time (tournament_expired_time ASC))
ENGINE = InnoDB;`

	//CreateTableRooms - query to create rooms table
	CreateTableRooms = `
CREATE TABLE IF NOT EXISTS rooms (
  id INT(11) NOT NULL AUTO_INCREMENT,
  tournament_id INT(11) NOT NULL,
  INDEX id (id ASC),
  INDEX fk_rooms_1_idx (tournament_id ASC),
  CONSTRAINT fk_rooms_1
    FOREIGN KEY (tournament_id)
    REFERENCES gamelink.tournaments (id)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION)
ENGINE = InnoDB;`

	//CreateTableRoomsUsers - query to create table rooms_users
	CreateTableRoomsUsers = `CREATE TABLE IF NOT EXISTS gamelink.rooms_users (
id INT(11) NOT NULL AUTO_INCREMENT,
tournament_id INT(11) NOT NULL,
tournament_expired_time INT(11) NOT NULL,
room_id INT(11) NOT NULL,
user_id INT(11) NOT NULL,
score INT(11) UNSIGNED NULL DEFAULT NULL,
  PRIMARY KEY (id),
  INDEX fk_rooms_users_rooms1_idx (room_id ASC),
  INDEX fk_rooms_users_users1_idx (user_id ASC),
  INDEX fk_rooms_users_1_idx (tournament_id ASC),
  INDEX index5 (score ASC),
  INDEX fk_rooms_users_2_idx (tournament_expired_time ASC),
  CONSTRAINT fk_rooms_users_1
    FOREIGN KEY (tournament_id)
    REFERENCES gamelink.tournaments (id)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT fk_rooms_users_2
    FOREIGN KEY (tournament_expired_time)
    REFERENCES gamelink.tournaments (tournament_expired_time)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT fk_rooms_users_rooms1
    FOREIGN KEY (room_id)
    REFERENCES gamelink.rooms (id)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT fk_rooms_users_users1
    FOREIGN KEY (user_id)
     REFERENCES gamelink.users (id)
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
        u.meta AS meta,
        IFNULL(u.lb%[1]d, 0) AS score
    FROM
        users u
    ORDER BY u.lb%[1]d DESC`
	//CreateUsersTournamentsTable - query to create table users in tournament
	CreateUsersTournamentsTable = `CREATE TABLE IF NOT EXISTS gamelink.users_tournaments (
  tournament_id INT(11) NOT NULL,
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

	//CreateTableDbVersion - create table for database versions
	CreateTableDbVersion = `CREATE TABLE IF NOT EXISTS gamelink.db_version (
 version INT NOT NULL,
 PRIMARY KEY (version));`
	//InsertVersionZero - insert 0 version of db
	InsertVersionZero = `INSERT IGNORE INTO gamelink.db_version (version) VALUES (0);`
	//ModifyBdateColumn - change varchar column to DATE
	ModifyBdateColumn = `ALTER TABLE gamelink.users MODIFY COLUMN bdate INT GENERATED ALWAYS AS (json_unquote(json_extract(data,'$.bdate'))) VIRTUAL ;`
	//InsertVersionOne - insert new db version
	InsertVersionOne = `INSERT IGNORE INTO gamelink.db_version (version) values (1);`
	//GetDbVersion - return max bd version
	GetDbVersion = `SELECT MAX(version) FROM gamelink.db_version`
)
