-- =====================================================
-- Astraios 用户服务数据库表结构
-- 数据库类型: MySQL 8.0+
-- 字符集: utf8mb4
-- 排序规则: utf8mb4_unicode_ci
-- =====================================================

-- 创建数据库（如不存在）
CREATE DATABASE IF NOT EXISTS `astraios_user` 
    DEFAULT CHARACTER SET utf8mb4 
    DEFAULT COLLATE utf8mb4_unicode_ci;

USE `astraios_user`;

-- =====================================================
-- 1. 用户主表 (t_user)
-- 说明: 存储用户核心认证信息，字段精简，便于高频查询
-- =====================================================
-- DROP TABLE IF EXISTS `t_user`;
CREATE TABLE `t_user` (
    `id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID（雪花算法生成）',
    `username` VARCHAR(50) NOT NULL COMMENT '用户名（唯一标识，用于登录）',
    `password` VARCHAR(255) NOT NULL COMMENT '密码（BCrypt加密存储）',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '账户状态：0-禁用，1-正常，2-待激活，3-已注销',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间（软删除）',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户主表';

-- =====================================================
-- 2. 用户资料表 (t_user_profile)
-- 说明: 存储用户个人详细资料，与主表 1:1 关系
-- =====================================================
-- DROP TABLE IF EXISTS `t_user_profile`;
CREATE TABLE `t_user_profile` (
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `nickname` VARCHAR(50) DEFAULT NULL COMMENT '昵称',
    `avatar` VARCHAR(500) DEFAULT NULL COMMENT '头像URL',
    `gender` TINYINT UNSIGNED DEFAULT 0 COMMENT '性别：0-未知，1-男，2-女',
    `birthday` DATE DEFAULT NULL COMMENT '生日',
    `bio` VARCHAR(500) DEFAULT NULL COMMENT '个人简介/签名',
    `background_image` VARCHAR(500) DEFAULT NULL COMMENT '个人主页背景图',
    `country` VARCHAR(50) DEFAULT NULL COMMENT '国家',
    `province` VARCHAR(50) DEFAULT NULL COMMENT '省份',
    `city` VARCHAR(50) DEFAULT NULL COMMENT '城市',
    `school` VARCHAR(100) DEFAULT NULL COMMENT '学校',
    `major` VARCHAR(100) DEFAULT NULL COMMENT '专业',
    `graduation_year` SMALLINT UNSIGNED DEFAULT NULL COMMENT '毕业年份',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`user_id`),
    KEY `idx_nickname` (`nickname`),
    KEY `idx_city` (`city`),
    KEY `idx_school` (`school`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户资料表';

-- =====================================================
-- 3. 用户认证方式表 (t_user_auth)
-- 说明: 支持多种认证方式（手机号、邮箱），一个用户可绑定多种
-- =====================================================
-- DROP TABLE IF EXISTS `t_user_auth`;
CREATE TABLE `t_user_auth` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `auth_type` TINYINT UNSIGNED NOT NULL COMMENT '认证类型：1-手机号，2-邮箱',
    `identifier` VARCHAR(100) NOT NULL COMMENT '认证标识（手机号/邮箱地址）',
    `verified` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否已验证：0-未验证，1-已验证',
    `verified_at` DATETIME DEFAULT NULL COMMENT '验证时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_auth_type_identifier` (`auth_type`, `identifier`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_identifier` (`identifier`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户认证方式表';

-- =====================================================
-- 4. 第三方登录绑定表 (t_user_oauth)
-- 说明: 存储第三方OAuth登录信息，支持微信、QQ等平台
-- =====================================================
-- DROP TABLE IF EXISTS `t_user_oauth`;
CREATE TABLE `t_user_oauth` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `platform` VARCHAR(20) NOT NULL COMMENT '第三方平台：wechat-微信，qq-QQ，weibo-微博，apple-苹果，google-谷歌',
    `openid` VARCHAR(100) NOT NULL COMMENT '第三方平台用户唯一标识',
    `unionid` VARCHAR(100) DEFAULT NULL COMMENT '第三方平台统一ID（如微信unionid）',
    `nickname` VARCHAR(100) DEFAULT NULL COMMENT '第三方平台昵称',
    `avatar` VARCHAR(500) DEFAULT NULL COMMENT '第三方平台头像',
    `access_token` VARCHAR(500) DEFAULT NULL COMMENT '访问令牌',
    `refresh_token` VARCHAR(500) DEFAULT NULL COMMENT '刷新令牌',
    `token_expires_at` DATETIME DEFAULT NULL COMMENT '令牌过期时间',
    `raw_data` JSON DEFAULT NULL COMMENT '原始返回数据（JSON格式）',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '首次绑定时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_platform_openid` (`platform`, `openid`),
    UNIQUE KEY `uk_platform_unionid` (`platform`, `unionid`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='第三方登录绑定表';

-- =====================================================
-- 5. 用户统计表 (t_user_stats)
-- 说明: 存储高频更新的统计数据，独立存储提升性能
--       预留评分系统相关字段（虎扑式评分）
-- =====================================================
-- DROP TABLE IF EXISTS `t_user_stats`;
CREATE TABLE `t_user_stats` (
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `following_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关注数',
    `follower_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '粉丝数',
    `likes_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '获赞总数',
    `video_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '发布视频数',
    `favorites_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收藏数',
    -- 评分系统相关字段（虎扑式评分预留）
    `total_score` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '累计获得评分总和',
    `score_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '被评分次数',
    `avg_score` DECIMAL(3,1) NOT NULL DEFAULT 0.0 COMMENT '平均评分（0-5分，保留1位小数）',
    -- 互动统计
    `comment_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '发布评论数',
    `share_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '分享次数',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`user_id`),
    KEY `idx_follower_count` (`follower_count`),
    KEY `idx_avg_score` (`avg_score`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户统计表';

-- =====================================================
-- 6. 用户设置表 (t_user_settings)
-- 说明: 存储用户个性化设置，使用JSON字段便于扩展
-- =====================================================
-- DROP TABLE IF EXISTS `t_user_settings`;
CREATE TABLE `t_user_settings` (
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    -- 隐私设置
    `privacy_level` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '隐私等级：0-私密，1-公开，2-仅粉丝可见',
    `allow_stranger_msg` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '允许陌生人私信：0-不允许，1-允许',
    `show_online_status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '显示在线状态：0-隐藏，1-显示',
    `show_likes` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '显示点赞列表：0-隐藏，1-显示',
    `show_following` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '显示关注列表：0-隐藏，1-显示',
    `show_followers` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '显示粉丝列表：0-隐藏，1-显示',
    -- 通知设置（使用JSON便于扩展）
    `notification_settings` JSON DEFAULT NULL COMMENT '通知设置（JSON格式）',
    -- 其他设置
    `language` VARCHAR(10) DEFAULT 'zh-CN' COMMENT '语言偏好',
    `timezone` VARCHAR(50) DEFAULT 'Asia/Shanghai' COMMENT '时区',
    `theme` VARCHAR(20) DEFAULT 'auto' COMMENT '主题：auto-自动，light-浅色，dark-深色',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户设置表';

-- =====================================================
-- 外键约束（可选，根据实际需求决定是否启用）
-- 在微服务架构中，通常不使用外键以提升性能和灵活性
-- 以下为参考，默认注释掉
-- =====================================================
-- ALTER TABLE `t_user_profile` ADD CONSTRAINT `fk_profile_user` 
--     FOREIGN KEY (`user_id`) REFERENCES `t_user`(`id`) ON DELETE CASCADE;
-- ALTER TABLE `t_user_auth` ADD CONSTRAINT `fk_auth_user` 
--     FOREIGN KEY (`user_id`) REFERENCES `t_user`(`id`) ON DELETE CASCADE;
-- ALTER TABLE `t_user_oauth` ADD CONSTRAINT `fk_oauth_user` 
--     FOREIGN KEY (`user_id`) REFERENCES `t_user`(`id`) ON DELETE CASCADE;
-- ALTER TABLE `t_user_stats` ADD CONSTRAINT `fk_stats_user` 
--     FOREIGN KEY (`user_id`) REFERENCES `t_user`(`id`) ON DELETE CASCADE;
-- ALTER TABLE `t_user_settings` ADD CONSTRAINT `fk_settings_user` 
--     FOREIGN KEY (`user_id`) REFERENCES `t_user`(`id`) ON DELETE CASCADE;
