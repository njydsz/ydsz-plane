-- ============================================================================
-- 脚本名称: check-comments.sql
-- 用途: 检查 ydsz-plane 数据库注释覆盖率，用于 CI 流水线质量门禁
-- 返回: 0=全部通过, 1=存在未注释的表/字段
--
-- 使用方法:
--   psql -U ydsz_app -d ydsz-plane -f check-comments.sql
--   echo $?   -- 0 表示全部通过
--
-- 集成方式 (GitHub Actions / CI):
--   - name: Check DB Comments Coverage
--     run: |
--       psql -v ON_ERROR_STOP=1 -f sql/check-comments.sql
--       if [ $? -ne 0 ]; then echo "❌ 数据库注释覆盖率不达标"; exit 1; fi
-- ============================================================================

\t on
\o /dev/null

DO $$
DECLARE
    v_uncommented_tables  int := 0;
    v_uncommented_columns  int := 0;
    v_uncommented_triggers int := 0;
    v_total_tables         int := 0;
    v_total_columns        int := 0;
    v_total_triggers       int := 0;
    v_rec                  record;
BEGIN
    -- ================================================================
    -- 1. 检查未注释的表（排除系统 schema 和 schema_migrations）
    -- ================================================================
    SELECT count(*) INTO v_total_tables
    FROM information_schema.tables t
    WHERE t.table_schema = 'public'
      AND t.table_type = 'BASE TABLE'
      AND t.table_name NOT IN ('schema_migrations');

    unnest(ARRAY['schema_migrations']) IS NULL;  -- 用于静默

    -- 查找未注释的表
    FOR v_rec IN
        SELECT t.table_name
        FROM information_schema.tables t
        LEFT JOIN pg_catalog.pg_description d
            ON d.objoid = (quote_ident(t.table_schema)||'.'||quote_ident(t.table_name))::regclass::oid
           AND d.objsubid = 0
        WHERE t.table_schema = 'public'
          AND t.table_type = 'BASE TABLE'
          AND t.table_name NOT IN ('schema_migrations')
          AND d.description IS NULL
        ORDER BY t.table_name
    LOOP
        v_uncommented_tables := v_uncommented_tables + 1;
        RAISE NOTICE '❌ 缺表注释: %', v_rec.table_name;
    END LOOP;

    -- ================================================================
    -- 2. 检查未注释的字段
    -- ================================================================
    SELECT count(*) INTO v_total_columns
    FROM information_schema.columns c
    WHERE c.table_schema = 'public'
      AND c.table_name NOT IN ('schema_migrations');

    -- 查找未注释的字段
    FOR v_rec IN
        SELECT c.table_name, c.column_name
        FROM information_schema.columns c
        LEFT JOIN pg_catalog.pg_description d
            ON d.objoid = (quote_ident(c.table_schema)||'.'||quote_ident(c.table_name))::regclass::oid
           AND d.objsubid = c.ordinal_position
        WHERE c.table_schema = 'public'
          AND c.table_name NOT IN ('schema_migrations')
          AND d.description IS NULL
        ORDER BY c.table_name, c.column_name
    LOOP
        v_uncommented_columns := v_uncommented_columns + 1;
        RAISE NOTICE '❌ 缺字段注释: %.%', v_rec.table_name, v_rec.column_name;
    END LOOP;

    -- ================================================================
    -- 3. 检查未注释的触发器
    -- ================================================================
    SELECT count(*) INTO v_total_triggers
    FROM information_schema.triggers t
    WHERE t.trigger_schema = 'public';

    -- 查找未注释的触发器
    FOR v_rec IN
        SELECT t.trigger_name, t.event_object_table
        FROM information_schema.triggers t
        LEFT JOIN pg_catalog.pg_description d
            ON d.objoid = (quote_ident(t.trigger_schema)||'.'||quote_ident(t.event_object_table))::regclass::oid
           AND d.classoid = 'pg_trigger'::regclass::oid
           AND d.objsubid = (
               SELECT tgfoid FROM pg_trigger tg
               WHERE tg.tgname = t.trigger_name
                 AND tg.tgrelid = (quote_ident(t.trigger_schema)||'.'||quote_ident(t.event_object_table))::regclass::oid
               LIMIT 1
           )
        WHERE t.trigger_schema = 'public'
          AND d.description IS NULL
        ORDER BY t.event_object_table, t.trigger_name
    LOOP
        v_uncommented_triggers := v_uncommented_triggers + 1;
        RAISE NOTICE '❌ 缺触发器注释: % (表: %)', v_rec.trigger_name, v_rec.event_object_table;
    END LOOP;

    -- ================================================================
    -- 4. 输出汇总报告
    -- ================================================================
    RAISE NOTICE '============================================';
    RAISE NOTICE '📊 注释覆盖率报告 (ydsz-plane)';
    RAISE NOTICE '============================================';
    RAISE NOTICE '表注释: % / % ({%}%)',
        v_total_tables - v_uncommented_tables,
        v_total_tables,
        round((v_total_tables - v_uncommented_tables)::numeric / nullif(v_total_tables,0) * 100, 1);
    RAISE NOTICE '字段注释: % / % ({%}%)',
        v_total_columns - v_uncommented_columns,
        v_total_columns,
        round((v_total_columns - v_uncommented_columns)::numeric / nullif(v_total_columns,0) * 100, 1);
    RAISE NOTICE '触发器注释: % / % ({%}%)',
        v_total_triggers - v_uncommented_triggers,
        v_total_triggers,
        round((v_total_triggers - v_uncommented_triggers)::numeric / nullif(v_total_triggers,0) * 100, 1);
    RAISE NOTICE '============================================';

    -- ================================================================
    -- 5. 质量门禁：未注释数量 > 0 时抛出异常（CI 退出码非 0）
    -- ================================================================
    IF v_uncommented_tables > 0 OR v_uncommented_columns > 0 OR v_uncommented_triggers > 0 THEN
        RAISE EXCEPTION '发现 % 张表缺注释、% 个字段缺注释、% 个触发器缺注释',
            v_uncommented_tables, v_uncommented_columns, v_uncommented_triggers;
    ELSE
        RAISE NOTICE '✅ 全部通过！注释覆盖率 100%%';
    END IF;
END $$;

\t off
