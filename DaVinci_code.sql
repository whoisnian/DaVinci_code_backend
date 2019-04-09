-- CREATE USER 'davinci'@'localhost' IDENTIFIED BY '4KzyzTL9gyQpycJ9';
-- CREATE DATABASE DaVinci_code;
-- GRANT ALL ON DaVinci_code.* TO 'davinci'@'localhost';
-- FLUSH PRIVILEGES;

CREATE TABLE `DaVinci_code`.`user`(
    `id` INT NOT NULL AUTO_INCREMENT,
    `openid` VARCHAR(255) NOT NULL,
    `nickname` VARCHAR(255) NOT NULL,
    `avatarurl` VARCHAR(255) NOT NULL,
    `gender` INT NOT NULL,
    `time` DATE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(`id`),
    UNIQUE(`openid`)
) ENGINE = InnoDB;

CREATE TABLE `DaVinci_code`.`score`(
    `id` INT NOT NULL AUTO_INCREMENT,
    `openid` VARCHAR(255) NOT NULL,
    `scoreall` INT NOT NULL,
    `num` INT NOT NULL,
    `numwin` INT NOT NULL,
    `num4` INT NOT NULL,
    `numwin4` INT NOT NULL,
    `num3` INT NOT NULL,
    `numwin3` INT NOT NULL,
    `num2` INT NOT NULL,
    `numwin2` INT NOT NULL,
    PRIMARY KEY(`id`)
) ENGINE = InnoDB;

CREATE TABLE `DaVinci_code`.`setting`(
    `id` INT NOT NULL AUTO_INCREMENT,
    `openid` VARCHAR(255) NOT NULL,
    `vol` INT NOT NULL,
    PRIMARY KEY(`id`)
) ENGINE = InnoDB;
