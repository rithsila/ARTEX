package db

import (
	"database/sql"
	"errors"
	"fmt"
)

// This file carries the English copy for the built-in intercept rules. The
// seeds in db.go use these strings for a fresh install; RefreshBuiltinInterceptCopy
// rewrites rows that an older install already seeded in Chinese. A rule the user
// edited is left alone: a field is only rewritten while it still contains CJK.

type interceptCopy struct {
	oldName string
	newName string
	oldMsg  string
	newMsg  string
}

// builtinInterceptCopy maps the original Chinese seed name to the English one.
// The old message is repeated here so the refresh can recognise an untouched row.
var builtinInterceptCopy = []interceptCopy{
	{
		oldName: "[内置] 递归强制删除 rm -rf",
		newName: "[builtin] recursive force delete rm -rf",
		oldMsg:  "禁止执行递归强制删除（rm -rf / rm --recursive），可能永久损坏系统或靶机环境",
		newMsg:  "Blocked: recursive force delete (rm -rf / rm --recursive) can permanently destroy the system or target environment",
	},
	{
		oldName: "[内置] 删除系统关键目录",
		newName: "[builtin] delete critical system directory",
		oldMsg:  "禁止删除系统关键路径",
		newMsg:  "Blocked: deleting critical system paths",
	},
	{
		oldName: "[内置] 磁盘格式化 mkfs",
		newName: "[builtin] disk format mkfs",
		oldMsg:  "禁止格式化磁盘（mkfs）",
		newMsg:  "Blocked: disk format (mkfs)",
	},
	{
		oldName: "[内置] 覆写磁盘设备 dd",
		newName: "[builtin] overwrite disk device dd",
		oldMsg:  "禁止使用 dd 覆写磁盘设备",
		newMsg:  "Blocked: dd overwrite of a disk device",
	},
	{
		oldName: "[内置] Fork 炸弹",
		newName: "[builtin] fork bomb",
		oldMsg:  "禁止执行 Fork 炸弹",
		newMsg:  "Blocked: fork bomb",
	},
	{
		oldName: "[内置] 关机 / 重启",
		newName: "[builtin] shutdown / reboot",
		oldMsg:  "禁止执行关机或重启命令",
		newMsg:  "Blocked: shutdown or reboot command",
	},
	{
		oldName: "[内置] 杀死全部进程",
		newName: "[builtin] kill all processes",
		oldMsg:  "禁止 kill -9 -1 或 killall -9（杀死所有进程）",
		newMsg:  "Blocked: kill -9 -1 or killall -9 (kills every process)",
	},
	{
		oldName: "[内置] 磁盘擦除 shred / wipe",
		newName: "[builtin] disk erase shred / wipe",
		oldMsg:  "禁止对磁盘设备执行 shred/wipe 擦除",
		newMsg:  "Blocked: shred/wipe on a disk device",
	},
	{
		oldName: "[内置] 清空防火墙规则",
		newName: "[builtin] flush firewall rules",
		oldMsg:  "禁止清空防火墙规则（iptables -F / nft flush）",
		newMsg:  "Blocked: flushing firewall rules (iptables -F / nft flush)",
	},
	{
		oldName: "[内置] SQL DROP DATABASE / TABLE / SCHEMA",
		newName: "[builtin] SQL DROP DATABASE / TABLE / SCHEMA",
		oldMsg:  "禁止执行 DROP 操作，可能不可逆地销毁数据库对象",
		newMsg:  "Blocked: DROP destroys database objects, often irreversibly",
	},
	{
		oldName: "[内置] SQL TRUNCATE",
		newName: "[builtin] SQL TRUNCATE",
		oldMsg:  "禁止执行 TRUNCATE，可能清空数据表所有数据",
		newMsg:  "Blocked: TRUNCATE empties every row of the table",
	},
	{
		oldName: "[内置] MongoDB drop / dropDatabase",
		newName: "[builtin] MongoDB drop / dropDatabase",
		oldMsg:  "禁止执行 MongoDB drop 操作",
		newMsg:  "Blocked: MongoDB drop operation",
	},
	{
		oldName: "[内置] Redis FLUSHALL / FLUSHDB",
		newName: "[builtin] Redis FLUSHALL / FLUSHDB",
		oldMsg:  "禁止执行 Redis FLUSHALL / FLUSHDB，可能清空全部缓存数据",
		newMsg:  "Blocked: Redis FLUSHALL / FLUSHDB clears all cached data",
	},
	{
		oldName: "[内置] curl / wget 发送 DELETE 请求",
		newName: "[builtin] curl / wget HTTP DELETE",
		oldMsg:  "禁止通过 curl/wget 发送 HTTP DELETE 请求，可能删除目标系统数据",
		newMsg:  "Blocked: HTTP DELETE via curl/wget can delete target-system data",
	},
	{
		oldName: "[内置] Python HTTP 客户端 DELETE（requests/httpx/aiohttp）",
		newName: "[builtin] Python HTTP client DELETE (requests/httpx/aiohttp)",
		oldMsg:  "禁止使用 Python HTTP 客户端发送 DELETE 请求",
		newMsg:  "Blocked: HTTP DELETE from a Python HTTP client",
	},
	{
		oldName: "[内置] 脚本中声明 HTTP DELETE 方法（JS/通用）",
		newName: "[builtin] script declares HTTP DELETE method (JS/generic)",
		oldMsg:  "禁止在脚本中声明并发送 HTTP DELETE 请求",
		newMsg:  "Blocked: script declares and sends an HTTP DELETE request",
	},
	{
		oldName: "[内置] 批量清空 / 清除接口路径",
		newName: "[builtin] bulk clear / purge endpoint path",
		oldMsg:  "禁止调用批量清空或销毁类接口（/clear /wipe /flush /purge 等）",
		newMsg:  "Blocked: bulk clear or destroy endpoint (/clear /wipe /flush /purge and similar)",
	},
	{
		oldName: "[内置] 破坏性系统命令",
		newName: "[builtin] destructive system command",
		oldMsg:  "破坏性命令被拒绝（rm -rf / / mkfs / dd / fork bomb / 关机重启 / 覆写磁盘设备）",
		newMsg:  "Destructive command denied (rm -rf / / mkfs / dd / fork bomb / shutdown / reboot / disk overwrite)",
	},
	{
		oldName: "[内置] 数据外泄管道",
		newName: "[builtin] data exfiltration pipe",
		oldMsg:  "疑似数据外泄管道被拒绝（命令输出经 curl/wget/nc 外传）",
		newMsg:  "Suspected data exfiltration pipe denied (command output piped out via curl/wget/nc)",
	},
	{
		oldName: "[内置] 删除类接口路径",
		newName: "[builtin] delete endpoint path",
		oldMsg:  "禁止调用删除类接口（/delete /remove /unlink /erase 等），不论使用哪种 HTTP 方法——多数应用的删除接口用 GET/POST 就能触发，同样会真实删除目标数据",
		newMsg:  "Blocked: delete endpoint paths (/delete /remove /unlink /erase and similar) under any HTTP method — most apps expose deletion over GET/POST, which still deletes target data",
	},
}

// builtinAssetInterceptCopy is the English copy for the built-in asset
// blocklist notes (government / education domains).
var builtinAssetInterceptCopy = []interceptCopy{
	{oldName: "[内置] 政府网站 (.gov)", newName: "[builtin] Government site (.gov)"},
	{oldName: "[内置] 政府网站 (.gov.cn)", newName: "[builtin] Government site (.gov.cn)"},
	{oldName: "[内置] 教育网站 (.edu)", newName: "[builtin] Education site (.edu)"},
	{oldName: "[内置] 教育网站 (.edu.cn)", newName: "[builtin] Education site (.edu.cn)"},
}

// RefreshBuiltinInterceptCopy rewrites the Chinese name/message of the built-in
// intercept rules to English. Only rows still carrying CJK are touched, so a user
// edit survives. Runs once, gated by a settings flag.
func (d *DB) RefreshBuiltinInterceptCopy() error {
	const flag = "intercept_copy_en_v1"
	if v, _, _ := d.GetSetting(flag); v == "done" {
		return nil
	}

	for _, c := range builtinInterceptCopy {
		name, message, err := d.interceptRuleText(c.oldName)
		if err != nil {
			return fmt.Errorf("read intercept rule %q: %w", c.oldName, err)
		}
		if name == "" {
			continue // rule absent or already renamed
		}
		newName, newMsg := name, message
		if name == c.oldName {
			newName = c.newName
		}
		if newMsg == c.oldMsg {
			newMsg = c.newMsg
		}
		if _, err := d.Exec(`UPDATE intercept_rules SET name=$1, message=$2 WHERE name=$3`,
			newName, newMsg, c.oldName); err != nil {
			return fmt.Errorf("update intercept rule %q: %w", c.oldName, err)
		}
	}

	for _, c := range builtinAssetInterceptCopy {
		var note string
		err := d.QueryRow(`SELECT note FROM asset_intercept_rules WHERE note=$1`, c.oldName).Scan(&note)
		if errors.Is(err, sql.ErrNoRows) {
			continue // absent or already renamed
		}
		if err != nil {
			return fmt.Errorf("read asset rule note %q: %w", c.oldName, err)
		}
		if _, err := d.Exec(`UPDATE asset_intercept_rules SET note=$1 WHERE note=$2`, c.newName, c.oldName); err != nil {
			return fmt.Errorf("update asset rule note %q: %w", c.oldName, err)
		}
	}

	return d.SetSetting(flag, "done")
}

// interceptRuleText returns the name and message of one rule, or empty strings
// when the rule no longer exists.
func (d *DB) interceptRuleText(name string) (string, string, error) {
	var n, m string
	err := d.QueryRow(`SELECT name, message FROM intercept_rules WHERE name=$1`, name).Scan(&n, &m)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	if err != nil {
		return "", "", err
	}
	return n, m, nil
}
