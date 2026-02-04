-- =====================================================
-- Astraios 用户服务初始化数据
-- 执行前请先执行 schema.sql 创建表结构
-- =====================================================

USE `astraios_user`;

-- =====================================================
-- 默认通知设置模板（JSON格式参考）
-- =====================================================
-- notification_settings JSON 字段结构说明：
-- {
--     "push": {
--         "enabled": true,           -- 是否开启推送
--         "likes": true,             -- 点赞通知
--         "comments": true,          -- 评论通知
--         "follows": true,           -- 关注通知
--         "mentions": true,          -- @提及通知
--         "messages": true,          -- 私信通知
--         "system": true             -- 系统通知
--     },
--     "email": {
--         "enabled": false,          -- 是否开启邮件通知
--         "weekly_digest": false,    -- 每周摘要
--         "security_alerts": true    -- 安全提醒
--     },
--     "sms": {
--         "enabled": false,          -- 是否开启短信通知
--         "security_alerts": true    -- 安全提醒
--     },
--     "quiet_hours": {
--         "enabled": false,          -- 是否开启免打扰
--         "start": "22:00",          -- 开始时间
--         "end": "08:00"             -- 结束时间
--     }
-- }

-- =====================================================
-- 测试数据（开发环境使用，生产环境请删除）
-- =====================================================

-- 测试用户1：使用雪花算法生成的ID示例
-- 密码为 "password123" 的 BCrypt 加密结果
INSERT INTO `t_user` (`id`, `username`, `password`, `status`) VALUES
(1891234567890123456, 'test_user_01', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z5EHsM8lE9lBOsl7iAt6Z5EH', 1);

-- 测试用户1的资料
INSERT INTO `t_user_profile` (`user_id`, `nickname`, `avatar`, `gender`, `birthday`, `bio`, `country`, `province`, `city`, `school`, `major`) VALUES
(1891234567890123456, '测试用户', 'https://example.com/avatars/default.png', 1, '2000-01-01', '这是一个测试账号', '中国', '广东省', '深圳市', '深圳大学', '计算机科学与技术');

-- 测试用户1的认证方式（手机号）
INSERT INTO `t_user_auth` (`user_id`, `auth_type`, `identifier`, `verified`, `verified_at`) VALUES
(1891234567890123456, 1, '13800138000', 1, NOW());

-- 测试用户1的统计数据
INSERT INTO `t_user_stats` (`user_id`, `following_count`, `follower_count`, `likes_count`, `video_count`) VALUES
(1891234567890123456, 0, 0, 0, 0);

-- 测试用户1的设置（使用默认通知设置）
INSERT INTO `t_user_settings` (`user_id`, `privacy_level`, `allow_stranger_msg`, `show_online_status`, `notification_settings`) VALUES
(1891234567890123456, 1, 1, 1, '{
    "push": {
        "enabled": true,
        "likes": true,
        "comments": true,
        "follows": true,
        "mentions": true,
        "messages": true,
        "system": true
    },
    "email": {
        "enabled": false,
        "weekly_digest": false,
        "security_alerts": true
    },
    "sms": {
        "enabled": false,
        "security_alerts": true
    },
    "quiet_hours": {
        "enabled": false,
        "start": "22:00",
        "end": "08:00"
    }
}');

-- =====================================================
-- 存储过程：创建新用户（事务保证数据一致性）
-- =====================================================
DELIMITER //

DROP PROCEDURE IF EXISTS `sp_create_user`//

CREATE PROCEDURE `sp_create_user`(
    IN p_user_id BIGINT UNSIGNED,
    IN p_username VARCHAR(50),
    IN p_password VARCHAR(255),
    IN p_auth_type TINYINT UNSIGNED,
    IN p_identifier VARCHAR(100)
)
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;
    
    START TRANSACTION;
    
    -- 插入用户主表
    INSERT INTO `t_user` (`id`, `username`, `password`, `status`)
    VALUES (p_user_id, p_username, p_password, 1);
    
    -- 插入用户资料表（使用默认值）
    INSERT INTO `t_user_profile` (`user_id`)
    VALUES (p_user_id);
    
    -- 插入认证方式
    IF p_auth_type IS NOT NULL AND p_identifier IS NOT NULL THEN
        INSERT INTO `t_user_auth` (`user_id`, `auth_type`, `identifier`, `verified`)
        VALUES (p_user_id, p_auth_type, p_identifier, 0);
    END IF;
    
    -- 插入统计表（初始化为0）
    INSERT INTO `t_user_stats` (`user_id`)
    VALUES (p_user_id);
    
    -- 插入设置表（使用默认值）
    INSERT INTO `t_user_settings` (`user_id`, `notification_settings`)
    VALUES (p_user_id, '{
        "push": {"enabled": true, "likes": true, "comments": true, "follows": true, "mentions": true, "messages": true, "system": true},
        "email": {"enabled": false, "weekly_digest": false, "security_alerts": true},
        "sms": {"enabled": false, "security_alerts": true},
        "quiet_hours": {"enabled": false, "start": "22:00", "end": "08:00"}
    }');
    
    COMMIT;
END//

DELIMITER ;

-- =====================================================
-- 存储过程：绑定第三方登录
-- =====================================================
DELIMITER //

DROP PROCEDURE IF EXISTS `sp_bind_oauth`//

CREATE PROCEDURE `sp_bind_oauth`(
    IN p_user_id BIGINT UNSIGNED,
    IN p_platform VARCHAR(20),
    IN p_openid VARCHAR(100),
    IN p_unionid VARCHAR(100),
    IN p_nickname VARCHAR(100),
    IN p_avatar VARCHAR(500)
)
BEGIN
    INSERT INTO `t_user_oauth` (`user_id`, `platform`, `openid`, `unionid`, `nickname`, `avatar`)
    VALUES (p_user_id, p_platform, p_openid, p_unionid, p_nickname, p_avatar)
    ON DUPLICATE KEY UPDATE
        `nickname` = p_nickname,
        `avatar` = p_avatar,
        `updated_at` = NOW();
END//

DELIMITER ;

-- =====================================================
-- 常用查询示例（供开发参考）
-- =====================================================

-- 根据用户名查询用户（登录验证）
-- SELECT u.id, u.username, u.password, u.status 
-- FROM t_user u 
-- WHERE u.username = ? AND u.deleted_at IS NULL;

-- 根据手机号查询用户
-- SELECT u.id, u.username, u.password, u.status 
-- FROM t_user u 
-- INNER JOIN t_user_auth ua ON u.id = ua.user_id 
-- WHERE ua.auth_type = 1 AND ua.identifier = ? AND u.deleted_at IS NULL;

-- 根据微信openid查询用户
-- SELECT u.id, u.username, u.status 
-- FROM t_user u 
-- INNER JOIN t_user_oauth uo ON u.id = uo.user_id 
-- WHERE uo.platform = 'wechat' AND uo.openid = ? AND u.deleted_at IS NULL;

-- 获取用户完整资料
-- SELECT u.id, u.username, u.status, u.created_at,
--        p.nickname, p.avatar, p.gender, p.birthday, p.bio, p.country, p.province, p.city, p.school, p.major,
--        s.following_count, s.follower_count, s.likes_count, s.video_count, s.avg_score
-- FROM t_user u
-- LEFT JOIN t_user_profile p ON u.id = p.user_id
-- LEFT JOIN t_user_stats s ON u.id = s.user_id
-- WHERE u.id = ? AND u.deleted_at IS NULL;
