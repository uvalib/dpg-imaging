START TRANSACTION;

CREATE TABLE `categories` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  `projects_count` int DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `problems` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  `label` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE `workflows` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  `description` text,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `active` tinyint(1) DEFAULT '1',
  PRIMARY KEY (`id`)
)  ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `steps` (
  `id` int NOT NULL AUTO_INCREMENT,
  `step_type` int DEFAULT '3',
  `name` varchar(255) DEFAULT NULL,
  `description` text,
  `workflow_id` int DEFAULT NULL,
  `next_step_id` int DEFAULT NULL,
  `fail_step_id` int DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `owner_type` int DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `index_steps_on_workflow_id` (`workflow_id`),
  KEY `index_steps_on_next_step_id` (`next_step_id`),
  KEY `index_steps_on_fail_step_id` (`fail_step_id`),
  CONSTRAINT `steps_workflow_id_fk` FOREIGN KEY (`workflow_id`) REFERENCES `workflows` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `workstations` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `status` int DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `projects` (
  `id` int NOT NULL AUTO_INCREMENT,
  `workflow_id` int DEFAULT NULL,
  `unit_id` int DEFAULT NULL,
  `owner_id` int DEFAULT NULL,
  `current_step_id` int DEFAULT NULL,
  `item_condition` int DEFAULT NULL,
  `added_at` datetime DEFAULT NULL,
  `started_at` datetime DEFAULT NULL,
  `finished_at` datetime DEFAULT NULL,
  `category_id` int DEFAULT NULL,
  `capture_resolution` int DEFAULT NULL,
  `resized_resolution` int DEFAULT NULL,
  `resolution_note` varchar(255) DEFAULT NULL,
  `workstation_id` int DEFAULT NULL,
  `condition_note` text,
  `container_type_id` bigint DEFAULT NULL,
  `total_duration_mins` int DEFAULT NULL,
  `image_count` int DEFAULT '0',
  `order_id` bigint DEFAULT NULL,
  `customer_id` bigint DEFAULT NULL,
  `agency_id` bigint DEFAULT NULL,
  `call_number` varchar(255) DEFAULT NULL,
  `title` text,
  `date_due` date DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `index_projects_on_workflow_id` (`workflow_id`),
  KEY `index_projects_on_unit_id` (`unit_id`),
  KEY `index_projects_on_category_id` (`category_id`),
  KEY `index_projects_on_workstation_id` (`workstation_id`),
  KEY `index_projects_on_container_type_id` (`container_type_id`),
  CONSTRAINT `projects_category_id_fk` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`),
  CONSTRAINT `projects_current_step_id_fk` FOREIGN KEY (`current_step_id`) REFERENCES `steps` (`id`),
  CONSTRAINT `projects_workflow_id_fk` FOREIGN KEY (`workflow_id`) REFERENCES `workflows` (`id`),
  CONSTRAINT `projects_workstation_id_fk` FOREIGN KEY (`workstation_id`) REFERENCES `workstations` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `assignments` (
  `id` int NOT NULL AUTO_INCREMENT,
  `project_id` int DEFAULT NULL,
  `step_id` int DEFAULT NULL,
  `staff_member_id` int DEFAULT NULL,
  `assigned_at` datetime DEFAULT NULL,
  `started_at` datetime DEFAULT NULL,
  `finished_at` datetime DEFAULT NULL,
  `status` int DEFAULT '0',
  `duration_minutes` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `index_assignments_on_step_id` (`step_id`),
  KEY `index_assignments_on_staff_member_id` (`staff_member_id`),
  KEY `index_assignments_on_project_id` (`project_id`),
  CONSTRAINT `assigmnents_project_id_fk` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`),
  CONSTRAINT `assigmnents_step_id_fk` FOREIGN KEY (`step_id`) REFERENCES `steps` (`id`)
)  ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `equipment` (
  `id` int NOT NULL AUTO_INCREMENT,
  `type` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `serial_number` varchar(255) DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `status` int DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `messages` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `subject` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `from_id` bigint DEFAULT NULL,
  `sent_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `index_messages_on_from_id` (`from_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `message_recipients` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `message_id` bigint DEFAULT NULL,
  `staff_id` bigint DEFAULT NULL,
  `read` tinyint(1) DEFAULT '0',
  `deleted` tinyint(1) DEFAULT '0',
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `index_message_recipients_on_staff_id` (`staff_id`),
  KEY `index_message_recipients_on_message_id` (`message_id`),
  CONSTRAINT `message_recipients_message_id_fk` FOREIGN KEY (`message_id`) REFERENCES `messages` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `notes` (
  `id` int NOT NULL AUTO_INCREMENT,
  `staff_member_id` int DEFAULT NULL,
  `project_id` int DEFAULT NULL,
  `note` text,
  `note_type` int DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `step_id` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `index_notes_on_staff_member_id` (`staff_member_id`),
  KEY `index_notes_on_project_id` (`project_id`),
  KEY `index_notes_on_step_id` (`step_id`),
  CONSTRAINT `notes_project_id_fk` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `notes_problems` (
  `note_id` int DEFAULT NULL,
  `problem_id` int DEFAULT NULL,
  KEY `index_notes_problems_on_note_id` (`note_id`),
  KEY `index_notes_problems_on_problem_id` (`problem_id`),
  CONSTRAINT `notes_problems_note_id_fk` FOREIGN KEY (`note_id`) REFERENCES `notes` (`id`),
  CONSTRAINT `notes_problems_problem_id_fk` FOREIGN KEY (`problem_id`) REFERENCES `problems` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `project_equipment` (
  `id` int NOT NULL AUTO_INCREMENT,
  `project_id` int DEFAULT NULL,
  `equipment_id` int DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `index_project_equipment_on_equipment_id` (`equipment_id`),
  KEY `index_project_equipment_on_project_id` (`project_id`),
  CONSTRAINT `project_equipment_project_id_fk` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`),
  CONSTRAINT `project_equipment_equipment_id_fk` FOREIGN KEY (`equipment_id`) REFERENCES `equipment` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE `workstation_equipment` (
  `id` int NOT NULL AUTO_INCREMENT,
  `workstation_id` int DEFAULT NULL,
  `equipment_id` int DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `index_workstation_equipment_on_workstation_id` (`workstation_id`),
  KEY `index_workstation_equipment_on_equipment_id` (`equipment_id`),
  CONSTRAINT `workstation_equipment_equipment_id_fk` FOREIGN KEY (`equipment_id`) REFERENCES `equipment` (`id`),
  CONSTRAINT `workstation_equipment_workstation_id_fk` FOREIGN KEY (`workstation_id`) REFERENCES `workstations` (`id`)
)  ENGINE=InnoDB DEFAULT CHARSET=utf8;

COMMIT;
