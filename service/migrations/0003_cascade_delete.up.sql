ALTER TABLE installation_configuration
DROP FOREIGN KEY installation_configuration_ibfk_1;
ALTER TABLE installation_configuration
ADD CONSTRAINT installation_configuration_ibfk_1 
    FOREIGN KEY (installation_id) 
    REFERENCES installations(id) ON DELETE CASCADE;

ALTER TABLE instances
DROP FOREIGN KEY instances_ibfk_1;
ALTER TABLE instances
ADD CONSTRAINT instances_ibfk_1
    FOREIGN KEY (installation_id) 
    REFERENCES installations(id) ON DELETE CASCADE;