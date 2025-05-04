-- Create the database if it doesn't exist
CREATE DATABASE IF NOT EXISTS `resonite-inventory`;
USE `resonite-inventory`;

-- Users table
CREATE TABLE `Users` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `username` text NOT NULL,
  `auth` varchar(256) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin;

-- Inventories table
CREATE TABLE `Inventories` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` text NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin;

-- Users inventories relationship table
CREATE TABLE `users_inventories` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `inventory_id` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `inventory_id` (`inventory_id`),
  KEY `user_id` (`user_id`),
  CONSTRAINT `users_inventories_ibfk_1` FOREIGN KEY (`inventory_id`) REFERENCES `Inventories` (`id`),
  CONSTRAINT `users_inventories_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `Users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin;

-- Folders table
CREATE TABLE `Folders` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` text NOT NULL,
  `parent_folder_id` int(11) NOT NULL,
  `inventory_id` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `inventoryId` (`inventory_id`),
  CONSTRAINT `Folders_ibfk_1` FOREIGN KEY (`inventory_id`) REFERENCES `Inventories` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin;

-- Items table
CREATE TABLE `Items` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` text NOT NULL,
  `folder_id` int(11) NOT NULL,
  `url` text NOT NULL,
  PRIMARY KEY (`id`),
  KEY `Items_ibfk_1` (`folder_id`),
  CONSTRAINT `Items_ibfk_1` FOREIGN KEY (`folder_id`) REFERENCES `Folders` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin;

-- Assets table
CREATE TABLE `Assets` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `hash` text NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `hash` (`hash`) USING HASH
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin;

-- Hash usage table
CREATE TABLE `hash-usage` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `asset_id` int(11) NOT NULL,
  `item_id` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `asset_id` (`asset_id`),
  KEY `item_id` (`item_id`),
  CONSTRAINT `hash-usage_ibfk_1` FOREIGN KEY (`asset_id`) REFERENCES `Assets` (`id`),
  CONSTRAINT `hash-usage_ibfk_2` FOREIGN KEY (`item_id`) REFERENCES `Items` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin;

-- Tags table
CREATE TABLE `Tags` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` text NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin;

-- Item tags relationship table
CREATE TABLE `item_tags` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tag_id` int(11) NOT NULL,
  `item_id` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `item_id` (`item_id`),
  KEY `tag_id` (`tag_id`),
  CONSTRAINT `item_tags_ibfk_1` FOREIGN KEY (`item_id`) REFERENCES `Items` (`id`),
  CONSTRAINT `item_tags_ibfk_2` FOREIGN KEY (`tag_id`) REFERENCES `Tags` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin;

-- Asset tags relationship table
CREATE TABLE `asset_tags` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `asset_id` int(11) NOT NULL,
  `tag_id` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `asset_id` (`asset_id`),
  KEY `tag_id` (`tag_id`),
  CONSTRAINT `asset_tags_ibfk_1` FOREIGN KEY (`asset_id`) REFERENCES `Assets` (`id`),
  CONSTRAINT `asset_tags_ibfk_2` FOREIGN KEY (`tag_id`) REFERENCES `item_tags` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin;