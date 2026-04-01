-- Don't allow rollback if data would be lost
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM assignment
        WHERE LENGTH(unit_name) > 20 
           OR LENGTH(name) > 20 
           OR LENGTH(required_filename) > 20
    ) THEN
        RAISE EXCEPTION 'Cannot rollback: Data exists that exceeds VARCHAR(20) limit';
    END IF;
END $$;

ALTER TABLE assignment
    ALTER COLUMN unit_name TYPE VARCHAR(20),
    ALTER COLUMN name TYPE VARCHAR(20),
    ALTER COLUMN required_filename TYPE VARCHAR(20);
