-- CREATE USER 'davinci'@'localhost' IDENTIFIED BY '4KzyzTL9gyQpycJ9';
-- CREATE DATABASE DaVinci_code;
-- GRANT ALL ON DaVinci_code.* TO 'davinci'@'localhost';
-- FLUSH PRIVILEGES;

CREATE TABLE `DaVinci_code`.`user`(
    `id` INT NOT NULL AUTO_INCREMENT,
    `openid` VARCHAR(255) NOT NULL,
    `nickname` VARCHAR(255) NOT NULL DEFAULT '初来乍到',
    `avatarurl` VARCHAR(255) NOT NULL DEFAULT 'https://whoisnian.com/public/avatar.jpg',
    `gender` INT NOT NULL DEFAULT 0,
    `time` DATE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(`id`),
    UNIQUE(`openid`)
) ENGINE = InnoDB;

CREATE TABLE `DaVinci_code`.`score`(
    `id` INT NOT NULL AUTO_INCREMENT,
    `openid` VARCHAR(255) NOT NULL,
    `scoreall` int(11) NOT NULL DEFAULT 0,
    `num` int(11) NOT NULL DEFAULT 0,
    `num4` int(11) NOT NULL DEFAULT 0,
    `num3` int(11) NOT NULL DEFAULT 0,
    `num2` int(11) NOT NULL DEFAULT 0,
    PRIMARY KEY(`id`),
    UNIQUE(`openid`)
) ENGINE = InnoDB;

CREATE TABLE `DaVinci_code`.`setting`(
    `id` INT NOT NULL AUTO_INCREMENT,
    `openid` VARCHAR(255) NOT NULL,
    `vol` INT NOT NULL DEFAULT 0,
    PRIMARY KEY(`id`),
    UNIQUE(`openid`)
) ENGINE = InnoDB;
