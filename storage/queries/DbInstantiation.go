package queries

//
//const (
//	//CreateSchema - querie to create schema
//	CreateSchema = `SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
//SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
//SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='TRADITIONAL,ALLOW_INVALID_DATES';
//CREATE SCHEMA IF NOT EXISTS gamelink DEFAULT CHARACTER SET utf8 ;
//USE gamelink ;`
//
//	//CreateTableUsers - querie to create table users
//	CreateTableUsers = `CREATE TABLE IF NOT EXISTS gamelink.users (
//  id INT(11) NOT NULL AUTO_INCREMENT,
//  vk_id VARCHAR(45) GENERATED ALWAYS AS (json_unquote(json_extract(data,'$.vk_id'))) VIRTUAL,
//  fb_id VARCHAR(45) GENERATED ALWAYS AS (json_unquote(json_extract(data,'$.fb_id'))) VIRTUAL,
//  name VARCHAR(45) GENERATED ALWAYS AS (json_unquote(json_extract(data,'$.name'))) VIRTUAL,
//  nickname VARCHAR(45) GENERATED ALWAYS AS (json_unquote(json_extract(data,'$.nickname'))),
//  sex ENUM('F', 'M', 'X') GENERATED ALWAYS AS (json_unquote(ifnull(json_extract(data,'$.sex'),'X'))) VIRTUAL,
//  lb1 INT(11) GENERATED ALWAYS AS (json_unquote(json_extract(data,'$.lb1'))) VIRTUAL,
//  lb2 INT(11) GENERATED ALWAYS AS (json_unquote(json_extract(data,'$.lb2'))) VIRTUAL,
//  lb3 INT(11) GENERATED ALWAYS AS (json_unquote(json_extract(data,'$.lb3'))),
//  lbmeta JSON GENERATED ALWAYS AS (json_unquote(json_extract(data,'$.lbmeta'))),
//  data JSON NOT NULL,
//  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
//  updated_at TIMESTAMP NULL DEFAULT NULL,
//  country VARCHAR(45) GENERATED ALWAYS AS (json_unquote(json_extract(data,'$.country'))),
//  PRIMARY KEY (id),
//  UNIQUE INDEX vk_id_UNIQUE (vk_id ASC),
//  UNIQUE INDEX fb_id_UNIQUE (fb_id ASC),
//  INDEX lb1 (lb1 ASC),
//  INDEX lb2 (lb2 ASC))
//ENGINE = InnoDB
//AUTO_INCREMENT = 49
//DEFAULT CHARACTER SET = utf8;`
//
//	//CreateTableFriends - querie to create table friends
//	CreateTableFriends = `CREATE TABLE IF NOT EXISTS gamelink.friends (
//  user_id1 INT(11) NOT NULL,
//  user_id2 INT(11) NOT NULL,
//  PRIMARY KEY (user_id1, user_id2),
//  INDEX fk_users_has_users_users2_idx (user_id2 ASC),
//  INDEX fk_users_has_users_users1_idx (user_id1 ASC),
//  CONSTRAINT fk_users_has_users_users1
//    FOREIGN KEY (user_id1)
//    REFERENCES gamelink.users (id)
//    ON DELETE NO ACTION
//    ON UPDATE NO ACTION,
//  CONSTRAINT fk_users_has_users_users2
//    FOREIGN KEY (user_id2)
//    REFERENCES gamelink.users (id)
//    ON DELETE NO ACTION
//    ON UPDATE NO ACTION)
//ENGINE = InnoDB
//DEFAULT CHARACTER SET = utf8;`
//
//	//CreateTableSaves - querie to create table saves
//	CreateTableSaves = `CREATE TABLE IF NOT EXISTS gamelink.saves (
//  id INT(11) NOT NULL AUTO_INCREMENT,
//  name VARCHAR(45) GENERATED ALWAYS AS (json_unquote(json_extract(data,'$.name'))) VIRTUAL,
//  data JSON NOT NULL,
//  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
//  updated_at TIMESTAMP NULL DEFAULT NULL,
//  user_id INT(11) NOT NULL,
//  PRIMARY KEY (id),
//  INDEX fk_saves_users_idx (user_id ASC),
//  CONSTRAINT fk_saves_users
//    FOREIGN KEY (user_id)
//    REFERENCES gamelink.users (id)
//    ON DELETE NO ACTION
//    ON UPDATE NO ACTION)
//ENGINE = InnoDB
//DEFAULT CHARACTER SET = utf8;
//
//USE gamelink;`
//
//	//CreateLbView - querie to create view for leaderboard
//	CreateLbView = `CREATE
//    ALGORITHM = UNDEFINED
//    DEFINER = root@localhost
//    SQL SECURITY DEFINER
//VIEW leader_board%[1]d AS
//    SELECT
//        u.id AS id,
//        u.name AS name,
//        u.nickname AS nickname,
//        u.country AS country,
//        u.lbmeta AS meta,
//        IFNULL(u.lb%[1]d, 0) AS score
//    FROM
//        users u
//    ORDER BY u.lb%[1]d DESC`
//)
