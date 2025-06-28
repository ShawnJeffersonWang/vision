-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS agricultural_vision;

-- 使用数据库
USE agricultural_vision;

-- 创建允许从任何主机连接的root用户
CREATE USER IF NOT EXISTS 'root'@'%' IDENTIFIED BY '325523';

-- 授予所有权限
GRANT ALL PRIVILEGES ON *.* TO 'root'@'%' WITH GRANT OPTION;

-- 刷新权限
FLUSH PRIVILEGES;

-- 你的其他表结构和数据...