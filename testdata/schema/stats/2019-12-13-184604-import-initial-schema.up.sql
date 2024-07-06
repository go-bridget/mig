-- Repository table
CREATE TABLE `repository` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT 'Repository ID',
    `name` VARCHAR(255) NOT NULL COMMENT 'Repository name',
    `url` VARCHAR(255) NOT NULL COMMENT 'Repository URL',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Record creation timestamp',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Record update timestamp'
) COMMENT='Stores basic information about repositories';

-- Branch table
CREATE TABLE `branch` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT 'Branch ID',
    `repository_id` INT NOT NULL COMMENT 'Repository ID',
    `name` VARCHAR(255) NOT NULL COMMENT 'Branch name',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Record creation timestamp',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Record update timestamp',
    FOREIGN KEY (`repository_id`) REFERENCES `repository`(`id`) ON DELETE CASCADE
) COMMENT='Stores information about branches in repositories';

-- Commit table
CREATE TABLE `commit` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT 'Commit ID',
    `branch_id` INT NOT NULL COMMENT 'Branch ID',
    `commit_hash` VARCHAR(40) NOT NULL COMMENT 'Commit hash',
    `author` VARCHAR(255) COMMENT 'Commit author',
    `message` TEXT COMMENT 'Commit message',
    `committed_at` TIMESTAMP COMMENT 'Commit timestamp',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Record creation timestamp',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Record update timestamp',
    FOREIGN KEY (`branch_id`) REFERENCES `branch`(`id`) ON DELETE CASCADE
) COMMENT='Stores information about commits in branches';

-- Asset table
CREATE TABLE `asset` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT 'Asset ID',
    `commit_id` INT NOT NULL COMMENT 'Commit ID',
    `filename` VARCHAR(255) NOT NULL COMMENT 'Filename',
    `contents` LONGTEXT COMMENT 'File contents',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Record creation timestamp',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Record update timestamp',
    FOREIGN KEY (`commit_id`) REFERENCES `commit`(`id`) ON DELETE CASCADE
) COMMENT='Stores asset information for each commit';
