SET FOREIGN_KEY_CHECKS=0;

CREATE DATABASE IF NOT EXISTS `exchange_rate`;
USE `exchange_rate`;

-- ----------------------------
-- Table structure for rate
-- ----------------------------
DROP TABLE IF EXISTS `rate`;
CREATE TABLE `rate` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `from` varchar(3) DEFAULT NULL,
  `to` varchar(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_from_to` (`from`,`to`)
) ENGINE=InnoDB AUTO_INCREMENT=27 DEFAULT CHARSET=latin1;

-- ----------------------------
-- Table structure for rate_data
-- ----------------------------
DROP TABLE IF EXISTS `rate_data`;
CREATE TABLE `rate_data` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `rate_id` int(11) NOT NULL,
  `date` date DEFAULT NULL,
  `rate` decimal(20,5) DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_rate_id_and_date` (`rate_id`,`date`),
  CONSTRAINT `rate_data_ibfk_1` FOREIGN KEY (`rate_id`) REFERENCES `rate` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
