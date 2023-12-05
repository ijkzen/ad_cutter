package ad_cutter

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
)

/*
@description 根据视频文件生成 PCM 数据
@param 返回 Raw 文件的绝对文件名
@param 返回异常
*/
func generateRawData(videoName string) (string, error) {
	if exists, err := exists(videoName); !exists {
		return "", err
	} else {
		if !isCommandExist("ffmpeg") {
			return "", errors.New("can not find ffmpeg")
		} else {
			output := "output.raw"
			os.Remove(output)
			args := make([]string, 0)
			args = append(args, "-ss", "00:00:00")
			args = append(args, "-to", "00:05:00")
			args = append(args, "-i", videoName)
			args = append(args, "-vn")
			args = append(args, "-f", "u8")
			args = append(args, "-acodec", "pcm_u8")
			args = append(args, "-ac", "1")
			args = append(args, "-ar", "8000")
			args = append(args, "-y")
			args = append(args, output)
			cmd := exec.Command("ffmpeg", args...)
			var errOut bytes.Buffer
			cmd.Stderr = &errOut
			err := cmd.Run()
			if err != nil {
				return "", errors.New(errOut.String())
			} else {
				if currentDir, err := os.Getwd(); err != nil {
					return "", nil
				} else {
					return filepath.Join(currentDir, "output.raw"), nil
				}
			}
		}
	}
}

func cutMoive(videoName string, cutPoint int) (bool, error) {
	if exists, err := exists(videoName); !exists {
		return false, err
	} else {
		if !isCommandExist("ffmpeg") {
			return false, errors.New("can not find ffmpeg")
		} else {
			output := "output.mp4"
			output = filepath.Join(filepath.Dir(videoName), output)
			args := make([]string, 0)
			args = append(args, "-ss", strconv.Itoa(cutPoint))
			args = append(args, "-i", videoName)
			args = append(args, "-c", "copy")
			args = append(args, "-y")
			args = append(args, output)
			cmd := exec.Command("ffmpeg", args...)
			var errOut bytes.Buffer
			cmd.Stderr = &errOut
			err := cmd.Run()
			if err != nil {
				return false, wrapError(errors.New(errOut.String()))
			} else {
				err = os.Remove(videoName)
				if err != nil {
					return false, wrapError(err)
				}

				err = os.Rename(output, videoName)
				if err != nil {
					err = tryAddWritePermission(filepath.Dir(videoName))
					if err == nil {
						err = os.Rename(output, videoName)
						return err == nil, wrapError(err)
					} else {
						return false, wrapError(err)
					}
				}
				return true, nil
			}
		}
	}
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}

// 指定文件是否存在
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}

// 指定命令是否存在
func isCommandExist(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// 指定文件夹是否有写权限
func dirPermission(dir string) (bool, error) {
	exists, err := exists(dir)
	if err != nil {
		return exists, err
	}
	fileInfo, err := os.Stat(dir)
	if err != nil {
		return false, err
	}
	if !fileInfo.IsDir() {
		return false, fmt.Errorf("%s, is not dir", dir)
	}
	fileMode := fileInfo.Mode()
	permission := fileMode.Perm()

	processUid := os.Getuid()
	processGid := os.Getgid()

	// 如果 windows 就不检查权限
	if isWindows() {
		return true, nil
	}

	unixInfo, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return false, fmt.Errorf("fail to get %s uid and gid", dir)
	} else {
		if unixInfo.Uid == uint32(processUid) {
			userPermission := (permission & 0200) == 0200
			if userPermission {
				return true, nil
			}
		}

		if unixInfo.Gid == uint32(processGid) {
			groupPermission := (permission & 0020) == 0020
			if groupPermission {
				return true, nil
			}
		}

		otherPermission := (permission & 0002) == 0002
		if otherPermission {
			return true, nil
		} else {
			return false, fmt.Errorf("%s has no permission to write", dir)
		}
	}
}

func tryAddWritePermission(dir string) error {
	hasWritePermission, _ := dirPermission(dir)
	if hasWritePermission {
		return nil
	}

	return os.Chmod(dir, fs.ModeDir+0777)
}

func wrapError(err error) error {
	_, file, line, _ := runtime.Caller(1)
	return fmt.Errorf("%s:%d %s", file, line, err.Error())
}
