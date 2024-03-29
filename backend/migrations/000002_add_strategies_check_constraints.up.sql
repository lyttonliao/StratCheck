ALTER TABLE strategies ADD CONSTRAINT strategies_fields_check CHECK (array_length(fields, 1) > 0);

ALTER TABLE strategies ADD CONSTRAINT strategies_criteria_check CHECK (array_length(criteria, 1) > 0);
