CREATE TABLE IF NOT EXISTS `migrations` (
 `project` text,
 `filename` text,
 `statement_index` integer,
 `status` text,
 PRIMARY KEY (project, filename)
);
