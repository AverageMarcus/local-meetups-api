SET NAMES utf8mb4;
CREATE TABLE meetup (
	ID varchar(50),
	Title VARCHAR(1000) CHARACTER SET utf8mb4,
	Created VARCHAR(100) CHARACTER SET utf8mb4,
	Updated VARCHAR(100) CHARACTER SET utf8mb4,
	Persisted VARCHAR(100) CHARACTER SET utf8mb4,
	Description TEXT CHARACTER SET utf8mb4,
	URL VARCHAR(1000) CHARACTER SET utf8mb4,
	RsvpCount int,
	RsvpLimit int,
	Time VARCHAR(100) CHARACTER SET utf8mb4,
	Status VARCHAR(100) CHARACTER SET utf8mb4,
  GroupName VARCHAR(500) CHARACTER SET utf8mb4,
  GroupUrlName VARCHAR(1000) CHARACTER SET utf8mb4,
	VenueName VARCHAR(1000) CHARACTER SET utf8mb4,
  VenueAddress TEXT CHARACTER SET utf8mb4,
  VenueCity VARCHAR(500) CHARACTER SET utf8mb4,
  VenueCountry VARCHAR(500) CHARACTER SET utf8mb4,
  PRIMARY KEY(ID)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
CREATE INDEX meetup_time ON meetup (Time);
CREATE INDEX meetup_group_name ON meetup (GroupName);