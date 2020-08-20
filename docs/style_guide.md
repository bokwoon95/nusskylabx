- [Go](#go)
- [JavaScript/TypeScript](#javascripttypescript)
- [CSS](#css)
- [SQL](#sql)
- [PL/pgSQL](#plpgsql)

## Go
- Use [gofmt](https://golang.org/cmd/gofmt/) to format your source code.

## JavaScript/TypeScript
- Use [prettier](https://prettier.io/docs/en/cli.html) to format your source code.
    - prettier should automatically pick up the config in `prettier.config.js`.

## HTML
- Manually format it yourself.
- Don't use prettier even though prettier does support HTML. This is because prettier will mangle Go's `{{}}` templating syntax ([or any other templating syntax](https://github.com/prettier/prettier/issues/5581#issuecomment-459417515)).

## CSS
- Use [prettier](https://prettier.io/docs/en/cli.html) to format your CSS scripts.

## SQL
- Avoid CamelCase in naming, use snake\_case. SQL is case insensitive, so when you 'CamelCase' it just ends internally translated to 'camelcase'. For this reason, uppercase letters in names should be avoided entirely.
- Uppercase keywords to visually differentiate it from names
    - You will often encounter SQL strings with no syntax highlighting available (e.g. logs), so this will help distinguishing between keywords and names.
- Include the optional [AS](https://stackoverflow.com/a/4164675) keyword for aliases.
- Use leading commas, not trailing commas.
    - It looks ugly but makes it so much easier to reorganize your columns.
    - It makes it painfully clear whenever you are missing a comma (all new lines start with a comma and they must line up).
```sql
-- Avoid
SELECT
    apple,
    banana,
    coconut,
    durian,
    eggplant
FROM
    table

-- Instead
SELECT
    -- The first column is unlikely to ever change. However, the last column may
    -- constantly keep changing as you add new columns to the select list. Leading
    -- commas are superior when it comes to appending, removing and reorganizing the
    -- columns.
    apple
    ,banana
    ,coconut
    ,durian
    ,eggplant
FROM
    table
```
- Don't bother trying to pretty-indent your SQL queries. It makes it hard to edit.
```sql
-- Avoid
-- This is an old-school SQL formatting style that looks aesthetic but is hard to maintain
SELECT  st.column_name_1, jt.column_name_2,
        sjt.column_name_3
FROM    source_table st
        INNER JOIN join_table jt USING (source_table_id)
        INNER JOIN second_join_table sjt
            ON st.source_table_id = sjt.source_table_id
            AND jt.column_3 = sjt.column_4
WHERE   st.source_table_id = X
        AND measurement_id IN (SELECT  measurement_id
                               FROM    measurements m
                                       INNER JOIN sites s ON m.site_id = st.x
                               WHERE   s.site_id = Y)
        AND jt.column_name_3 = Z

-- Instead
-- Each subcomponent is simply indented one extra level
SELECT
    st.column_name_1
    ,jt.column_name_2
    ,sjt.column_name_3
FROM
    source_table AS st
    INNER JOIN join_table AS jt USING (source_table_id)
    INNER JOIN second_join_table AS sjt
        ON st.source_table_id = sjt.source_table_id
        AND jt.column_3 = sjt.column_4
WHERE
    st.source_table_id = X
    AND measurement_id IN (
        SELECT
            measurement_id
        FROM
            measurements AS m
            INNER JOIN sites AS s ON m.site_id = st.x
        WHERE
            s.site_id = Y
    )
    AND jt.column_name_3 = Z
```
