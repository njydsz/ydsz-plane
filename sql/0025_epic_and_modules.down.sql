-- 迁移 0025 回滚：Epic 类型支持 + Module 模块体系

-- 回滚触发器
DROP TRIGGER IF EXISTS trg_modules_updated_at ON modules;
-- 注意：set_updated_at 函数若被其他表共享则不删除

-- 回滚关联表
DROP TABLE IF EXISTS module_issues;

-- 回滚模块表
DROP TABLE IF EXISTS modules;

-- 回滚 issues 类型约束
ALTER TABLE issues DROP CONSTRAINT IF EXISTS issues_type_code_check;
ALTER TABLE issues ADD CONSTRAINT issues_type_code_check
    CHECK (type_code = ANY (ARRAY['requirement'::text, 'task'::text, 'defect'::text]));

COMMENT ON COLUMN issues.type_code IS '工作项类型: requirement(需求) / task(任务) / defect(缺陷)';
