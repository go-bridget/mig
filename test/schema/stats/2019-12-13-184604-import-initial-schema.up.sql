CREATE TABLE `incoming` (
 `id` bigint(20) unsigned NOT NULL COMMENT 'Tracking ID',
 `property` varchar(32) COLLATE utf8_slovenian_ci NOT NULL COMMENT 'Property name (human readable, a-z)',-- comment here
 `property_section` int(11) unsigned NOT NULL COMMENT 'Property Section ID',
 `property_id` int(11) unsigned NOT NULL COMMENT 'Property Item ID', -- comments can go anywhere
 `remote_ip` varchar(255) COLLATE utf8_slovenian_ci NOT NULL COMMENT 'Remote IP from user making request',
 `stamp` datetime NOT NULL COMMENT 'Timestamp of request',
 PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_slovenian_ci COMMENT='Incoming stats log, writes only'; -- we can add comments after statements

-- we can add comments betweeb statements

CREATE TABLE `incoming_proc` LIKE `incoming`; -- comment after statement

-- trailing comment
