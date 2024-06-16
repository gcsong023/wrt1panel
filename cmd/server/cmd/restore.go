package cmd

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	cmdUtils "github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(restoreCmd)
}

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "回滚 1Panel 服务及数据",
	RunE:  performRollback,
}

func performRollback(cmd *cobra.Command, args []string) error {
	if !isRoot() {
		fmt.Println("请使用 sudo 1pctl restore 或者切换到 root 用户")
		return nil
	}

	baseDir, err := getBaseDir()
	if err != nil {
		return err
	}

	upgradeDir := path.Join(baseDir, "1panel", "tmp", "upgrade")
	tmpPath, err := loadRestorePath(upgradeDir)
	if err != nil {
		return err
	}
	if tmpPath == "暂无可回滚文件" {
		fmt.Println("暂无可回滚文件")
		return nil
	}

	binDir := "/usr/local/bin"
	if err := ensureDir(binDir); err != nil {
		return err
	}

	// var serviceDir string
	var serviceTarget string
	serviceTarget, err = ensureServiceDir()
	if err != nil {
		return err
	}

	tmpPath = path.Join(upgradeDir, tmpPath, "original")
	fmt.Printf("(0/4) 开始从 %s 目录回滚 1Panel 服务及数据... \n", tmpPath)

	checkPointOfWal()
	if err := restoreFiles(tmpPath, binDir, serviceTarget, baseDir); err != nil {
		return err
	}

	fmt.Println("回滚成功！正在重启服务，请稍候...")
	return nil
}

func getBaseDir() (string, error) {
	stdout, err := cmdUtils.Exec("grep '^BASE_DIR=' /usr/local/bin/1pctl | cut -d'=' -f2")
	if err != nil {
		return "", fmt.Errorf("handle load `BASE_DIR` failed, err: %v", err)
	}
	return strings.TrimSpace(stdout), nil
}

func ensureDir(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("创建目录 %s 失败: %v", dirPath, err)
		}
	}
	return nil
}

func ensureServiceDir() (string, error) {
	// 确保服务目录存在并选择正确的目录
	serviceDir := "/etc/systemd/system"
	initdDir := "/etc/init.d/"
	serviceTarget := path.Join(serviceDir, "1panel.service")

	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		serviceDir = initdDir
		serviceTarget = path.Join(serviceDir, "1paneld")
	}

	return serviceTarget, nil
}

func restoreFiles(tmpPath, binDir, serviceTarget, baseDir string) error {
	var serviceFileName = "1paneld"
	// 检查1paneld是否存在，若不存在则尝试1panel.service
	if _, err := os.Stat(path.Join(tmpPath, serviceFileName)); os.IsNotExist(err) {
		serviceFileName = "1panel.service"
		if _, err := os.Stat(path.Join(tmpPath, serviceFileName)); os.IsNotExist(err) {
			return fmt.Errorf("服务文件 %s 或 %s 未找到", "1panel.service", "1paneld")
		}
	}
	filesToRestore := []struct {
		source string
		dest   string
	}{
		{path.Join(tmpPath, "1panel"), binDir},
		{path.Join(tmpPath, "1pctl"), binDir},
		{path.Join(tmpPath, serviceFileName), serviceTarget},
		{path.Join(tmpPath, "1Panel.db"), path.Join(baseDir, "1panel/db")},
		{path.Join(tmpPath, "db.tar.gz"), path.Join(baseDir, "1panel")},
	}

	for i, file := range filesToRestore {
		if err := common.CopyFile(file.source, file.dest); err != nil {
			return err
		}
		fmt.Printf("步骤 %d/%d: %s 已成功回滚\n", i+1, len(filesToRestore), file.dest)
	}

	return nil
}
func checkPointOfWal() {
	db, err := loadDBConn()
	if err != nil {
		return
	}
	_ = db.Exec("PRAGMA wal_checkpoint(TRUNCATE);").Error
}

func loadRestorePath(upgradeDir string) (string, error) {
	if _, err := os.Stat(upgradeDir); err != nil && os.IsNotExist(err) {
		return "暂无可回滚文件", nil
	}
	files, err := os.ReadDir(upgradeDir)
	if err != nil {
		return "", err
	}
	var folders []string
	for _, file := range files {
		if file.IsDir() {
			folders = append(folders, file.Name())
		}
	}
	if len(folders) == 0 {
		return "暂无可回滚文件", nil
	}
	sort.Slice(folders, func(i, j int) bool {
		return folders[i] > folders[j]
	})
	return folders[0], nil
}

func handleUnTar(sourceFile, targetDir string) error {
	if _, err := os.Stat(targetDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return err
		}
	}

	commands := fmt.Sprintf("tar zxvf %s -C %s", sourceFile, targetDir)
	stdout, err := cmdUtils.ExecWithTimeOut(commands, 20*time.Second)
	if err != nil {
		return errors.New(stdout)
	}
	return nil
}
