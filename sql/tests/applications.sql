DROP FUNCTION IF EXISTS t.test_applications;
CREATE OR REPLACE FUNCTION t.test_applications()
RETURNS SETOF TEXT AS $$ DECLARE
    var_application_form_id INT;
    var_application_id INT;
BEGIN
    -- Get an form_id for an application form
    SELECT form_id
    INTO var_application_form_id
    FROM forms AS f JOIN periods AS p ON p.period_id = f.period_id
    WHERE p.stage = 'application'
    LIMIT 1
    ;

    RETURN NEXT ok(
        (SELECT var_application_form_id IS NOT NULL)
        ,'there should be at least one application form created before running this test'
    );

    INSERT INTO applications (application_form_id)
    VALUES (var_application_form_id)
    RETURNING application_id INTO var_application_id
    ;
    RETURN NEXT is(
        (SELECT project_level FROM applications WHERE application_id = var_application_id)
        ,'gemini'
        ,'project_level column defaults to gemini when application_data is not provided'
    );

    INSERT INTO applications (application_form_id, application_data)
    VALUES (var_application_form_id, '{"project_level":["artemis"]}')
    RETURNING application_id INTO var_application_id
    ;
    RETURN NEXT is(
        (SELECT project_level FROM applications WHERE application_id = var_application_id)
        ,'artemis'
        ,'On INSERT, application_data.project_level[0] gets copied to project_level column'
    );

    UPDATE applications
    SET application_data = jsonb_set(application_data, '{project_level}', '["vostok"]')
    WHERE application_id = var_application_id
    ;
    RETURN NEXT is(
        (SELECT project_level FROM applications WHERE application_id = var_application_id)
        ,'vostok'
        ,'On UPDATE, application_data.project_level[0] gets copied to project_level column'
    );
END $$ LANGUAGE plpgsql;
