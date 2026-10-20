ALTER TABLE assignment
    ADD COLUMN category_id INTEGER REFERENCES assignment_category (id) ON DELETE SET NULL;
