-- Create the database if it doesn't exist
CREATE DATABASE IF NOT EXISTS `resonite-inventory`;
USE `resonite-inventory`;

-- Users table for authentication and user management
CREATE TABLE IF NOT EXISTS `Users` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `username` VARCHAR(255) NOT NULL UNIQUE,
  `auth` VARCHAR(255) NOT NULL, -- Stores bcrypt hashed passwords
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Inventories table to group folders and assets
CREATE TABLE IF NOT EXISTS `Inventories` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `name` VARCHAR(255) NOT NULL,
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Many-to-many relationship between users and inventories
CREATE TABLE IF NOT EXISTS `users_inventories` (
  `user_id` INT NOT NULL,
  `inventory_id` INT NOT NULL,
  `access_level` ENUM('owner', 'editor', 'viewer') DEFAULT 'owner',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`, `inventory_id`),
  FOREIGN KEY (`user_id`) REFERENCES `Users` (`id`) ON DELETE CASCADE,
  FOREIGN KEY (`inventory_id`) REFERENCES `Inventories` (`id`) ON DELETE CASCADE
);

-- Folders table for organizing items
CREATE TABLE IF NOT EXISTS `Folders` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `name` VARCHAR(255) NOT NULL,
  `parent_folder_id` INT NULL,
  `inventory_id` INT NOT NULL,
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (`parent_folder_id`) REFERENCES `Folders` (`id`) ON DELETE CASCADE,
  FOREIGN KEY (`inventory_id`) REFERENCES `Inventories` (`id`) ON DELETE CASCADE
);

-- Items table for actual assets
CREATE TABLE IF NOT EXISTS `Items` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `name` VARCHAR(255) NOT NULL,
  `url` VARCHAR(512) NOT NULL,  -- Path/URL to the asset file
  `folder_id` INT NOT NULL,
  `type` VARCHAR(50) NOT NULL DEFAULT 'asset',  -- Type of asset (e.g., model, texture, sound)
  `size` BIGINT UNSIGNED NULL,  -- File size in bytes
  `hash` VARCHAR(64) NULL,      -- For deduplication purposes
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (`folder_id`) REFERENCES `Folders` (`id`) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX `idx_folders_parent` ON `Folders` (`parent_folder_id`);
CREATE INDEX `idx_folders_inventory` ON `Folders` (`inventory_id`);
CREATE INDEX `idx_items_folder` ON `Items` (`folder_id`);
CREATE INDEX `idx_items_url` ON `Items` (`url`);

-- Create a root inventory and folder for testing
INSERT INTO `Inventories` (`id`, `name`) VALUES (1, 'Default Inventory');

-- Create a root folder for the default inventory
INSERT INTO `Folders` (`id`, `name`, `parent_folder_id`, `inventory_id`) 
VALUES (1, 'Root', NULL, 1);

-- Procedure to create a new user with default inventory
DELIMITER //
CREATE PROCEDURE create_user_with_inventory(IN username VARCHAR(255), IN auth_hash VARCHAR(255))
BEGIN
    DECLARE new_user_id INT;
    DECLARE new_inventory_id INT;
    
    START TRANSACTION;
    
    -- Create user
    INSERT INTO `Users` (`username`, `auth`) VALUES (username, auth_hash);
    SET new_user_id = LAST_INSERT_ID();
    
    -- Create personal inventory
    INSERT INTO `Inventories` (`name`) VALUES (CONCAT(username, '\'s Inventory'));
    SET new_inventory_id = LAST_INSERT_ID();
    
    -- Associate user with inventory
    INSERT INTO `users_inventories` (`user_id`, `inventory_id`, `access_level`) 
    VALUES (new_user_id, new_inventory_id, 'owner');
    
    -- Create root folder
    INSERT INTO `Folders` (`name`, `parent_folder_id`, `inventory_id`) 
    VALUES ('Root', NULL, new_inventory_id);
    
    COMMIT;
END //
DELIMITER ;