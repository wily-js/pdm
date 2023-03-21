-- 创建管理员表
DROP TABLE IF EXISTS admins;
CREATE TABLE admins
(
    id         INTEGER PRIMARY KEY AUTO_INCREMENT, -- 自增主键
    created_at DATETIME,                           -- 创建时间
    updated_at DATETIME,                           -- 更新时间
    username   VARCHAR(128),                       -- 用户名
    password   VARCHAR(512),                       -- 口令加盐Hash结果 16进制字符串
    salt       VARCHAR(512),                       -- 盐值 16进制字符串
    role       TINYINT,                            -- 角色类型
    cert       TEXT                                -- 证书
);
-- 创建admin
INSERT INTO `admins`
VALUES (1, '2022-11-07 09:19:44', '2022-11-07 03:25:36', 'admin',
        'ba182cee746bc776a9bec5c73293dc730d517acf4a5f9c88213184739ef54693', 'a79e9fc93a41399c0e2a87971434655f',
        0,NULL);

INSERT INTO `admins`
VALUES (2, '2022-11-07 09:19:44', '2022-11-07 10:25:36', 'audit',
        '9f1a7062905d2a4e208f92ecb56f967569762bdbd09f7e020414736e79067893', '9946f8047b368c6219ec246e3f4638cb',
        1,NULL);

-- 创建用户表
DROP TABLE IF EXISTS users;
CREATE TABLE users
(
    id          INTEGER PRIMARY KEY AUTO_INCREMENT, -- 自增主键
    created_at  DATETIME,                           -- 创建时间
    updated_at  DATETIME,                           -- 更新时间
    openid     VARCHAR(512),                        -- 工号
    name        VARCHAR(256),                       -- 用户真实姓名
    name_pinyin VARCHAR(32),                        -- 姓名拼音缩写
    password    VARCHAR(512),                       -- 口令加盐Hash结果 16进制字符串
    salt        VARCHAR(512),                       -- 盐值 16进制字符串
    username    VARCHAR(128),                       -- 用户登录时输入的账户名称
    phone       VARCHAR(256),                       -- 手机号
    email       VARCHAR(256),                       -- 邮箱
    sn					VARCHAR(128),												-- 身份证号
    qq_openid   VARCHAR(200),												-- QQ_openID
    wechat_openid VARCHAR(200),											-- 微信openID
    avatar      VARCHAR(512),												-- 头像文件名
    is_delete   TINYINT-- 是否删除 0 - 未删除（默认值） 1 - 删除
);

-- 创建项目表
DROP TABLE IF EXISTS projects;
CREATE TABLE projects
(
    id            INTEGER PRIMARY KEY AUTO_INCREMENT,-- 自增主键
    created_at    DATETIME,-- 创建时间
    updated_at    DATETIME,-- 更新时间
    name          VARCHAR(256) NOT NULL,-- 项目名称
    name_pinyin   VARCHAR(32),-- 项目名称拼音缩写
    description   VARCHAR(256),-- 简介
    manager       INTEGER,-- 项目负责人ID
    version       VARCHAR(256),-- 版本号 默认为空表示没有，在发布版本时更新该字段
    is_delete     TINYINT-- 是否删除 0 - 未删除（默认值） 1 - 删除
);


-- 创建项目成员表
DROP TABLE IF EXISTS project_members;
CREATE TABLE project_members
(
    id          INTEGER PRIMARY KEY AUTO_INCREMENT, -- 自增主键
    created_at  DATETIME,                           -- 创建时间
    updated_at  DATETIME,                           -- 更新时间
    role        TINYINT,                            -- 角色 角色类型包括：0 - 开发者，1 - 对接者，2 - 负责人，3 - 管理员
    project_id  INTEGER,                            -- 项目ID
    user_id     INTEGER                            -- 用户ID
);

-- 创建接口管理表
DROP TABLE IF EXISTS api_management;
CREATE TABLE api_management
(
    id         INTEGER PRIMARY KEY AUTO_INCREMENT,-- 自增主键
    created_at DATETIME,-- 创建时间
    updated_at DATETIME,-- 更新时间
    project_id INTEGER,-- 所属项目ID
    title      VARCHAR(512) NOT NULL,-- 文档名
    filename   VARCHAR(512)                       -- 文件名称
);

-- 创建接口分类表
DROP TABLE IF EXISTS api_categorize;
CREATE TABLE api_categorize
(
    id         INTEGER PRIMARY KEY AUTO_INCREMENT,-- 自增主键
    created_at DATETIME,-- 创建时间
    updated_at DATETIME,-- 更新时间
    parent_id  INTEGER, -- 父分类ID
    name       VARCHAR(512) NOT NULL, -- 分类名称
    project_id INTEGER,-- 所属项目ID
    user_id    INTEGER-- 创建人ID
);

-- 创建接口用例表
DROP TABLE IF EXISTS api_cases;
CREATE TABLE api_cases
(
    id         INTEGER PRIMARY KEY AUTO_INCREMENT,-- 自增主键
    created_at DATETIME,-- 创建时间
    updated_at DATETIME,-- 更新时间
    name       VARCHAR(512) NOT NULL, -- 接口名称
    user_id  INTEGER, -- 创建人ID
    categorize_id  INTEGER, -- 所属分类ID
    description TEXT, -- 接口描述
    method INTEGER ,-- 请求方法 0-GET，1-POST，2-PUT，3-DELETE
    path VARCHAR(512),-- 请求路径
    params TEXT,-- 请求参数
    headers TEXT,-- 请求头
    body_type INTEGER, -- 请求体类型  0-none，1-json，2-form，3-binary
    body TEXT -- 请求体
);

-- 创建对接文档表
DROP TABLE IF EXISTS docking_documents;
CREATE TABLE docking_documents
(
    id         INTEGER PRIMARY KEY AUTO_INCREMENT, -- 自增主键
    created_at DATETIME,                           -- 创建时间
    updated_at DATETIME,                           -- 更新时间
    name    VARCHAR(256) NOT NULL,              	 -- 对接文档名
    user_id    INTEGER,                            -- 发布者ID
    project_id INTEGER,                            -- 项目ID
    content    TEXT,                               -- 对接文档描述
    assets     TEXT                                -- 附件列表,"JSON MAP:- key 文件名称- value 文件下载URL附件删除或添加时更新该字段。"

);

-- 创建技术方案表
DROP TABLE IF EXISTS technical_proposal;
CREATE TABLE technical_proposal
(
    id         INTEGER PRIMARY KEY AUTO_INCREMENT, -- 自增主键
    created_at DATETIME,                           -- 创建时间
    updated_at DATETIME,                           -- 更新时间
    name       VARCHAR(512) NOT NULL              -- 技术方案名称
);

-- 创建日志表
DROP TABLE IF EXISTS logs;
CREATE TABLE logs
(
    id         INTEGER PRIMARY KEY AUTO_INCREMENT, -- 自增主键
    created_at DATETIME,                           -- 创建时间
    op_type    TINYINT,                            -- 操作者类型 类型如下包括：0 - 匿名，1 - 管理员，2 - 用户 若不知道用户或没有用户信息，则使用匿名。
    op_id      INTEGER,                            -- 操作者记录ID 0 表示匿名
    op_name    VARCHAR(512) NOT NULL,              -- 操作名称
    op_param   TEXT NULL                           -- 操作的关键参数 可选参数，例如删除用户时，删除的用户ID，复杂参数请使用JSON对象字符串，如{id: 1}
);


-- 创建版本号表
DROP TABLE IF EXISTS configs;
CREATE TABLE configs
(
    id        INTEGER PRIMARY KEY AUTO_INCREMENT, -- 自增主键
    item_name VARCHAR(256),
    content   VARCHAR(256)                        -- 版本号时间
);

-- 创建版本号记录
-- INSERT INTO configs(item_name, content)
-- VALUES ("db_version", "2023010501");