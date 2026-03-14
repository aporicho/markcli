package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/aporicho/markcli/internal/annotation"
	"github.com/aporicho/markcli/internal/ipc"
	mcpkg "github.com/aporicho/markcli/internal/mcp"
	"github.com/aporicho/markcli/internal/tui"
)

// version is set via ldflags: -X main.version=x.y.z
var version = "dev"

const repo = "aporicho/markcli"

func main() {
	root := &cobra.Command{
		Use:           "mark",
		Short:         "Mark - annotation tool for Claude Code",
		Version:       version,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return runBrowse(cmd)
			}
			if len(args) == 1 {
				return runOpen(cmd, args)
			}
			return cmd.Help()
		},
	}

	openCmd := &cobra.Command{
		Use:          "open <file>",
		Short:        "Open a Markdown file for annotation",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE:         runOpen,
	}

	listCmd := &cobra.Command{
		Use:          "list <file>",
		Short:        "List annotations in JSON format",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE:         runList,
	}

	showCmd := &cobra.Command{
		Use:          "show <file>",
		Short:        "Show annotations in human-readable format",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE:         runShow,
	}

	clearCmd := &cobra.Command{
		Use:          "clear <file>",
		Short:        "Clear all annotations for a file",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE:         runClear,
	}

	mcpCmd := &cobra.Command{
		Use:          "mcp",
		Short:        "Start MCP server (for Claude Code integration)",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE:         runMcp,
	}

	updateCmd := &cobra.Command{
		Use:          "update",
		Short:        "Check for updates and self-update",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE:         runUpdate,
	}

	doctorCmd := &cobra.Command{
		Use:          "doctor",
		Short:        "Check environment configuration",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE:         runDoctor,
	}

	completionCmd := &cobra.Command{
		Use:       "completion [bash|zsh|fish]",
		Short:     "Generate shell completion script",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return root.GenBashCompletion(os.Stdout)
			case "zsh":
				return root.GenZshCompletion(os.Stdout)
			case "fish":
				return root.GenFishCompletion(os.Stdout, true)
			default:
				return fmt.Errorf("不支持的 shell: %s（可选: bash, zsh, fish）", args[0])
			}
		},
	}

	root.AddCommand(openCmd, listCmd, showCmd, clearCmd, mcpCmd, updateCmd, doctorCmd, completionCmd)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runOpen(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Start IPC server
	sock := ipc.SocketPath()
	srv := ipc.NewServer(sock)
	ipcCh, err := srv.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: IPC server failed to start: %v\n", err)
		ipcCh = nil
	} else {
		defer srv.Stop()
	}

	m := tui.New(filePath, ipcCh)
	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err = p.Run()
	tui.ResetPointer()
	return err
}

func runBrowse(cmd *cobra.Command) error {
	// Start IPC server
	sock := ipc.SocketPath()
	srv := ipc.NewServer(sock)
	ipcCh, err := srv.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: IPC server failed to start: %v\n", err)
		ipcCh = nil
	} else {
		defer srv.Stop()
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	m := tui.NewBrowse(dir, ipcCh)
	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err = p.Run()
	tui.ResetPointer()
	return err
}

func runList(cmd *cobra.Command, args []string) error {
	filePath, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}
	af, err := annotation.Load(filePath)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(af, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func runShow(cmd *cobra.Command, args []string) error {
	filePath, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}
	af, err := annotation.Load(filePath)
	if err != nil {
		return err
	}

	fmt.Printf("File: %s\n", filepath.Base(filePath))
	fmt.Println("---")

	if len(af.Annotations) == 0 {
		fmt.Println("（暂无批注）")
		fmt.Println("---")
		return nil
	}

	for _, ann := range af.Annotations {
		var rangeStr string
		if ann.StartLine == ann.EndLine {
			rangeStr = fmt.Sprintf("Line %d", ann.StartLine)
		} else {
			rangeStr = fmt.Sprintf("Line %d-%d", ann.StartLine, ann.EndLine)
		}

		textPreview := ann.SelectedText
		if len([]rune(textPreview)) > 60 {
			textPreview = string([]rune(textPreview)[:60]) + "..."
		}

		fmt.Printf("[%s] \"%s\"\n", rangeStr, textPreview)
		fmt.Printf("批注: %s\n", ann.Comment)
		fmt.Println()
	}

	fmt.Println("---")
	return nil
}

func runClear(cmd *cobra.Command, args []string) error {
	filePath, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}
	if err := annotation.Clear(filePath); err != nil {
		return err
	}
	fmt.Printf("已清除 %s 的所有批注\n", filepath.Base(filePath))
	return nil
}

func runMcp(cmd *cobra.Command, args []string) error {
	ipcClient := ipc.NewClient(ipc.SocketPath())
	mcpServer := mcpkg.NewServer(ipcClient)
	return server.ServeStdio(mcpServer)
}

// assetArch maps Go's GOARCH to the release asset architecture name.
func assetArch() string {
	if runtime.GOARCH == "amd64" {
		return "x64"
	}
	return runtime.GOARCH // arm64 stays arm64
}

func runUpdate(cmd *cobra.Command, args []string) error {
	platform := runtime.GOOS
	arch := assetArch()
	asset := fmt.Sprintf("mark-%s-%s", platform, arch)

	fmt.Printf("当前版本: %s\n", version)
	fmt.Println("检查最新版本...")

	// Query GitHub latest release
	out, err := exec.Command("curl", "-fsSL",
		fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo),
	).Output()
	if err != nil {
		return fmt.Errorf("获取最新版本失败: %w", err)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(out, &release); err != nil {
		return fmt.Errorf("解析版本信息失败: %w", err)
	}

	latest := release.TagName
	latestVersion := strings.TrimPrefix(latest, "v")

	if version == latestVersion {
		fmt.Printf("已是最新版本 (%s)\n", version)
		return nil
	}

	fmt.Printf("发现新版本: %s\n", latest)

	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取二进制路径失败: %w", err)
	}
	binPath, err = filepath.EvalSymlinks(binPath)
	if err != nil {
		return fmt.Errorf("解析二进制路径失败: %w", err)
	}

	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, latest, asset)

	fmt.Println("下载中...")
	dlCmd := exec.Command("curl", "-fsSL", "-o", binPath, url)
	dlCmd.Stdout = os.Stdout
	dlCmd.Stderr = os.Stderr
	if err := dlCmd.Run(); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	os.Chmod(binPath, 0o755)

	// macOS: clear quarantine attributes and re-sign
	if platform == "darwin" {
		exec.Command("xattr", "-dr", "com.apple.quarantine", binPath).Run()
		exec.Command("xattr", "-dr", "com.apple.provenance", binPath).Run()
		exec.Command("codesign", "--force", "--sign", "-", binPath).Run()
	}

	fmt.Printf("已更新到 %s\n", latest)
	return nil
}

func runDoctor(cmd *cobra.Command, args []string) error {
	ok := true

	// Version
	fmt.Printf("mark %s\n\n", version)

	// PATH check
	binPath, _ := os.Executable()
	binDir := filepath.Dir(binPath)
	pathDirs := filepath.SplitList(os.Getenv("PATH"))
	inPath := false
	for _, d := range pathDirs {
		if d == binDir {
			inPath = true
			break
		}
	}
	if inPath {
		fmt.Printf("✅ PATH: %s 已在 PATH 中\n", binDir)
	} else {
		fmt.Printf("❌ PATH: %s 不在 PATH 中\n", binDir)
		ok = false
	}

	// Claude Code check
	hasClaude := false
	if _, err := exec.LookPath("claude"); err == nil {
		hasClaude = true
	}
	if hasClaude {
		fmt.Println("✅ Claude Code: 已安装")
	} else {
		fmt.Println("⚠️  Claude Code: 未安装")
	}

	// MCP configuration check
	if hasClaude {
		mcpOk := checkMcpConfig()
		if mcpOk {
			fmt.Println("✅ MCP 集成: 已配置")
		} else {
			fmt.Println("❌ MCP 集成: 未配置（运行 claude mcp add mark -- mark mcp）")
			ok = false
		}
	}

	// macOS code signing check
	if runtime.GOOS == "darwin" {
		signOk := false
		if err := exec.Command("codesign", "-v", binPath).Run(); err == nil {
			signOk = true
		}
		if signOk {
			fmt.Println("✅ 代码签名: 有效")
		} else {
			fmt.Printf("❌ 代码签名: 无效（运行 codesign --force --sign - %s）\n", binPath)
			ok = false
		}
	}

	fmt.Println()
	if ok {
		fmt.Println("一切正常 👍")
	} else {
		fmt.Println("存在问题，请按提示修复")
		os.Exit(1)
	}
	return nil
}

// checkMcpConfig reads ~/.claude.json and checks if "mark" MCP server is configured.
func checkMcpConfig() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	data, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	if err != nil {
		return false
	}

	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		return false
	}

	// Check top-level mcpServers
	if servers, ok := cfg["mcpServers"].(map[string]any); ok {
		if _, has := servers["mark"]; has {
			return true
		}
	}

	// Check per-project mcpServers
	projects, _ := cfg["projects"].(map[string]any)
	for _, proj := range projects {
		projMap, _ := proj.(map[string]any)
		if servers, ok := projMap["mcpServers"].(map[string]any); ok {
			if _, has := servers["mark"]; has {
				return true
			}
		}
	}

	return false
}
