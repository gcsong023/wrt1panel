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

// var restoreCmd = &cobra.Command{
// 	Use:   "restore",
// 	Short: "回滚 1Panel 服务及数据",
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		if !isRoot() {
// 			fmt.Println("请使用 sudo 1pctl restore 或者切换到 root 用户")
// 			return nil
// 		}
// 		stdout, err := cmdUtils.Exec("grep '^BASE_DIR=' /usr/local/bin/1pctl | cut -d'=' -f2")
// 		if err != nil {
// 			return fmt.Errorf("handle load `BASE_DIR` failed, err: %v", err)
// 		}
// 		baseDir := strings.ReplaceAll(stdout, "\n", "")
// 		upgradeDir := path.Join(baseDir, "1panel", "tmp", "upgrade")

// 		tmpPath, err := loadRestorePath(upgradeDir)
// 		if err != nil {
// 			return err
// 		}
// 		if tmpPath == "暂无可回滚文件" {
// 			fmt.Println("暂无可回滚文件")
// 			return nil
// 		}
// 		// 确保 /usr/local/bin 目录存在
// 		binDir := "/usr/local/bin"
// 		if _, err := os.Stat(binDir); os.IsNotExist(err) {
// 			if err := os.MkdirAll(binDir, 0755); err != nil {
// 				return fmt.Errorf("创建目录 %s 失败: %v", binDir, err)
// 			}
// 		}

// 		// 确保目标服务文件目录存在，如果/etc/systemd/system不存在则转向/etc/init.d/
// 		serviceDir := "/etc/systemd/system"
// 		initdDir := "/etc/init.d/"
// 		serviceTarget := path.Join(serviceDir, "1panel.service")
// 		if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
// 			serviceDir = initdDir
// 			serviceTarget = path.Join(serviceDir, "1panel")
// 		}

// 		tmpPath = path.Join(upgradeDir, tmpPath, "original")
// 		fmt.Printf("(0/4) 开始从 %s 目录回滚 1Panel 服务及数据... \n", tmpPath)

// 		if err := common.CopyFile(path.Join(tmpPath, "1panel"), binDir); err != nil {
// 			return err
// 		}
// 		fmt.Println("(1/4) 1panel 二进制回滚成功")
// 		if err := common.CopyFile(path.Join(tmpPath, "1pctl"), binDir); err != nil {
// 			return err
// 		}
// 		fmt.Println("(2/4) 1panel 脚本回滚成功")
// 		if err := common.CopyFile(path.Join(tmpPath, "1panel.service"), serviceTarget); err != nil {
// 			return err
// 		}
// 		fmt.Println("(3/4) 1panel 服务回滚成功")
// 		checkPointOfWal()
// 		if _, err := os.Stat(path.Join(tmpPath, "1Panel.db")); err == nil {
// 			if err := common.CopyFile(path.Join(tmpPath, "1Panel.db"), path.Join(baseDir, "1panel/db")); err != nil {
// 				return err
// 			}
// 		}
// 		if _, err := os.Stat(path.Join(tmpPath, "db.tar.gz")); err == nil {
// 			if err := handleUnTar(path.Join(tmpPath, "db.tar.gz"), path.Join(baseDir, "1panel")); err != nil {
// 				return err
// 			}
// 		}
// 		fmt.Printf("(4/4) 1panel 数据回滚成功 \n\n")

// 		fmt.Println("回滚成功！正在重启服务，请稍候...")
// 		return nil
// 	},
// }

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

	var serviceDir string
	var serviceTarget string
	serviceDir, serviceTarget, err = ensureServiceDir()
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

func ensureServiceDir() (string, string, error) {
	// 确保服务目录存在并选择正确的目录
	serviceDir := "/etc/systemd/system"
	initdDir := "/etc/init.d/"
	serviceTarget := path.Join(serviceDir, "1panel.service")

	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		serviceDir = initdDir
		serviceTarget = path.Join(serviceDir, "1panel")
	}

	if err := ensureDir(serviceDir); err != nil {
		return "", "", err
	}

	return serviceDir, serviceTarget, nil
}

func restoreFiles(tmpPath, binDir, serviceTarget, baseDir string) error {
	filesToRestore := []struct {
		source string
		dest   string
	}{
		{path.Join(tmpPath, "1panel"), binDir},
		{path.Join(tmpPath, "1pctl"), binDir},
		{path.Join(tmpPath, "1panel.service"), serviceTarget},
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
