-- ****************** SqlDBM: MySQL ******************;
-- ***************************************************;

DROP TABLE user, translation, highlight, location, event_schedule, event, category, trip_itinerary_event, trip_participant, trip_itinerary, trip_invite, trip;


-- ************************************** `user`

CREATE TABLE `user`
(
 `id`                   varchar(45) NOT NULL ,
 `active`               smallint NOT NULL DEFAULT 1 ,
 `principal_trip_id`    varchar(45) NOT NULL ,
 `first_name`           text ,
 `last_name`            text ,
 `image_path`           text ,
 `country_id`           varchar(45) ,
 `region_id`            varchar(45) ,
 `city_id`              varchar(45) ,
 `about_me`             mediumtext ,
PRIMARY KEY (`id`)
);



-- ************************************** `trip`

CREATE TABLE `trip`
(
 `id`           varchar(45) NOT NULL ,
 `scope`        tinytext NOT NULL ,
 `itinerary_id` varchar(45) NOT NULL ,
 `created_by`   varchar(45) NOT NULL ,
 `created_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ,
 `updated_by`   varchar(45) NOT NULL ,
 `updated_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP ,
PRIMARY KEY (`id`)
);






-- ************************************** `translation`

CREATE TABLE `translation`
(
 `id`        varchar(45) NOT NULL ,
 `parent_id` varchar(45) NOT NULL ,
 `table`     tinytext NOT NULL ,
 `field`     tinytext NOT NULL ,
 `pt`        text ,
 `es`        text ,
 `en`        text ,
PRIMARY KEY (`id`)
);






-- ************************************** `highlight`

CREATE TABLE `highlight`
(
 `id`              varchar(45) NOT NULL ,
 `image_path`      text ,
 `active`          smallint NOT NULL DEFAULT 1 ,
 `schedule_date`   timestamp ,
 `filter`          text ,
 `trip_ids`        text ,
 `event_ids`       text ,
 `created_by`      varchar(45) NOT NULL ,
 `created_date`    timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ,
 `updated_by`      varchar(45) NOT NULL ,
 `updated_date`    timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP ,
PRIMARY KEY (`id`)
);






-- ************************************** `location`

CREATE TABLE `location`
(
 `id`         varchar(45) NOT NULL ,
 `country_id` varchar(45) ,
 `region_id`  varchar(45) ,
PRIMARY KEY (`id`)
);






-- ************************************** `event`

CREATE TABLE `event`
(
 `id`                    varchar(45) NOT NULL ,
 `active`                smallint NOT NULL DEFAULT 1,
 `main_category_id`      varchar(45) ,
 `secondary_category_id` varchar(45) ,
 `country_id`            varchar(45) ,
 `region_id`             varchar(45) ,
 `city_id`               varchar(45) ,
 `address`               text ,
 `created_by`   		 varchar(45) NOT NULL ,
 `created_date` 		 timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ,
 `updated_by`   		 varchar(45) NOT NULL ,
 `updated_date` 		 timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP ,
PRIMARY KEY (`id`)
);






-- ************************************** `category`

CREATE TABLE `category`
(
 `id`           varchar(45) NOT NULL ,
 `parent_id`    varchar(45) ,
 `active`       smallint NOT NULL DEFAULT 1 ,
 `created_by`   varchar(45) NOT NULL ,
 `created_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ,
 `updated_by`   varchar(45) NOT NULL ,
 `updated_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP ,
PRIMARY KEY (`id`)
);






-- ************************************** `trip_participant`

CREATE TABLE `trip_participant`
(
 `id`      		varchar(45) NOT NULL ,
 `trip_id` 		varchar(45) NOT NULL ,
 `user_id` 		varchar(45) NOT NULL ,
 `role`    		varchar(45) NOT NULL ,
 `created_by`   varchar(45) NOT NULL ,
 `created_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ,
 `updated_by`   varchar(45) NOT NULL ,
 `updated_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP ,
PRIMARY KEY (`id`, `trip_id`),
KEY `fk_trip` (`trip_id`),
UNIQUE KEY `unique_participant` (`trip_id`,`user_id`),
CONSTRAINT `FK_127` FOREIGN KEY `fk_trip` (`trip_id`) REFERENCES `trip` (`id`) ON DELETE CASCADE
);






-- ************************************** `trip_itinerary`

CREATE TABLE `trip_itinerary`
(
 `id`         	varchar(45) NOT NULL ,
 `trip_id`    	varchar(45) NOT NULL ,
 `owner_id`   	varchar(45) NOT NULL ,
 `start_date` 	timestamp NOT NULL ,
 `end_date`   	timestamp NOT NULL ,
 `created_by`   varchar(45) NOT NULL ,
 `created_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ,
 `updated_by`   varchar(45) NOT NULL ,
 `updated_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP ,
PRIMARY KEY (`id`, `trip_id`),
KEY `fk_trip` (`trip_id`),
CONSTRAINT `FK_75` FOREIGN KEY `fk_trip` (`trip_id`) REFERENCES `trip` (`id`) ON DELETE CASCADE
);






-- ************************************** `trip_invite`

CREATE TABLE `trip_invite`
(
 `id`      		varchar(45) NOT NULL ,
 `email`   		text NOT NULL ,
 `trip_id` 		varchar(45) NOT NULL ,
 `created_by`   varchar(45) NOT NULL ,
 `created_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ,
PRIMARY KEY (`id`, `trip_id`),
KEY `fk_trip` (`trip_id`),
CONSTRAINT `FK_135` FOREIGN KEY `fk_trip` (`trip_id`) REFERENCES `trip` (`id`) ON DELETE CASCADE
);






-- ************************************** `event_schedule`

CREATE TABLE `event_schedule`
(
 `id`           varchar(45) NOT NULL ,
 `event_id`     varchar(45) NOT NULL ,
 `annually`     smallint NOT NULL DEFAULT 0,
 `fixed_date`   smallint NOT NULL DEFAULT 0,
 `fixed_period` smallint NOT NULL DEFAULT 0,
 `closed`       smallint NOT NULL DEFAULT 0,
 `start_date`   timestamp ,
 `end_date`     timestamp ,
 `week_days`    varchar(45) ,
 `created_by`   varchar(45) NOT NULL ,
 `created_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ,
 `updated_by`   varchar(45) NOT NULL ,
 `updated_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP ,
PRIMARY KEY (`id`, `event_id`),
KEY `fk_event` (`event_id`),
CONSTRAINT `FK_103` FOREIGN KEY `fk_event` (`event_id`) REFERENCES `event` (`id`) ON DELETE CASCADE
);






-- ************************************** `itinerary_event`

CREATE TABLE `trip_itinerary_event`
(
 `id`                    varchar(45) NOT NULL ,
 `itinerary_id`          varchar(45) NOT NULL ,
 `trip_id`               varchar(45) NOT NULL ,
 `global_event_id`       varchar(45) ,
 `begin_offset`          bigint NOT NULL DEFAULT 0,
 `duration`              bigint NOT NULL DEFAULT 0,
 `main_category_id`      varchar(45) ,
 `secondary_category_id` varchar(45) ,
 `country_id`            varchar(45) ,
 `region_id`             varchar(45) ,
 `city_id`               varchar(45) ,
 `address`               text ,
 `created_by`   		 varchar(45) NOT NULL ,
 `created_date` 		 timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ,
 `updated_by`   		 varchar(45) NOT NULL ,
 `updated_date` 		 timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP ,
 `evaluated_by`   		 varchar(45) ,
 `evaluated_date` 		 timestamp ,
 `evaluated_comment`     text ,
PRIMARY KEY (`id`, `itinerary_id`, `trip_id`),
KEY `fk_itinerary_trip` (`itinerary_id`, `trip_id`),
CONSTRAINT `FK_84` FOREIGN KEY `fk_itinerary_trip` (`itinerary_id`, `trip_id`) REFERENCES `trip_itinerary` (`id`, `trip_id`) ON DELETE CASCADE
);

