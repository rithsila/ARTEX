package selfupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// smokeEnv 让被冒烟测试拉起的子进程直接跳过 Bootstrap。
//
// 严格来说不加也不会出事：子进程的 os.Executable() 是 artex.new，推导出来的
// 全部路径都带 .new 前缀，碰不到真正的升级文件。但依赖这种巧合太脆弱，
// 显式短路一目了然，也省掉子进程一次无谓的磁盘探测。
const smokeEnv = "ARTEX_SELFUPDATE_SMOKE"

// Action 是 Bootstrap 给 main 的指令。
type Action int

const (
	// Continue：照常启动 server。
	Continue Action = iota
	// Restart：立刻以 ExitRestart 退出，让守护脚本重新拉起。
	Restart
)

// State 描述本次启动时的升级状态，供 /api/update/check 如实告诉前端
// "上一次升级是成功了还是被回滚了"。
type State struct {
	Pending     bool   // 换装后尚未确认稳定
	RolledBack  bool   // 本次启动刚刚执行过自动回滚
	FailedStage bool   // 暂存件校验/冒烟未通过，已丢弃
	Detail      string // 面向用户的一句话说明
}

// Bootstrap 在 main 的最开头运行，必须在任何监听端口、打开数据库之前调用。
//
// 三种局面：
//
//	① 存在暂存件 artex.new  → 校验 + 冒烟，通过则换装并要求重启；不通过则丢弃继续跑旧版
//	② 只剩标记文件          → 说明刚换装完，累计一次尝试；连续失败够多次则回滚
//	③ 什么都没有            → 正常启动
func Bootstrap() (Action, State) {
	if os.Getenv(smokeEnv) != "" {
		return Continue, State{}
	}
	p, err := ResolvePaths()
	if err != nil {
		log.Printf("[update] bootstrap skipped: %v", err)
		return Continue, State{}
	}

	if _, err := os.Stat(p.New); err == nil {
		return applyStaged(p)
	}

	m, ok := readMarker(p.Marker)
	if !ok {
		return Continue, State{}
	}
	return confirmOrRollback(p, m)
}

// applyStaged 处理"存在暂存件"的局面：校验通过就换装，失败就丢弃。
//
// 这里是整个升级链路唯一会覆盖可执行文件的地方，也是最后一道闸门——冒烟测试挡掉
// 下载损坏、架构选错、动态链接缺失这类问题。一旦放行一个跑不起来的二进制，
// 守护脚本会不知疲倦地反复拉起它，而 Go 代码根本没机会运行，自动回滚也就无从谈起。
func applyStaged(p Paths) (Action, State) {
	m, _ := readMarker(p.Marker)

	if err := verifyStaged(p); err != nil {
		log.Printf("[update] staged binary verification failed, discarded, continuing with current version: %v", err)
		cleanStaged(p)
		_ = os.Remove(p.Marker)
		return Continue, State{FailedStage: true, Detail: "New version verification failed, discarded: " + err.Error()}
	}

	if err := swap(p); err != nil {
		log.Printf("[update] binary swap failed, continuing with current version: %v", err)
		cleanStaged(p)
		_ = os.Remove(p.Marker)
		return Continue, State{FailedStage: true, Detail: "Binary swap failed: " + err.Error()}
	}

	// Swap succeeded. Keep marker for next launch to confirm stability.
	m.Attempts = 0
	if m.StagedAt == 0 {
		m.StagedAt = time.Now().Unix()
	}
	if err := writeMarker(p.Marker, m); err != nil {
		log.Printf("[update] write upgrade marker failed (auto-rollback disabled): %v", err)
	}
	log.Printf("[update] binary swapped to %s, exiting to restart (exit %d)", orUnknown(m.To), ExitRestart)
	return Restart, State{Pending: true}
}

// confirmOrRollback handles post-swap launch: tracks attempt counter and rolls back on failure.
func confirmOrRollback(p Paths, m marker) (Action, State) {
	m.Attempts++
	if m.Attempts > maxAttempts {
		if err := rollback(p); err != nil {
			log.Printf("[update] new version failed %d consecutive startups and rollback failed: %v", maxAttempts, err)
			_ = os.Remove(p.Marker)
			return Continue, State{Detail: "New version startup failed and rollback failed: " + err.Error()}
		}
		log.Printf("[update] new version failed %d consecutive startups, rolled back to %s, exiting to restart (exit %d)",
			maxAttempts, orUnknown(m.From), ExitRestart)
		_ = os.Remove(p.Marker)
		return Restart, State{RolledBack: true, Detail: fmt.Sprintf("New version failed to start, rolled back to %s", orUnknown(m.From))}
	}
	if err := writeMarker(p.Marker, m); err != nil {
		log.Printf("[update] update upgrade marker failed: %v", err)
	}
	log.Printf("[update] new version starting (attempt %d/%d), awaiting stability confirmation",
		m.Attempts, maxAttempts)
	return Continue, State{Pending: true}
}

// Settle confirms the new version is stable and clears the upgrade marker.
func Settle() {
	p, err := ResolvePaths()
	if err != nil {
		return
	}
	settle(p)
}

func settle(p Paths) {
	if _, ok := readMarker(p.Marker); !ok {
		return
	}
	if err := os.Remove(p.Marker); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("[update] clear upgrade marker failed: %v", err)
		return
	}
	log.Printf("[update] new version running stable, upgrade completed (previous version retained as %s)", p.Old)
}

// SettleDelay 是判定"新版本活下来了"所需的运行时长。
const SettleDelay = 30 * time.Second

// verifyStaged 校验暂存件：先比对 SHA256，再真正把它拉起来跑一次。
func verifyStaged(p Paths) error {
	want, err := os.ReadFile(p.Sum)
	if err != nil {
		return fmt.Errorf("read checksum: %w", err)
	}
	got, err := fileSHA256(p.New)
	if err != nil {
		return fmt.Errorf("checksum new binary: %w", err)
	}
	if !strings.EqualFold(strings.TrimSpace(string(want)), got) {
		return errors.New("SHA256 mismatch (corrupt download or tampering)")
	}
	return smokeTest(p.New)
}

// smokeTest 用 -h 拉起新二进制，确认它在当前系统上真的能执行。
// 这能挡掉下载截断、架构选错（exec format error）、缺依赖等一大类问题。
func smokeTest(bin string) error {
	if err := os.Chmod(bin, 0o755); err != nil {
		return fmt.Errorf("make new binary executable: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, "-h")
	cmd.Env = append(os.Environ(), smokeEnv+"=1")
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return errors.New("smoke test timed out (new binary unresponsive)")
	}
	if err != nil {
		snippet := strings.TrimSpace(string(out))
		if len(snippet) > 300 {
			snippet = snippet[:300] + "…"
		}
		return fmt.Errorf("smoke test failed: %v: %s", err, snippet)
	}
	return nil
}

// swap 把当前二进制换成暂存的新版本。
//
// Unix 和 Windows 都允许 rename 一个正在运行的可执行文件（Windows 禁止的是删除和
// 覆盖，rename 不在其列），所以这里不需要分平台，也不需要先停掉自己。
// 例外：把二进制以单文件 bind-mount 挂进容器（-v /host/artex:/app/artex）时，
// /app/artex 是挂载点 —— rename 必然 EBUSY，且挂载点既不能被 rename 覆盖也不能
// 原地写入，原位换装在这种部署形态下结构上不可行。此时给出可操作的指引而不是
// 裸报错（换目录挂载 / 宿主机替换文件后重启容器 / 走镜像升级）。
func swap(p Paths) error {
	// Windows 的 rename 不会覆盖已存在的目标，上一轮升级留下的 .old 必须先清掉。
	if err := os.Remove(p.Old); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove old backup %s: %w", p.Old, err)
	}
	if err := os.Rename(p.Current, p.Old); err != nil {
		if busyErr(err) {
			return fmt.Errorf(
				"cannot swap in place: %s is a single-file bind mount (device busy); a mount point cannot be renamed or overwritten. "+
					"Fix: 1) replace the mounted file on the host, then docker restart; "+
					"2) mount a directory instead of a single file; 3) image deployments: docker compose pull artex && docker compose up -d artex: %w",
				p.Current, err)
		}
		return fmt.Errorf("backup current binary: %w", err)
	}
	if err := os.Rename(p.New, p.Current); err != nil {
		// 换装失败但当前版本已经被挪走了，必须原样放回去，否则下次启动没有可执行文件。
		if rerr := os.Rename(p.Old, p.Current); rerr != nil {
			return fmt.Errorf("install new binary failed (%v); restoring the current binary also failed: %w", err, rerr)
		}
		return fmt.Errorf("install new binary: %w", err)
	}
	_ = os.Remove(p.Sum)
	return nil
}

// rollback 把 swap 备份的旧版本换回来。
func rollback(p Paths) error {
	if _, err := os.Stat(p.Old); err != nil {
		return fmt.Errorf("no backup to roll back to (%s): %w", p.Old, err)
	}
	// 把起不来的新版挪到 .failed 留作排查，而不是直接删掉。
	failed := p.Current + ".failed"
	_ = os.Remove(failed)
	if err := os.Rename(p.Current, failed); err != nil {
		return fmt.Errorf("move failed binary aside: %w", err)
	}
	if err := os.Rename(p.Old, p.Current); err != nil {
		return fmt.Errorf("restore previous binary: %w", err)
	}
	return nil
}

// Rollback 是 /api/update/rollback 的实现：主动退回上一版本。
// 只做换装，重启同样交给守护脚本（调用方随后以 ExitRestart 退出）。
func Rollback() error {
	p, err := ResolvePaths()
	if err != nil {
		return err
	}
	if _, err := os.Stat(p.Old); err != nil {
		return errors.New("no previous binary to roll back to (" + p.Old + " does not exist)")
	}
	cleanStaged(p)
	if err := smokeTest(p.Old); err != nil {
		return fmt.Errorf("previous binary is not executable; refusing to roll back: %w", err)
	}
	// 交换当前与备份：回滚之后还能再滚回来。
	tmp := p.Current + ".swap"
	_ = os.Remove(tmp)
	if err := os.Rename(p.Current, tmp); err != nil {
		return fmt.Errorf("move current binary aside: %w", err)
	}
	if err := os.Rename(p.Old, p.Current); err != nil {
		_ = os.Rename(tmp, p.Current)
		return fmt.Errorf("install previous binary: %w", err)
	}
	if err := os.Rename(tmp, p.Old); err != nil {
		log.Printf("[update] post-rollback backup cleanup failed: %v", err)
	}
	_ = os.Remove(p.Marker)
	return nil
}

// HasBackup 报告是否存在可回滚的上一版本，供前端决定要不要显示回滚按钮。
func HasBackup() bool {
	p, err := ResolvePaths()
	if err != nil {
		return false
	}
	_, err = os.Stat(p.Old)
	return err == nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func orUnknown(s string) string {
	if strings.TrimSpace(s) == "" {
		return "unknown version"
	}
	return s
}
