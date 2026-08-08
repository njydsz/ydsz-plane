-- 工作项数据迁移回滚脚本：删除新表中迁移过来的数据，恢复旧表读写逻辑
-- 仅在新表数据还没有被修改的情况下可以回滚，如果有业务已经通过新表写入了数据，需要手动确认后再回滚

-- 1. 清空关联关系迁移数据
DELETE FROM biz_entity_relation WHERE id IN (
    SELECT ir.id FROM biz_entity_relation ir
    JOIN issues i1 ON ir.source_id = i1.id
    JOIN issues i2 ON ir.target_id = i2.id
    WHERE ir.source_type = i1.type_code::text AND ir.target_type = i2.type_code::text
);

-- 2. 清空defect表迁移数据
DELETE FROM defect WHERE public_id IN (SELECT public_id FROM issues WHERE type_code = 'defect');

-- 3. 清空requirement表迁移数据
DELETE FROM requirement WHERE public_id IN (SELECT public_id FROM issues WHERE type_code = 'requirement');

-- 4. 清空task表迁移数据
DELETE FROM task WHERE public_id IN (SELECT public_id FROM issues WHERE type_code = 'task');
